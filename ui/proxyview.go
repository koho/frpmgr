package ui

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	frpconfig "github.com/fatedier/frp/pkg/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
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
	mc := movingConditions()
	return ToolBar{
		AssignTo:    &pv.toolbar,
		ButtonStyle: ToolBarButtonImageBeforeText,
		Orientation: Horizontal,
		Items: []MenuItem{
			Action{
				AssignTo: &pv.newAction,
				Text:     i18n.Sprintf("Add"),
				Image:    loadIcon(res.IconCreate, 16),
				OnTriggered: func() {
					pv.onEdit(false, nil)
				},
			},
			Menu{
				Text:  i18n.Sprintf("Quick Add"),
				Image: loadIcon(res.IconQuickAdd, 16),
				Items: []MenuItem{
					Action{
						AssignTo: &pv.portAction,
						Text:     i18n.Sprintf("Open Port"),
						Image:    loadIcon(res.IconOpenPort, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPortProxyDialog())
						},
					},
					Action{
						AssignTo: &pv.rdAction,
						Text:     i18n.Sprintf("Remote Desktop"),
						Image:    loadIcon(res.IconRemote, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog(i18n.Sprintf("Remote Desktop"), loadIcon(res.IconRemote, 32),
								"rdp", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":3389"))
						},
					},
					Action{
						AssignTo: &pv.vncAction,
						Text:     "VNC",
						Image:    loadIcon(res.IconVNC, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("VNC", loadIcon(res.IconVNC, 32),
								"vnc", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":5900"))
						},
					},
					Action{
						AssignTo: &pv.sshAction,
						Text:     "SSH",
						Image:    loadIcon(res.IconSSH, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("SSH", loadIcon(res.IconSSH, 32),
								"ssh", []string{consts.ProxyTypeTCP}, ":22"))
						},
					},
					Action{
						AssignTo: &pv.webAction,
						Text:     "Web",
						Image:    loadIcon(res.IconWeb, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("Web", loadIcon(res.IconWeb, 32),
								"web", []string{consts.ProxyTypeTCP}, ":80"))
						},
					},
					Action{
						AssignTo: &pv.dnsAction,
						Text:     "DNS",
						Image:    loadIcon(res.IconDns, 16),
						OnTriggered: func() {
							systemDns := util.GetSystemDnsServer()
							if systemDns == "" {
								systemDns = "114.114.114.114"
							}
							pv.onQuickAdd(NewSimpleProxyDialog("DNS", loadIcon(res.IconDns, 32),
								"dns", []string{consts.ProxyTypeUDP}, systemDns+":53"))
						},
					},
					Action{
						AssignTo: &pv.vpnAction,
						Text:     "OpenVPN",
						Image:    loadIcon(res.IconLock, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("OpenVPN", loadIcon(res.IconLock, 32),
								"openvpn", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":1194"))
						},
					},
					Action{
						AssignTo: &pv.ftpAction,
						Text:     "FTP",
						Image:    loadIcon(res.IconFtp, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("FTP", loadIcon(res.IconFtp, 32),
								"ftp", []string{consts.ProxyTypeTCP}, ":21"))
						},
					},
					Action{
						AssignTo: &pv.httpFileAction,
						Text:     i18n.Sprintf("HTTP File Server"),
						Image:    loadIcon(res.IconHttpFile, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog(i18n.Sprintf("HTTP File Server"), loadIcon(res.IconHttpFile, 32),
								consts.PluginStaticFile))
						},
					},
					Action{
						AssignTo: &pv.httpProxyAction,
						Text:     i18n.Sprintf("HTTP Proxy"),
						Image:    loadIcon(res.IconHttpProxy, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog(i18n.Sprintf("HTTP Proxy"), loadIcon(res.IconHttpProxy, 32),
								consts.PluginHttpProxy))
						},
					},
				},
			},
			Action{
				AssignTo: &pv.editAction,
				Image:    loadIcon(res.IconEdit, 16),
				Text:     i18n.Sprintf("Edit"),
				Enabled:  Bind("proxy.CurrentIndex >= 0"),
				OnTriggered: func() {
					pv.onEdit(true, nil)
				},
			},
			Action{
				AssignTo:    &pv.toggleAction,
				Image:       loadIcon(res.IconDisable, 16),
				Text:        i18n.Sprintf("Disable"),
				Enabled:     Bind("proxy.CurrentIndex >= 0 && switchable(proxy.CurrentIndex)"),
				OnTriggered: pv.onToggleProxy,
			},
			Action{
				Image:   loadIcon(res.IconArrowUp, 16),
				Text:    i18n.Sprintf("Move Up"),
				Enabled: mc[0],
				OnTriggered: func() {
					pv.onMove(-1)
				},
			},
			Action{
				Image:   flipIcon(res.IconArrowUp, 16),
				Text:    i18n.Sprintf("Move Down"),
				Enabled: mc[1],
				OnTriggered: func() {
					pv.onMove(1)
				},
			},
			Action{
				AssignTo:    &pv.deleteAction,
				Image:       loadIcon(res.IconDelete, 16),
				Text:        i18n.Sprintf("Delete"),
				Enabled:     Bind("proxy.CurrentIndex >= 0"),
				OnTriggered: pv.onDelete,
			},
		},
	}
}

