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

type PluginProxyDialog struct {
	*walk.Dialog

	Proxies []*config.Proxy
	binder  *quickAddBinder
	db      *walk.DataBinder

	// title of the dialog
	title string
	// icon of the dialog
	icon *walk.Icon
	// plugin of the proxy
	plugin string
}

// NewPluginProxyDialog creates proxy with given plugin
func NewPluginProxyDialog(title string, icon *walk.Icon, plugin string) *PluginProxyDialog {
	return &PluginProxyDialog{
		title:   title,
		icon:    icon,
		plugin:  plugin,
		Proxies: make([]*config.Proxy, 0),
		binder:  &quickAddBinder{},
	}
}

func (pp *PluginProxyDialog) Run(owner walk.Form) (int, error) {
	widgets := []Widget{
		Label{Text: i18n.SprintfColon("Remote Port")},
		NumberEdit{Value: Bind("RemotePort"), MaxValue: 65535, MinSize: Size{Width: 280}},
	}
	switch pp.plugin {
	case consts.PluginHttpProxy, consts.PluginSocks5:
		pp.binder.Plugin = consts.PluginHttpProxy
		widgets = append([]Widget{
			Label{Text: i18n.SprintfColon("Type")},
			NewRadioButtonGroup("Plugin", nil, nil, []RadioButton{
				{Text: "HTTP", Value: consts.PluginHttpProxy},
				{Text: "SOCKS5", Value: consts.PluginSocks5},
			}),
		}, widgets...)
	case consts.PluginStaticFile:
		widgets = append(widgets,
			Label{Text: i18n.SprintfColon("Local Directory")},
			NewBrowseLineEdit(nil, true, true, Bind("Dir", res.ValidateNonEmpty),
				i18n.Sprintf("Select a folder for directory listing."), "", false),
		)
	}
	return NewBasicDialog(&pp.Dialog, pp.title, pp.icon, DataBinder{
		AssignTo:   &pp.db,
		DataSource: pp.binder,
	}, pp.onSave, append(widgets, VSpacer{})...).Run(owner)
}

func (pp *PluginProxyDialog) GetProxies() []*config.Proxy {
	return pp.Proxies
}

func (pp *PluginProxyDialog) onSave() {
	if err := pp.db.Submit(); err != nil {
		return
	}
	if pp.binder.Plugin != "" {
		pp.plugin = pp.binder.Plugin
	}
	pp.Proxies = append(pp.Proxies, &config.Proxy{
		BaseProxyConf: config.BaseProxyConf{
			Name:   fmt.Sprintf("%s_%d", pp.plugin, pp.binder.RemotePort),
			Type:   "tcp",
			Plugin: pp.plugin,
			PluginParams: config.PluginParams{
				PluginLocalPath: pp.binder.Dir,
			},
		},
		RemotePort: strconv.Itoa(pp.binder.RemotePort),
	})
	pp.Accept()
}
