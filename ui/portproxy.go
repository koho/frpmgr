package ui

import (
	"fmt"
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
)

type portProxyBinder struct {
	quickAddBinder
	Name string
	TCP  bool
	UDP  bool
}

type PortProxyDialog struct {
	*walk.Dialog

	Proxies []*config.Proxy
	binder  *portProxyBinder
	db      *walk.DataBinder
}

func NewPortProxyDialog() *PortProxyDialog {
	dlg := new(PortProxyDialog)
	dlg.binder = &portProxyBinder{
		quickAddBinder: quickAddBinder{
			LocalAddr: "127.0.0.1",
		},
		TCP: true,
		UDP: true,
	}
	return dlg
}

func (pp *PortProxyDialog) Run(owner walk.Form) (int, error) {
	widgets := []Widget{
		Label{Text: i18n.SprintfColon("Name"), ColumnSpan: 2},
		LineEdit{Text: Bind("Name"), CueBanner: "open_xxx", ColumnSpan: 2},
		Label{Text: i18n.SprintfColon("Remote Port"), ColumnSpan: 2},
		NumberEdit{Value: Bind("RemotePort"), MaxValue: 65535, ColumnSpan: 2},
		Label{Text: i18n.SprintfColon("Protocol"), ColumnSpan: 2},
		Composite{
			Layout:     HBox{MarginsZero: true},
			ColumnSpan: 2,
			Children: []Widget{
				CheckBox{Text: "TCP", Checked: Bind("TCP")},
				CheckBox{Text: "UDP", Checked: Bind("UDP")},
			},
		},
		Label{Text: i18n.SprintfColon("Local Address")},
		Label{Text: i18n.SprintfColon("Port")},
		LineEdit{Text: Bind("LocalAddr", res.ValidateNonEmpty), StretchFactor: 2},
		NumberEdit{Value: Bind("LocalPort", Range{Min: 1, Max: 65535}), MaxValue: 65535, MinSize: Size{Width: 90}},
	}
	return NewBasicDialog(&pp.Dialog, i18n.Sprintf("Open Port"), loadIcon(res.IconOpenPort, 32), DataBinder{
		AssignTo:   &pp.db,
		DataSource: pp.binder,
	}, pp.onSave, Composite{
		Layout:   Grid{Columns: 2, MarginsZero: true},
		MinSize:  Size{Width: 280},
		Children: widgets,
	}, VSpacer{}).Run(owner)
}

func (pp *PortProxyDialog) GetProxies() []*config.Proxy {
	return pp.Proxies
}

func (pp *PortProxyDialog) onSave() {
	if err := pp.db.Submit(); err != nil {
		return
	}
	name := pp.binder.Name
	if name == "" {
		name = fmt.Sprintf("open_%d", pp.binder.RemotePort)
	}
	proxy := config.Proxy{
		BaseProxyConf: config.BaseProxyConf{
			Name:      name,
			LocalIP:   pp.binder.LocalAddr,
			LocalPort: strconv.Itoa(pp.binder.LocalPort),
		},
		RemotePort: strconv.Itoa(pp.binder.RemotePort),
	}
	if pp.binder.TCP {
		tcpProxy := proxy
		tcpProxy.Name += "_tcp"
		tcpProxy.Type = consts.ProxyTypeTCP
		pp.Proxies = append(pp.Proxies, &tcpProxy)
	}
	if pp.binder.UDP {
		udpProxy := proxy
		udpProxy.Name += "_udp"
		udpProxy.Type = consts.ProxyTypeUDP
		pp.Proxies = append(pp.Proxies, &udpProxy)
	}
	pp.Accept()
}
