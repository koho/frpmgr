package ui

import (
	"fmt"
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
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
	p, _ := strconv.Atoi(port)
	v.binder = &quickAddBinder{LocalAddr: ip, LocalPort: p}
	if v.binder.LocalAddr == "" {
		v.binder.LocalAddr = "127.0.0.1"
	}
	return v
}

func (sp *SimpleProxyDialog) Run(owner walk.Form) (int, error) {
	widgets := []Widget{
		Label{Text: i18n.SprintfColon("Remote Port"), ColumnSpan: 2},
		NumberEdit{Value: Bind("RemotePort"), MaxValue: 65535, ColumnSpan: 2},
		Label{Text: i18n.SprintfColon("Local Address")},
		Label{Text: i18n.SprintfColon("Port")},
		LineEdit{Text: Bind("LocalAddr", res.ValidateNonEmpty), StretchFactor: 2},
		NumberEdit{Value: Bind("LocalPort", Range{Min: 1, Max: 65535}), MaxValue: 65535, MinSize: Size{Width: 90}},
	}
	switch sp.service {
	case "ftp":
		var minPort, maxPort *walk.NumberEdit
		widgets = append(widgets, Label{Text: i18n.SprintfColon("Passive Port Range"), ColumnSpan: 2}, Composite{
			Layout: HBox{MarginsZero: true},
			Children: []Widget{
				NumberEdit{
					AssignTo:           &minPort,
					Value:              Bind("LocalPortMin", Range{Min: 1, Max: 65535}),
					MaxValue:           65535,
					MinSize:            Size{Width: 80},
					SpinButtonsVisible: true,
					OnValueChanged: func() {
						maxPort.SetRange(minPort.Value(), 65535)
					},
				},
				Label{Text: "-"},
				NumberEdit{
					AssignTo:           &maxPort,
					Value:              Bind("LocalPortMax"),
					MaxValue:           65535,
					MinSize:            Size{Width: 80},
					SpinButtonsVisible: true,
				},
				HSpacer{},
			},
		})
	}
	return NewBasicDialog(&sp.Dialog, fmt.Sprintf("%s %s", i18n.Sprintf("Add"), sp.title), sp.icon, DataBinder{
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
				Name:      fmt.Sprintf("%s_%s_%d", sp.service, proto, sp.binder.RemotePort),
				Type:      proto,
				LocalIP:   sp.binder.LocalAddr,
				LocalPort: strconv.Itoa(sp.binder.LocalPort),
			},
			RemotePort: strconv.Itoa(sp.binder.RemotePort),
		}
		if sp.binder.LocalPortMin > 0 && sp.binder.LocalPortMax > 0 {
			portRange := fmt.Sprintf("%d-%d", sp.binder.LocalPortMin, sp.binder.LocalPortMax)
			proxy.LocalPort += "," + portRange
			proxy.RemotePort += "," + portRange
		}
		sp.Proxies = append(sp.Proxies, &proxy)
	}
	sp.Accept()
}
