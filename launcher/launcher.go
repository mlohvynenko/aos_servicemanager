// Package launcher provides set of API to controls services lifecycle
package launcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/opencontainers/runtime-spec/specs-go"
	log "github.com/sirupsen/logrus"
)

/*******************************************************************************
 * Consts
 ******************************************************************************/

const (
	// services database name
	serviceDatabase = "services.db"
	// services directory
	serviceDir = "services"
)

var (
	statusStr []string = []string{"OK", "Error"}
	stateStr  []string = []string{"Init", "Running", "Stopped"}
)

// Service status
const (
	statusOk = iota
	statusError
)

// Service state
const (
	stateInit = iota
	stateRunning
	stateStopped
)

/*******************************************************************************
 * Types
 ******************************************************************************/

type serviceStatus int
type serviceState int

type ServiceInfo struct {
	Id      string `json:id`
	Version uint   `json:version`
	Status  string `json:status`
}

// Launcher instance
type Launcher struct {
	db              *database
	systemd         *dbus.Conn
	closeChannel    chan bool
	services        sync.Map
	serviceTemplate string
	workingDir      string
	runcPath        string
	netnsPath       string
}

/*******************************************************************************
 * Public
 ******************************************************************************/

// New creates new launcher object
func New(workingDir string) (launcher *Launcher, err error) {
	log.Debug("New launcher")

	var localLauncher Launcher

	localLauncher.workingDir = workingDir
	localLauncher.closeChannel = make(chan bool)

	// Check and create service dir
	dir := path.Join(workingDir, serviceDir)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return launcher, err
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return launcher, err
		}
	}

	// Create new database instance
	localLauncher.db, err = newDatabase(path.Join(workingDir, serviceDatabase))
	if err != nil {
		return launcher, err
	}

	// Load all installed services
	services, err := localLauncher.db.getServices()
	if err != nil {
		return launcher, err
	}
	for _, service := range services {
		localLauncher.services.Store(service.serviceName, service.id)
	}

	// Create systemd connection
	localLauncher.systemd, err = dbus.NewSystemConnection()
	if err != nil {
		return launcher, err
	}
	if err = localLauncher.systemd.Subscribe(); err != nil {
		return launcher, err
	}

	localLauncher.handleSystemdSubscription()

	// Get systemd service template
	localLauncher.serviceTemplate, err = getSystemdServiceTemplate(workingDir)
	if err != nil {
		return launcher, err
	}

	// Retreive runc abs path
	localLauncher.runcPath, err = exec.LookPath("runc")
	if err != nil {
		return launcher, err
	}

	// Retreive netns abs path
	localLauncher.netnsPath, _ = filepath.Abs(path.Join(workingDir, "netns"))
	if _, err := os.Stat(localLauncher.netnsPath); err != nil {
		// check system PATH
		localLauncher.netnsPath, err = exec.LookPath("netns")
		if err != nil {
			return launcher, err
		}
	}

	launcher = &localLauncher

	return launcher, nil
}

// Close closes launcher
func (launcher *Launcher) Close() {
	log.Debug("Close launcher")

	if err := launcher.systemd.Unsubscribe(); err != nil {
		log.Warn("Can't unsubscribe from systemd: ", err)
	}

	launcher.closeChannel <- true

	launcher.systemd.Close()
	launcher.db.close()
}

// GetServiceVersion returns installed version of requested service
func (launcher *Launcher) GetServiceVersion(id string) (version uint, err error) {
	log.WithField("id", id).Debug("Get service version")

	service, err := launcher.db.getService(id)
	if err != nil {
		return version, err
	}

	version = service.version

	return version, nil
}

// InstallService installs and runs service
func (launcher *Launcher) InstallService(image string, id string, version uint) (err error) {
	log.WithFields(log.Fields{"path": image, "id": id, "version": version}).Debug("Install service")

	// TODO: do we need install to /tmp dir first?
	// In case something wrong, artifacts will be removed after system reboot
	// but it will introduce additional io operations.

	// create install dir
	installDir, err := ioutil.TempDir(path.Join(launcher.workingDir, serviceDir), "")
	if err != nil {
		return err
	}
	log.WithField("dir", installDir).Debug("Create install dir")

	// unpack image there
	if err := UnpackImage(image, installDir); err != nil {
		return err
	}

	configFile := path.Join(installDir, "config.json")

	// get service spec
	spec, err := GetServiceSpec(configFile)
	if err != nil {
		return err
	}

	// update config.json
	if err := launcher.updateServiceSpec(&spec); err != nil {
		return err
	}

	// update config.json
	if err := WriteServiceSpec(&spec, configFile); err != nil {
		return err
	}

	// check if service already installed
	// TODO: check version?
	service, err := launcher.db.getService(id)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		return err
	}

	// remove if exists
	if err == nil {
		log.WithField("name", id).Debug("Service exists.")

		if err := launcher.RemoveService(id); err != nil {
			return err
		}
	}

	serviceName := "aos_" + id + ".service"

	serviceFile, err := launcher.createSystemdService(installDir, serviceName, id)
	if err != nil {
		return err
	}

	if err := launcher.startService(serviceFile, serviceName); err != nil {
		return err
	}

	service = serviceEntry{
		id:          id,
		version:     version,
		path:        installDir,
		serviceName: serviceName,
		state:       stateInit,
		status:      statusOk}

	// add to database
	if err := launcher.db.addService(service); err != nil {
		if err := launcher.stopService(serviceName); err != nil {
			log.WithField("name", serviceName).Warn("Can't stop service: ", err)
		}
		return err
	}

	launcher.services.Store(serviceName, id)

	log.WithFields(log.Fields{"id": id, "version": version}).Info("Service successfully installed")

	return nil
}

