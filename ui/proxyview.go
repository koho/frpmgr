package ui

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
)

type ProxyView struct {
	*walk.Composite

	model   *ProxyModel
	toolbar *walk.ToolBar
	table   *walk.TableView

	// Actions
	newAction       *walk.Action
	portAction      *walk.Action
	rdAction        *walk.Action
	sshAction       *walk.Action
	webAction       *walk.Action
	vncAction       *walk.Action
	dnsAction       *walk.Action
	ftpAction       *walk.Action
	httpFileAction  *walk.Action
	httpProxyAction *walk.Action
	vpnAction       *walk.Action
	editAction      *walk.Action
	deleteAction    *walk.Action
	openConfAction  *walk.Action
	showConfAction  *walk.Action
	toggleAction    *walk.Action
}

func NewProxyView() *ProxyView {
	return new(ProxyView)
}

func (pv *ProxyView) View() Widget {
	return Composite{
		AssignTo: &pv.Composite,
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout:          HBox{MarginsZero: true, SpacingZero: true},
				Alignment:       AlignHNearVNear,
				DoubleBuffering: true,
				Children: []Widget{
					pv.createToolbar(),
				},
			},
			pv.createProxyTable(),
		},
		Functions: map[string]func(args ...interface{}) (interface{}, error){
			// switchable make sure that at least one proxy is enabled
			"switchable": func(args ...interface{}) (interface{}, error) {
				if conf, proxy := pv.getConfigProxy(int(args[0].(float64))); conf != nil {
					// We can't disable all proxies
					return proxy.Disabled || conf.CountStart() > 1, nil
				}
				return false, nil
			},
		},
	}
}

func (pv *ProxyView) OnCreate() {
	pv.editAction.SetDefault(true)
	pv.table.CurrentIndexChanged().Attach(pv.switchToggleAction)
}

func (pv *ProxyView) Invalidate() {
	if conf := getCurrentConf(); conf != nil {
		if _, ok := conf.Data.(*config.ClientConfig); ok {
			pv.model = NewProxyModel(conf)
			pv.table.SetModel(pv.model)
			return
		}
	}
	pv.model = nil
	pv.table.SetModel(nil)
}

