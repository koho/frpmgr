package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"time"
	"unsafe"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/util/log"

	"github.com/koho/frpmgr/pkg/util"
)

const defaultDnsCheckInterval = 300 * time.Second

type FrpClientSVCBService struct {
	*FrpClientService
	cfg        reflect.Value
	ctx        context.Context
	cancel     context.CancelFunc
	serverAddr string
	dnsAddr    string
	addrBuf    []byte
}

func NewFrpClientSVCBService(cfgFile string) (*FrpClientSVCBService, error) {
	service := new(FrpClientSVCBService)
	cfg, pxyCfgs, visitorCfgs, _, err := config.LoadClientConfig(cfgFile, false)
	if err != nil {
		return nil, err
	}
	if net.ParseIP(cfg.ServerAddr) != nil {
		return nil, fmt.Errorf("server address is not a domain")
	}

	service.serverAddr = cfg.ServerAddr
	// Expand server address buffer, so that
	// there's no out-of-bound error for concurrent access.
	service.addrBuf = make([]byte, 255)
	for i := range service.addrBuf {
		service.addrBuf[i] = '.'
	}
	copy(service.addrBuf, cfg.ServerAddr)
	newAddr := service.addrBuf[:len(cfg.ServerAddr)]
	cfg.ServerAddr = *(*string)(unsafe.Pointer(&newAddr))

	svr, err := client.NewService(client.ServiceOptions{
		Common:         cfg,
		ProxyCfgs:      pxyCfgs,
		VisitorCfgs:    visitorCfgs,
		ConfigFilePath: cfgFile,
	})
	if err != nil {
		return nil, err
	}
	log.InitLog(cfg.Log.To, cfg.Log.Level, cfg.Log.MaxDays, cfg.Log.DisablePrintColor)

	service.ctx, service.cancel = context.WithCancel(context.Background())
	service.cfg = reflect.ValueOf(svr).Elem().FieldByName("cfg")
	service.dnsAddr = cfg.DNSServer
	service.FrpClientService = &FrpClientService{
		svr:  svr,
		file: cfgFile,
	}
	return service, nil
}

// Run periodically resolves the server domain SVCB record
// and updates the config in an unsafe way.
func (s *FrpClientSVCBService) Run() {
	var run bool
	var ip string
	var port uint16
	var tryCnt int

	timer := time.NewTimer(time.Duration(0))
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			newIP, newPort, err := util.ResolveSVCB(s.ctx, s.serverAddr, s.dnsAddr)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Warn("lookup %s SVCB error: %v", s.serverAddr, err)
				// Backs off if the resolution has failed in some way.
				timer.Reset(s.backoff(tryCnt))
				// Prevent counter overflow.
				if tryCnt < tryCnt+1 {
					tryCnt++
				}
				continue
			}
			// As the domain has been resolved, removes the need for backoff
			// for the next retry by resetting the try count.
			tryCnt = 0

			if newIP != ip {
				ip = newIP
				s.setAddress(newIP)
			}
			if newPort == 0 {
				newPort = 7000
			}
			if newPort != port {
				port = newPort
				s.setPort(newPort)
			}
			if !run {
				run = true
				go s.FrpClientService.Run()
			}
			timer.Reset(defaultDnsCheckInterval)
		case <-s.ctx.Done():
			return
		}
	}
}

// Although string assignment is not thread-safe, addrBuf has enough space to store
// ip address, preventing from out-of-bounds errors. If an incorrect ip address is read,
// frp will reconnect with another read.
func (s *FrpClientSVCBService) setAddress(addr string) {
	copy(s.addrBuf, addr)
	newAddr := s.addrBuf[:len(addr)]
	reflect.NewAt(s.cfg.Type(), unsafe.Pointer(s.cfg.UnsafeAddr())).
		Elem().FieldByName("ServerAddr").SetString(*(*string)(unsafe.Pointer(&newAddr)))
}

func (s *FrpClientSVCBService) setPort(port uint16) {
	reflect.NewAt(s.cfg.Type(), unsafe.Pointer(s.cfg.UnsafeAddr())).
		Elem().FieldByName("ServerPort").SetInt(int64(port))
}

func (s *FrpClientSVCBService) Stop(wait bool) {
	s.cancel()
	s.FrpClientService.Stop(wait)
}

// backoff returns the amount of time to wait before the next retry given the
// number of retries.
func (s *FrpClientSVCBService) backoff(retries int) time.Duration {
	baseDelay, multiplier, jitter, maxDelay := 1.0*time.Second, 1.6, 0.2, defaultDnsCheckInterval
	if retries == 0 {
		return baseDelay
	}
	backoff, max := float64(baseDelay), float64(maxDelay)
	for backoff < max && retries > 0 {
		backoff *= multiplier
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays
	backoff *= 1 + jitter*(rand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}
