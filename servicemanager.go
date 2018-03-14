package main

import (
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"

	amqp "gitpct.epam.com/epmd-aepr/aos_servicemanager/amqphandler"
	"gitpct.epam.com/epmd-aepr/aos_servicemanager/downloadmanager"
	"gitpct.epam.com/epmd-aepr/aos_servicemanager/launcher"
)

//TODO
//- add tls to downloadmanager
//- add encript image

type appInfo struct {
	Name string
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		FullTimestamp:    true})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
}

func sendInitalSetup(launcher *launcher.Launcher) {
	initialList, err := launcher.GetServicesInfo()
	if err != nil {
		log.Error("erro get inital list ", err)
		//todo return
	}
	amqp.SendInitialSetup(initialList)
}

func processAmqpReturn(data interface{}, launcher *launcher.Launcher, output chan string) bool {
	switch data := data.(type) {
	case error:
		log.Warning("receive error from amqp ", data)
		amqp.CloseAllConnections()
		return false
	case amqp.ServiseInfoFromCloud:
		version, err := launcher.GetServiceVersion(data.Id)
		if err != nil {
			log.Warning("error get version ", err)
			break
		}
		if data.Version > version {
			log.Debug("send download request url ", data.Url)
			go downloadmanager.DownloadPkg("/tmp", data.Url, output)
		}

		return true
	default:
		log.Info("receive some data amqp")

		return true
	}
	return true
}

func main() {
	log.Info("Start service manager")
	defer func() {
		log.Info("Stop service manager")
	}()

	amqp.CloseAllConnections()
	out := make(chan string)

	//go downloadmanager.DownloadPkg("./", "https://kor.ill.in.ua/m/610x385/2122411.jpg", out)
	//go downloadmanager.DownloadPkg("./test/", "http://speedtest.tele2.net/100MB.zip", out)

	launcher, err := launcher.New(path.Join(os.Getenv("GOPATH"), "aos"))
	if err != nil {
		log.Fatal("Can't create launcher")
	}
	defer launcher.Close()
	sendInitalSetup(launcher)

	for {
		log.Debug("start connection")
		amqpChan, err := amqp.InitAmqphandler("serviseDiscoveryURL")

		if err != nil {
			log.Error("Can't esablish connection ", err)
			time.Sleep(3 * time.Second)
			continue
		}
		sendInitalSetup(launcher)
		for {
			log.Debug("start select")
			select {
			case amqpReturn := <-amqpChan:
				isContinue := processAmqpReturn(amqpReturn, launcher, out)
				if isContinue != true {
					break
				}
			case msg := <-out:
				log.Debug("Save file here: %v", msg)
				launcher.InstallService(msg) //todo add erro handling
			}
		}
	}

}
