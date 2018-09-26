package config_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"gitpct.epam.com/epmd-aepr/aos_servicemanager/config"
)

/*******************************************************************************
 * Private
 ******************************************************************************/

func createConfigFile() (err error) {
	configContent := `{
	"fcrypt" : {
		"CACert" : "CACert",
		"ClientCert" : "ClientCert",
		"ClientKey" : "ClientKey",
		"OfflinePrivKey" : "OfflinePrivKey",
		"OfflineCert" : "OfflineCert"	
	},
	"serviceDiscovery" : "www.aos.com",
	"workingDir" : "workingDir",
	"visServer" : "wss://localhost:8088",
	"defaultServiceTTLDays" : 30,
	"monitoring": {
		"sendPeriod": "00:05:00",
		"pollPeriod": "00:00:01",
		"maxOfflineMessages": 25,
		"maxAlertsPerMessage": 128,
		"ram": {
			"minTimeout": "00:00:10",
			"minThreshold": 10,
			"maxThreshold": 150
		},
		"outTraffic": {
			"minTimeout": "00:00:20",
			"minThreshold": 10,
			"maxThreshold": 150
		}
	}
}`

	if err := ioutil.WriteFile(path.Join("tmp", "aos_servicemanager.cfg"), []byte(configContent), 0644); err != nil {
		return err
	}

	return nil
}

func setup() (err error) {
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return err
	}

	if err = createConfigFile(); err != nil {
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

func TestGetCrypt(t *testing.T) {
	config, err := config.New("tmp/aos_servicemanager.cfg")
	if err != nil {
		t.Fatalf("Error opening config file: %s", err)
	}

	if config.Crypt.CACert != "CACert" {
		t.Errorf("Wrong CACert value: %s", config.Crypt.CACert)
	}

	if config.Crypt.ClientCert != "ClientCert" {
		t.Errorf("Wrong ClientCert value: %s", config.Crypt.ClientCert)
	}

	if config.Crypt.ClientKey != "ClientKey" {
		t.Errorf("Wrong ClientKey value: %s", config.Crypt.ClientKey)
	}

	if config.Crypt.OfflinePrivKey != "OfflinePrivKey" {
		t.Errorf("Wrong OfflinePrivKey value: %s", config.Crypt.OfflinePrivKey)
	}

	if config.Crypt.OfflineCert != "OfflineCert" {
		t.Errorf("Wrong OfflineCert value: %s", config.Crypt.OfflineCert)
	}
}

func TestGetServiceDiscoveryURL(t *testing.T) {
	config, err := config.New("tmp/aos_servicemanager.cfg")
	if err != nil {
		t.Fatalf("Error opening config file: %s", err)
	}

	if config.ServiceDiscoveryURL != "www.aos.com" {
		t.Errorf("Wrong server URL value: %s", config.ServiceDiscoveryURL)
	}
}

func TestGetWorkingDir(t *testing.T) {
	config, err := config.New("tmp/aos_servicemanager.cfg")
	if err != nil {
		t.Fatalf("Error opening config file: %s", err)
	}

	if config.WorkingDir != "workingDir" {
		t.Errorf("Wrong workingDir value: %s", config.WorkingDir)
	}
}

func TestGetVisServerURL(t *testing.T) {
	config, err := config.New("tmp/aos_servicemanager.cfg")
	if err != nil {
		t.Fatalf("Error opening config file: %s", err)
	}

	if config.VISServerURL != "wss://localhost:8088" {
		t.Errorf("Wrong VIS server value: %s", config.VISServerURL)
	}
}

func TestGetDefaultServiceTTL(t *testing.T) {
	config, err := config.New("tmp/aos_servicemanager.cfg")
	if err != nil {
		t.Fatalf("Error opening config file: %s", err)
	}

	if config.DefaultServiceTTL != 30 {
		t.Errorf("Wrong default service TTL value: %d", config.DefaultServiceTTL)
	}
}

func TestDurationMarshal(t *testing.T) {
	d := config.Duration{Duration: 32 * time.Second}

	result, err := json.Marshal(d)
	if err != nil {
		t.Errorf("Can't marshal: %s", err)
	}

	if string(result) != `"00:00:32"` {
		t.Errorf("Wrong value: %s", result)
	}
}

func TestGetMonitoringConfig(t *testing.T) {
	config, err := config.New("tmp/aos_servicemanager.cfg")
	if err != nil {
		t.Fatalf("Error opening config file: %s", err)
	}

	if config.Monitoring.SendPeriod.Duration != 5*time.Minute {
		t.Errorf("Wrong send period value: %s", config.Monitoring.SendPeriod)
	}

	if config.Monitoring.PollPeriod.Duration != 1*time.Second {
		t.Errorf("Wrong poll period value: %s", config.Monitoring.PollPeriod)
	}

	if config.Monitoring.RAM.MinTimeout.Duration != 10*time.Second {
		t.Errorf("Wrong value: %s", config.Monitoring.RAM.MinTimeout)
	}

	if config.Monitoring.OutTraffic.MinTimeout.Duration != 20*time.Second {
		t.Errorf("Wrong value: %s", config.Monitoring.RAM.MinTimeout)
	}
}