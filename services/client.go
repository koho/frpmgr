package services

import (
	"time"

	_ "github.com/fatedier/frp/assets/frpc"
	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/util/log"
)

type FrpClientService struct {
	svr  *client.Service
	file string
}

func NewFrpClientService(cfgFile string) (*FrpClientService, error) {
	cfg, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(cfgFile)
	if err != nil {
		return nil, err
	}
	svr, err := client.NewService(cfg, pxyCfgs, visitorCfgs, cfgFile)
	if err != nil {
		return nil, err
	}
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel,
		cfg.LogMaxDays, cfg.DisableLogColor)
	return &FrpClientService{svr: svr, file: cfgFile}, nil
}

// Run starts frp client service in blocking mode.
func (s *FrpClientService) Run() {
	if s.file != "" {
		log.Trace("start frpc service for config file [%s]", s.file)
		defer log.Trace("frpc service for config file [%s] stopped", s.file)
	}

	// There's no guarantee that this function will return after a close call.
	// So we can't wait for the Run function to finish.
	if err := s.svr.Run(); err != nil {
		log.Error("run service error: %v", err)
	}
}

// Stop closes all frp connections.
func (s *FrpClientService) Stop(wait bool) {
	// Close client service.
	if wait {
		s.svr.GracefulClose(500 * time.Millisecond)
	} else {
		s.svr.Close()
	}
}

// Reload creates or updates or removes proxies of frpc.
func (s *FrpClientService) Reload() error {
	_, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(s.file)
	if err != nil {
		return err
	}
	return s.svr.ReloadConf(pxyCfgs, visitorCfgs)
}
