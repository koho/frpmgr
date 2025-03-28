package services

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/fatedier/frp/pkg/util/log"
	"golang.org/x/sys/windows/svc"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"
)

type Service interface {
	// Run service in blocking mode.
	Run()
	// Reload config file.
	Reload() error
	// Stop service and cleanup resources.
	Stop(wait bool)
	// Done returns a channel that's closed when work done.
	Done() <-chan struct{}
}

func ServiceNameOfClient(configPath string) string {
	return fmt.Sprintf("frpmgr_%x", md5.Sum([]byte(util.FileNameWithoutExt(configPath))))
}

func DisplayNameOfClient(name string) string {
	return "FRP Manager: " + name
}

type frpService struct {
	configPath string
}

func (service *frpService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	path, err := os.Executable()
	if err != nil {
		return
	}
	if err = os.Chdir(filepath.Dir(path)); err != nil {
		return
	}
	changes <- svc.Status{State: svc.StartPending}

	defer func() {
		changes <- svc.Status{State: svc.StopPending}
	}()

	cc, err := config.UnmarshalClientConf(service.configPath)
	if err != nil {
		return
	}
	var expired <-chan time.Time
	t, err := config.Expiry(service.configPath, cc.AutoDelete)
	switch err {
	case nil:
		if t <= 0 {
			deleteFrpConfig(args[0], service.configPath, cc)
			return
		}
		expired = time.After(t)
	case os.ErrNoDeadline:
		break
	default:
		return
	}

	var svr Service
	if cc.SVCBEnable && net.ParseIP(cc.ServerAddress) == nil {
		// WARNING: Experimental feature.
		svr, err = NewFrpClientSVCBService(service.configPath)
	} else {
		svr, err = NewFrpClientService(service.configPath)
	}
	if err != nil {
		return
	}

	go svr.Run()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptParamChange}

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				svr.Stop(false)
				return
			case svc.ParamChange:
				// Reload service
				if err = svr.Reload(); err != nil {
					log.Errorf("reload frp config error: %v", err)
				}
			case svc.Interrogate:
				changes <- c.CurrentStatus
			default:
			}
		case <-svr.Done():
			return
		case <-expired:
			svr.Stop(false)
			deleteFrpConfig(args[0], service.configPath, cc)
			return
		}
	}
}

// Run executes frp service in background service process.
func Run(configPath string) error {
	serviceName := ServiceNameOfClient(configPath)
	return svc.Run(serviceName, &frpService{configPath})
}

// ReloadService sends a reload event to the frp service
// which triggers hot-reloading of frp configuration.
func ReloadService(configPath string) error {
	m, err := serviceManager()
	if err != nil {
		return err
	}

	svcName := ServiceNameOfClient(configPath)
	service, err := m.OpenService(svcName)
	if err != nil {
		return err
	}
	defer service.Close()
	_, err = service.Control(svc.ParamChange)
	return err
}
