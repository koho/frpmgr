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
	serviceName := ServiceNameOfClient(configPath)
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
		for i := 0; i < 2; i++ {
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
func UninstallService(configPath string, wait bool) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}
	serviceName := ServiceNameOfClient(configPath)
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
				service.Close()
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

// QueryStartInfo returns the start type and process id of the given service.
func QueryStartInfo(configPath string) (uint32, uint32, error) {
	m, err := serviceManager()
	if err != nil {
		return 0, 0, err
	}
	serviceName := ServiceNameOfClient(configPath)
	service, err := m.OpenService(serviceName)
	if err != nil {
		return 0, 0, err
	}
	defer service.Close()
	cfg, err := service.Config()
	if err != nil {
		return 0, 0, err
	}
	var pid uint32
	if status, err := service.Query(); err == nil {
		pid = status.ProcessId
	}
	return cfg.StartType, pid, nil
}
