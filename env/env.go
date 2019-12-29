package env

import (
	network "github/efjoubert/lnksys/network"
)

var wrapupcalls []func()

func ShutdownEnvironment() {
	network.ShutdownListener()
	if len(wrapupcalls) > 0 {
		for _, wrpupcall := range wrapupcalls {
			wrpupcall()
		}
	}
}

func WrapupCall(wrpupcall ...func()) {
	if len(wrpupcall) > 0 {
		if len(wrapupcalls) == 0 {
			wrapupcalls = []func(){}
		}
		wrapupcalls = append(wrapupcalls, wrpupcall...)
	}
}

func init() {
	network.RegisterShutdownEnv(ShutdownEnvironment)
}
