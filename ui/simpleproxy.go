package ui

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/pkg/validators"
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
		Label{Text: i18n.SprintfColon("Remote Port"), ColumnSpan: 2},
		LineEdit{Text: Bind("RemotePort", consts.ValidatePortRange...), ColumnSpan: 2},
		Label{Text: i18n.SprintfColon("Local Address")},
		Label{Text: i18n.SprintfColon("Port")},
		LineEdit{Text: Bind("LocalAddr", consts.ValidateNonEmpty), StretchFactor: 2},
		LineEdit{Text: Bind("LocalPort", consts.ValidatePortRange...), StretchFactor: 1},
	}
	switch sp.service {
	case "ftp":
		var lPortMinEdit *walk.LineEdit
		widgets = append(widgets, Label{Text: i18n.SprintfColon("Passive Port Range"), ColumnSpan: 2}, Composite{
			Layout: HBox{MarginsZero: true},
			Children: []Widget{
				LineEdit{AssignTo: &lPortMinEdit, Text: Bind("LocalPortMin", consts.ValidatePortRange...)},
				Label{Text: "-"},
				LineEdit{Text: Bind("LocalPortMax", append(consts.ValidatePortRange, validators.GTE{Value: &lPortMinEdit})...)},
			},
		})
	}
	return NewBasicDialog(&sp.Dialog, fmt.Sprintf("%s %s", i18n.Sprintf("Add"), sp.title), sp.icon, DataBinder{
		AssignTo:       &sp.db,
		DataSource:     sp.binder,
		ErrorPresenter: validators.SilentToolTipErrorPresenter{},
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
				Name:      fmt.Sprintf("%s_%s_%s", sp.service, proto, sp.binder.RemotePort),
				Type:      proto,
				LocalIP:   sp.binder.LocalAddr,
				LocalPort: sp.binder.LocalPort,
			},
			RemotePort: sp.binder.RemotePort,
		}
		if sp.binder.LocalPortMin != "" && sp.binder.LocalPortMax != "" {
			proxy.Name = consts.RangePrefix + proxy.Name
			portRange := fmt.Sprintf("%s-%s", sp.binder.LocalPortMin, sp.binder.LocalPortMax)
			proxy.LocalPort += "," + portRange
			proxy.RemotePort += "," + portRange
		}
		sp.Proxies = append(sp.Proxies, &proxy)
	}
	sp.Accept()
}
