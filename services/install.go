package services

import (
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
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

// InstallService runs the program as Windows service
func InstallService(name string, configPath string, manual bool) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}
	path, err := os.Executable()
	if err != nil {
		return err
	}
	if configPath, err = filepath.Abs(configPath); err != nil {
		return err
	}
	serviceName := ServiceNameOfClient(name)
	service, err := m.OpenService(serviceName)
	if err == nil {
		_, err = service.Query()
		if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			service.Close()
			return err
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
			if service != nil {
				service.Close()
			}
			time.Sleep(time.Second / 3)
		}
	}

	conf := mgr.Config{
		ServiceType:  windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:    mgr.StartAutomatic,
		ErrorControl: mgr.ErrorNormal,
		DisplayName:  DisplayNameOfClient(name),
		Description:  "FRP Runtime Service for FRP Manager.",
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

// UninstallService stops and removes the given service
func UninstallService(name string, wait bool) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}
	serviceName := ServiceNameOfClient(name)
	service, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	service.Control(svc.Stop)
	if wait {
		try := 0
		for {
			time.Sleep(time.Second / 3)
			try++
			status, err := service.Query()
			if err != nil {
				return err
			}
			if status.ProcessId == 0 || try >= 3 {
				break
			}
		}
	}
	err = service.Delete()
	err2 := service.Close()
	if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
		return err
	}
	return err2
}

// QueryService returns whether the given service is running
func QueryService(name string) (bool, error) {
	if name == "" {
		return false, os.ErrInvalid
	}
	m, err := serviceManager()
	if err != nil {
		return false, err
	}

	serviceName := ServiceNameOfClient(name)
	service, err := m.OpenService(serviceName)
	if err != nil {
		return false, err
	}
	defer service.Close()
	status, err := service.Query()
	if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
		return false, err
	}
	return status.State != svc.Stopped && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE, nil
}
