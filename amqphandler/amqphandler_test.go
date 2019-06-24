package amqphandler_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gitpct.epam.com/epmd-aepr/aos_servicemanager/amqphandler"
)

/*******************************************************************************
 * Const
 ******************************************************************************/

const (
	inQueueName  = "in_queue"
	outQueueName = "out_queue"
	consumerName = "test_consumer"
	exchangeName = "test_exchange"
)

/*******************************************************************************
 * Types
 ******************************************************************************/

type backendClient struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	delivery   <-chan amqp.Delivery
	errChannel chan *amqp.Error
}

type sendData struct {
	version     uint64
	messageType string
	data        interface{}
}

/*******************************************************************************
 * Vars
 ******************************************************************************/

var testClient backendClient

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

	if testClient.conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/"); err != nil {
		return err
	}

	if testClient.channel, err = testClient.conn.Channel(); err != nil {
		return err
	}

	if _, err = testClient.channel.QueueDeclare(inQueueName, false, false, false, false, nil); err != nil {
		return err
	}

	if _, err = testClient.channel.QueueDeclare(outQueueName, false, false, false, false, nil); err != nil {
		return err
	}

	if err = testClient.channel.ExchangeDeclare(exchangeName, "fanout", false, false, false, false, nil); err != nil {
		return err
	}

	if err = testClient.channel.QueueBind(inQueueName, "", exchangeName, false, nil); err != nil {
		return err
	}

	if testClient.delivery, err = testClient.channel.Consume(inQueueName, "", true, false, false, false, nil); err != nil {
		return err
	}

	testClient.errChannel = testClient.conn.NotifyClose(make(chan *amqp.Error, 1))

	return nil
}

func cleanup() {
	if testClient.channel != nil {
		testClient.channel.QueueDelete(inQueueName, false, false, false)
		testClient.channel.QueueDelete(outQueueName, false, false, false)
		testClient.channel.ExchangeDelete(exchangeName, false, false)
		testClient.channel.Close()
	}

	if testClient.conn != nil {
		testClient.conn.Close()
	}

	if err := os.RemoveAll("tmp"); err != nil {
		log.Errorf("Can't remove tmp folder: %s", err)
	}
}

func sendMessage(correlationID string, version uint64, messageType string, data interface{}) (err error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	tmpData := make(map[string]interface{})

	if err = json.Unmarshal(dataJSON, &tmpData); err != nil {
		return err
	}

	tmpData["version"] = version
	tmpData["messageType"] = messageType

	if dataJSON, err = json.Marshal(tmpData); err != nil {
		return err
	}

	log.Debug(string(dataJSON))

	return testClient.channel.Publish(
		"",
		outQueueName,
		false,
		false,
		amqp.Publishing{
			CorrelationId: correlationID,
			ContentType:   "text/plain",
			Body:          dataJSON})
}

/*******************************************************************************
 * Main
 ******************************************************************************/

func TestMain(m *testing.M) {

	if err := setup(); err != nil {
		log.Fatalf("Error creating service images: %s", err)
	}

	ret := m.Run()

	cleanup()

	os.Exit(ret)
}

/*******************************************************************************
 * Tests
 ******************************************************************************/

func TestSendMessages(t *testing.T) {
	amqpHandler, err := amqphandler.New()
	if err != nil {
		t.Fatalf("Can't create amqp: %s", err)
	}
	defer amqpHandler.Close()

	if err = amqpHandler.ConnectRabbit("localhost", "guest", "guest",
		exchangeName, consumerName, outQueueName); err != nil {
		t.Fatalf("Can't connect to server: %s", err)
	}

	testData := []sendData{
		sendData{1, "stateAcceptance", &amqphandler.StateAcceptance{
			ServiceID: "service0", Checksum: "0123456890", Result: "accepted", Reason: "just because"}},
		sendData{1, "updateState", &amqphandler.UpdateState{
			ServiceID: "service1", Checksum: "0993478847", State: "This is new state"}},
		sendData{1, "requestServiceLog", &amqphandler.RequestServiceLog{
			ServiceID: "service2", LogID: uuid.New().String(), From: &time.Time{}, Till: &time.Time{}}},
		sendData{1, "requestServiceCrashLog", &amqphandler.RequestServiceCrashLog{
			ServiceID: "service3", LogID: uuid.New().String()}},
	}

	for _, message := range testData {
		correlationID := uuid.New().String()

		if err = sendMessage(correlationID, message.version, message.messageType, message.data); err != nil {
			t.Errorf("Can't send message: %s", err)
			continue
		}

		select {
		case receiveMessage := <-amqpHandler.MessageChannel:
			if !reflect.DeepEqual(message.data, receiveMessage.Data) {
				t.Errorf("Wrong data received: %v %v", message.data, receiveMessage.Data)
				continue
			}

			if correlationID != receiveMessage.CorrelationID {
				t.Errorf("Wrong correlation ID received: %s %s", correlationID, receiveMessage.CorrelationID)
				continue
			}

		case err = <-testClient.errChannel:
			t.Fatalf("AMQP error: %s", err)
			return

		case <-time.After(5 * time.Second):
			t.Error("Waiting data timeout")
			continue
		}
	}
}

