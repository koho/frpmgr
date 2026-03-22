package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/client/configmgmt"
	"github.com/fatedier/frp/client/proxy"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/source"
	"github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/util/log"
	_ "github.com/fatedier/frp/web/frpc"
	glog "github.com/fatedier/golib/log"

	"github.com/koho/frpmgr/pkg/consts"
)

type FrpClientService struct {
	svr            *client.Service
	file           string
	cfg            *v1.ClientCommonConfig
	done           chan struct{}
	statusExporter client.StatusExporter
	logger         *glog.RotateFileWriter
}

func NewFrpClientService(cfgFile string) (*FrpClientService, error) {
	result, err := config.LoadClientConfigResult(cfgFile, false)
	if err != nil {
		return nil, err
	}
	configSource := source.NewConfigSource()
	if err := configSource.ReplaceAll(result.Proxies, result.Visitors); err != nil {
		return nil, fmt.Errorf("failed to set config source: %w", err)
	}

	var storeSource *source.StoreSource

	if result.Common.Store.IsEnabled() {
		storePath := result.Common.Store.Path
		if storePath != "" && cfgFile != "" && !filepath.IsAbs(storePath) {
			storePath = filepath.Join(filepath.Dir(cfgFile), storePath)
		}

		s, err := source.NewStoreSource(source.StoreSourceConfig{
			Path: storePath,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create store source: %w", err)
		}
		storeSource = s
	}

	aggregator := source.NewAggregator(configSource)
	if storeSource != nil {
		aggregator.SetStoreSource(storeSource)
	}

	proxyCfgs, visitorCfgs, err := aggregator.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from sources: %w", err)
	}

	proxyCfgs, visitorCfgs = config.FilterClientConfigurers(result.Common, proxyCfgs, visitorCfgs)
	proxyCfgs = config.CompleteProxyConfigurers(proxyCfgs)
	visitorCfgs = config.CompleteVisitorConfigurers(visitorCfgs)

	_, err = validation.ValidateAllClientConfig(result.Common, proxyCfgs, visitorCfgs, nil)
	if err != nil {
		return nil, err
	}

	svr, err := client.NewService(client.ServiceOptions{
		Common:                 result.Common,
		ConfigSourceAggregator: aggregator,
		ConfigFilePath:         cfgFile,
	})
	if err != nil {
		return nil, err
	}
	logger := initLogger(result.Common.Log.To, result.Common.Log.Level, int(result.Common.Log.MaxDays))
	return &FrpClientService{
		svr:            svr,
		file:           cfgFile,
		cfg:            result.Common,
		done:           make(chan struct{}),
		statusExporter: svr.StatusExporter(),
		logger:         logger,
	}, nil
}

// Run starts frp client service in blocking mode.
func (s *FrpClientService) Run() {
	defer close(s.done)
	if s.file != "" {
		log.Infof("start frpc service for config file [%s] with aggregated configuration", s.file)
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
	result, err := config.LoadClientConfigResult(s.file, false)
	if err != nil {
		return fmt.Errorf("%w: %v", configmgmt.ErrInvalidArgument, err)
	}

	proxyCfgsForValidation, visitorCfgsForValidation := config.FilterClientConfigurers(
		result.Common,
		result.Proxies,
		result.Visitors,
	)
	proxyCfgsForValidation = config.CompleteProxyConfigurers(proxyCfgsForValidation)
	visitorCfgsForValidation = config.CompleteVisitorConfigurers(visitorCfgsForValidation)

	if _, err := validation.ValidateAllClientConfig(result.Common, proxyCfgsForValidation, visitorCfgsForValidation, nil); err != nil {
		return fmt.Errorf("%w: %v", configmgmt.ErrInvalidArgument, err)
	}

	if err := s.svr.UpdateConfigSource(result.Common, result.Proxies, result.Visitors); err != nil {
		return fmt.Errorf("%w: %v", configmgmt.ErrApplyConfig, err)
	}
	return nil
}

func (s *FrpClientService) Done() <-chan struct{} {
	return s.done
}

func (s *FrpClientService) GetProxyStatus(name string) (status *proxy.WorkingStatus, ok bool) {
	status, ok = s.statusExporter.GetProxyStatus(name)
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

func initLogger(logPath string, levelStr string, maxDays int) *glog.RotateFileWriter {
	var options []glog.Option
	writer := glog.NewRotateFileWriter(glog.RotateFileConfig{
		FileName: logPath,
		Mode:     glog.RotateFileModeDaily,
		MaxDays:  maxDays,
	})
	writer.Init()
	options = append(options, glog.WithOutput(writer))
	level, err := glog.ParseLevel(levelStr)
	if err != nil {
		level = glog.InfoLevel
	}
	options = append(options, glog.WithLevel(level))
	log.Logger = log.Logger.WithOptions(options...)
	return writer
}