// RemoveService stops and removes service
func (launcher *Launcher) RemoveService(id string) (err error) {
	service, err := launcher.db.getService(id)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"id": service.id, "version": service.version}).Debug("Remove service")

	launcher.services.Delete(service.serviceName)

	if err := launcher.stopService(service.serviceName); err != nil {
		log.WithField("name", service.serviceName).Warn("Can't stop service: ", err)
	}

	if err := launcher.db.removeService(service.id); err != nil {
		log.WithField("name", service.serviceName).Warn("Can't remove service from db: ", err)
	}

	if err := os.RemoveAll(service.path); err != nil {
		log.WithField("path", service.path).Error("Can't remove service path")
	}

	log.WithFields(log.Fields{"id": id, "version": service.version}).Info("Service successfully removed")

	return nil
}

// GetServicesInfo returns informaion about all installed services
func (launcher *Launcher) GetServicesInfo() (info []ServiceInfo, err error) {
	log.Debug("Get services info")

	services, err := launcher.db.getServices()
	if err != nil {
		return info, err
	}

	info = make([]ServiceInfo, len(services))

	for i, service := range services {
		info[i] = ServiceInfo{service.id, service.version, statusStr[service.status]}
	}

	return info, nil
}

/*******************************************************************************
 * Private
 ******************************************************************************/

func getSystemdServiceTemplate(workingDir string) (template string, err error) {
	template = `[Unit]
Description=AOS Service
After=network.target

[Service]
Type=forking
Restart=always
RestartSec=1
ExecStart=%s
ExecStop=%s
ExecStopPost=%s
PIDFile=%s

[Install]
WantedBy=multi-user.target
`
	fileName := path.Join(workingDir, "template.service")
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		if !os.IsNotExist(err) {
			return template, err
		}

		log.Warnf("Service template file does not exist. Creating %s", fileName)

		if err = ioutil.WriteFile(fileName, []byte(template), 0644); err != nil {
			return template, err
		}
	} else {
		template = string(fileContent)
	}

	return template, nil
}

func (launcher *Launcher) handleSystemdSubscription() {
	serviceChannel, errorChannel := launcher.systemd.SubscribeUnitsCustom(time.Millisecond*1000,
		2,
		func(u1, u2 *dbus.UnitStatus) bool { return *u1 != *u2 },
		func(serviceName string) bool {
			if _, exist := launcher.services.Load(serviceName); exist {
				return false
			}
			return true
		})
	go func() {
		for {
			select {
			case services := <-serviceChannel:
				for _, service := range services {
					var (
						state  serviceState
						status serviceStatus
					)

					if service == nil {
						continue
					}

					log.WithField("name", service.Name).Debugf(
						"Service state changed. Load state: %s, active state: %s, sub state: %s",
						service.LoadState,
						service.ActiveState,
						service.SubState)

					switch service.SubState {
					case "running":
						state = stateRunning
						status = statusOk
					default:
						state = stateStopped
						status = statusError
					}

					log.WithField("name", service.Name).Debugf("Set service state: %s, status: %s", stateStr[state], statusStr[status])

					id, exist := launcher.services.Load(service.Name)

					if exist {
						if err := launcher.db.setServiceState(id.(string), state); err != nil {
							log.WithField("name", service.Name).Error("Can't set service state: ", err)
						}

						if err := launcher.db.setServiceStatus(id.(string), status); err != nil {
							log.WithField("name", service.Name).Error("Can't set service status: ", err)
						}
					} else {
						log.WithField("name", service.Name).Warning("Can't update state or status. Service is not installed.")
					}
				}
			case err := <-errorChannel:
				log.Error("Subscription error: ", err)
			case <-launcher.closeChannel:
				return
			}
		}
	}()
}

