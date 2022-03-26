package ui

import (
	"fmt"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type SimpleProxyDialog struct {
	*walk.Dialog

	Proxies []*config.Proxy
	binder  *quickAddBinder
	db      *walk.DataBinder

	// title of the dialog
	title string
	// icon of the dialog
	icon *walk.Icon
	// The local service name
	service string
	// types of the proxy
	types []string
}

// NewSimpleProxyDialog creates proxies connecting to the local service
func NewSimpleProxyDialog(title string, icon *walk.Icon, service string, types []string, localAddr string) *SimpleProxyDialog {
	v := &SimpleProxyDialog{title: title, icon: icon, service: service, types: types}
	v.Proxies = make([]*config.Proxy, 0)
	ip, sep, port := util.Partition(localAddr, ":")
	if sep == "" {
		return nil
	}
	v.binder = &quickAddBinder{LocalAddr: ip, LocalPort: port}
	if v.binder.LocalAddr == "" {
		v.binder.LocalAddr = "127.0.0.1"
	}
	return v
}

func (sp *SimpleProxyDialog) Run(owner walk.Form) (int, error) {
	widgets := []Widget{
		Label{Text: "远程端口", ColumnSpan: 2},
		LineEdit{Text: Bind("RemotePort", Regexp{"^\\d+$"}), ColumnSpan: 2},
		Label{Text: "本地地址"},
		Label{Text: "端口"},
		LineEdit{Text: Bind("LocalAddr", Regexp{".+"}), StretchFactor: 2},
		LineEdit{Text: Bind("LocalPort", Regexp{"^\\d+$"}), StretchFactor: 1},
	}
	switch sp.service {
	case "ftp":
		widgets = append(widgets, Label{Text: "被动端口范围", ColumnSpan: 2}, Composite{
			Layout: HBox{MarginsZero: true},
			Children: []Widget{
				LineEdit{Text: Bind("LocalPortMin", Regexp{"^\\d+$"})},
				Label{Text: "-"},
				LineEdit{Text: Bind("LocalPortMax", Regexp{"^\\d+$"})},
			},
		})
	}
	return NewBasicDialog(&sp.Dialog, "添加 "+sp.title, sp.icon, DataBinder{
		AssignTo:   &sp.db,
		DataSource: sp.binder,
	}, sp.onSave, Composite{
		Layout:   Grid{Columns: 2, MarginsZero: true},
		MinSize:  Size{Width: 280},
		Children: widgets,
	}, VSpacer{}).Run(owner)
}

func (sp *SimpleProxyDialog) GetProxies() []*config.Proxy {
	return sp.Proxies
}

func (sp *SimpleProxyDialog) onSave() {
	if err := sp.db.Submit(); err != nil {
		return
	}
	for _, proto := range sp.types {
		proxy := config.Proxy{
			BaseProxyConf: config.BaseProxyConf{
				Name:      fmt.Sprintf("%s-%s-%s", sp.service, proto, sp.binder.RemotePort),
				Type:      proto,
				LocalIP:   sp.binder.LocalAddr,
				LocalPort: sp.binder.LocalPort,
			},
			RemotePort: sp.binder.RemotePort,
		}
		if sp.binder.LocalPortMin != "" && sp.binder.LocalPortMax != "" {
			proxy.Name = "range:" + proxy.Name
			portRange := fmt.Sprintf("%s-%s", sp.binder.LocalPortMin, sp.binder.LocalPortMax)
			proxy.LocalPort += "," + portRange
			proxy.RemotePort += "," + portRange
		}
		sp.Proxies = append(sp.Proxies, &proxy)
	}
	sp.Accept()
}
