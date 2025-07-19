package services

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/fatedier/frp/pkg/util/log"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/ipc"
	"github.com/koho/frpmgr/pkg/util"
)

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
			deleteFrpFiles(args[0], service.configPath, cc.LogFile)
			return
		}
		expired = time.After(t)
	case os.ErrNoDeadline:
		break
	default:
		return
	}

	svr, err := NewFrpClientService(service.configPath)
	if err != nil {
		return
	}

	is, err := ipc.NewServer(args[0], svr)
	if err != nil {
		return
	}
	defer is.Close()

	go svr.Run()
	go is.Run()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptParamChange}

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				svr.Stop(false)
				if code := shutdownReason(path); code > 0 {
					return false, code
				}
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
			svr.logger.Close()
			deleteFrpFiles(args[0], service.configPath, cc.LogFile)
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

func shutdownReason(path string) uint32 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	fileID, err := winio.GetFileID(f)
	f.Close()
	if err != nil {
		return 0
	}
	name, err := syscall.UTF16PtrFromString(fmt.Sprintf("Global\\%x%x", fileID.VolumeSerialNumber, fileID.FileID))
	if err != nil {
		return 0
	}
	if h, err := windows.OpenEvent(windows.READ_CONTROL, false, name); err == nil {
		windows.CloseHandle(h)
		return uint32(windows.ERROR_FAIL_NOACTION_REBOOT)
	}
	return 0
}
