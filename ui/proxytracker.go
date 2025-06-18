package ui

import (
	"context"
	"sync"
	"time"

	"github.com/fatedier/frp/client/proxy"
	"github.com/lxn/walk"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/ipc"
	"github.com/koho/frpmgr/services"
)

type ProxyTracker struct {
	sync.RWMutex
	owner              walk.Form
	model              *ProxyModel
	cache              map[string]*config.Proxy
	ctx                context.Context
	cancel             context.CancelFunc
	refreshTimer       *time.Timer
	client             ipc.Client
	rowsInsertedHandle int
	beforeRemoveHandle int
	rowEditedHandle    int
	rowRenamedHandle   int
}

func NewProxyTracker(owner walk.Form, model *ProxyModel, refresh bool) (tracker *ProxyTracker) {
	cache := make(map[string]*config.Proxy)
	ctx, cancel := context.WithCancel(context.Background())
	client := ipc.NewPipeClient(services.ServiceNameOfClient(model.conf.Path), func() []string {
		tracker.RLock()
		defer tracker.RUnlock()
		names := make([]string, 0, len(cache))
		for k := range cache {
			if !cache[k].Disabled {
				names = append(names, k)
			}
		}
		return names
	})
	tracker = &ProxyTracker{
		owner:  owner,
		model:  model,
		cache:  cache,
		ctx:    ctx,
		cancel: cancel,
		client: client,
		rowsInsertedHandle: model.RowsInserted().Attach(func(from, to int) {
			tracker.Lock()
			defer tracker.Unlock()
			for i := from; i <= to; i++ {
				for _, key := range model.items[i].GetAlias() {
					cache[key] = model.items[i].Proxy
				}
			}
			client.Probe(ctx)
		}),
		beforeRemoveHandle: model.BeforeRemove().Attach(func(i int) {
			tracker.Lock()
			defer tracker.Unlock()
			for _, key := range model.items[i].GetAlias() {
				delete(cache, key)
			}
		}),
		rowEditedHandle: model.RowEdited().Attach(func(i int) {
			client.Probe(ctx)
		}),
		rowRenamedHandle: model.RowRenamed().Attach(func(i int) {
			tracker.buildCache()
		}),
	}
	tracker.buildCache()
	client.SetCallback(tracker.onMessage)
	go client.Run(ctx)
	// If no status information is received within a certain period of time,
	// we need to refresh the view to make the icon visible.
	if refresh {
		tracker.refreshTimer = time.AfterFunc(300*time.Millisecond, func() {
			owner.Synchronize(func() {
				if ctx.Err() != nil {
					return
				}
				model.PublishRowsChanged(0, len(model.items)-1)
			})
		})
	}
	return
}

func (pt *ProxyTracker) onMessage(msg []ipc.ProxyMessage) {
	pt.RLock()
	defer pt.RUnlock()
	stat := make(map[*config.Proxy]ipc.ProxyMessage)
	for _, pm := range msg {
		pxy, ok := pt.cache[pm.Name]
		if !ok {
			continue
		}
		_, priority := proxyPhaseToProxyState(pm.Status)
		s, ok := stat[pxy]
		if ok {
			_, prevPriority := proxyPhaseToProxyState(s.Status)
			if prevPriority < priority || (prevPriority == priority && pm.Name < s.Name) {
				stat[pxy] = pm
			}
		} else {
			stat[pxy] = pm
		}
	}
	pt.owner.Synchronize(func() {
		if pt.ctx.Err() != nil {
			return
		}
		for i, item := range pt.model.items {
			if item.Disabled {
				continue
			}
			var statusInfo ProxyStatusInfo
			if m, ok := stat[item.Proxy]; ok {
				state, _ := proxyPhaseToProxyState(m.Status)
				statusInfo = ProxyStatusInfo{
					State:       state,
					Error:       m.Err,
					StateSource: m.Name,
					RemoteAddr:  m.RemoteAddr,
				}
			}
			if item.ProxyStatusInfo != statusInfo {
				item.ProxyStatusInfo = statusInfo
				item.UpdateRemotePort()
				pt.model.PublishRowChanged(i)
				if pt.refreshTimer != nil {
					pt.refreshTimer.Stop()
					pt.refreshTimer = nil
				}
			}
		}
	})
}

func (pt *ProxyTracker) buildCache() {
	pt.Lock()
	defer pt.Unlock()
	clear(pt.cache)
	for _, item := range pt.model.items {
		for _, name := range item.GetAlias() {
			pt.cache[name] = item.Proxy
		}
	}
}

func (pt *ProxyTracker) Close() {
	pt.model.RowsInserted().Detach(pt.rowsInsertedHandle)
	pt.model.BeforeRemove().Detach(pt.beforeRemoveHandle)
	pt.model.RowEdited().Detach(pt.rowEditedHandle)
	pt.model.RowRenamed().Detach(pt.rowRenamedHandle)
	pt.cancel()
	if pt.refreshTimer != nil {
		pt.refreshTimer.Stop()
		pt.refreshTimer = nil
	}
}

func proxyPhaseToProxyState(phase string) (consts.ProxyState, int) {
	switch phase {
	case proxy.ProxyPhaseRunning:
		return consts.ProxyStateRunning, 0
	case proxy.ProxyPhaseStartErr, proxy.ProxyPhaseCheckFailed, proxy.ProxyPhaseClosed:
		return consts.ProxyStateError, 2
	default:
		return consts.ProxyStateUnknown, 1
	}
}
