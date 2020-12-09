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

package visidentifier_test

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"gitpct.epam.com/epmd-aepr/aos_common/visprotocol"
	"gitpct.epam.com/epmd-aepr/aos_common/wsserver"

	"aos_servicemanager/identification/visidentifier"
	"aos_servicemanager/launcher"
	"aos_servicemanager/pluginprovider"
)

/*******************************************************************************
 * Consts
 ******************************************************************************/

const serverURL = "wss://localhost:8088"

/*******************************************************************************
 * Types
 ******************************************************************************/

type testServiceProvider struct {
	services map[string]*launcher.Service
}

type clientHandler struct {
}

/*******************************************************************************
 * Vars
 ******************************************************************************/

var vis pluginprovider.Identifier
var server *wsserver.Server

var subscriptionID = "test_subscription"

var serviceProvider = testServiceProvider{services: make(map[string]*launcher.Service)}

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

func setup() (err error) {
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())

	url, err := url.Parse(serverURL)
	if err != nil {
		return err
	}

	if server, err = wsserver.New("TestServer", url.Host,
		"../../ci/crt.pem",
		"../../ci/key.pem", new(clientHandler)); err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	if vis, err = visidentifier.New([]byte(`{"VisServer": "wss://localhost:8088"}`)); err != nil {
		return err
	}

	return nil
}

func cleanup() (err error) {
	if err := os.RemoveAll("tmp"); err != nil {
		return err
	}

	if err = vis.Close(); err != nil {
		return err
	}

	return nil
}

/*******************************************************************************
 * Main
 ******************************************************************************/

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		log.Fatalf("Setup error: %s", err)
	}

	ret := m.Run()

	if err := cleanup(); err != nil {
		log.Fatalf("Cleanup error: %s", err)
	}

	os.Exit(ret)
}

/*******************************************************************************
 * Tests
 ******************************************************************************/

func TestGetSystemID(t *testing.T) {
	systemID, err := vis.GetSystemID()
	if err != nil {
		t.Fatalf("Error getting system ID: %s", err)
	}

	if systemID == "" {
		t.Fatalf("Wrong system ID value: %s", systemID)
	}
}

func TestGetUsers(t *testing.T) {
	users, err := vis.GetUsers()
	if err != nil {
		t.Fatalf("Error getting users: %s", err)
	}

	if users == nil {
		t.Fatalf("Wrong users value: %s", users)
	}
}

func TestUsersChanged(t *testing.T) {
	newUsers := []string{generateRandomString(10), generateRandomString(10)}

	message, err := json.Marshal(&visprotocol.SubscriptionNotification{
		Action:         "subscription",
		SubscriptionID: subscriptionID,
		Value:          map[string][]string{"Attribute.Vehicle.UserIdentification.Users": newUsers}})
	if err != nil {
		t.Fatalf("Error marshal request: %s", err)
	}

	clients := server.GetClients()

	for _, client := range clients {
		if err := client.SendMessage(websocket.TextMessage, message); err != nil {
			t.Fatalf("Error send message: %s", err)
		}
	}

	select {
	case users := <-vis.UsersChangedChannel():
		if len(users) != len(newUsers) {
			t.Errorf("Wrong users len: %d", len(users))
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Waiting for users changed timeout")
	}
}

/*******************************************************************************
 * Interfaces
 ******************************************************************************/

func (handler clientHandler) ProcessMessage(client *wsserver.Client, messageType int, message []byte) (response []byte, err error) {
	var header visprotocol.MessageHeader

	if err = json.Unmarshal(message, &header); err != nil {
		return nil, err
	}

	var rsp interface{}

	switch header.Action {
	case visprotocol.ActionSubscribe:
		rsp = &visprotocol.SubscribeResponse{
			MessageHeader:  header,
			SubscriptionID: subscriptionID}

	case visprotocol.ActionGet:
		var getReq visprotocol.GetRequest

		getRsp := visprotocol.GetResponse{
			MessageHeader: header}

		if err = json.Unmarshal(message, &getReq); err != nil {
			return nil, err
		}

		switch getReq.Path {
		case "Attribute.Vehicle.VehicleIdentification.VIN":
			getRsp.Value = map[string]string{getReq.Path: "VIN1234567890"}

		case "Attribute.Vehicle.UserIdentification.Users":
			getRsp.Value = map[string][]string{getReq.Path: []string{"user1", "user2", "user3"}}
		}

		rsp = &getRsp

	default:
		return nil, errors.New("unknown action")
	}

	if response, err = json.Marshal(rsp); err != nil {
		return
	}

	return response, nil
}

func (handler clientHandler) ClientConnected(client *wsserver.Client) {

}

func (handler clientHandler) ClientDisconnected(client *wsserver.Client) {

}

/*******************************************************************************
 * Private
 ******************************************************************************/

func generateRandomString(size uint) (result string) {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	tmp := make([]rune, size)
	for i := range tmp {
		tmp[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(tmp)
}
