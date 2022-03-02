package services

import (
	"errors"
	"github.com/koho/frpmgr/config"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
	"os"
	"time"
)

var cachedServiceManager *mgr.Mgr

func serviceManager() (*mgr.Mgr, error) {
	if cachedServiceManager != nil {
		return cachedServiceManager, nil
	}
	m, err := mgr.Connect()
	if err != nil {
		return nil, err
	}
	cachedServiceManager = m
	return cachedServiceManager, nil
}

func InstallService(configPath string, manual bool) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}
	path, err := os.Executable()
	if err != nil {
		return err
	}

	name := config.NameFromPath(configPath)
	serviceName := ServiceNameOfConf(name)
	service, err := m.OpenService(serviceName)
	if err == nil {
		status, err := service.Query()
		if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			service.Close()
			return err
		}
		if status.State != svc.Stopped && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			service.Close()
			return errors.New("service already installed and running")
		}
		err = service.Delete()
		service.Close()
		if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			return err
		}
		for {
			service, err = m.OpenService(serviceName)
			if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
				break
			}
			service.Close()
			time.Sleep(time.Second / 3)
		}
	}

	conf := mgr.Config{
		ServiceType:  windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:    mgr.StartAutomatic,
		ErrorControl: mgr.ErrorNormal,
		DisplayName:  "FRP Client: " + name,
		Description:  "FRP Client Daemon Service",
		SidType:      windows.SERVICE_SID_TYPE_UNRESTRICTED,
	}
	if manual {
		conf.StartType = mgr.StartManual
	}
	service, err = m.CreateService(serviceName, path, conf, "-c", configPath)
	if err != nil {
		return err
	}

	err = service.Start()
	service.Close()
	return err
}

func UninstallService(name string) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}
	serviceName := ServiceNameOfConf(name)
	service, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	service.Control(svc.Stop)
	err = service.Delete()
	err2 := service.Close()
	if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
		return err
	}
	return err2
}

func QueryService(name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	m, err := serviceManager()
	if err != nil {
		return false, err
	}

	serviceName := ServiceNameOfConf(name)
	service, err := m.OpenService(serviceName)
	if err == nil {
		defer service.Close()
		status, err := service.Query()
		if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			return false, err
		}
		return status.State != svc.Stopped && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE, nil
	}
	return false, err
}
