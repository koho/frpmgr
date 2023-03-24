package services

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"

	_ "github.com/fatedier/frp/assets/frpc"
	"github.com/fatedier/frp/client"
	frpconfig "github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/util/log"
)

func runFrpClient(data []byte) {
	cfg, pxyCfgs, visitorCfgs, err := parseClientConfig(data)
	if err != nil {
		os.Exit(1)
	}
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel,
		cfg.LogMaxDays, cfg.DisableLogColor)

	svr, err := client.NewService(cfg, pxyCfgs, visitorCfgs, "")
	if err != nil {
		os.Exit(1)
	}

	closedDoneCh := make(chan struct{})
	shouldGracefulClose := cfg.Protocol == "kcp" || cfg.Protocol == "quic"
	// Capture the exit signal if we use kcp or quic.
	if shouldGracefulClose {
		go handleSignal(svr, closedDoneCh)
	}

	err = svr.Run()
	if err == nil && shouldGracefulClose {
		<-closedDoneCh
	}
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
	data, err := config.ReadFile(path)
	if err != nil {
		return err
	}
	_, _, _, err = parseClientConfig(data)
	return
}

// VerifyClientProxy validates the frp proxy
func VerifyClientProxy(source []byte) (err error) {
	_, _, err = frpconfig.LoadAllProxyConfsFromIni("", source, nil)
	return
}

func parseClientConfig(data []byte) (
	cfg frpconfig.ClientCommonConf,
	pxyCfgs map[string]frpconfig.ProxyConf,
	visitorCfgs map[string]frpconfig.VisitorConf,
	err error,
) {
	var content []byte
	content, err = frpconfig.RenderContent(data)
	if err != nil {
		return
	}
	configBuffer := bytes.NewBuffer(nil)
	configBuffer.Write(content)

	// Parse common section.
	cfg, err = frpconfig.UnmarshalClientConfFromIni(content)
	if err != nil {
		return
	}
	cfg.Complete()
	if err = cfg.Validate(); err != nil {
		err = fmt.Errorf("parse config error: %v", err)
		return
	}

	// Aggregate proxy configs from include files.
	var buf []byte
	buf, err = getIncludeContents(cfg.IncludeConfigFiles)
	if err != nil {
		err = fmt.Errorf("getIncludeContents error: %v", err)
		return
	}
	configBuffer.WriteString("\n")
	configBuffer.Write(buf)

	// Parse all proxy and visitor configs.
	pxyCfgs, visitorCfgs, err = frpconfig.LoadAllProxyConfsFromIni(cfg.User, configBuffer.Bytes(), cfg.Start)
	if err != nil {
		return
	}
	return
}

func handleSignal(svr *client.Service, doneCh chan struct{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svr.GracefulClose(500 * time.Millisecond)
	close(doneCh)
}

// getIncludeContents renders all configs from paths.
// files format can be a single file path or directory or regex path.
func getIncludeContents(paths []string) ([]byte, error) {
	out := bytes.NewBuffer(nil)
	for _, path := range paths {
		absDir, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			return nil, err
		}
		files, err := os.ReadDir(absDir)
		if err != nil {
			return nil, err
		}
		for _, fi := range files {
			if fi.IsDir() {
				continue
			}
			absFile := filepath.Join(absDir, fi.Name())
			if matched, _ := filepath.Match(filepath.Join(absDir, filepath.Base(path)), absFile); matched {
				tmpContent, err := frpconfig.GetRenderedConfFromFile(absFile)
				if err != nil {
					return nil, fmt.Errorf("render extra config %s error: %v", absFile, err)
				}
				out.Write(tmpContent)
				out.WriteString("\n")
			}
		}
	}
	return out.Bytes(), nil
}
