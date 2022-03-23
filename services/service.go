package services

import (
	"fmt"
	"github.com/koho/frpmgr/pkg/util"
	"golang.org/x/sys/windows/svc"
	"log"
	"os"
	"path/filepath"
)

func ServiceNameOfClient(name string) string {
	return fmt.Sprintf("FRPC$%s", name)
}

func ServiceNameOfServer(name string) string {
	return fmt.Sprintf("FRPS$%s", name)
}

type frpService struct {
	configPath string
}

func (service *frpService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	if err := os.Chdir(filepath.Dir(service.configPath)); err != nil {
		return
	}
	changes <- svc.Status{State: svc.StartPending}

	defer func() {
		changes <- svc.Status{State: svc.StopPending}
		log.Println("Shutting down")
	}()

	go runFrpClient()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	log.Println("Startup complete")

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				return
			case svc.Interrogate:
				changes <- c.CurrentStatus
			default:
				log.Printf("Unexpected services control request #%d\n", c)
			}
		}
	}
}

func Run(configPath string) error {
	baseName, _ := util.SplitExt(configPath)
	serviceName := ServiceNameOfClient(baseName)
	return svc.Run(serviceName, &frpService{configPath})
}