func (pv *ProxyView) createToolbar() ToolBar {
	return ToolBar{
		AssignTo:    &pv.toolbar,
		ButtonStyle: ToolBarButtonImageBeforeText,
		Orientation: Horizontal,
		Items: []MenuItem{
			Action{
				AssignTo: &pv.newAction,
				Text:     i18n.Sprintf("Add"),
				Image:    loadSysIcon("shell32", consts.IconCreate, 16),
				OnTriggered: func() {
					pv.onEdit(false, nil)
				},
			},
			Menu{
				Text:  i18n.Sprintf("Quick Add"),
				Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
				Items: []MenuItem{
					Action{
						AssignTo: &pv.portAction,
						Text:     i18n.Sprintf("Open Port"),
						Image:    loadSysIcon("shell32", consts.IconOpenPort, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPortProxyDialog())
						},
					},
					Action{
						AssignTo: &pv.rdAction,
						Text:     i18n.Sprintf("Remote Desktop"),
						Image:    loadSysIcon("imageres", consts.IconRemote, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog(i18n.Sprintf("Remote Desktop"), loadSysIcon("imageres", consts.IconRemote, 32),
								"rdp", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":3389"))
						},
					},
					Action{
						AssignTo: &pv.vncAction,
						Text:     "VNC",
						Image:    loadSysIcon("imageres", consts.IconVNC, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("VNC", loadSysIcon("imageres", consts.IconVNC, 32),
								"vnc", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":5900"))
						},
					},
					Action{
						AssignTo: &pv.sshAction,
						Text:     "SSH",
						Image:    loadResourceIcon(consts.IconSSH, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("SSH", loadResourceIcon(consts.IconSSH, 32),
								"ssh", []string{consts.ProxyTypeTCP}, ":22"))
						},
					},
					Action{
						AssignTo: &pv.webAction,
						Text:     "Web",
						Image:    loadSysIcon("shell32", consts.IconWeb, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("Web", loadSysIcon("shell32", consts.IconWeb, 32),
								"web", []string{consts.ProxyTypeTCP}, ":80"))
						},
					},
					Action{
						AssignTo: &pv.dnsAction,
						Text:     "DNS",
						Image:    loadSysIcon("imageres", consts.IconDns, 16),
						OnTriggered: func() {
							systemDns := util.GetSystemDnsServer()
							if systemDns == "" {
								systemDns = "114.114.114.114"
							}
							pv.onQuickAdd(NewSimpleProxyDialog("DNS", loadSysIcon("imageres", consts.IconDns, 32),
								"dns", []string{consts.ProxyTypeUDP}, systemDns+":53"))
						},
					},
					Action{
						AssignTo: &pv.vpnAction,
						Text:     "OpenVPN",
						Image:    loadSysIcon("shell32", consts.IconVpn, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("OpenVPN", loadSysIcon("shell32", consts.IconVpn, 32),
								"openvpn", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":1194"))
						},
					},
					Action{
						AssignTo: &pv.ftpAction,
						Text:     "FTP",
						Image:    loadSysIcon("imageres", consts.IconFtp, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("FTP", loadSysIcon("imageres", consts.IconFtp, 32),
								"ftp", []string{consts.ProxyTypeTCP}, ":21"))
						},
					},
					Action{
						AssignTo: &pv.httpFileAction,
						Text:     i18n.Sprintf("HTTP File Server"),
						Image:    loadSysIcon("imageres", consts.IconHttpFile, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog(i18n.Sprintf("HTTP File Server"), loadSysIcon("imageres", consts.IconHttpFile, 32),
								consts.PluginStaticFile))
						},
					},
					Action{
						AssignTo: &pv.httpProxyAction,
						Text:     i18n.Sprintf("HTTP Proxy"),
						Image:    loadSysIcon("imageres", consts.IconHttpProxy, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog(i18n.Sprintf("HTTP Proxy"), loadSysIcon("imageres", consts.IconHttpProxy, 32),
								consts.PluginHttpProxy))
						},
					},
				},
			},
			Action{
				AssignTo: &pv.editAction,
				Image:    loadSysIcon("shell32", consts.IconEdit, 16),
				Text:     i18n.Sprintf("Edit"),
				Enabled:  Bind("proxy.CurrentIndex >= 0"),
				OnTriggered: func() {
					pv.onEdit(true, nil)
				},
			},
			Action{
				AssignTo:    &pv.toggleAction,
				Image:       loadResourceIcon(consts.IconEnable, 16),
				Text:        i18n.Sprintf("Enable"),
				Enabled:     Bind("proxy.CurrentIndex >= 0 && switchable(proxy.CurrentIndex)"),
				OnTriggered: pv.onToggleProxy,
			},
			Action{
				AssignTo:    &pv.deleteAction,
				Image:       loadSysIcon("shell32", consts.IconDelete, 16),
				Text:        i18n.Sprintf("Delete"),
				Enabled:     Bind("proxy.CurrentIndex >= 0"),
				OnTriggered: pv.onDelete,
			},
			Menu{
				Text:  i18n.Sprintf("Open Config"),
				Image: loadResourceIcon(consts.IconOpen, 16),
				Items: []MenuItem{
					Action{
						AssignTo: &pv.openConfAction,
						Text:     i18n.Sprintf("Direct Edit"),
						OnTriggered: func() {
							pv.onOpenConfig(false)
						},
					},
					Action{
						AssignTo: &pv.showConfAction,
						Text:     i18n.Sprintf("Show in Folder"),
						OnTriggered: func() {
							pv.onOpenConfig(true)
						},
					},
				},
			},
		},
	}
}

