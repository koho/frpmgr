package services

import (
	"fmt"
	"github.com/koho/frpmgr/config"
	"github.com/koho/frpmgr/utils"
	"golang.org/x/sys/windows/svc"
	"log"
	"os"
	"path/filepath"
)

func ServiceNameOfConf(name string) string {
	return fmt.Sprintf("FRPC$%s", name)
}

type frpService struct {
	Path string
}

func (service *frpService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	os.Chdir(filepath.Dir(service.Path))
	changes <- svc.Status{State: svc.StartPending}

	defer func() {
		changes <- svc.Status{State: svc.StopPending}
		log.Println("Shutting down")
	}()

	go utils.RunFrpClient(service.Path)

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

func Run(confPath string) error {
	name := config.NameFromPath(confPath)
	serviceName := ServiceNameOfConf(name)
	return svc.Run(serviceName, &frpService{confPath})
}
