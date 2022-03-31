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

type SectionView struct {
	*walk.Composite

	model   *SectionModel
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
}

func NewSectionView() *SectionView {
	return new(SectionView)
}

func (sv *SectionView) View() Widget {
	return Composite{
		AssignTo: &sv.Composite,
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			sv.createToolbar(),
			sv.createSectionTable(),
		},
	}
}

func (sv *SectionView) OnCreate() {

}

func (sv *SectionView) Invalidate() {
	if conf := getCurrentConf(); conf == nil {
		sv.model = nil
		sv.table.SetModel(nil)
	} else {
		sv.model = NewSectionModel(conf)
		sv.table.SetModel(sv.model)
	}
}

func (sv *SectionView) createToolbar() ToolBar {
	return ToolBar{
		AssignTo:    &sv.toolbar,
		ButtonStyle: ToolBarButtonImageBeforeText,
		Orientation: Horizontal,
		Items: []MenuItem{
			Action{
				AssignTo: &sv.newAction,
				Text:     "添加",
				Image:    loadSysIcon("shell32", consts.IconCreate, 16),
				OnTriggered: func() {
					sv.onEdit(false)
				},
			},
			Menu{
				Text:  "快速添加",
				Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
				Items: []MenuItem{
					Action{
						AssignTo: &sv.rdAction,
						Text:     "远程桌面",
						Image:    loadSysIcon("imageres", consts.IconRemote, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewSimpleProxyDialog("远程桌面", loadSysIcon("imageres", consts.IconRemote, 32),
								"rdp", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":3389"))
						},
					},
					Action{
						AssignTo: &sv.vncAction,
						Text:     "VNC",
						Image:    loadSysIcon("imageres", consts.IconVNC, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewSimpleProxyDialog("VNC", loadSysIcon("imageres", consts.IconVNC, 32),
								"vnc", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":5900"))
						},
					},
					Action{
						AssignTo: &sv.sshAction,
						Text:     "SSH",
						Image:    loadResourceIcon(consts.IconSSH, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewSimpleProxyDialog("SSH", loadResourceIcon(consts.IconSSH, 32),
								"ssh", []string{consts.ProxyTypeTCP}, ":22"))
						},
					},
					Action{
						AssignTo: &sv.webAction,
						Text:     "Web",
						Image:    loadSysIcon("shell32", consts.IconWeb, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewSimpleProxyDialog("Web", loadSysIcon("shell32", consts.IconWeb, 32),
								"web", []string{consts.ProxyTypeTCP}, ":80"))
						},
					},
					Action{
						AssignTo: &sv.dnsAction,
						Text:     "DNS",
						Image:    loadSysIcon("imageres", consts.IconDns, 16),
						OnTriggered: func() {
							systemDns := util.GetSystemDnsServer()
							if systemDns == "" {
								systemDns = "114.114.114.114"
							}
							sv.onQuickAdd(NewSimpleProxyDialog("DNS", loadSysIcon("imageres", consts.IconDns, 32),
								"dns", []string{consts.ProxyTypeUDP}, systemDns+":53"))
						},
					},
					Action{
						AssignTo: &sv.ftpAction,
						Text:     "FTP",
						Image:    loadSysIcon("imageres", consts.IconFtp, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewSimpleProxyDialog("FTP", loadSysIcon("imageres", consts.IconFtp, 32),
								"ftp", []string{consts.ProxyTypeTCP}, ":21"))
						},
					},
					Action{
						AssignTo: &sv.httpFileAction,
						Text:     "HTTP 文件服务",
						Image:    loadSysIcon("imageres", consts.IconHttpFile, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewPluginProxyDialog("HTTP 文件服务", loadSysIcon("imageres", consts.IconHttpFile, 32),
								consts.PluginStaticFile))
						},
					},
					Action{
						AssignTo: &sv.httpProxyAction,
						Text:     "HTTP 代理",
						Image:    loadSysIcon("imageres", consts.IconHttpProxy, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewPluginProxyDialog("HTTP 代理", loadSysIcon("imageres", consts.IconHttpProxy, 32),
								consts.PluginHttpProxy))
						},
					},
					Action{
						AssignTo: &sv.socks5Action,
						Text:     "SOCKS5 代理",
						Image:    loadSysIcon("imageres", consts.IconSocks5, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewPluginProxyDialog("SOCKS5 代理", loadSysIcon("imageres", consts.IconSocks5, 32),
								consts.PluginSocks5))
						},
					},
					Action{
						AssignTo: &sv.vpnAction,
						Text:     "OpenVPN",
						Image:    loadSysIcon("shell32", consts.IconVpn, 16),
						OnTriggered: func() {
							sv.onQuickAdd(NewSimpleProxyDialog("OpenVPN", loadSysIcon("shell32", consts.IconVpn, 32),
								"openvpn", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":1194"))
						},
					},
				},
			},
			Action{
				AssignTo: &sv.editAction,
				Image:    loadSysIcon("shell32", consts.IconEdit, 16),
				Text:     "编辑",
				Enabled:  Bind("section.CurrentIndex >= 0"),
				OnTriggered: func() {
					sv.onEdit(true)
				},
			},
			Action{
				AssignTo:    &sv.deleteAction,
				Image:       loadSysIcon("shell32", consts.IconDelete, 16),
				Text:        "删除",
				Enabled:     Bind("section.CurrentIndex >= 0"),
				OnTriggered: sv.onDelete,
			},
			Action{
				AssignTo: &sv.openConfAction,
				Image:    loadResourceIcon(consts.IconOpen, 16),
				Text:     "打开配置文件",
				OnTriggered: func() {
					if sv.model == nil {
						return
					}
					if path, err := filepath.Abs(sv.model.conf.Path); err == nil {
						openPath(path)
					}
				},
			},
		},
	}
}

