package ipc

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/Microsoft/go-winio"
)

type PipeClient struct {
	path    string
	payload func() []string
	ch      chan struct{}
	cb      func([]ProxyMessage)
}

func NewPipeClient(name string, payload func() []string) *PipeClient {
	return &PipeClient{
		path:    `\\.\pipe\` + name,
		payload: payload,
		ch:      make(chan struct{}),
	}
}

func (p *PipeClient) SetCallback(cb func([]ProxyMessage)) {
	p.cb = cb
}

func (p *PipeClient) Run(ctx context.Context) {
	conn, err := winio.DialPipeContext(ctx, p.path)
	if err != nil {
		return
	}
	defer conn.Close()
	timer := time.NewTimer(0)
	defer timer.Stop()

	seq := []time.Duration{100 * time.Millisecond, 500 * time.Millisecond, time.Second, 2 * time.Second, 5 * time.Second}
	index := -1

	query := func() {
		if err = gob.NewEncoder(conn).Encode(p.payload()); err != nil {
			return
		}
		var msg []ProxyMessage
		if err = gob.NewDecoder(conn).Decode(&msg); err != nil {
			return
		}
		if p.cb != nil {
			p.cb(msg)
		}
	}
	for {
		select {
		case <-timer.C:
			query()
			if index < len(seq)-1 {
				index++
			}
			timer.Reset(seq[index])
		case <-p.ch:
			index = 0
			timer.Reset(seq[index])
		case <-ctx.Done():
			return
		}
	}
}

func (p *PipeClient) Probe(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case p.ch <- struct{}{}:
	default:
		return
	}
}
