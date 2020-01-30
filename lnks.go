package main

import (
	lnks "github.com/efjoubert/lnksys/lnks"
	//runtime "runtime"
	//runtimedbg "runtime/debug"
)

func main() {
	//runtimedbg.SetGCPercent(25)
	//runtime.GOMAXPROCS(runtime.NumCPU() * 8)

	lnks.RunService()
}
