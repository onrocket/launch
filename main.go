package main

import (
	"github.com/onrocket/launch/binfiles"
	"github.com/onrocket/launch/config"
	"github.com/onrocket/launch/get"

	"flag"
	"fmt"
)

func main() {

	binfilesPtr := flag.Bool("binfiles", false, "load binfiles - not implemented yet")
	getPtr := flag.Bool("get", false, "get config from etcd")
	listenPtr := flag.Bool("listen", false, "listen for jobs from beanstalkd")
	loadPtr := flag.Bool("load", false, "load config to etcd")
	runPtr := flag.Bool("run", false, "get and run config from etcd")

	nodeNamePtr := flag.String("nodeName", "ajbc.co", "name of node")
	serviceNamePtr := flag.String("serviceName", "WBcMTtwPdXC435m42", "name of service")

	flag.Parse()

	fmt.Println("   node : ", *nodeNamePtr)
	fmt.Println("service : ", *serviceNamePtr)
	d := new(get.OnRocket)
	switch {
	case *binfilesPtr:
		fmt.Println("loading binary files ...")
		binfiles.LoadBinFiles()
	case *getPtr:
		fmt.Println("getting config ... ")
		d.DownloadConfig(*nodeNamePtr, *serviceNamePtr)
	case *listenPtr:
		fmt.Println("listening for jobs from beanstalkd ...")
		d.LentilListener()
	case *loadPtr:
		fmt.Println("loading config ...")
		config.LoadConfig()
	case *runPtr:
		fmt.Println("getting and running config ... ")
		d.DownloadConfig(*nodeNamePtr, *serviceNamePtr)
	}

}
