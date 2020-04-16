// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2019 Renesas Inc.
// Copyright 2019 EPAM Systems Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path"
	"reflect"
	"syscall"
	"time"

	"github.com/docker/docker/pkg/reexec"
	log "github.com/sirupsen/logrus"

	"aos_servicemanager/alerts"
	amqp "aos_servicemanager/amqphandler"
	"aos_servicemanager/config"
	"aos_servicemanager/database"
	"aos_servicemanager/fcrypt"
	_ "aos_servicemanager/identification"
	"aos_servicemanager/launcher"
	"aos_servicemanager/layermanager"
	"aos_servicemanager/logging"
	"aos_servicemanager/monitoring"
	"aos_servicemanager/networkmanager"
	"aos_servicemanager/pluginprovider"
	"aos_servicemanager/umclient"
)

/*******************************************************************************
 * Consts
 ******************************************************************************/

const reconnectTimeout = 10 * time.Second

const dbFileName = "servicemanager.db"

// IMPORTANT: if new functionality doesn't allow existing services to work
// properly, this value should be increased. It will force to remove all
// services and their storages before first start.
const operationVersion = 2

/*******************************************************************************
 * Types
 ******************************************************************************/

type serviceManager struct {
	alerts     *alerts.Alerts
	amqp       *amqp.AmqpHandler
	cfg        *config.Config
	crypt      *fcrypt.CryptoContext
	db         *database.Database
	identifier pluginprovider.Identifier
	launcher   *launcher.Launcher
	logging    *logging.Logging
	monitor    *monitoring.Monitor
	network    *networkmanager.NetworkManager
	um         *umclient.Client
	layerMgr   *layermanager.LayerManager
}

/*******************************************************************************
 * Vars
 ******************************************************************************/

// GitSummary provided by govvv at compile-time
var GitSummary = "Unknown"

/*******************************************************************************
 * Init
 ******************************************************************************/

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		FullTimestamp:    true})
	log.SetOutput(os.Stdout)
}

/*******************************************************************************
 * ServiceManager
 ******************************************************************************/

func cleanup(cfg *config.Config, dbFile string) {
	log.Info("System cleanup")

	if err := launcher.Cleanup(cfg); err != nil {
		log.Errorf("Can't cleanup launcher: %s", err)
	}

	log.WithField("file", dbFile).Debug("Delete DB file")
	if err := os.RemoveAll(dbFile); err != nil {
		log.Errorf("Can't cleanup database: %s", err)
	}

	log.WithField("file", dbFile).Debug("Delete networks")

	network, err := networkmanager.New(cfg)
	if err != nil {
		log.Errorf("Can't create network: %s", err)
	}

	if err = network.DeleteAllNetworks(); err != nil {
		log.Errorf("Can't delete networks: %s", err)
	}
}

