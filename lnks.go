package main

import (
	_ "github.com/efjoubert/lnksys/db"
	_ "github.com/efjoubert/lnksys/iorw/active"
	lnks "github.com/efjoubert/lnksys/lnks"
	os "os"
)

func main() {
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
	for argn, arg := range os.Args {
		if arg == "broker" {
			var brokerargs = append(os.Args[:argn], os.Args[argn+1:]...)
			RunBroker(brokerargs...)
			return
		}
	}
	RunService(os.Args...)
}

func RunService(args ...string) {
	var lnksrvs, err = lnks.NewLnkService()
	if err == nil {
		err = lnksrvs.Execute(args...)
	}
	if err != nil {
		println(err)
	}
}

func RunBroker(args ...string) {

}
