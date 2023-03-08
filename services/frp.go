package services

import (
	"os"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"

	_ "github.com/fatedier/frp/assets/frpc"
	frpc "github.com/fatedier/frp/cmd/frpc/sub"
	frpconfig "github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/util/log"
)

func runFrpClient() {
	// Change program arguments for frpc to parse
	// No need to change it for now
	frpc.Execute()
}

func deleteFrpConfig(serviceName string, configPath string, c config.Config) {
	// Delete logs
	log.Log.Close()
	if logs, _, err := util.FindLogFiles(c.GetLogFile()); err == nil {
		util.DeleteFiles(logs)
	}
	// Delete config file
	os.Remove(configPath)
	// Delete service
	m, err := serviceManager()
	if err != nil {
		return
	}
	defer m.Disconnect()
	service, err := m.OpenService(serviceName)
	if err != nil {
		return
	}
	defer service.Close()
	service.Delete()
}

// VerifyClientConfig validates the frp client config file
func VerifyClientConfig(path string) (err error) {
	_, _, _, err = frpconfig.ParseClientConfig(path)
	return
}

// VerifyClientProxy validates the frp proxy
func VerifyClientProxy(source []byte) (err error) {
	_, _, err = frpconfig.LoadAllProxyConfsFromIni("", source, nil)
	return
}
