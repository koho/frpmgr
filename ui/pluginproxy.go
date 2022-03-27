package ui

import (
	"fmt"
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
		Label{Text: "远程端口:"},
		LineEdit{Text: Bind("RemotePort", Regexp{"^\\d+$"})},
	}
	switch pp.plugin {
	case consts.PluginStaticFile:
		// Make the dialog wider
		remoteView := widgets[1].(LineEdit)
		remoteView.MinSize = Size{Width: 250}
		widgets[1] = remoteView
		widgets = append(widgets,
			Label{Text: "本地目录:"},
			NewBrowseLineEdit(nil, true, true, Bind("Dir", Regexp{".+"}),
				"选择本地文件夹", "", false),
		)
	}
	return NewBasicDialog(&pp.Dialog, "添加 "+pp.title, pp.icon, DataBinder{
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
			Name:   fmt.Sprintf("%s-%s", pp.plugin, pp.binder.RemotePort),
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
