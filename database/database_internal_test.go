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

package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"aos_servicemanager/launcher"
	"aos_servicemanager/umcontroller"
)

/*******************************************************************************
 * Variables
 ******************************************************************************/

var tmpDir string
var db *Database

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
 * Main
 ******************************************************************************/

func TestMain(m *testing.M) {
	var err error

	tmpDir, err = ioutil.TempDir("", "sm_")
	if err != nil {
		log.Fatalf("Error create temporary dir: %s", err)
	}

	dbPath := path.Join(tmpDir, "test.db")
	db, err = New(dbPath, tmpDir, tmpDir)
	if err != nil {
		log.Fatalf("Can't create database: %s", err)
	}

	ret := m.Run()

	if err = os.RemoveAll(tmpDir); err != nil {
		log.Fatalf("Error cleaning up: %s", err)
	}

	db.Close()

	os.Exit(ret)
}

/*******************************************************************************
 * Tests
 ******************************************************************************/

func TestAddService(t *testing.T) {
	// AddService
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// GetService
	service, err := db.GetService("service1")
	if err != nil {
		t.Errorf("Can't get service: %s", err)
	}

	if !reflect.DeepEqual(service, service1) {
		t.Error("service1 doesn't match stored one")
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestUpdateService(t *testing.T) {
	// AddService
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	service1 = launcher.Service{"service1", 2, "", "new_sp", "to/new_service1", "new_service1.service", 5001, 5001, "new_host",
		`{"*":"rw", "new":"r"}`, 1, 1, time.Now().UTC().Add(time.Hour * 10), 0, "{}", 123, 456, 342, 696, 789, 1024,
		[]string{"path1", "path2"}, "", []string{"dbus"}, "", []string{"service10/1000/tcp"}, []string{"1002/udp"}}

	// UpdateService
	err = db.UpdateService(service1)
	if err != nil {
		t.Errorf("Can't update service: %s", err)
	}

	// GetService
	service, err := db.GetService("service1")
	if err != nil {
		t.Errorf("Can't get service: %s", err)
	}
	if !reflect.DeepEqual(service, service1) {
		t.Errorf("service1 doesn't match updated one: %v", service)
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestNotExistService(t *testing.T) {
	// GetService
	_, err := db.GetService("service3")
	if err == nil {
		t.Error("Error in non existed service")
	} else if err != ErrNotExist {
		t.Errorf("Can't get service: %s", err)
	}
}

func TestSetServiceStatus(t *testing.T) {
	// AddService
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// SetServiceStatus
	err = db.SetServiceStatus("service1", 1)
	if err != nil {
		t.Errorf("Can't set service status: %s", err)
	}
	service, err := db.GetService("service1")
	if err != nil {
		t.Errorf("Can't get service: %s", err)
	}
	if service.Status != 1 {
		t.Errorf("Service status mismatch")
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestSetServiceState(t *testing.T) {
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// SetServiceState
	err = db.SetServiceState("service1", 1)
	if err != nil {
		t.Errorf("Can't set service state: %s", err)
	}
	service, err := db.GetService("service1")
	if err != nil {
		t.Errorf("Can't get service: %s", err)
	}
	if service.State != 1 {
		t.Errorf("Service state mismatch")
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestSetServiceStartTime(t *testing.T) {
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, `{"*":"rw"}`, "host", 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	time := time.Date(2018, 1, 1, 15, 35, 49, 0, time.UTC)
	// SetServiceState
	err = db.SetServiceStartTime("service1", time)
	if err != nil {
		t.Errorf("Can't set service state: %s", err)
	}
	service, err := db.GetService("service1")
	if err != nil {
		t.Errorf("Can't get service: %s", err)
	}
	if service.StartAt != time {
		t.Errorf("Service start time mismatch")
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestRemoveService(t *testing.T) {
	// AddService
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// RemoveService
	err = db.RemoveService("service1")
	if err != nil {
		t.Errorf("Can't remove service: %s", err)
	}
	_, err = db.GetService("service1")
	if err == nil {
		t.Errorf("Error deleteing service")
	}
}

func TestGetServices(t *testing.T) {
	// Add service 1
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// Add service 2
	service2 := launcher.Service{"service2", 1, "", "sp1", "to/service2", "service2.service", 5002, 5002, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err = db.AddService(service2)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// GetServices
	services, err := db.GetServices()
	if err != nil {
		t.Errorf("Can't get services: %s", err)
	}
	if len(services) != 2 {
		t.Error("Wrong service count")
	}
	for _, service := range services {
		if !reflect.DeepEqual(service, service1) && !reflect.DeepEqual(service, service2) {
			t.Error("Error getting services")
		}
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestSetGetDevicesAtService(t *testing.T) {
	deviceNames := []string{"gpu0", "mic0", "camera0"}
	deviceNamesForService, err := json.Marshal(deviceNames)
	if err != nil {
		t.Errorf("Can't convert device resources to text view: %s", err)
	}

	// Add service 1
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, string(deviceNamesForService), []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err = db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// GetServices
	service, err := db.GetService("service1")
	if err != nil {
		t.Errorf("Can't get services: %s", err)
	}

	if !reflect.DeepEqual(service, service1) {
		t.Error("service1 doesn't match stored one")
	}

	var storedDeviceNames []string
	if err := json.Unmarshal([]byte(service.Devices), &storedDeviceNames); err != nil {
		t.Errorf("Can't convert text view to device resources: %s", err)
	}

	if reflect.DeepEqual(deviceNames, storedDeviceNames) != true {
		t.Errorf("Stored device resources are not equal to requested")
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestGetServiceProviderServices(t *testing.T) {
	// Add service 1
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// Add service 2
	service2 := launcher.Service{"service2", 1, "", "sp1", "to/service2", "service2.service", 5002, 5002, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1002/udp"}}
	err = db.AddService(service2)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// Add service 3
	service3 := launcher.Service{"service3", 1, "", "sp2", "to/service3", "service3.service", 5003, 5003, "host3", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1003/udp"}}
	err = db.AddService(service3)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// Add service 4
	service4 := launcher.Service{"service4", 1, "", "sp2", "to/service4", "service4.service", 5004, 5004, "host4", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1004/udp"}}
	err = db.AddService(service4)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// Get sp1 services
	servicesSp1, err := db.GetServiceProviderServices("sp1")
	if err != nil {
		t.Errorf("Can't get services: %s", err)
	}
	if len(servicesSp1) != 2 {
		t.Error("Wrong service count")
	}
	for _, service := range servicesSp1 {
		if !reflect.DeepEqual(service, service1) && !reflect.DeepEqual(service, service2) {
			t.Error("Error getting services")
		}
	}

	// Get sp2 services
	servicesSp2, err := db.GetServiceProviderServices("sp2")
	if err != nil {
		t.Errorf("Can't get services: %s", err)
	}
	if len(servicesSp2) != 2 {
		t.Error("Wrong service count")
	}
	for _, service := range servicesSp2 {
		if !reflect.DeepEqual(service, service3) && !reflect.DeepEqual(service, service4) {
			t.Error("Error getting services")
		}
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestAddUsersService(t *testing.T) {
	// Add services
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	service2 := launcher.Service{"service2", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err = db.AddService(service2)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// Add service to users
	err = db.AddServiceToUsers([]string{"user1"}, "service1")
	if err != nil {
		t.Errorf("Can't add users service: %s", err)
	}

	err = db.AddServiceToUsers([]string{"user2"}, "service2")
	if err != nil {
		t.Errorf("Can't add users service: %s", err)
	}

	// Check user1
	services, err := db.GetUsersServices([]string{"user1"})
	if err != nil {
		t.Errorf("Can't get users services: %s", err)
	}

	if len(services) != 1 {
		t.Error("Wrong service count")
	}

	if services[0].ID != "service1" {
		t.Errorf("Wrong service id: %s", services[0].ID)
	}

	// Check user2
	services, err = db.GetUsersServices([]string{"user2"})
	if err != nil {
		t.Errorf("Can't get users services: %s", err)
	}

	if len(services) != 1 {
		t.Error("Wrong service count")
	}

	if services[0].ID != "service2" {
		t.Errorf("Wrong service id: %s", services[0].ID)
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}

	if err = db.removeAllUsers(); err != nil {
		t.Errorf("Can't remove all users: %s", err)
	}
}

func TestAddSameUsersService(t *testing.T) {
	// Add service
	err := db.AddServiceToUsers([]string{"user0", "user1"}, "service1")
	if err != nil {
		t.Errorf("Can't add users service: %s", err)
	}

	// Add service
	err = db.AddServiceToUsers([]string{"user0", "user1"}, "service1")
	if err == nil {
		t.Error("Error adding same users service")
	}

	// Clear DB
	if err = db.removeAllUsers(); err != nil {
		t.Errorf("Can't remove all users: %s", err)
	}
}

func TestNotExistUsersServices(t *testing.T) {
	// GetService
	_, err := db.GetUsersService([]string{"user2", "user3"}, "service18")
	if err != nil && err != ErrNotExist {
		t.Fatalf("Can't check if service in users: %s", err)
	}

	if err == nil {
		t.Errorf("Error users service: %s", err)
	}
}

func TestRemoveUsersService(t *testing.T) {
	// Add service
	err := db.AddServiceToUsers([]string{"user0", "user1"}, "service1")
	if err != nil {
		t.Errorf("Can't add users service: %s", err)
	}

	// Remove service
	err = db.RemoveServiceFromUsers([]string{"user0", "user1"}, "service1")
	if err != nil {
		t.Errorf("Can't remove users service: %s", err)
	}

	_, err = db.GetUsersService([]string{"user0", "user1"}, "service1")
	if err != nil && err != ErrNotExist {
		t.Fatalf("Can't check if service in users: %s", err)
	}

	if err == nil {
		t.Errorf("Error users service: %s", err)
	}
}

func TestAddUsersList(t *testing.T) {
	numUsers := 5
	numServices := 3

	for i := 0; i < numUsers; i++ {
		users := []string{fmt.Sprintf("user%d", i)}
		for j := 0; j < numServices; j++ {
			err := db.AddServiceToUsers(users, fmt.Sprintf("service%d", j))
			if err != nil {
				t.Errorf("Can't add users service: %s", err)
			}
		}
	}

	// Check users list
	usersList, err := db.getUsersList()
	if err != nil {
		t.Fatalf("Can't get users list: %s", err)
	}

	if len(usersList) != numUsers {
		t.Fatal("Wrong users count")
	}

	for _, users := range usersList {
		ok := false

		for i := 0; i < numUsers; i++ {
			if users[0] == fmt.Sprintf("user%d", i) {
				ok = true
				break
			}
		}

		if !ok {
			t.Errorf("Invalid users: %s", users)
		}
	}

	for j := 0; j < numServices; j++ {
		serviceID := fmt.Sprintf("service%d", j)

		usersServices, err := db.GetUsersServicesByServiceID(serviceID)
		if err != nil {
			t.Errorf("Can't get users services: %s", err)
		}

		for _, userService := range usersServices {
			if userService.ServiceID != serviceID {
				t.Errorf("Invalid serviceID: %s", userService.ServiceID)
			}

			ok := false

			for i := 0; i < numUsers; i++ {
				if userService.Users[0] == fmt.Sprintf("user%d", i) {
					ok = true
					break
				}
			}

			if !ok {
				t.Errorf("Invalid users: %s", userService.Users)
			}
		}

		err = db.RemoveServiceFromAllUsers(serviceID)
		if err != nil {
			t.Errorf("Can't delete users: %s", err)
		}
	}

	usersList, err = db.getUsersList()
	if err != nil {
		t.Fatalf("Can't get users list: %s", err)
	}

	if len(usersList) != 0 {
		t.Fatal("Wrong users count")
	}

	// Clear DB
	if err = db.removeAllUsers(); err != nil {
		t.Errorf("Can't remove all users: %s", err)
	}
}

func TestUsersStorage(t *testing.T) {
	// Add users service
	err := db.AddServiceToUsers([]string{"user1"}, "service1")
	if err != nil {
		t.Errorf("Can't add users service: %s", err)
	}

	// Check default values
	usersService, err := db.GetUsersService([]string{"user1"}, "service1")
	if err != nil {
		t.Errorf("Can't get users service: %s", err)
	}

	if usersService.StorageFolder != "" || len(usersService.StateChecksum) != 0 {
		t.Error("Wrong users service value")
	}

	if err = db.SetUsersStorageFolder([]string{"user1"}, "service1", "stateFolder1"); err != nil {
		t.Errorf("Can't set users storage folder: %s", err)
	}

	if err = db.SetUsersStateChecksum([]string{"user1"}, "service1", []byte{0, 1, 2, 3, 4, 5}); err != nil {
		t.Errorf("Can't set users state checksum: %s", err)
	}

	usersService, err = db.GetUsersService([]string{"user1"}, "service1")
	if err != nil {
		t.Errorf("Can't get users service: %s", err)
	}

	if usersService.StorageFolder != "stateFolder1" || !reflect.DeepEqual(usersService.StateChecksum, []byte{0, 1, 2, 3, 4, 5}) {
		t.Error("Wrong users service value")
	}

	// Clear DB
	if err = db.removeAllUsers(); err != nil {
		t.Errorf("Can't remove all users: %s", err)
	}
}
func TestTrafficMonitor(t *testing.T) {
	setTime := time.Now()
	setValue := uint64(100)

	if err := db.SetTrafficMonitorData("chain1", setTime, setValue); err != nil {
		t.Fatalf("Can't set traffic monitor: %s", err)
	}

	getTime, getValue, err := db.GetTrafficMonitorData("chain1")
	if err != nil {
		t.Fatalf("Can't get traffic monitor: %s", err)
	}

	if !getTime.Equal(setTime) || getValue != setValue {
		t.Fatalf("Wrong value time: %s, value %d", getTime, getValue)
	}

	if err := db.RemoveTrafficMonitorData("chain1"); err != nil {
		t.Fatalf("Can't remove traffic monitor: %s", err)
	}

	if _, _, err := db.GetTrafficMonitorData("chain1"); err == nil {
		t.Fatal("Entry should be removed")
	}

	// Clear DB
	if err := db.removeAllTrafficMonitor(); err != nil {
		t.Errorf("Can't remove all traffic monitor: %s", err)
	}
}

func TestOperationVersion(t *testing.T) {
	var setOperationVersion uint64 = 123

	if err := db.SetOperationVersion(setOperationVersion); err != nil {
		t.Fatalf("Can't set operation version: %s", err)
	}

	getOperationVersion, err := db.GetOperationVersion()
	if err != nil {
		t.Fatalf("Can't get operation version: %s", err)
	}

	if setOperationVersion != getOperationVersion {
		t.Errorf("Wrong operation version: %d", getOperationVersion)
	}
}

func TestCursor(t *testing.T) {
	setCursor := "cursor123"

	if err := db.SetJournalCursor(setCursor); err != nil {
		t.Fatalf("Can't set logging cursor: %s", err)
	}

	getCursor, err := db.GetJournalCursor()
	if err != nil {
		t.Fatalf("Can't get logger cursor: %s", err)
	}

	if getCursor != setCursor {
		t.Fatalf("Wrong cursor value: %s", getCursor)
	}
}

func TestGetServiceByUnitName(t *testing.T) {
	// AddService
	service1 := launcher.Service{"service1", 1, "", "sp1", "to/service1", "service1.service", 5001, 5001, "host1", `{"*":"rw"}`, 0, 0,
		time.Now().UTC(), 0, "", 0, 0, 0, 0, 0, 0, []string{"path1", "path2"}, "", []string{"dbus", "bluez"}, "",
		[]string{"service10/1000/tcp"}, []string{"1001/udp"}}
	err := db.AddService(service1)
	if err != nil {
		t.Errorf("Can't add service: %s", err)
	}

	// GetService
	service, err := db.GetServiceByUnitName("service1.service")
	if err != nil {
		t.Errorf("Can't get service: %s", err)
	}

	if !reflect.DeepEqual(service, service1) {
		t.Error("service1 doesn't match stored one")
	}

	// Clear DB
	if err = db.removeAllServices(); err != nil {
		t.Errorf("Can't remove all services: %s", err)
	}
}

func TestComponentsUpdateInfo(t *testing.T) {
	testData := []umcontroller.SystemComponent{
		{ID: "component1", VendorVersion: "v1", AosVersion: 1,
			Annotations: "Some anotation", URL: "url12", Sha512: []byte{1, 3, 90, 42}},
		{ID: "component2", VendorVersion: "v1", AosVersion: 1, URL: "url12", Sha512: []byte{1, 3, 90, 42}},
	}

	if err := db.SetComponentsUpdateInfo(testData); err != nil {
		t.Fatal("Can't set update manager's update info ", err)
	}

	getUpdateInfo, err := db.GetComponentsUpdateInfo()
	if err != nil {
		t.Fatal("Can't get update manager's update info ", err)
	}

	if !reflect.DeepEqual(testData, getUpdateInfo) {
		t.Fatalf("Wrong update info value: %v", getUpdateInfo)
	}

	testData = []umcontroller.SystemComponent{}

	if err := db.SetComponentsUpdateInfo(testData); err != nil {
		t.Fatal("Can't set update manager's update info ", err)
	}

	getUpdateInfo, err = db.GetComponentsUpdateInfo()
	if err != nil {
		t.Fatal("Can't get update manager's update info ", err)
	}

	if len(getUpdateInfo) != 0 {
		t.Fatalf("Wrong count of update elements 0 != %d", len(getUpdateInfo))
	}
}

func TestMultiThread(t *testing.T) {
	const numIterations = 1000

	var wg sync.WaitGroup

	wg.Add(4)

	go func() {
		defer wg.Done()

		for i := 0; i < numIterations; i++ {
			if err := db.SetOperationVersion(uint64(i)); err != nil {
				t.Fatalf("Can't set operation version: %s", err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		_, err := db.GetOperationVersion()
		if err != nil {
			t.Fatalf("Can't get Operation Version : %s", err)
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < numIterations; i++ {
			if err := db.SetJournalCursor(strconv.Itoa(i)); err != nil {
				t.Fatalf("Can't set journal cursor: %s", err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < numIterations; i++ {
			if _, err := db.GetJournalCursor(); err != nil {
				t.Fatalf("Can't get journal cursor: %s", err)
			}
		}
	}()

	wg.Wait()
}

func TestLayers(t *testing.T) {
	if err := db.AddLayer("sha256:1", "id1", "path1", "1", "1.0", "some layer 1", 1); err != nil {
		t.Errorf("Can't add layer %s", err)
	}

	if err := db.AddLayer("sha256:2", "id2", "path2", "1", "2.0", "some layer 2", 2); err != nil {
		t.Errorf("Can't add layer %s", err)
	}

	if err := db.AddLayer("sha256:3", "id3", "path3", "1", "1.0", "some layer 3", 3); err != nil {
		t.Errorf("Can't add layer %s", err)
	}

	path, err := db.GetLayerPathByDigest("sha256:2")
	if err != nil {
		t.Errorf("Can't get layer path %s", err)
	}

	if path != "path2" {
		t.Errorf("Path form db %s != path2", path)
	}

	if _, err := db.GetLayerPathByDigest("sha256:12345"); err == nil {
		t.Errorf("Should be error: entry does not exist")
	}

	if _, err := db.GetLayerPathByDigest("sha256:12345"); err == nil {
		t.Errorf("Should be error: entry does not exist")
	}

	if err := db.DeleteLayerByDigest("sha256:2"); err != nil {
		t.Errorf("Can't delete layer %s", err)
	}

	layers, err := db.GetLayersInfo()
	if err != nil {
		t.Errorf("Can't get layers info %s", err)
	}

	if len(layers) != 2 {
		t.Errorf("Count of layers in DB %d != 2", len(layers))
	}

	if layers[0].AosVersion != 1 {
		t.Errorf("Layer AosVersion should be 1")
	}
}

func TestMigrationToV1(t *testing.T) {
	migrationDb := path.Join(tmpDir, "test_migration.db")
	mergedMigrationDir := path.Join(tmpDir, "mergedMigration")

	if err := os.MkdirAll(mergedMigrationDir, 0755); err != nil {
		t.Fatalf("Error creating merged migration dir: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(mergedMigrationDir); err != nil {
			t.Fatalf("Error removing merged migration dir: %s", err)
		}
	}()

	if err := createDatabaseV0(migrationDb); err != nil {
		t.Fatalf("Can't create initial database %s", err)
	}

	// Migration upward
	db, err := newDatabase(migrationDb, "migration", mergedMigrationDir, 1)
	if err != nil {
		t.Fatalf("Can't create database: %s", err)
	}

	if err = isDatabaseVer1(db.sql); err != nil {
		t.Fatalf("Error checking db version: %s", err)
	}

	db.Close()

	// Migration downward
	db, err = newDatabase(migrationDb, "migration", mergedMigrationDir, 0)
	if err != nil {
		t.Fatalf("Can't create database: %s", err)
	}

	if err = isDatabaseVer0(db.sql); err != nil {
		t.Fatalf("Error checking db version: %s", err)
	}

	db.Close()
}

/*******************************************************************************
 * Private
 ******************************************************************************/

func (db *Database) getUsersList() (usersList [][]string, err error) {
	rows, err := db.sql.Query("SELECT DISTINCT users FROM users")
	if err != nil {
		return usersList, err
	}
	defer rows.Close()

	usersList = make([][]string, 0)

	for rows.Next() {
		var usersJSON []byte
		err = rows.Scan(&usersJSON)
		if err != nil {
			return usersList, err
		}

		var users []string

		if err = json.Unmarshal(usersJSON, &users); err != nil {
			return usersList, err
		}

		usersList = append(usersList, users)
	}

	return usersList, rows.Err()
}

func createDir(t *testing.T, name string, errorMessage string) {
	if err := os.MkdirAll(name, 0755); err != nil {
		t.Fatalf("%s: %s", errorMessage, err)
	}
}

func removeDir(t *testing.T, name string, errorMessage string) {
	if err := os.RemoveAll(name); err != nil {
		t.Fatalf("%s: %s", errorMessage, err)
	}
}

func createDatabaseV0(name string) (err error) {
	sqlite, err := sql.Open("sqlite3", fmt.Sprintf("%s?_busy_timeout=%d&_journal_mode=%s&_sync=%s",
		name, busyTimeout, journalMode, syncMode))
	if err != nil {
		return err
	}
	defer sqlite.Close()

	if _, err = sqlite.Exec(
		`CREATE TABLE config (
			operationVersion INTEGER,
			cursor TEXT)`); err != nil {
		return err
	}

	if _, err = sqlite.Exec(
		`INSERT INTO config (
			operationVersion,
			cursor) values(?, ?)`, launcher.OperationVersion, ""); err != nil {
		return err
	}

	_, err = sqlite.Exec(`CREATE TABLE IF NOT EXISTS services (id TEXT NOT NULL PRIMARY KEY,
															   aosVersion INTEGER,
															   serviceProvider TEXT,
															   path TEXT,
															   unit TEXT,
															   uid INTEGER,
															   gid INTEGER,
															   hostName TEXT,
															   permissions TEXT,
															   state INTEGER,
															   status INTEGER,
															   startat TIMESTAMP,
															   ttl INTEGER,
															   alertRules TEXT,
															   ulLimit INTEGER,
															   dlLimit INTEGER,
															   ulSpeed INTEGER,
															   dlSpeed INTEGER,
															   storageLimit INTEGER,
															   stateLimit INTEGER,
															   layerList TEXT,
															   deviceResources TEXT,
															   boardResources TEXT,
															   vendorVersion TEXT,
															   description TEXT)`)

	_, err = sqlite.Exec(`CREATE TABLE IF NOT EXISTS users (users TEXT NOT NULL,
															serviceid TEXT NOT NULL,
															storageFolder TEXT,
															stateCheckSum BLOB,
															PRIMARY KEY(users, serviceid))`)

	_, err = sqlite.Exec(`CREATE TABLE IF NOT EXISTS trafficmonitor (chain TEXT NOT NULL PRIMARY KEY,
																	 time TIMESTAMP,
																	 value INTEGER)`)

	_, err = sqlite.Exec(`CREATE TABLE IF NOT EXISTS layers (digest TEXT NOT NULL PRIMARY KEY,
															 layerId TEXT,
															 path TEXT,
															 osVersion TEXT,
															 vendorVersion TEXT,
															 description TEXT,
															 aosVersion INTEGER)`)

	return nil
}

func isDatabaseVer1(sqlite *sql.DB) (err error) {
	rows, err := sqlite.Query("SELECT COUNT(*) AS CNTREC FROM pragma_table_info('config') WHERE name='componentsUpdateInfo'")
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return err
		}

		if count == 0 {
			return ErrNotExist
		}

		break
	}

	servicesRows, err := sqlite.Query("SELECT COUNT(*) AS CNTREC FROM pragma_table_info('services') WHERE name='exposedPorts'")
	if err != nil {
		return err
	}
	defer servicesRows.Close()

	count = 0
	for servicesRows.Next() {
		if err = servicesRows.Scan(&count); err != nil {
			return err
		}

		if count == 0 {
			return ErrNotExist
		}

		return nil
	}

	return ErrNotExist
}

func isDatabaseVer0(sqlite *sql.DB) (err error) {
	rows, err := sqlite.Query("SELECT COUNT(*) AS CNTREC FROM pragma_table_info('config') WHERE name='componentsUpdateInfo'")
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return err
		}

		if count != 0 {
			return ErrNotExist
		}

		break
	}

	servicesRows, err := sqlite.Query("SELECT COUNT(*) AS CNTREC FROM pragma_table_info('config') WHERE name='exposedPorts'")
	if err != nil {
		return err
	}
	defer servicesRows.Close()

	count = 0
	for servicesRows.Next() {
		if err = servicesRows.Scan(&count); err != nil {
			return err
		}

		if count != 0 {
			return ErrNotExist
		}

		return nil
	}

	return ErrNotExist
}
