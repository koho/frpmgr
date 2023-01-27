package ui

import (
	"fmt"
	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
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
	v := &PluginProxyDialog{title: title, icon: icon, plugin: plugin}
	v.Proxies = make([]*config.Proxy, 0)
	v.binder = &quickAddBinder{}
	return v
}

func (pp *PluginProxyDialog) Run(owner walk.Form) (int, error) {
	widgets := []Widget{
		Label{Text: i18n.SprintfColon("Remote Port")},
		LineEdit{Text: Bind("RemotePort", consts.ValidateRequireInteger)},
	}
	switch pp.plugin {
	case consts.PluginStaticFile:
		// Make the dialog wider
		remoteView := widgets[1].(LineEdit)
		remoteView.MinSize = Size{Width: 300}
		widgets[1] = remoteView
		widgets = append(widgets,
			Label{Text: i18n.SprintfColon("Local Directory")},
			NewBrowseLineEdit(nil, true, true, Bind("Dir", consts.ValidateNonEmpty),
				i18n.Sprintf("Select a folder for directory listing."), "", false),
		)
	}
	return NewBasicDialog(&pp.Dialog, fmt.Sprintf("%s %s", i18n.Sprintf("Add"), pp.title), pp.icon, DataBinder{
		AssignTo:   &pp.db,
		DataSource: pp.binder,
	}, pp.onSave, Composite{
		Layout:   Grid{Columns: 2, MarginsZero: true},
		Children: widgets,
	}, VSpacer{}).Run(owner)
}

func (pp *PluginProxyDialog) GetProxies() []*config.Proxy {
	return pp.Proxies
}

func (pp *PluginProxyDialog) onSave() {
	if err := pp.db.Submit(); err != nil {
		return
	}
	pp.Proxies = append(pp.Proxies, &config.Proxy{
		BaseProxyConf: config.BaseProxyConf{
			Name:   fmt.Sprintf("%s_%s", pp.plugin, pp.binder.RemotePort),
			Type:   "tcp",
			Plugin: pp.plugin,
			PluginParams: config.PluginParams{
				PluginLocalPath: pp.binder.Dir,
			},
		},
		RemotePort: pp.binder.RemotePort,
	})
	pp.Accept()
}