func newServiceManager(cfg *config.Config) (sm *serviceManager, err error) {
	defer func() {
		if err != nil {
			sm.close()
			sm = nil
		}
	}()

	var names []string
	sm = &serviceManager{cfg: cfg}

	// Create DB
	dbFile := path.Join(cfg.WorkingDir, dbFileName)

	if sm.db, err = database.New(dbFile); err != nil {
		return sm, err
	}

	// Check operation version

	version, err := sm.db.GetOperationVersion()
	if err != nil {
		return sm, err
	}

	if operationVersion != version {
		log.Warning("Unsupported operation version")

		sm.db.Close()

		cleanup(cfg, dbFile)

		if sm.db, err = database.New(dbFile); err != nil {
			return sm, err
		}

		if err = sm.db.SetOperationVersion(operationVersion); err != nil {
			return sm, err
		}
	}

	// Initialize fcrypt
	if err = fcrypt.Init(cfg.Crypt); err != nil {
		return sm, err
	}

	// Get organization names from certificate and use it as discovery URL
	names, err = fcrypt.GetCertificateOrganizations(cfg.Crypt.ClientCert)
	if err != nil {
		log.Warningf("Organization name will be taken from config file: %s", err)
	} else {
		// We use the first member of organization list
		// The certificate should contain only one organization
		if len(names) == 1 && names[0] != "" {
			url := url.URL{
				Scheme: "https",
				Host:   names[0],
			}
			cfg.ServiceDiscoveryURL = url.String() + ":9000"
		} else {
			log.Error("Certificate organization name is empty or organization is not a single")
		}
	}

	// Create crypto context
	if sm.crypt, err = fcrypt.CreateContext(cfg.Crypt); err != nil {
		return sm, err
	}

	if err = sm.crypt.LoadOfflineKey(); err != nil {
		return sm, err
	}

	if err = sm.crypt.LoadOnlineKey(); err != nil {
		return sm, err
	}

	// Create alerts
	if sm.alerts, err = alerts.New(cfg, sm.db, sm.db); err != nil {
		if err == alerts.ErrDisabled {
			log.Warn(err)
		} else {
			return sm, err
		}
	}

	// Create network
	if sm.network, err = networkmanager.New(cfg); err != nil {
		return sm, err
	}

	// Create monitor
	if sm.monitor, err = monitoring.New(cfg, sm.db, sm.alerts); err != nil {
		if err == monitoring.ErrDisabled {
			log.Warn(err)
		} else {
			return sm, err
		}
	}

	// Create amqp
	if sm.amqp, err = amqp.New(); err != nil {
		return sm, err
	}

	if sm.layerMgr, err = layermanager.New(cfg.LayersDir, sm.crypt, sm.db, sm.amqp); err != nil {
		return sm, err
	}

	// Create launcher
	if sm.launcher, err = launcher.New(cfg, sm.amqp, sm.db, sm.layerMgr, sm.monitor, sm.network); err != nil {
		return sm, err
	}

	// Create identifier
	// Use plugged in identifier
	if sm.identifier, err = pluginprovider.GetIdentifier(cfg.Identifier.Type, cfg.Identifier.Config, sm.db); err != nil {
		return sm, err
	}

	// Create UM client
	if cfg.UMServerURL != "" {
		if sm.um, err = umclient.New(cfg, sm.crypt, sm.amqp, sm.db); err != nil {
			return sm, err
		}
	}

	// Create logging
	if sm.logging, err = logging.New(cfg, sm.db); err != nil {
		return sm, err
	}

	return sm, nil
}

func (sm *serviceManager) close() {
	// Close logging
	if sm.logging != nil {
		sm.logging.Close()
	}

	// Close amqp
	if sm.amqp != nil {
		sm.amqp.Close()
	}

	// Close UM client
	if sm.um != nil {
		sm.um.Close()
	}

	// Close identifier
	if sm.identifier != nil {
		sm.identifier.Close()
	}

	// Close launcher
	if sm.launcher != nil {
		sm.launcher.Close()
	}

	// Close monitor
	if sm.monitor != nil {
		sm.monitor.Close()
	}

	// Close network
	if sm.network != nil {
		sm.monitor.Close()
	}

	// Close alerts
	if sm.alerts != nil {
		sm.alerts.Close()
	}

	// Close DB
	if sm.db != nil {
		sm.db.Close()
	}
}

func (sm *serviceManager) sendInitialSetup() (err error) {
	initialList, err := sm.launcher.GetServicesInfo()
	if err != nil {
		log.Fatalf("Can't get services: %s", err)
	}

	initialLayerList, err := sm.layerMgr.GetLayersInfo()
	if err != nil {
		log.Fatalf("Can't get layers list: %s", err)
	}

	if err = sm.amqp.SendInitialSetup(initialList, initialLayerList); err != nil {
		return err
	}

	return nil
}

