package ui

import (
	"fmt"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"net"
	"path/filepath"
	"strings"
)

type ProxyView struct {
	*walk.Composite

	model   *ProxyModel
	toolbar *walk.ToolBar
	table   *walk.TableView

	// Actions
	newAction       *walk.Action
	rdAction        *walk.Action
	sshAction       *walk.Action
	webAction       *walk.Action
	vncAction       *walk.Action
	dnsAction       *walk.Action
	ftpAction       *walk.Action
	httpFileAction  *walk.Action
	httpProxyAction *walk.Action
	socks5Action    *walk.Action
	vpnAction       *walk.Action
	editAction      *walk.Action
	deleteAction    *walk.Action
	openConfAction  *walk.Action
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
			pv.createToolbar(),
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
		MaxSize:     Size{Width: 300},
		ButtonStyle: ToolBarButtonImageBeforeText,
		Orientation: Horizontal,
		Items: []MenuItem{
			Action{
				AssignTo: &pv.newAction,
				Text:     "添加",
				Image:    loadSysIcon("shell32", consts.IconCreate, 16),
				OnTriggered: func() {
					pv.onEdit(false, nil)
				},
			},
			Menu{
				Text:  "快速添加",
				Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
				Items: []MenuItem{
					Action{
						AssignTo: &pv.rdAction,
						Text:     "远程桌面",
						Image:    loadSysIcon("imageres", consts.IconRemote, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("远程桌面", loadSysIcon("imageres", consts.IconRemote, 32),
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
						Text:     "HTTP 文件服务",
						Image:    loadSysIcon("imageres", consts.IconHttpFile, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog("HTTP 文件服务", loadSysIcon("imageres", consts.IconHttpFile, 32),
								consts.PluginStaticFile))
						},
					},
					Action{
						AssignTo: &pv.httpProxyAction,
						Text:     "HTTP 代理",
						Image:    loadSysIcon("imageres", consts.IconHttpProxy, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog("HTTP 代理", loadSysIcon("imageres", consts.IconHttpProxy, 32),
								consts.PluginHttpProxy))
						},
					},
					Action{
						AssignTo: &pv.socks5Action,
						Text:     "SOCKS5 代理",
						Image:    loadSysIcon("imageres", consts.IconSocks5, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog("SOCKS5 代理", loadSysIcon("imageres", consts.IconSocks5, 32),
								consts.PluginSocks5))
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
				},
			},
			Action{
				AssignTo: &pv.editAction,
				Image:    loadSysIcon("shell32", consts.IconEdit, 16),
				Text:     "编辑",
				Enabled:  Bind("proxy.CurrentIndex >= 0"),
				OnTriggered: func() {
					pv.onEdit(true, nil)
				},
			},
			Action{
				AssignTo:    &pv.toggleAction,
				Image:       loadResourceIcon(consts.IconEnable, 16),
				Text:        "启用",
				Enabled:     Bind("proxy.CurrentIndex >= 0 && switchable(proxy.CurrentIndex)"),
				OnTriggered: pv.onToggleProxy,
			},
			Action{
				AssignTo:    &pv.deleteAction,
				Image:       loadSysIcon("shell32", consts.IconDelete, 16),
				Text:        "删除",
				Enabled:     Bind("proxy.CurrentIndex >= 0"),
				OnTriggered: pv.onDelete,
			},
			Action{
				AssignTo: &pv.openConfAction,
				Image:    loadResourceIcon(consts.IconOpen, 16),
				Text:     "打开配置文件",
				OnTriggered: func() {
					if pv.model == nil {
						return
					}
					if path, err := filepath.Abs(pv.model.conf.Path); err == nil {
						openPath(path)
					}
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
			{Title: "名称", DataMember: "Name", Width: 105},
			{Title: "类型", DataMember: "Type", Width: 60},
			{Title: "本地地址", DataMember: "LocalIP", Width: 110},
			{Title: "本地端口", DataMember: "LocalPort", Width: 90},
			{Title: "远程端口", DataMember: "RemotePort", Width: 90},
			{Title: "子域名", DataMember: "SubDomain", Width: 70},
			{Title: "自定义域名", DataMember: "CustomDomains", Width: 85},
			{Title: "插件", DataMember: "Plugin", Width: 95},
		},
		ContextMenuItems: []MenuItem{
			ActionRef{&pv.editAction},
			ActionRef{&pv.toggleAction},
			Separator{},
			ActionRef{&pv.newAction},
			Menu{
				Text:  "快速添加",
				Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
				Items: []MenuItem{
					ActionRef{&pv.rdAction},
					ActionRef{&pv.vncAction},
					ActionRef{&pv.sshAction},
					ActionRef{&pv.webAction},
					ActionRef{&pv.dnsAction},
					ActionRef{&pv.ftpAction},
					ActionRef{&pv.httpFileAction},
					ActionRef{&pv.httpProxyAction},
					ActionRef{&pv.socks5Action},
					ActionRef{&pv.vpnAction},
				},
			},
			Action{
				Text:        "从剪贴板导入",
				Image:       loadSysIcon("shell32", consts.IconClipboard, 16),
				OnTriggered: pv.onClipboardImport,
			},
			Separator{},
			Action{
				Enabled:     Bind("proxy.CurrentIndex >= 0"),
				Text:        "复制访问地址",
				Image:       loadSysIcon("shell32", consts.IconSysCopy, 16),
				OnTriggered: pv.onCopyAccessAddr,
			},
			ActionRef{&pv.openConfAction},
			Separator{},
			ActionRef{&pv.deleteAction},
		},
		OnItemActivated: func() {
			pv.onEdit(true, nil)
		},
		StyleCell: func(style *walk.CellStyle) {
			if _, proxy := pv.getConfigProxy(style.Row()); proxy != nil && proxy.Disabled {
				style.TextColor = consts.ColorGray
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
	if walk.MsgBox(pv.Form(), fmt.Sprintf("删除代理「%s」", proxy.Name),
		fmt.Sprintf("确定要删除代理「%s」吗?", proxy.Name),
		walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
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
		ep := NewEditProxyDialog(proxy, true)
		if ret, _ := ep.Run(pv.Form()); ret == walk.DlgCmdOK {
			if conf.CountStart() == 0 {
				ep.Proxy.Disabled = false
			}
			pv.commit()
			pv.table.SetCurrentIndex(idx)
		}
	} else {
		ep := NewEditProxyDialog(fill, false)
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
		if walk.MsgBox(pv.Form(), fmt.Sprintf("禁用代理「%s」", proxy.Name),
			fmt.Sprintf("确定要禁用代理「%s」吗?", proxy.Name),
			walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
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
	if res, _ := qa.Run(pv.Form()); res == walk.DlgCmdOK {
		for _, proxy := range qa.GetProxies() {
			if !pv.model.data.AddItem(proxy) {
				showWarningMessage(pv.Form(), "代理已存在", fmt.Sprintf("代理名「%s」已存在。", proxy.Name))
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

// switchToggleAction updates the toggle action based on the current selected proxy
func (pv *ProxyView) switchToggleAction() {
	conf, proxy := pv.getConfigProxy(pv.table.CurrentIndex())
	if conf == nil {
		return
	}
	if proxy.Disabled {
		pv.toggleAction.SetText("启用")
		pv.toggleAction.SetImage(loadResourceIcon(consts.IconEnable, 16))
	} else {
		pv.toggleAction.SetText("禁用")
		pv.toggleAction.SetImage(loadResourceIcon(consts.IconDisable, 16))
	}
}

// commit will update the views and save the config to disk, then reload service
func (pv *ProxyView) commit() {
	pv.Invalidate()
	if pv.model != nil {
		commitConf(pv.model.conf, false)
	}
}

func (pv *ProxyView) scrollToBottom() {
	if tm := pv.table.TableModel(); tm != nil && tm.RowCount() > 0 {
		pv.table.EnsureItemVisible(tm.RowCount() - 1)
	}
}