func (sv *SectionView) createSectionTable() TableView {
	return TableView{
		Name:     "section",
		AssignTo: &sv.table,
		Columns: []TableViewColumn{
			{Title: "名称", DataMember: "Name", Width: 105},
			{Title: "类型", DataMember: "Type", Width: 60},
			{Title: "本地地址", DataMember: "LocalIP", Width: 110},
			{Title: "本地端口", DataMember: "LocalPort", Width: 90},
			{Title: "远程端口", DataMember: "RemotePort", Width: 90},
			{Title: "子域名", DataMember: "SubDomain", Width: 90},
			{Title: "自定义域名", DataMember: "CustomDomains", Width: 90},
			{Title: "插件", DataMember: "Plugin", Width: 100},
		},
		ContextMenuItems: []MenuItem{
			ActionRef{&sv.editAction},
			ActionRef{&sv.newAction},
			Menu{
				Text:  "快速添加",
				Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
				Items: []MenuItem{
					ActionRef{&sv.rdAction},
					ActionRef{&sv.vncAction},
					ActionRef{&sv.sshAction},
					ActionRef{&sv.webAction},
					ActionRef{&sv.dnsAction},
					ActionRef{&sv.ftpAction},
					ActionRef{&sv.httpFileAction},
					ActionRef{&sv.httpProxyAction},
					ActionRef{&sv.socks5Action},
					ActionRef{&sv.vpnAction},
				},
			},
			Action{
				Enabled:     Bind("section.CurrentIndex >= 0"),
				Text:        "复制访问地址",
				Image:       loadSysIcon("shell32", consts.IconSysCopy, 16),
				OnTriggered: sv.onCopyAccessAddr,
			},
			ActionRef{&sv.openConfAction},
			ActionRef{&sv.deleteAction},
		},
		OnCurrentIndexChanged: func() {
			if db := sv.DataBinder(); db != nil {
				db.Reset()
			}
		},
		OnItemActivated: func() {
			sv.onEdit(true)
		},
	}
}

func (sv *SectionView) onCopyAccessAddr() {
	if sv.model == nil {
		return
	}
	index := sv.table.CurrentIndex()
	if index < 0 {
		return
	}
	conf, ok := sv.model.conf.Data.(*config.ClientConfig)
	if !ok {
		return
	}
	var access string
	proxy := conf.Proxies[index]
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
			access = proxy.SubDomain + "." + conf.ServerAddress
		} else if proxy.CustomDomains != "" {
			access = strings.Split(proxy.CustomDomains, ",")[0]
		}
	case consts.ProxyTypeTCPMUX:
		access = util.GetOrElse(proxy.LocalIP, "127.0.0.1") + ":" + proxy.LocalPort
	}
	walk.Clipboard().SetText(access)
}

func (sv *SectionView) onDelete() {
	if sv.model == nil {
		return
	}
	index := sv.table.CurrentIndex()
	if index < 0 {
		return
	}
	section, ok := sv.model.conf.Data.ItemAt(index).(config.Section)
	if !ok {
		return
	}
	if walk.MsgBox(sv.Form(), fmt.Sprintf("删除项目「%s」", section.GetName()),
		fmt.Sprintf("确定要删除项目「%s」吗?", section.GetName()),
		walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
		return
	}
	sv.model.conf.Data.DeleteItem(index)
	sv.commit()
}

func (sv *SectionView) onEdit(edit bool) {
	if sv.model == nil {
		return
	}
	var ret int
	if edit {
		index := sv.table.CurrentIndex()
		if index < 0 {
			return
		}
		if ret, _ = NewEditProxyDialog(sv.model.conf.Data.ItemAt(index).(*config.Proxy)).Run(sv.Form()); ret == walk.DlgCmdOK {
			sv.commit()
			sv.table.SetCurrentIndex(index)
		}
	} else {
		ep := NewEditProxyDialog(nil)
		if ret, _ = ep.Run(sv.Form()); ret == walk.DlgCmdOK {
			if sv.model.conf.Data.AddItem(ep.Proxy) {
				sv.commit()
				sv.scrollToBottom()
			}
		}
	}
}

func (sv *SectionView) onQuickAdd(qa QuickAdd) {
	if sv.model == nil {
		return
	}
	added := false
	if res, _ := qa.Run(sv.Form()); res == walk.DlgCmdOK {
		for _, proxy := range qa.GetProxies() {
			if !sv.model.conf.Data.AddItem(proxy) {
				showWarningMessage(sv.Form(), "代理已存在", fmt.Sprintf("代理名「%s」已存在。", proxy.Name))
			} else {
				added = true
			}
		}
		if added {
			sv.commit()
			sv.scrollToBottom()
		}
	}
}

// commit will update the views and save the config to disk, then reload service
func (sv *SectionView) commit() {
	sv.Invalidate()
	commitConf(sv.model.conf, false)
}

func (sv *SectionView) scrollToBottom() {
	if sv.model == nil {
		return
	}
	if tm := sv.table.TableModel(); tm != nil && tm.RowCount() > 0 {
		sv.table.EnsureItemVisible(tm.RowCount() - 1)
	}
}