func (sm *serviceManager) processAmqpMessage(message amqp.Message) (err error) {
	switch data := message.Data.(type) {
	case amqp.DecodedDesiredStatus:
		if err := sm.layerMgr.ProcessDesiredLayersList(data.Layers, data.CertificateChains, data.Certificates); err != nil {
			log.Error("Can't process layer list: ", err)
			return err
		}

		log.WithField("len", len(data.Services)).Info("Receive services info")

		currentList, err := sm.launcher.GetServicesInfo()
		if err != nil {
			return err
		}

		type serviceDesc struct {
			serviceInfo          *amqp.ServiceInfo
			serviceInfoFromCloud *amqp.ServiceInfoFromCloud
		}

		servicesMap := make(map[string]*serviceDesc)

		for _, item := range currentList {
			serviceInfo := item

			servicesMap[serviceInfo.ID] = &serviceDesc{serviceInfo: &serviceInfo}
		}

		for _, item := range data.Services {
			serviceInfoFromCloud := item

			if _, ok := servicesMap[serviceInfoFromCloud.ID]; !ok {
				servicesMap[serviceInfoFromCloud.ID] = &serviceDesc{}
			}

			servicesMap[serviceInfoFromCloud.ID].serviceInfoFromCloud = &serviceInfoFromCloud
		}

		for _, service := range servicesMap {
			if service.serviceInfoFromCloud != nil && service.serviceInfo != nil {
				// Update
				if service.serviceInfoFromCloud.Version > service.serviceInfo.Version {
					log.WithFields(log.Fields{
						"id":   service.serviceInfo.ID,
						"from": service.serviceInfo.Version,
						"to":   service.serviceInfoFromCloud.Version}).Info("Update service")

					sm.launcher.InstallService(*service.serviceInfoFromCloud)
				}
			} else if service.serviceInfoFromCloud != nil {
				// Install
				log.WithFields(log.Fields{
					"id":      service.serviceInfoFromCloud.ID,
					"version": service.serviceInfoFromCloud.Version}).Info("Install service")

				sm.launcher.InstallService(*service.serviceInfoFromCloud)
			} else if service.serviceInfo != nil {
				// Remove
				log.WithFields(log.Fields{
					"id":      service.serviceInfo.ID,
					"version": service.serviceInfo.Version}).Info("Remove service")

				sm.launcher.UninstallService(service.serviceInfo.ID)
			}
		}

		sm.launcher.FinishProcessingLayers()

	case *amqp.StateAcceptance:
		log.WithFields(log.Fields{
			"serviceID": data.ServiceID,
			"result":    data.Result}).Info("Receive state acceptance")

		sm.launcher.StateAcceptance(*data, message.CorrelationID)

	case *amqp.UpdateState:
		log.WithFields(log.Fields{
			"serviceID": data.ServiceID,
			"checksum":  data.Checksum}).Info("Receive update state")

		sm.launcher.UpdateState(*data)

	case *amqp.RequestServiceLog:
		log.WithFields(log.Fields{
			"serviceID": data.ServiceID,
			"from":      data.From,
			"till":      data.Till}).Info("Receive request service log")

		sm.logging.GetServiceLog(*data)

	case *amqp.RequestServiceCrashLog:
		log.WithFields(log.Fields{
			"serviceID": data.ServiceID}).Info("Receive request service crash log")

		sm.logging.GetServiceCrashLog(*data)

	case *amqp.SystemUpgrade:
		log.WithFields(log.Fields{
			"imageVersion": data.ImageVersion}).Info("Receive system upgrade request")

		sm.um.SystemUpgrade(*data)

	case *amqp.SystemRevert:
		log.WithFields(log.Fields{
			"imageVersion": data.ImageVersion}).Info("Receive system revert request")

		sm.um.SystemRevert(data.ImageVersion)

	default:
		log.Warnf("Receive unsupported amqp message: %s", reflect.TypeOf(data))
	}

	return nil
}