func (launcher *Launcher) updateServiceSpec(spec *specs.Spec) (err error) {
	// disable terminal
	spec.Process.Terminal = false

	mounts := []specs.Mount{
		specs.Mount{"/bin", "bind", "/bin", []string{"bind", "ro"}},
		specs.Mount{"/sbin", "bind", "/sbin", []string{"bind", "ro"}},
		specs.Mount{"/lib", "bind", "/lib", []string{"bind", "ro"}},
		specs.Mount{"/usr", "bind", "/usr", []string{"bind", "ro"}},
		// TODO: mount individual tmp
		// "destination": "/tmp",
		// "type": "tmpfs",
		// "source": "tmpfs",
		// "options": ["nosuid","strictatime","mode=755","size=65536k"]
		specs.Mount{"/tmp", "bind", "/tmp", []string{"bind", "rw"}}}
	spec.Mounts = append(spec.Mounts, mounts...)
	// add lib64 if exists
	if _, err := os.Stat("/lib64"); err == nil {
		spec.Mounts = append(spec.Mounts, specs.Mount{"/lib64", "bind", "/lib64", []string{"bind", "ro"}})
	}
	// add hosts
	hosts, _ := filepath.Abs(path.Join(launcher.workingDir, "etc", "hosts"))
	if _, err := os.Stat(hosts); err != nil {
		hosts = "/etc/hosts"
	}
	spec.Mounts = append(spec.Mounts, specs.Mount{path.Join("/etc", "hosts"), "bind", hosts, []string{"bind", "ro"}})
	// add resolv.conf
	resolvConf, _ := filepath.Abs(path.Join(launcher.workingDir, "etc", "resolv.conf"))
	if _, err := os.Stat(resolvConf); err != nil {
		resolvConf = "/etc/resolv.conf"
	}
	spec.Mounts = append(spec.Mounts, specs.Mount{path.Join("/etc", "resolv.conf"), "bind", resolvConf, []string{"bind", "ro"}})
	// add nsswitch.conf
	nsswitchConf, _ := filepath.Abs(path.Join(launcher.workingDir, "etc", "nsswitch.conf"))
	if _, err := os.Stat(nsswitchConf); err != nil {
		nsswitchConf = "/etc/nsswitch.conf"
	}
	spec.Mounts = append(spec.Mounts, specs.Mount{path.Join("/etc", "nsswitch.conf"), "bind", nsswitchConf, []string{"bind", "ro"}})

	// TODO: all services should have their own certificates
	// this mound for demo only and should be removed 
	// mount /etc/ssl 
	spec.Mounts = append(spec.Mounts, specs.Mount{path.Join("/etc", "ssl"), "bind", path.Join("/etc", "ssl"), []string{"bind", "ro"}})

	// add netns hook
	if spec.Hooks == nil {
		spec.Hooks = &specs.Hooks{}
	}
	spec.Hooks.Prestart = append(spec.Hooks.Prestart, specs.Hook{Path: launcher.netnsPath})

	return nil
}

func (launcher *Launcher) startService(serviceFile, serviceName string) (err error) {
	if _, _, err := launcher.systemd.EnableUnitFiles([]string{serviceFile}, false, true); err != nil {
		return err
	}

	if err := launcher.systemd.Reload(); err != nil {
		return err
	}

	channel := make(chan string)
	if _, err = launcher.systemd.StartUnit(serviceName, "replace", channel); err != nil {
		return err
	}
	status := <-channel

	log.WithFields(log.Fields{"name": serviceName, "status": status}).Debug("Start service")

	return nil
}

func (launcher *Launcher) stopService(serviceName string) (err error) {
	channel := make(chan string)
	if _, err := launcher.systemd.StopUnit(serviceName, "replace", channel); err != nil {
		return err
	}
	status := <-channel

	log.WithFields(log.Fields{"name": serviceName, "status": status}).Debug("Stop service")

	if _, err := launcher.systemd.DisableUnitFiles([]string{serviceName}, false); err != nil {
		return err
	}

	if err := launcher.systemd.Reload(); err != nil {
		return err
	}

	return nil
}

func (launcher *Launcher) createSystemdService(installDir, serviceName, id string) (fileName string, err error) {
	f, err := os.Create(path.Join(installDir, serviceName))
	if err != nil {
		return fileName, err
	}
	defer f.Close()

	absServicePath, err := filepath.Abs(installDir)
	if err != nil {
		return fileName, err
	}

	pidFile := path.Join(absServicePath, id+".pid")
	execStartString := launcher.runcPath + " run -d --pid-file " + pidFile + " -b " + absServicePath + " " + id
	execStopString := launcher.runcPath + " kill -a " + id + " SIGKILL"
	execStopPostString := launcher.runcPath + " delete " + id

	lines := strings.SplitAfter(launcher.serviceTemplate, "\n")
	for _, line := range lines {
		switch {
		// the order is important for example: execstoppost should be evaluated
		// before execstop as execstop is substring of execstoppost
		case strings.Contains(strings.ToLower(line), "execstart"):
			if _, err := fmt.Fprintf(f, line, execStartString); err != nil {
				return fileName, err
			}
		case strings.Contains(strings.ToLower(line), "execstoppost"):
			if _, err := fmt.Fprintf(f, line, execStopPostString); err != nil {
				return fileName, err
			}
		case strings.Contains(strings.ToLower(line), "execstop"):
			if _, err := fmt.Fprintf(f, line, execStopString); err != nil {
				return fileName, err
			}
		case strings.Contains(strings.ToLower(line), "pidfile"):
			if _, err := fmt.Fprintf(f, line, pidFile); err != nil {
				return fileName, err
			}
		default:
			fmt.Fprint(f, line)
		}
	}

	if fileName, err = filepath.Abs(f.Name()); err != nil {
		return fileName, err
	}

	return fileName, nil
}