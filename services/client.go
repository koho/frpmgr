package services

import (
	"context"
	"time"

	_ "github.com/fatedier/frp/assets/frpc"
	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/client/proxy"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/util/log"

	"github.com/koho/frpmgr/pkg/consts"
)

type FrpClientService struct {
	svr            *client.Service
	file           string
	cfg            *v1.ClientCommonConfig
	done           chan struct{}
	statusExporter client.StatusExporter
}

func NewFrpClientService(cfgFile string) (*FrpClientService, error) {
	cfg, pxyCfgs, visitorCfgs, _, err := config.LoadClientConfig(cfgFile, false)
	if err != nil {
		return nil, err
	}
	svr, err := client.NewService(client.ServiceOptions{
		Common:         cfg,
		ProxyCfgs:      pxyCfgs,
		VisitorCfgs:    visitorCfgs,
		ConfigFilePath: cfgFile,
	})
	if err != nil {
		return nil, err
	}
	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)
	return &FrpClientService{
		svr:            svr,
		file:           cfgFile,
		cfg:            cfg,
		done:           make(chan struct{}),
		statusExporter: svr.StatusExporter(),
	}, nil
}

// Run starts frp client service in blocking mode.
func (s *FrpClientService) Run() {
	defer close(s.done)
	if s.file != "" {
		log.Infof("start frpc service for config file [%s]", s.file)
		defer log.Infof("frpc service for config file [%s] stopped", s.file)
	}

	// There's no guarantee that this function will return after a close call.
	// So we can't wait for the Run function to finish.
	if err := s.svr.Run(context.Background()); err != nil {
		log.Errorf("run service error: %v", err)
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
	_, pxyCfgs, visitorCfgs, _, err := config.LoadClientConfig(s.file, false)
	if err != nil {
		return err
	}
	return s.svr.UpdateAllConfigurer(pxyCfgs, visitorCfgs)
}

func (s *FrpClientService) Done() <-chan struct{} {
	return s.done
}

func (s *FrpClientService) GetProxyStatus(name string) (status *proxy.WorkingStatus, ok bool) {
	proxyName := name
	if s.cfg.User != "" {
		proxyName = s.cfg.User + "." + name
	}
	status, ok = s.statusExporter.GetProxyStatus(proxyName)
	if ok {
		status.Name = name
		if status.Err == "" {
			if status.Type == consts.ProxyTypeTCP || status.Type == consts.ProxyTypeUDP {
				status.RemoteAddr = s.cfg.ServerAddr + status.RemoteAddr
			}
		} else {
			status.RemoteAddr = ""
		}
	}
	return
}