func TestReceiveMessages(t *testing.T) {
	amqpHandler, err := amqphandler.New()
	if err != nil {
		t.Fatalf("Can't create amqp: %s", err)
	}
	defer amqpHandler.Close()

	if err = amqpHandler.ConnectRabbit("localhost", "guest", "guest",
		exchangeName, consumerName, outQueueName); err != nil {
		t.Fatalf("Can't connect to server: %s", err)
	}

	type messageDesc struct {
		correlationID string
		call          func() error
		data          interface{}
		getDataType   func() interface{}
	}

	type messageHeader struct {
		Version     uint64
		MessageType string
	}

	initialSetupData := []amqphandler.ServiceInfo{
		amqphandler.ServiceInfo{ID: "service0", Version: 1, Status: "running", Error: "", StateChecksum: "1234567890"},
		amqphandler.ServiceInfo{ID: "service1", Version: 2, Status: "stopped", Error: "crash", StateChecksum: "1234567890"},
		amqphandler.ServiceInfo{ID: "service2", Version: 3, Status: "unknown", Error: "unknown", StateChecksum: "1234567890"},
	}

	monitoringData := amqphandler.MonitoringData{Timestamp: time.Now().Local()}
	monitoringData.Data.Global.RAM = 1024
	monitoringData.Data.Global.CPU = 50
	monitoringData.Data.Global.UsedDisk = 2048
	monitoringData.Data.Global.InTraffic = 8192
	monitoringData.Data.Global.OutTraffic = 4096
	monitoringData.Data.ServicesData = []amqphandler.ServiceMonitoringData{
		amqphandler.ServiceMonitoringData{ServiceID: "service0", RAM: 1024, CPU: 50, UsedDisk: 100000},
		amqphandler.ServiceMonitoringData{ServiceID: "service1", RAM: 128, CPU: 60, UsedDisk: 200000},
		amqphandler.ServiceMonitoringData{ServiceID: "service2", RAM: 256, CPU: 70, UsedDisk: 300000},
		amqphandler.ServiceMonitoringData{ServiceID: "service3", RAM: 512, CPU: 80, UsedDisk: 400000}}

	sendNewStateCorrelationID := uuid.New().String()

	pushServiceLogError := "Error"
	var pushServiceLogPartCount uint64 = 2
	var pushServiceLogPart uint64 = 1
	pushServiceLogData := amqphandler.PushServiceLog{
		LogID:     "log0",
		PartCount: &pushServiceLogPartCount,
		Part:      &pushServiceLogPart,
		Data:      &[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Error:     &pushServiceLogError}

	var alertVersion uint64 = 2

	alertsData := amqphandler.Alerts{
		Data: []amqphandler.AlertItem{
			amqphandler.AlertItem{
				Timestamp: time.Now().Local(),
				Tag:       amqphandler.AlertTagSystemError,
				Source:    "system",
				Payload:   map[string]interface{}{"Message": "System error"},
			},
			amqphandler.AlertItem{
				Timestamp: time.Now().Local(),
				Tag:       amqphandler.AlertTagSystemError,
				Source:    "service 1",
				Version:   &alertVersion,
				Payload:   map[string]interface{}{"Message": "Service crashed"},
			},
			amqphandler.AlertItem{
				Timestamp: time.Now().Local(),
				Tag:       amqphandler.AlertTagResource,
				Source:    "system",
				Payload:   map[string]interface{}{"Parameter": "cpu", "Value": float64(100)},
			},
		}}

	testData := []messageDesc{
		messageDesc{
			call: func() error {
				return amqpHandler.SendInitialSetup(initialSetupData)
			},
			data: &struct {
				messageHeader
				Services []amqphandler.ServiceInfo
			}{
				messageHeader: messageHeader{
					Version:     1,
					MessageType: "vehicleStatus"},
				Services: initialSetupData},
			getDataType: func() interface{} {
				return &struct {
					messageHeader
					Services []amqphandler.ServiceInfo
				}{}
			},
		},

		messageDesc{
			call: func() error {
				return amqpHandler.SendServiceStatus(initialSetupData[0])
			},
			data: &struct {
				messageHeader
				Services []amqphandler.ServiceInfo
			}{
				messageHeader: messageHeader{
					Version:     1,
					MessageType: "serviceStatus"},
				Services: []amqphandler.ServiceInfo{initialSetupData[0]}},
			getDataType: func() interface{} {
				return &struct {
					messageHeader
					Services []amqphandler.ServiceInfo
				}{}
			},
		},

		messageDesc{
			call: func() error {
				return amqpHandler.SendMonitoringData(monitoringData)
			},
			data: &struct {
				messageHeader
				amqphandler.MonitoringData
			}{
				messageHeader: messageHeader{
					Version:     1,
					MessageType: "monitoringData"},
				MonitoringData: monitoringData},
			getDataType: func() interface{} {
				return &struct {
					messageHeader
					amqphandler.MonitoringData
				}{}
			},
		},

		messageDesc{
			correlationID: sendNewStateCorrelationID,
			call: func() error {
				return amqpHandler.SendNewState("service0", "This is state", "12345679", sendNewStateCorrelationID)
			},
			data: &struct {
				messageHeader
				ServiceID     string
				StateChecksum string
				State         string
			}{
				messageHeader: messageHeader{
					Version:     1,
					MessageType: "newState"},
				ServiceID:     "service0",
				StateChecksum: "12345679",
				State:         "This is state"},
			getDataType: func() interface{} {
				return &struct {
					messageHeader
					ServiceID     string
					StateChecksum string
					State         string
				}{}
			},
		},

		messageDesc{
			call: func() error {
				return amqpHandler.SendStateRequest("service1", true)
			},
			data: &struct {
				messageHeader
				ServiceID string
				Default   bool
			}{
				messageHeader: messageHeader{
					Version:     1,
					MessageType: "stateRequest"},
				ServiceID: "service1",
				Default:   true},
			getDataType: func() interface{} {
				return &struct {
					messageHeader
					ServiceID string
					Default   bool
				}{}
			},
		},

		messageDesc{
			call: func() error {
				return amqpHandler.SendServiceLog(pushServiceLogData)
			},
			data: &struct {
				messageHeader
				amqphandler.PushServiceLog
			}{
				messageHeader: messageHeader{
					Version:     1,
					MessageType: "pushServiceLog"},
				PushServiceLog: pushServiceLogData},
			getDataType: func() interface{} {
				return &struct {
					messageHeader
					amqphandler.PushServiceLog
				}{}
			},
		},

		messageDesc{
			call: func() error {
				return amqpHandler.SendAlerts(alertsData)
			},
			data: &struct {
				messageHeader
				amqphandler.Alerts
			}{
				messageHeader: messageHeader{
					Version:     1,
					MessageType: "alerts"},
				Alerts: alertsData},
			getDataType: func() interface{} {
				return &struct {
					messageHeader
					amqphandler.Alerts
				}{}
			},
		},
	}

	for _, message := range testData {
		if err = message.call(); err != nil {
			t.Errorf("Can't perform call: %s", err)
			continue
		}

		select {
		case delivery := <-testClient.delivery:
			receiveData := message.getDataType()

			if err = json.Unmarshal(delivery.Body, receiveData); err != nil {
				t.Errorf("Error parsing message: %s", err)
				continue
			}

			if message.correlationID != delivery.CorrelationId {
				t.Errorf("Wrong correlation ID received: %s %s", message.correlationID, delivery.CorrelationId)
			}

			if !reflect.DeepEqual(receiveData, message.data) {
				t.Errorf("Wrong data received: %v %v", message.data, receiveData)
			}

		case err = <-testClient.errChannel:
			t.Fatalf("AMQP error: %s", err)
			return

		case <-time.After(5 * time.Second):
			t.Error("Waiting data timeout")
			continue
		}
	}
}
