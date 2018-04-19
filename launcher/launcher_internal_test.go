package launcher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	amqp "gitpct.epam.com/epmd-aepr/aos_servicemanager/amqphandler"
)

/*******************************************************************************
 * Types
 ******************************************************************************/

type TestLauncher struct {
	*Launcher
}

/*******************************************************************************
 * Init
 ******************************************************************************/

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		FullTimestamp:    true})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
}

/*******************************************************************************
 * Private
 ******************************************************************************/

func newTestLauncher() (testLauncher *TestLauncher, statusChannel <-chan ActionStatus, err error) {
	instance, statusChannel, err := New("tmp")

	testLauncher = &TestLauncher{instance}
	testLauncher.downloader = testLauncher

	return testLauncher, statusChannel, err
}

func (launcher *TestLauncher) downloadService(serviceInfo amqp.ServiceInfoFromCloud) (outputFile string, err error) {
	imageDir, err := ioutil.TempDir("", "aos_")
	if err != nil {
		log.Error("Can't create image dir : ", err)
		return outputFile, err
	}
	defer os.RemoveAll(imageDir)

	if err := generateImage(imageDir); err != nil {
		return outputFile, err
	}

	specFile := path.Join(imageDir, "config.json")

	spec, err := getServiceSpec(specFile)
	if err != nil {
		return outputFile, err
	}

	spec.Process.Args = []string{"python3", "/home/service.py", serviceInfo.Id}

	if err := writeServiceSpec(&spec, specFile); err != nil {
		return outputFile, err
	}

	imageFile, err := ioutil.TempFile("", "aos_")
	if err != nil {
		return outputFile, err
	}
	outputFile = imageFile.Name()
	imageFile.Close()

	if err = packImage(imageDir, outputFile); err != nil {
		return outputFile, err
	}

	return outputFile, nil
}

func setup() (err error) {
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return err
	}

	return nil
}

func cleanup() (err error) {
	if err := os.RemoveAll("tmp"); err != nil {
		return err
	}

	return nil
}

func generateImage(imagePath string) (err error) {
	// create dir
	if err := os.MkdirAll(path.Join(imagePath, "rootfs", "home"), 0755); err != nil {
		return err
	}

	serviceContent := `#!/usr/bin/python

import time
import sys

i = 0
serviceName = sys.argv[1]

print(">>>> Start", serviceName)
while True:
	print(">>>> aos", serviceName, "count", i)
	i = i + 1
	time.sleep(5)`

	if err := ioutil.WriteFile(path.Join(imagePath, "rootfs", "home", "service.py"), []byte(serviceContent), 0644); err != nil {
		return err
	}

	// remove json
	if err := os.Remove(path.Join(imagePath, "config.json")); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	// generate config spec
	out, err := exec.Command("runc", "spec", "-b", imagePath).CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

/*******************************************************************************
 * Main
 ******************************************************************************/

func TestMain(m *testing.M) {

	if err := setup(); err != nil {
		log.Fatalf("Error creating service images: %s", err)
	}

	ret := m.Run()

	if err := cleanup(); err != nil {
		log.Fatalf("Error cleaning up: %s", err)
	}

	os.Exit(ret)
}

/*******************************************************************************
 * Tests
 ******************************************************************************/

func TestInstallRemove(t *testing.T) {
	launcher, statusChan, err := newTestLauncher()
	if err != nil {
		t.Fatalf("Can't create launcher: %s", err)
	}
	defer launcher.Close()

	numInstallServices := 30
	numRemoveServices := 10

	// install services
	for i := 0; i < numInstallServices; i++ {
		launcher.InstallService(amqp.ServiceInfoFromCloud{Id: fmt.Sprintf("service%d", i)})
	}
	// remove services
	for i := 0; i < numRemoveServices; i++ {
		launcher.RemoveService(fmt.Sprintf("service%d", i))
	}

	for i := 0; i < numInstallServices+numRemoveServices; i++ {
		if status := <-statusChan; status.Err != nil {
			if status.Action == ActionInstall {
				t.Error("Can't install service: ", status.Err)
			} else {
				t.Error("Can't remove service: ", status.Err)
			}
		}
	}

	services, err := launcher.GetServicesInfo()
	if err != nil {
		t.Error("Can't get services info: ", err)
	}
	if len(services) != numInstallServices-numRemoveServices {
		t.Errorf("Wrong service quantity")
	}
	for _, service := range services {
		if service.Status != "OK" {
			t.Errorf("Service %s error status: %s", service.Id, service.Status)
		}
	}

	time.Sleep(time.Second * 5)

	// remove remaining services
	for i := numRemoveServices; i < numInstallServices; i++ {
		launcher.RemoveService(fmt.Sprintf("service%d", i))
	}

	for i := 0; i < numInstallServices-numRemoveServices; i++ {
		if status := <-statusChan; status.Err != nil {
			t.Error("Can't remove service: ", status.Err)
		}
	}

	services, err = launcher.GetServicesInfo()
	if err != nil {
		t.Error("Can't get services info: ", err)
	}
	if len(services) != 0 {
		t.Errorf("Wrong service quantity")
	}
}

func TestAutoStart(t *testing.T) {
	launcher, statusChan, err := newTestLauncher()
	if err != nil {
		t.Fatalf("Can't create launcher: %s", err)
	}

	numServices := 10

	// install services
	for i := 0; i < numServices; i++ {
		launcher.InstallService(amqp.ServiceInfoFromCloud{Id: fmt.Sprintf("service%d", i)})
	}

	for i := 0; i < numServices; i++ {
		if status := <-statusChan; status.Err != nil {
			t.Error("Can't install service: ", status.Err)
		}
	}

	launcher.Close()

	time.Sleep(time.Second * 5)

	launcher, statusChan, err = newTestLauncher()
	if err != nil {
		t.Fatalf("Can't create launcher: %s", err)
	}
	defer launcher.Close()

	time.Sleep(time.Second * 5)

	services, err := launcher.GetServicesInfo()
	if err != nil {
		t.Error("Can't get services info: ", err)
	}
	if len(services) != numServices {
		t.Errorf("Wrong service quantity")
	}
	for _, service := range services {
		if service.Status != "OK" {
			t.Errorf("Service %s error status: %s", service.Id, service.Status)
		}
	}

	// remove services
	for i := 0; i < numServices; i++ {
		launcher.RemoveService(fmt.Sprintf("service%d", i))
	}

	for i := 0; i < numServices; i++ {
		if status := <-statusChan; status.Err != nil {
			t.Error("Can't remove service: ", status.Err)
		}
	}

	services, err = launcher.GetServicesInfo()
	if err != nil {
		t.Error("Can't get services info: ", err)
	}
	if len(services) != 0 {
		t.Errorf("Wrong service quantity")
	}
}
