package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
	"unsafe"

	"github.com/fatedier/frp/pkg/util/log"

	"github.com/koho/frpmgr/pkg/util"
)

const defaultDnsCheckInterval = 300 * time.Second

type FrpClientSVCBService struct {
	*FrpClientService
	ctx        context.Context
	cancel     context.CancelFunc
	serverAddr string
	addrBuf    []byte
}

func NewFrpClientSVCBService(cfgFile string) (*FrpClientSVCBService, error) {
	cs, err := NewFrpClientService(cfgFile)
	if err != nil {
		return nil, err
	}
	s := &FrpClientSVCBService{
		FrpClientService: cs,
		serverAddr:       cs.cfg.ServerAddr,
		addrBuf:          make([]byte, 255),
	}
	if net.ParseIP(cs.cfg.ServerAddr) != nil {
		return nil, fmt.Errorf("server address is not a domain")
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Expand server address buffer, so that
	// there's no out-of-bound error for concurrent access.
	for i := range s.addrBuf {
		s.addrBuf[i] = '.'
	}
	copy(s.addrBuf, s.serverAddr)
	newAddr := s.addrBuf[:len(s.serverAddr)]
	s.cfg.ServerAddr = *(*string)(unsafe.Pointer(&newAddr))

	return s, nil
}

// Run periodically resolves the server domain SVCB record
// and updates the config in an unsafe way.
func (s *FrpClientSVCBService) Run() {
	defer s.cancel()
	var run bool
	var ip string
	var port uint16
	var tryCnt int

	timer := time.NewTimer(time.Duration(0))
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			newIP, newPort, err := util.ResolveSVCB(s.ctx, s.serverAddr, s.cfg.DNSServer)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Warnf("lookup %s SVCB error: %v", s.serverAddr, err)
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
		case <-s.done:
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
	s.cfg.ServerAddr = *(*string)(unsafe.Pointer(&newAddr))
}

func (s *FrpClientSVCBService) setPort(port uint16) {
	s.cfg.ServerPort = int(port)
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
