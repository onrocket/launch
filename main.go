package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/onrocket/launch/binfiles"
	"github.com/onrocket/launch/config"
	"github.com/onrocket/launch/get"
)

func main() {

	var hostName string
	envHost, isSet := os.LookupEnv("HOST")
	var err error
	if isSet {
		hostName = envHost
	} else {
		hostName, err = os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
	}
	var serviceName = "WBcMTtwPdXC435m42"
	envServiceName, isSet := os.LookupEnv("SERVICE_NAME")
	if isSet {
		serviceName = envServiceName
	}

	binfilesPtr := flag.Bool("binfiles", false, "load binfiles - not implemented yet")
	getPtr := flag.Bool("get", false, "get config from etcd")
	listenJobPtr := flag.Bool("listen-job", false, "listen for jobs from beanstalkd")
	listenLogPtr := flag.Bool("listen-log", false, "listen for logs of jobs from beanstalkd")
	loadPtr := flag.Bool("load", false, "load config to etcd")
	runPtr := flag.Bool("run", false, "get and run config from etcd")

	nodeNamePtr := flag.String("nodeName", hostName, "name of node")

	flag.Parse()

	fmt.Println("   node : ", *nodeNamePtr)
	fmt.Println("service : ", serviceName)
	d := new(get.OnRocket)
	switch {
	case *binfilesPtr:
		fmt.Println("loading binary files ...")
		binfiles.LoadBinFiles()
	case *getPtr:
		fmt.Println("getting config ... ")
		d.DownloadConfig(*nodeNamePtr, serviceName)
	case *listenJobPtr:
		fmt.Println("listening for jobs from beanstalkd ...")
		d.LentilJobListener()
	case *listenLogPtr:
		fmt.Println("listening for logs of jobs from beanstalkd ...")
		d.LentilLogListener()
	case *loadPtr:
		fmt.Println("loading config ...")
		config.LoadConfig()
	case *runPtr:
		fmt.Println("getting and running config ... ")
		d.DownloadConfig(*nodeNamePtr, serviceName)
	}
}