func (sm *serviceManager) handleChannels() (err error) {
	var monitorDataChannel chan amqp.MonitoringData
	var alertsChannel chan amqp.Alerts
	var umErrChannel chan error

	if sm.monitor != nil {
		monitorDataChannel = sm.monitor.DataChannel
	}

	if sm.alerts != nil {
		alertsChannel = sm.alerts.AlertsChannel
	}

	if sm.um != nil {
		umErrChannel = sm.um.ErrorChannel
	}

	for {
		select {
		case amqpMessage := <-sm.amqp.MessageChannel:
			if err, ok := amqpMessage.Data.(error); ok {
				return err
			}

			if err = sm.processAmqpMessage(amqpMessage); err != nil {
				log.Errorf("Error processing amqp result: %s", err)
			}

		case newState := <-sm.launcher.NewStateChannel:
			if err := sm.amqp.SendNewState(newState.ServiceID, newState.State,
				newState.Checksum, newState.CorrelationID); err != nil {
				log.Errorf("Error send new state message: %s", err)
			}

		case data := <-monitorDataChannel:
			if err := sm.amqp.SendMonitoringData(data); err != nil {
				log.Errorf("Error send monitoring data: %s", err)
			}

		case logData := <-sm.logging.LogChannel:
			if err := sm.amqp.SendServiceLog(logData); err != nil {
				log.Errorf("Error send service log: %s", err)
			}

		case alerts := <-alertsChannel:
			if err := sm.amqp.SendAlerts(alerts); err != nil {
				log.Errorf("Error send alerts: %s", err)
			}

		case users := <-sm.identifier.UsersChangedChannel():
			log.WithField("users", users).Info("Users changed")
			return nil

		case err := <-sm.identifier.ErrorChannel():
			return err

		case err := <-umErrChannel:
			return err
		}
	}
}

func (sm *serviceManager) run() {
	for {
		var users []string
		var systemID string
		var err error

		// Get system id
		if systemID, err = sm.identifier.GetSystemID(); err != nil {
			log.Errorf("Can't get system id: %s", err)
			goto reconnect
		}

		// Get users
		if users, err = sm.identifier.GetUsers(); err != nil {
			log.Errorf("Can't get users: %s", err)
			goto reconnect
		}

		if err = sm.launcher.SetUsers(users); err != nil {
			log.Fatalf("Can't set users: %s", err)
		}

		if sm.um != nil {
			if err = sm.um.Connect(sm.cfg.UMServerURL); err != nil {
				log.Errorf("Can't connect to UM: %s", err)
				goto reconnect
			}
		}

		// Connect
		if err = sm.amqp.Connect(sm.cfg.ServiceDiscoveryURL, systemID, users); err != nil {
			log.Errorf("Can't establish connection: %s", err)
			goto reconnect
		}

		if sm.um != nil {
			version, err := sm.um.GetSystemVersion()
			if err != nil {
				log.Errorf("Can't get system version: %s", err)
				goto reconnect
			}

			if err = sm.amqp.SendSystemVersion(version); err != nil {
				log.Errorf("Can't send system version: %s", err)
				goto reconnect
			}
		}

		if err = sm.sendInitialSetup(); err != nil {
			log.Errorf("Can't send initial setup: %s", err)
			goto reconnect
		}

		if err = sm.handleChannels(); err != nil {
			log.Errorf("Runtime error: %s", err)
		}

	reconnect:
		sm.amqp.Disconnect()
		if sm.um != nil {
			sm.um.Disconnect()
		}

		<-time.After(reconnectTimeout)

		log.Debug("Reconnecting...")
	}
}

/*******************************************************************************
 * Main
 ******************************************************************************/

func main() {
	// This is network manager reexec initialization
	if reexec.Init() {
		return
	}

	// Initialize command line flags
	configFile := flag.String("c", "aos_servicemanager.cfg", "path to config file")
	strLogLevel := flag.String("v", "info", `log level: "debug", "info", "warn", "error", "fatal", "panic"`)
	doCleanup := flag.Bool("reset", false, `Removes all services, wipes services and storages and remove DB`)
	showVersion := flag.Bool("version", false, `Show service manager version`)

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("Version: %s\n", GitSummary)
		return
	}

	// Set log level
	logLevel, err := log.ParseLevel(*strLogLevel)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.SetLevel(logLevel)

	log.WithFields(log.Fields{"configFile": *configFile, "version": GitSummary}).Info("Start service manager")

	cfg, err := config.New(*configFile)
	if err != nil {
		log.Fatalf("Can't create config: %s", err)
	}

	if *doCleanup {
		cleanup(cfg, path.Join(cfg.WorkingDir, dbFileName))
		return
	}

	sm, err := newServiceManager(cfg)
	if err != nil {
		log.Fatalf("Can't create service manager: %s", err)
	}
	defer sm.close()

	// Handle SIGTERM
	terminateChannel := make(chan os.Signal, 1)
	signal.Notify(terminateChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-terminateChannel

		sm.close()

		os.Exit(1)
	}()

	sm.run()
}