func (pv *ProxyView) createProxyTable() TableView {
	return TableView{
		Name:     "proxy",
		AssignTo: &pv.table,
		Columns: []TableViewColumn{
			{Title: i18n.Sprintf("Name"), DataMember: "Name", Width: 100},
			{Title: i18n.Sprintf("Type"), DataMember: "Type", Width: 55},
			{Title: i18n.Sprintf("Local Address"), DataMember: "DisplayLocalIP", Width: 110},
			{Title: i18n.Sprintf("Local Port"), DataMember: "DisplayLocalPort", Width: 90},
			{Title: i18n.Sprintf("Remote Port"), DataMember: "RemotePort", Width: 90},
			{Title: i18n.Sprintf("Domains"), DataMember: "Domains", Width: 80},
			{Title: i18n.Sprintf("Plugin"), DataMember: "Plugin", Width: 80},
		},
		ContextMenuItems: []MenuItem{
			ActionRef{Action: &pv.editAction},
			ActionRef{Action: &pv.toggleAction},
			Separator{},
			ActionRef{Action: &pv.newAction},
			Menu{
				Text:  i18n.Sprintf("Quick Add"),
				Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
				Items: []MenuItem{
					ActionRef{Action: &pv.portAction},
					ActionRef{Action: &pv.rdAction},
					ActionRef{Action: &pv.vncAction},
					ActionRef{Action: &pv.sshAction},
					ActionRef{Action: &pv.webAction},
					ActionRef{Action: &pv.dnsAction},
					ActionRef{Action: &pv.vpnAction},
					ActionRef{Action: &pv.ftpAction},
					ActionRef{Action: &pv.httpFileAction},
					ActionRef{Action: &pv.httpProxyAction},
				},
			},
			Action{
				Text:        i18n.Sprintf("Import from Clipboard"),
				Image:       loadSysIcon("shell32", consts.IconClipboard, 16),
				OnTriggered: pv.onClipboardImport,
			},
			Separator{},
			Action{
				Enabled:     Bind("proxy.CurrentIndex >= 0"),
				Text:        i18n.Sprintf("Copy Access Address"),
				Image:       loadSysIcon("shell32", consts.IconSysCopy, 16),
				OnTriggered: pv.onCopyAccessAddr,
			},
			Menu{
				Text:  i18n.Sprintf("Open Config"),
				Image: loadResourceIcon(consts.IconOpen, 16),
				Items: []MenuItem{
					ActionRef{Action: &pv.openConfAction},
					ActionRef{Action: &pv.showConfAction},
				},
			},
			Separator{},
			ActionRef{Action: &pv.deleteAction},
		},
		OnItemActivated: func() {
			pv.onEdit(true, nil)
		},
		StyleCell: func(style *walk.CellStyle) {
			if _, proxy := pv.getConfigProxy(style.Row()); proxy != nil {
				if proxy.Disabled {
					// Disabled proxy
					style.TextColor = consts.ColorGray
					style.BackgroundColor = consts.ColorGrayBG
				} else if proxy.IsVisitor() {
					// Visitor proxy
					style.TextColor = consts.ColorBlue
				}
				// Normal proxy is default black text
			}
		},
	}
}

// getConfigProxy returns the config object and the proxy object of the given index
func (pv *ProxyView) getConfigProxy(idx int) (*config.ClientConfig, *config.Proxy) {
	if pv.model == nil {
		return nil, nil
	}
	if idx < 0 || idx >= len(pv.model.data.Proxies) {
		return nil, nil
	}
	return pv.model.data, pv.model.data.Proxies[idx]
}

func (pv *ProxyView) onCopyAccessAddr() {
	conf, proxy := pv.getConfigProxy(pv.table.CurrentIndex())
	if conf == nil {
		return
	}
	var access string
	switch proxy.Type {
	case consts.ProxyTypeTCP, consts.ProxyTypeUDP:
		if proxy.RemotePort != "" {
			access = conf.ServerAddress + ":" + strings.Split(strings.Split(proxy.RemotePort, ",")[0], "-")[0]
		}
	case consts.ProxyTypeXTCP, consts.ProxyTypeSTCP, consts.ProxyTypeSUDP:
		if proxy.Role == "visitor" {
			access = util.GetOrElse(proxy.BindAddr, "127.0.0.1") + ":" + proxy.BindPort
		} else {
			access = util.GetOrElse(proxy.LocalIP, "127.0.0.1") + ":" + proxy.LocalPort
		}
	case consts.ProxyTypeHTTP, consts.ProxyTypeHTTPS:
		if proxy.SubDomain != "" && net.ParseIP(conf.ServerAddress) == nil {
			// Assume subdomain_host is equal to server_address
			access = fmt.Sprintf("%s://%s.%s", proxy.Type, proxy.SubDomain, conf.ServerAddress)
		} else if proxy.CustomDomains != "" {
			access = fmt.Sprintf("%s://%s", proxy.Type, strings.Split(proxy.CustomDomains, ",")[0])
		}
	case consts.ProxyTypeTCPMUX:
		access = util.GetOrElse(proxy.LocalIP, "127.0.0.1") + ":" + proxy.LocalPort
	}
	walk.Clipboard().SetText(access)
}

func (pv *ProxyView) onClipboardImport() {
	text, err := walk.Clipboard().Text()
	if err != nil || strings.TrimSpace(text) == "" {
		return
	}
	proxy, err := config.UnmarshalProxyFromIni([]byte(text))
	if err != nil {
		showError(err, pv.Form())
		return
	}
	pv.onEdit(false, proxy)
}

func (pv *ProxyView) onDelete() {
	idx := pv.table.CurrentIndex()
	conf, proxy := pv.getConfigProxy(idx)
	if conf == nil {
		return
	}
	oldConf := pv.model.conf.Name
	if walk.MsgBox(pv.Form(), i18n.Sprintf("Delete proxy \"%s\"", proxy.Name),
		i18n.Sprintf("Are you sure you would like to delete proxy \"%s\"?", proxy.Name),
		walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
		return
	}
	if !hasConf(oldConf) {
		warnConfigRemoved(pv.Form(), oldConf)
		return
	}
	conf.DeleteItem(idx)
	pv.commit()
}

