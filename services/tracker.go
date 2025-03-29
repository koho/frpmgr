package services

import (
	"sync"
	"sync/atomic"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/koho/frpmgr/pkg/consts"
)

type ConfigStateCallback func(path string, state consts.ConfigState)

type tracker struct {
	service *mgr.Service
	done    sync.WaitGroup
	once    atomic.Uint32
}

var (
	trackedConfigs     = make(map[string]*tracker)
	trackedConfigsLock = sync.Mutex{}
)

func trackExistingConfigs(paths func() []string, cb ConfigStateCallback) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}
	for _, path := range paths() {
		trackedConfigsLock.Lock()
		if ctx := trackedConfigs[path]; ctx != nil {
			cfg, err := ctx.service.Config()
			trackedConfigsLock.Unlock()
			if (err != nil || cfg.StartType == windows.SERVICE_DISABLED) && ctx.once.CompareAndSwap(0, 1) {
				ctx.done.Done()
				cb(path, consts.StateStopped)
			}
			continue
		}
		trackedConfigsLock.Unlock()
		serviceName := ServiceNameOfClient(path)
		service, err := m.OpenService(serviceName)
		if err != nil {
			continue
		}
		go trackService(service, path, cb)
	}
	return nil
}

func WatchConfigServices(paths func() []string, cb ConfigStateCallback) (func() error, error) {
	m, err := serviceManager()
	if err != nil {
		return nil, err
	}
	var subscription uintptr
	err = windows.SubscribeServiceChangeNotifications(m.Handle, windows.SC_EVENT_DATABASE_CHANGE,
		windows.NewCallback(func(notification uint32, context uintptr) uintptr {
			trackExistingConfigs(paths, cb)
			return 0
		}), 0, &subscription)
	if err == nil {
		if err = trackExistingConfigs(paths, cb); err != nil {
			windows.UnsubscribeServiceChangeNotifications(subscription)
			return nil, err
		}
		return func() error {
			err := windows.UnsubscribeServiceChangeNotifications(subscription)
			trackedConfigsLock.Lock()
			for _, tc := range trackedConfigs {
				tc.done.Done()
			}
			trackedConfigsLock.Unlock()
			return err
		}, nil
	}
	return nil, err
}

func trackService(service *mgr.Service, path string, cb ConfigStateCallback) {
	trackedConfigsLock.Lock()
	if _, found := trackedConfigs[path]; found {
		trackedConfigsLock.Unlock()
		service.Close()
		return
	}

	defer func() {
		service.Close()
	}()
	ctx := &tracker{service: service}
	ctx.done.Add(1)
	trackedConfigs[path] = ctx
	trackedConfigsLock.Unlock()
	defer func() {
		trackedConfigsLock.Lock()
		delete(trackedConfigs, path)
		trackedConfigsLock.Unlock()
	}()

	var subscription uintptr
	lastState := consts.StateUnknown
	var updateState = func(state consts.ConfigState) {
		if state != lastState {
			cb(path, state)
			lastState = state
		}
	}
	err := windows.SubscribeServiceChangeNotifications(service.Handle, windows.SC_EVENT_STATUS_CHANGE,
		windows.NewCallback(func(notification uint32, context uintptr) uintptr {
			if ctx.once.Load() != 0 {
				return 0
			}
			configState := consts.StateUnknown
			if notification == 0 {
				status, err := service.Query()
				if err == nil {
					configState = svcStateToConfigState(uint32(status.State))
				}
			} else {
				configState = notifyStateToConfigState(notification)
			}
			updateState(configState)
			return 0
		}), 0, &subscription)
	if err == nil {
		defer windows.UnsubscribeServiceChangeNotifications(subscription)
		status, err := service.Query()
		if err == nil {
			updateState(svcStateToConfigState(uint32(status.State)))
		}
		ctx.done.Wait()
	} else {
		cb(path, consts.StateStopped)
		service.Control(svc.Stop)
	}
}

func svcStateToConfigState(s uint32) consts.ConfigState {
	switch s {
	case windows.SERVICE_STOPPED:
		return consts.StateStopped
	case windows.SERVICE_START_PENDING:
		return consts.StateStarting
	case windows.SERVICE_STOP_PENDING:
		return consts.StateStopping
	case windows.SERVICE_RUNNING:
		return consts.StateStarted
	case windows.SERVICE_NO_CHANGE:
		return 0
	default:
		return 0
	}
}

func notifyStateToConfigState(s uint32) consts.ConfigState {
	if s&(windows.SERVICE_NOTIFY_STOPPED|windows.SERVICE_NOTIFY_DELETED|windows.SERVICE_NOTIFY_DELETE_PENDING) != 0 {
		return consts.StateStopped
	} else if s&windows.SERVICE_NOTIFY_STOP_PENDING != 0 {
		return consts.StateStopping
	} else if s&windows.SERVICE_NOTIFY_RUNNING != 0 {
		return consts.StateStarted
	} else if s&windows.SERVICE_NOTIFY_START_PENDING != 0 {
		return consts.StateStarting
	} else {
		return consts.StateUnknown
	}
}
