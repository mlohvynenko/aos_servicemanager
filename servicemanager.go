package main

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	amqp "gitpct.epam.com/epmd-aepr/aos_servicemanager/amqphandler"
	"gitpct.epam.com/epmd-aepr/aos_servicemanager/downloadmanager"
	"gitpct.epam.com/epmd-aepr/aos_servicemanager/launcher"
)

const (
	aosReconnectTimeSec = 3
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		FullTimestamp:    true})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
}

func sendInitalSetup(launcher *launcher.Launcher, handler *amqp.AmqpHandler) (err error) {
	initialList, err := launcher.GetServicesInfo()
	if err != nil {
		log.Error("Error getting initial list: ", err)
		return err
	}

	if handler.SendInitialSetup(initialList) != nil {
		log.Error("Error sending initial setup: ", err)
		return err
	}

	return nil
}

func installService(launcher *launcher.Launcher, servInfo amqp.ServiceInfoFromCloud) {
	imageFile, err := downloadmanager.DownloadPkg(servInfo)
	if imageFile != "" {
		defer os.Remove(imageFile)
	}
	if err != nil {
		log.Error("Can't download package: ", err)
		return
	}

	if err := launcher.InstallService(imageFile, servInfo.Id, servInfo.Version); err != nil {
		log.WithFields(log.Fields{"id": servInfo.Id, "version": servInfo.Version}).Error("Can't install service: ", err)
	}
}

func removeService(launcher *launcher.Launcher, id string) {
	if err := launcher.RemoveService(id); err != nil {
		log.WithField("id", id).Error("Can't remove service: ", err)
	}
}

func processAmqpMessage(data interface{}, handler *amqp.AmqpHandler, launcher *launcher.Launcher) (err error) {
	switch data := data.(type) {
	case []amqp.ServiceInfoFromCloud:
		log.WithField("len", len(data)).Info("Recive services info")

		currenList, err := launcher.GetServicesInfo()
		if err != nil {
			log.Error("Error getting services info: ", err)
			return err
		}

		for iCur := len(currenList) - 1; iCur >= 0; iCur-- {
			for iDes := len(data) - 1; iDes >= 0; iDes-- {
				if data[iDes].Id == currenList[iCur].Id {
					if data[iDes].Version > currenList[iCur].Version {
						log.Info("Update ", data[iDes].Id, " from ", currenList[iCur].Version, " to ", data[iDes].Version)

						go installService(launcher, data[iDes])
					}

					data = append(data[:iDes], data[iDes+1:]...)
					currenList = append(currenList[:iCur], currenList[iCur+1:]...)
				}
			}
		}

		for _, deleteElemnt := range currenList {
			go removeService(launcher, deleteElemnt.Id)
		}

		for _, newElement := range data {
			go installService(launcher, newElement)
		}

		return nil

	default:
		log.Warn("Receive unsupported amqp message: ", reflect.TypeOf(data))
		return nil
	}
}

func main() {
	log.Info("Start service manager")

	launcher, err := launcher.New("data")
	if err != nil {
		log.Fatal("Can't create launcher: ", err)
	}
	defer launcher.Close()

	amqpHandler, err := amqp.New()
	if err != nil {
		log.Fatal("Can't create amqpHandler: ", err)
	}
	defer amqpHandler.CloseAllConnections()

	// handle SIGTERM
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		launcher.Close()
		amqpHandler.CloseAllConnections()
		os.Exit(1)
	}()

	for {
		log.Debug("Start connection")

		amqpChan, err := amqpHandler.InitAmqphandler("https://fusion-poc-2.cloudapp.net:9000")
		if err != nil {
			log.Error("Can't esablish connection: ", err)
			log.Debug("Reconnecting...")
			time.Sleep(time.Second * aosReconnectTimeSec)
			continue
		}

		sendInitalSetup(launcher, amqpHandler)

		for {
			amqpMessage := <-amqpChan

			// check for error
			if err, ok := amqpMessage.(error); ok {
				log.Error("Receive amqp error: ", err)
				log.Debug("Reconnecting...")
				break
			}

			if err := processAmqpMessage(amqpMessage, amqpHandler, launcher); err != nil {
				log.Error("Error processing amqp result: ", err)
			}
		}
	}
}