package main

import (
	_ "github.com/efjoubert/lnksys/db"
	_ "github.com/efjoubert/lnksys/iorw/active"
	lnks "github.com/efjoubert/lnksys/lnks"
	network "github.com/efjoubert/lnksys/network"
	os "os"
	runtime "runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	/*var tlkr=network.NewTalker()
	tlkr.Send("https://www.google.com")
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	*/
	/*
		network.MapRoots("/", "./", "resources/", "./resources", "apps/", "./apps")

		network.DefaultServeHttp(os.Stdout, "GET", "/@lnks.conf@.js", nil)

		//db.DBMSManager().RegisterDbms("avon","driver=sqlserver","username=PTOOLS","password=PTOOLS","host=134.65.204.106/PRESENCE")
		//db.DBMSManager().RegisterDbms("lnks", "driver=postgres", "username=lnksys", "password=lnksyslnksys", "host=127.0.0.1:5432", "sslmode=disable")
		//db.DBMSManager().RegisterDbms("tidb","driver=mysql","username=root","host=127.0.0.1:4000")
		//network.InvokeServer("0.0.0.0:1111")
		var d = make(chan bool, 1)

		active.MapGlobal("SHUTDOWNENV", func() {
			d <- true
		})
		var running = true
		for running {
			select {
			case e := <-d:
				if e {
					running = false
					break
				}
			}
		}*/
	RunService(os.Args...)
}

func RunService(args ...string) {
	var lnksrvs, err = lnks.NewLnkService(RunBroker)
	if err == nil {
		err = lnksrvs.Execute(args...)
	}
	if err != nil {
		println(err)
	}
}

func RunBroker(exename string, exealias string, args ...string) {
	network.BrokerServeHttp(os.Stdout, os.Stdin, exename, exealias, args...)
}
