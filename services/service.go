package services

import (
	"fmt"
	"github.com/fatedier/golib/crypto"
	"github.com/koho/frpmgr/config"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	frpc "github.com/fatedier/frp/cmd/frpc/sub"
	"golang.org/x/sys/windows/svc"
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

	go func() {
		crypto.DefaultSalt = "frp"
		rand.Seed(time.Now().UnixNano())
		err := frpc.RunClient(service.Path)
		if err != nil {
			os.Exit(1)
		}
	}()

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
