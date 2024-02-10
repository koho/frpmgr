package services

import (
	"os"

	frpconfig "github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/legacy"
	"github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/util/log"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"
)

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
func VerifyClientConfig(path string) error {
	cfg, proxyCfgs, visitorCfgs, _, err := frpconfig.LoadClientConfig(path, false)
	if err != nil {
		return err
	}
	_, err = validation.ValidateAllClientConfig(cfg, proxyCfgs, visitorCfgs)
	return err
}

// VerifyClientProxy validates the frp proxy
func VerifyClientProxy(source []byte) error {
	proxyCfgs, visitorCfgs, err := legacy.LoadAllProxyConfsFromIni("", source, nil)
	if err != nil {
		return err
	}
	for _, c := range proxyCfgs {
		v1Cfg := legacy.Convert_ProxyConf_To_v1(c)
		v1Cfg.Complete("")
		if err = validation.ValidateProxyConfigurerForClient(v1Cfg); err != nil {
			return err
		}
	}
	for _, c := range visitorCfgs {
		v1Cfg := legacy.Convert_VisitorConf_To_v1(c)
		v1Cfg.Complete(new(v1.ClientCommonConfig))
		if err = validation.ValidateVisitorConfigurer(v1Cfg); err != nil {
			return err
		}
	}
	return nil
}
