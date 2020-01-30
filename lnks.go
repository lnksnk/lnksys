package main

import (
	_ "github.com/efjoubert/lnksys/db"
	_ "github.com/efjoubert/lnksys/iorw/active"
	lnks "github.com/efjoubert/lnksys/lnks"
	network "github.com/efjoubert/lnksys/network"
	os "os"
	//runtime "runtime"
	//runtimedbg "runtime/debug"
)

func main() {
	//runtimedbg.SetGCPercent(25)
	//runtime.GOMAXPROCS(runtime.NumCPU() * 8)

	lnks.RunService(os.Args...)
}