func (pv *ProxyView) createProxyTable() TableView {
	mc := movingConditions()
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
			Menu{
				Text:    i18n.Sprintf("Move"),
				Image:   loadIcon(res.IconMove, 16),
				Enabled: Bind("proxy.CurrentIndex >= 0 && proxy.ItemCount > 1"),
				Items: []MenuItem{
					Action{
						Text:    i18n.Sprintf("Up"),
						Enabled: mc[0],
						OnTriggered: func() {
							pv.onMove(-1)
						},
					},
					Action{
						Text:    i18n.Sprintf("Down"),
						Enabled: mc[1],
						OnTriggered: func() {
							pv.onMove(1)
						},
					},
					Action{
						Text:    i18n.Sprintf("To Top"),
						Enabled: mc[0],
						OnTriggered: func() {
							pv.onMove(-pv.table.CurrentIndex())
						},
					},
					Action{
						Text:    i18n.Sprintf("To Bottom"),
						Enabled: mc[1],
						OnTriggered: func() {
							if pv.model == nil {
								return
							}
							pv.onMove(len(pv.model.items) - pv.table.CurrentIndex() - 1)
						},
					},
				},
			},
			Separator{},
			ActionRef{Action: &pv.newAction},
			Menu{
				Text:  i18n.Sprintf("Quick Add"),
				Image: loadIcon(res.IconQuickAdd, 16),
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
				Image:       loadIcon(res.IconClipboard, 16),
				OnTriggered: pv.onClipboardImport,
			},
			Separator{},
			Action{
				Enabled:     Bind("proxy.CurrentIndex >= 0"),
				Text:        i18n.Sprintf("Copy Access Address"),
				Image:       loadIcon(res.IconSysCopy, 16),
				OnTriggered: pv.onCopyAccessAddr,
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
					style.TextColor = res.ColorGray
					style.BackgroundColor = res.ColorGrayBG
				} else if proxy.IsVisitor() {
					// Visitor proxy
					style.TextColor = res.ColorBlue
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
			if proxy.BindPort > 0 {
				access = util.GetOrElse(proxy.BindAddr, "127.0.0.1") + ":" + strconv.Itoa(proxy.BindPort)
			}
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
	var proxy *config.Proxy
	if strings.HasPrefix(text, "[[proxies]]") {
		var proxies struct {
			C []config.TypedProxyConfig `json:"proxies"`
		}
		if err = frpconfig.LoadConfigure([]byte(text), &proxies, false); err == nil && len(proxies.C) > 0 {
			proxy = config.ClientProxyFromV1(proxies.C[0])
		}
	} else if strings.HasPrefix(text, "[[visitors]]") {
		var visitors struct {
			C []config.TypedVisitorConfig `json:"visitors"`
		}
		if err = frpconfig.LoadConfigure([]byte(text), &visitors, false); err == nil && len(visitors.C) > 0 {
			proxy = config.ClientVisitorFromV1(visitors.C[0])
		}
	} else if strings.HasPrefix(text, "[") {
		proxy, err = config.UnmarshalProxyFromIni([]byte(text))
	} else {
		showErrorMessage(pv.Form(), "", i18n.Sprintf("This feature only supports text in INI or TOML format."))
		return
	}
	if err != nil {
		showError(err, pv.Form())
		return
	}
	if proxy == nil {
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
		walk.MsgBoxYesNo|walk.MsgBoxIconWarning) == walk.DlgCmdNo {
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
		ep := NewEditProxyDialog(pv.model.conf.Name, proxy, pv.visitors(proxy), true, pv.model.data.LegacyFormat)
		if ret, _ := ep.Run(pv.Form()); ret == walk.DlgCmdOK {
			if conf.CountStart() == 0 {
				ep.Proxy.Disabled = false
			}
			pv.commit()
			pv.table.SetCurrentIndex(idx)
		}
	} else {
		ep := NewEditProxyDialog(pv.model.conf.Name, fill, pv.visitors(nil), false, pv.model.data.LegacyFormat)
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
		if cc := getCurrentConf(); cc != nil && cc.State == consts.StateStarted {
			if walk.MsgBox(pv.Form(), i18n.Sprintf("Disable proxy \"%s\"", proxy.Name),
				i18n.Sprintf("Are you sure you would like to disable proxy \"%s\"?", proxy.Name),
				walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdNo {
				return
			}
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
	if r, _ := qa.Run(pv.Form()); r == walk.DlgCmdOK {
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

func (pv *ProxyView) onMove(delta int) {
	curIdx := pv.table.CurrentIndex()
	conf, _ := pv.getConfigProxy(curIdx)
	if conf == nil {
		return
	}
	targetIdx := curIdx + delta
	conf, _ = pv.getConfigProxy(targetIdx)
	if conf == nil {
		return
	}
	pv.model.Move(curIdx, targetIdx)
	commitConf(pv.model.conf, runFlagReload)
	pv.table.SetCurrentIndex(targetIdx)
}

// switchToggleAction updates the toggle action based on the current selected proxy
func (pv *ProxyView) switchToggleAction() {
	conf, proxy := pv.getConfigProxy(pv.table.CurrentIndex())
	if conf == nil {
		return
	}
	if proxy.Disabled {
		pv.toggleAction.SetText(i18n.Sprintf("Enable"))
		pv.toggleAction.SetImage(loadIcon(res.IconEnable, 16))
	} else {
		pv.toggleAction.SetText(i18n.Sprintf("Disable"))
		pv.toggleAction.SetImage(loadIcon(res.IconDisable, 16))
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

// Conditions for moving up/down proxy.
func movingConditions() [2]Property {
	return [2]Property{
		Bind("proxy.CurrentIndex > 0"),
		Bind("proxy.CurrentIndex >= 0 && proxy.CurrentIndex < proxy.ItemCount - 1"),
	}
}