func (pv *ProxyView) onEdit(current bool, fill *config.Proxy) {
	if pv.model == nil {
		return
	}
	if current {
		idx := pv.table.CurrentIndex()
		conf, proxy := pv.getConfigProxy(idx)
		if conf == nil {
			return
		}
		ep := NewEditProxyDialog(pv.model.conf.Name, proxy, pv.visitors(proxy), true)
		if ret, _ := ep.Run(pv.Form()); ret == walk.DlgCmdOK {
			if conf.CountStart() == 0 {
				ep.Proxy.Disabled = false
			}
			pv.commit()
			pv.table.SetCurrentIndex(idx)
		}
	} else {
		ep := NewEditProxyDialog(pv.model.conf.Name, fill, pv.visitors(nil), false)
		if ret, _ := ep.Run(pv.Form()); ret == walk.DlgCmdOK {
			if pv.model.data.AddItem(ep.Proxy) {
				if pv.model.data.CountStart() == 0 {
					ep.Proxy.Disabled = false
				}
				pv.commit()
				pv.scrollToBottom()
			}
		}
	}
}

func (pv *ProxyView) onToggleProxy() {
	conf, proxy := pv.getConfigProxy(pv.table.CurrentIndex())
	if conf == nil {
		return
	}
	if !proxy.Disabled {
		// We can't disable all proxies
		if conf.CountStart() <= 1 {
			return
		}
		oldConf := pv.model.conf.Name
		if walk.MsgBox(pv.Form(), i18n.Sprintf("Disable proxy \"%s\"", proxy.Name),
			i18n.Sprintf("Are you sure you would like to disable proxy \"%s\"?", proxy.Name),
			walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
			return
		}
		if !hasConf(oldConf) {
			warnConfigRemoved(pv.Form(), oldConf)
			return
		}
	}
	proxy.Disabled = !proxy.Disabled
	pv.commit()
}

func (pv *ProxyView) onQuickAdd(qa QuickAdd) {
	if pv.model == nil {
		return
	}
	added := false
	oldConf := pv.model.conf.Name
	if res, _ := qa.Run(pv.Form()); res == walk.DlgCmdOK {
		if pv.model == nil || pv.model.conf.Name != oldConf {
			warnConfigRemoved(pv.Form(), oldConf)
			return
		}
		for _, proxy := range qa.GetProxies() {
			if !pv.model.data.AddItem(proxy) {
				showWarningMessage(pv.Form(), i18n.Sprintf("Proxy already exists"), i18n.Sprintf("The proxy name \"%s\" already exists.", proxy.Name))
			} else {
				added = true
			}
		}
		if added {
			pv.commit()
			pv.scrollToBottom()
		}
	}
}

func (pv *ProxyView) onOpenConfig(folder bool) {
	if pv.model == nil {
		return
	}
	if path, err := filepath.Abs(pv.model.conf.Path); err == nil {
		if folder {
			openFolder(path)
		} else {
			openPath(path)
		}
	}
}

// switchToggleAction updates the toggle action based on the current selected proxy
func (pv *ProxyView) switchToggleAction() {
	conf, proxy := pv.getConfigProxy(pv.table.CurrentIndex())
	if conf == nil {
		return
	}
	if proxy.Disabled {
		pv.toggleAction.SetText(i18n.Sprintf("Enable"))
		pv.toggleAction.SetImage(loadResourceIcon(consts.IconEnable, 16))
	} else {
		pv.toggleAction.SetText(i18n.Sprintf("Disable"))
		pv.toggleAction.SetImage(loadResourceIcon(consts.IconDisable, 16))
	}
}

// commit will update the views and save the config to disk, then reload service
func (pv *ProxyView) commit() {
	pv.Invalidate()
	if pv.model != nil {
		commitConf(pv.model.conf, runFlagReload)
	}
}

func (pv *ProxyView) scrollToBottom() {
	if tm := pv.table.TableModel(); tm != nil && tm.RowCount() > 0 {
		pv.table.EnsureItemVisible(tm.RowCount() - 1)
	}
}

// visitors returns a list of visitor names except the given proxy.
func (pv *ProxyView) visitors(except *config.Proxy) (visitors []string) {
	for _, proxy := range pv.model.data.Proxies {
		if proxy != except && proxy.IsVisitor() {
			visitors = append(visitors, proxy.Name)
		}
	}
	return
}
