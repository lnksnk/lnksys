package service

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/service"
)

//Service struct
type Service struct {
	isService   bool
	isConsole   bool
	isBroker	bool
	start       func(*Service, ...string)
	run         func(*Service, ...string)
	stop        func(*Service, ...string)
	execname    string
	execfolder  string
	name        string
	displayName string
	description string
	svcConfig   *service.Config
	args        []string
}

//Start Service
func (svr *Service) Start(s service.Service) error {
	if svr.start != nil {
		if svr.isService {
			go svr.start(svr, svr.args...)
		} else if svr.isConsole {
			svr.start(svr, svr.args...)
		}
	}

	if svr.isService {
		go svr.exec()
	} else if svr.isConsole {
		svr.exec()
	}
	return nil
}

func (svr *Service) exec() {
	if svr.run != nil {
		if svr.isService || svr.isConsole || svr.isBroker {
			svr.run(svr, svr.args...)
		}
	}
}

//Stop Service
func (svr *Service) Stop(s service.Service) error {
	if svr.stop != nil {
		if svr.isService || svr.isConsole || svr.isBroker {
			svr.stop(svr, svr.args...)
		}
	}
	return nil
}

//IsBroker Service
func (svr *Service) IsBroker() bool {
	return svr.isBroker
}

//IsConsole Service
func (svr *Service) IsConsole() bool {
	return svr.isConsole
}

//IsService Service
func (svr *Service) IsService() bool {
	return svr.isService
}

//ServiceExeName Service Executable Name
func (svr *Service) ServiceExeName() string {
	return svr.execname
}

//ServiceName Service Name
func (svr *Service) ServiceName() string {
	return svr.name
}

//ServiceExeFolder local folder where Service Executable resides
func (svr *Service) ServiceExeFolder() string {
	return svr.execfolder
}

//ServiceDisplayName Service Display Name
func (svr *Service) ServiceDisplayName() string {
	return svr.displayName
}

//ServiceDescription Service Description
func (svr *Service) ServiceDescription() string {
	return svr.description
}

//NewService invoke new *Service
//name - ServiceName
//displayName - ServiceDisplayName
//description - ServiceDescription
//start - func(*Service, ...string) implementation, gets invoked when Service Start
//run - func(*Service, ...string) implementation, gets invoked when Service Run
//stop - func(*Service, ...string) implementation, gets invoked when Service Stop
func NewService(name string, displayName string, description string, start func(*Service, ...string),
	run func(*Service, ...string),
	stop func(*Service, ...string)) (svr *Service, err error) {
	if run != nil {
		execname, _ := os.Executable()
		execname = strings.Replace(execname, "\\", "/", -1)
		execfolder, _ := ExecutableFolder()
		execfolder = strings.Replace(execfolder, "\\", "/", -1)
		if name == "" {
			if execname != "" && execfolder != "" {
				execname = execname[len(execfolder)+1:]
			}
			name = execname
			if si := strings.Index(name, "."); si > -1 {
				name = name[0:si]
			}
		}

		if displayName == "" {
			displayName = name
		}

		if description == "" {
			description = strings.ToUpper(displayName)
		}
		//svcargs := []string{}

		svcConfig := &service.Config{
			Name:        name,
			DisplayName: displayName,
			Description: description,
		}

		svr = &Service{execfolder: execfolder, execname: execname, start: start, run: run, stop: stop, name: name, displayName: displayName, description: description, svcConfig: svcConfig}
	}
	return svr, err
}

var logger service.Logger

//Execute main Service Execute method when executing Service
//called in main() func of golang app,
//args - args from os gets passed into here
func (svr *Service) Execute(args ...string) (err error) {
	svcargs := []string{}
	canappendargs := false
	if len(args) > 0 {
		if len(args[1:]) > 0 {
			for _, arg := range args[1:] {
				if arg == "install" {
					canappendargs = true
				} else if strings.Index(",start,stop,restart,install,uninstall,console,", ","+arg+",") > -1 {
					canappendargs = false
					break
				} else {
					svcargs = append(svcargs, arg)
				}
			}
		} else {
			//svr.isConsole = true
		}
	}
	if len(svcargs) > 0 {
		if canappendargs {
			svr.svcConfig.Arguments = svcargs
		} else {
			svr.isConsole = true
		}
	}

	if s, serr := service.New(svr, svr.svcConfig); serr == nil {
		if logger, err = s.Logger(nil); err == nil {
			argFound := ""
			svr.args = args[:]
			for _, arg := range svr.args {
				if strings.Index(",start,stop,restart,install,uninstall,", ","+arg+",") > -1 {
					argFound = arg
					svr.isService = true
					if err = service.Control(s, argFound); err == nil {
						break
					}
				} else if strings.Index(",console,", ","+arg+",") > -1 {
					svr.isConsole = true
					break
				} else if strings.Index(",broker,", ","+arg+",") > -1 {
					svr.isBroker = true
					break
				}
				if err != nil {
					break
				}
			}
			if err == nil && argFound == "" {
				if !svr.isService {
					svr.isService = !svr.isConsole
				}
				if svr.isService {
					err = s.Run()
				} else if svr.isConsole {
					svr.Start(s)
					svr.Stop(s)
				} else if svr.isBroker {
					svr.Start(s)
					svr.Stop(s)
				}
			}
		}
	} else {
		err = serr
	}

	if err != nil {
		logger.Error(err)
	}

	return err
}

var cx, ce = executableClean()

func executableClean() (string, error) {
	p, err := executable()
	return filepath.Clean(p), err
}

func executable() (string, error) {
	return os.Executable()
}

// Executable returns an absolute path that can be used to
// re-invoke the current program.
// It may not be valid after the current program exits.
func Executable() (string, error) {
	return cx, ce
}

// Returns same path as Executable, returns just the folder
// path. Excludes the executable name and any trailing slash.
func ExecutableFolder() (string, error) {
	p, err := Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(p), nil
}
