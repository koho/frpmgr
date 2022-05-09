package ui

import (
	"fmt"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/services"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/thoas/go-funk"
	"strings"
)

type EditProxyDialog struct {
	*walk.Dialog

	Proxy *config.Proxy
	// Whether we are editing an existing proxy
	exist bool

	// View models
	binder    *editProxyBinder
	dbs       [2]*walk.DataBinder
	vmDB      *walk.DataBinder
	viewModel proxyViewModel

	// Views
	nameView       *walk.LineEdit
	localPortView  *walk.LineEdit
	remotePortView *walk.LineEdit
	customText     *walk.TextEdit
	typeView       *walk.ComboBox
	roleView       *walk.CheckBox
	pluginView     *walk.ComboBox
}

// View model for ui logics
type proxyViewModel struct {
	LocalAddrVisible      bool
	LocalPortVisible      bool
	RemotePortVisible     bool
	RoleVisible           bool
	SKVisible             bool
	ServerNameVisible     bool
	BindAddrVisible       bool
	BindPortVisible       bool
	DomainVisible         bool
	HTTPVisible           bool
	MuxVisible            bool
	PluginEnable          bool
	PluginUnixVisible     bool
	PluginHTTPAuthVisible bool
	PluginAuthVisible     bool
	PluginStaticVisible   bool
	PluginHTTPFwdVisible  bool
	PluginCertVisible     bool
	HealthCheckEnable     bool
	HealthCheckVisible    bool
	HealthCheckURLVisible bool
}

// Data binder contains a copy of proxy
type editProxyBinder struct {
	config.Proxy

	// Extra fields needed for ui display
	Visitor       bool
	BandwidthNum  string
	BandwidthUnit string
}

func NewEditProxyDialog(proxy *config.Proxy, exist bool) *EditProxyDialog {
	v := &EditProxyDialog{exist: exist}
	if proxy == nil {
		proxy = config.NewDefaultProxyConfig("")
		v.exist = false
	}
	v.Proxy = proxy
	v.binder = &editProxyBinder{
		Proxy:   *v.Proxy,
		Visitor: proxy.Role == "visitor",
	}
	v.binder.BandwidthNum, v.binder.BandwidthUnit = splitBandwidth(v.Proxy.BandwidthLimit)
	return v
}

func (pd *EditProxyDialog) View() Dialog {
	dlg := NewBasicDialog(&pd.Dialog, "编辑代理", loadSysIcon("imageres", consts.IconEditDialog, 32), DataBinder{
		AssignTo:   &pd.vmDB,
		Name:       "vm",
		DataSource: &pd.viewModel,
	}, pd.onSave,
		Composite{
			DataBinder: DataBinder{
				AssignTo:   &pd.dbs[0],
				DataSource: pd.binder,
				OnCanSubmitChanged: func() {
					pd.DefaultButton().SetEnabled(pd.dbs[0].CanSubmit())
				},
			},
			Layout: Grid{Columns: 2, SpacingZero: false, Margins: Margins{0, 4, 0, 4}},
			Children: []Widget{
				Label{Text: "名称:", Alignment: AlignHNearVCenter},
				Composite{
					Layout: HBox{MarginsZero: true},
					Children: []Widget{
						LineEdit{AssignTo: &pd.nameView, Text: Bind("Name", consts.ValidateNonEmpty)},
						PushButton{Text: " 随机名称", Image: loadResourceIcon(consts.IconRefresh, 16), OnClicked: func() {
							pd.nameView.SetText(funk.RandomString(8))
						}},
					},
				},
				Label{Text: "类型:", Alignment: AlignHNearVCenter},
				ComboBox{
					AssignTo:              &pd.typeView,
					Model:                 consts.ProxyTypes,
					Value:                 Bind("Type"),
					OnCurrentIndexChanged: pd.switchType,
				},
			},
		},
		Composite{
			DataBinder: DataBinder{
				AssignTo:   &pd.dbs[1],
				DataSource: pd.binder,
			},
			Layout: VBox{MarginsZero: true, SpacingZero: true},
			Children: []Widget{
				TabWidget{
					MinSize: Size{0, 240},
					Pages: []TabPage{
						pd.baseProxyPage(),
						pd.advancedProxyPage(),
						pd.pluginProxyPage(),
						pd.loadBalanceProxyPage(),
						pd.healthCheckProxyPage(),
						pd.customProxyPage(),
					},
				},
			},
		},
	)
	dlg.Layout = VBox{Margins: Margins{7, 9, 7, 9}}
	return dlg
}

func (pd *EditProxyDialog) baseProxyPage() TabPage {
	return TabPage{
		Title:  "基本",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Visible: Bind("vm.RoleVisible"), Text: "角色:", MinSize: Size{Width: 55}},
			CheckBox{
				AssignTo: &pd.roleView,
				Visible:  Bind("vm.RoleVisible"), Text: "访问者",
				Checked: Bind("Visitor"), OnCheckedChanged: pd.switchType,
			},
			Label{Visible: Bind("vm.SKVisible"), Text: "私钥:"},
			LineEdit{Visible: Bind("vm.SKVisible"), Text: Bind("SK")},
			Label{Visible: Bind("vm.LocalAddrVisible"), Text: "本地地址:"},
			LineEdit{Visible: Bind("vm.LocalAddrVisible"), Text: Bind("LocalIP")},
			Label{Visible: Bind("vm.LocalPortVisible"), Text: "本地端口:"},
			LineEdit{
				AssignTo: &pd.localPortView, Visible: Bind("vm.LocalPortVisible"),
				Text: Bind("LocalPort"), OnTextChanged: pd.watchRangePort,
			},
			Label{Visible: Bind("vm.RemotePortVisible"), Text: "远程端口:"},
			LineEdit{
				AssignTo: &pd.remotePortView, Visible: Bind("vm.RemotePortVisible"),
				Text: Bind("RemotePort"), OnTextChanged: pd.watchRangePort,
			},
			Label{Visible: Bind("vm.BindAddrVisible"), Text: "绑定地址:"},
			LineEdit{Visible: Bind("vm.BindAddrVisible"), Text: Bind("BindAddr")},
			Label{Visible: Bind("vm.BindPortVisible"), Text: "绑定端口:"},
			LineEdit{Visible: Bind("vm.BindPortVisible"), Text: Bind("BindPort")},
			Label{Visible: Bind("vm.ServerNameVisible"), Text: "服务名称:"},
			LineEdit{Visible: Bind("vm.ServerNameVisible"), Text: Bind("ServerName")},
			Label{Visible: Bind("vm.DomainVisible"), Text: "子域名:"},
			LineEdit{Visible: Bind("vm.DomainVisible"), Text: Bind("SubDomain")},
			Label{Visible: Bind("vm.DomainVisible"), Text: "自定义域名:"},
			LineEdit{Visible: Bind("vm.DomainVisible"), Text: Bind("CustomDomains")},
			Label{Visible: Bind("vm.HTTPVisible"), Text: "URL 路由:"},
			LineEdit{Visible: Bind("vm.HTTPVisible"), Text: Bind("Locations")},
			Label{Visible: Bind("vm.MuxVisible"), Text: "复用器:"},
			LineEdit{Visible: Bind("vm.MuxVisible"), Text: Bind("Multiplexer")},
		},
	}
}

func (pd *EditProxyDialog) advancedProxyPage() TabPage {
	return TabPage{
		Title:  "高级",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Visible: Bind("vm.PluginEnable"), Text: "带宽限流:"},
			Composite{
				Visible: Bind("vm.PluginEnable"),
				Layout:  HBox{MarginsZero: true},
				Children: []Widget{
					LineEdit{Text: Bind("BandwidthNum")},
					ComboBox{Model: []string{"MB", "KB"}, Value: Bind("BandwidthUnit")},
				},
			},
			Label{Visible: Bind("vm.PluginEnable"), Text: "代理版本:"},
			ComboBox{
				Visible:       Bind("vm.PluginEnable"),
				Model:         NewDefaultListModel([]string{"v1", "v2"}, "", "空"),
				BindingMember: "Name",
				DisplayMember: "DisplayName",
				Value:         Bind("ProxyProtocolVersion"),
			},
			CheckBox{Text: "加密传输", Checked: Bind("UseEncryption"), MaxSize: Size{Width: 75}},
			CheckBox{Text: "压缩传输", Checked: Bind("UseCompression")},
			Label{Visible: Bind("vm.HTTPVisible"), Text: "HTTP 用户:"},
			LineEdit{Visible: Bind("vm.HTTPVisible"), Text: Bind("HTTPUser")},
			Label{Visible: Bind("vm.HTTPVisible"), Text: "HTTP 密码:"},
			LineEdit{Visible: Bind("vm.HTTPVisible"), Text: Bind("HTTPPwd")},
			Label{Visible: Bind("vm.HTTPVisible"), Text: "Host 替换:"},
			LineEdit{Visible: Bind("vm.HTTPVisible"), Text: Bind("HostHeaderRewrite")},
		},
	}
}

func (pd *EditProxyDialog) pluginProxyPage() TabPage {
	return TabPage{
		Title:  "插件",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "插件名称:", MinSize: Size{Width: 65}, Enabled: Bind("vm.PluginEnable")},
			ComboBox{
				AssignTo:              &pd.pluginView,
				Enabled:               Bind("vm.PluginEnable"),
				MinSize:               Size{Width: 250},
				Model:                 NewDefaultListModel(consts.PluginTypes, "", "无"),
				Value:                 Bind("Plugin"),
				BindingMember:         "Name",
				DisplayMember:         "DisplayName",
				OnCurrentIndexChanged: pd.switchType,
			},
			Label{Visible: Bind("vm.PluginUnixVisible"), Text: "Unix 路径:"},
			NewBrowseLineEdit(nil, Bind("vm.PluginUnixVisible"), true, Bind("PluginUnixPath"),
				"选择 Unix 路径", "", true),
			Label{Visible: Bind("vm.PluginStaticVisible"), Text: "本地路径:"},
			NewBrowseLineEdit(nil, Bind("vm.PluginStaticVisible"), true, Bind("PluginLocalPath"),
				"选择本地文件夹", "", false),
			Label{Visible: Bind("vm.PluginStaticVisible"), Text: "移除前缀:"},
			LineEdit{Visible: Bind("vm.PluginStaticVisible"), Text: Bind("PluginStripPrefix")},
			Label{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: "HTTP 用户:"},
			LineEdit{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: Bind("PluginHttpUser")},
			Label{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: "HTTP 密码:"},
			LineEdit{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: Bind("PluginHttpPasswd")},
			Label{Visible: Bind("vm.PluginAuthVisible"), Text: "用户名:"},
			LineEdit{Visible: Bind("vm.PluginAuthVisible"), Text: Bind("PluginUser")},
			Label{Visible: Bind("vm.PluginAuthVisible"), Text: "密码:"},
			LineEdit{Visible: Bind("vm.PluginAuthVisible"), Text: Bind("PluginPasswd")},
			Label{Visible: Bind("vm.PluginHTTPFwdVisible"), Text: "本地地址:"},
			LineEdit{Visible: Bind("vm.PluginHTTPFwdVisible"), Text: Bind("PluginLocalAddr")},
			Label{Visible: Bind("vm.PluginCertVisible"), Text: "证书路径:"},
			NewBrowseLineEdit(nil, Bind("vm.PluginCertVisible"), true, Bind("PluginCrtPath"),
				"选择证书文件", consts.FilterCert, true),
			Label{Visible: Bind("vm.PluginCertVisible"), Text: "密钥路径:"},
			NewBrowseLineEdit(nil, Bind("vm.PluginCertVisible"), true, Bind("PluginKeyPath"),
				"选择密钥文件", consts.FilterKey, true),
			Label{Visible: Bind("vm.PluginHTTPFwdVisible"), Text: "Host 替换:"},
			LineEdit{Visible: Bind("vm.PluginHTTPFwdVisible"), Text: Bind("PluginHostHeaderRewrite")},
		},
	}
}

func (pd *EditProxyDialog) loadBalanceProxyPage() TabPage {
	return TabPage{
		Title:  "负载均衡",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Enabled: Bind("vm.PluginEnable"), Text: "分组名称:"},
			LineEdit{Enabled: Bind("vm.PluginEnable"), Text: Bind("Group")},
			Label{Enabled: Bind("vm.PluginEnable"), Text: "分组密钥:"},
			LineEdit{Enabled: Bind("vm.PluginEnable"), Text: Bind("GroupKey")},
		},
	}
}

func (pd *EditProxyDialog) healthCheckProxyPage() TabPage {
	return TabPage{
		Title:  "健康检查",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "检查类型:", Enabled: Bind("vm.HealthCheckEnable"), MinSize: Size{Width: 55}},
			NewRadioButtonGroup("HealthCheckType", &DataBinder{DataSource: pd.binder, AutoSubmit: true}, []RadioButton{
				{Text: "tcp", Value: "tcp", Enabled: Bind("vm.HealthCheckEnable"), OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
				{Text: "http", Value: "http", Enabled: Bind("vm.HealthCheckEnable"), OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
				{Text: "无", Value: "", Enabled: Bind("vm.HealthCheckEnable"), OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
			}),
			Label{Visible: Bind("vm.HealthCheckURLVisible"), Text: "URL:"},
			LineEdit{Visible: Bind("vm.HealthCheckURLVisible"), Text: Bind("HealthCheckURL")},
			Label{Visible: Bind("vm.HealthCheckVisible"), Text: "超时时间:"},
			NumberEdit{Visible: Bind("vm.HealthCheckVisible"), Value: Bind("HealthCheckTimeoutS"), Suffix: " 秒"},
			Label{Visible: Bind("vm.HealthCheckVisible"), Text: "错误次数:"},
			NumberEdit{Visible: Bind("vm.HealthCheckVisible"), Value: Bind("HealthCheckMaxFailed")},
			Label{Visible: Bind("vm.HealthCheckVisible"), Text: "检查周期:"},
			NumberEdit{Visible: Bind("vm.HealthCheckVisible"), Value: Bind("HealthCheckIntervalS"), Suffix: " 秒"},
		},
	}
}

func (pd *EditProxyDialog) customProxyPage() TabPage {
	return TabPage{
		Title:  "自定义",
		Layout: VBox{},
		Children: []Widget{
			Label{Text: "*参考 FRP 支持的参数，每行格式为 a = b"},
			TextEdit{AssignTo: &pd.customText, Text: util.Map2String(pd.binder.Custom), VScroll: true},
		},
	}
}

func (pd *EditProxyDialog) Run(owner walk.Form) (int, error) {
	if err := pd.View().Create(owner); err != nil {
		return 0, err
	}
	pd.DefaultButton().SetEnabled(pd.dbs[0].CanSubmit())
	return pd.Dialog.Run(), nil
}

func (pd *EditProxyDialog) onSave() {
	for _, db := range pd.dbs {
		if err := db.Submit(); err != nil {
			return
		}
	}
	if pd.exist {
		// Change proxy name
		if pd.binder.Name != pd.Proxy.Name && pd.hasProxy(pd.binder.Name) {
			return
		}
	} else if pd.hasProxy(pd.binder.Name) {
		return
	}
	// Update custom options
	pd.binder.Proxy.Custom = util.String2Map(pd.customText.Text())
	// Update role
	if pd.binder.Visitor {
		pd.binder.Proxy.Role = "visitor"
	} else {
		pd.binder.Proxy.Role = ""
	}
	// Update bandwidth
	if pd.binder.BandwidthNum != "" {
		pd.binder.Proxy.BandwidthLimit = pd.binder.BandwidthNum + pd.binder.BandwidthUnit
	} else {
		pd.binder.Proxy.BandwidthLimit = ""
	}
	pb, err := pd.binder.Proxy.Marshal()
	if err != nil {
		showError(err, pd.Form())
		return
	}
	if err = services.VerifyClientProxy(pb); err != nil {
		showError(err, pd.Form())
		return
	}
	*pd.Proxy = pd.binder.Proxy
	pd.Accept()
}

func (pd *EditProxyDialog) hasProxy(name string) bool {
	if conf := getCurrentConf(); conf != nil {
		if funk.Contains(conf.Data.Items(), func(proxy *config.Proxy) bool { return proxy.Name == name }) {
			showWarningMessage(pd.Form(), "代理已存在", fmt.Sprintf("代理名「%s」已存在。", name))
			return true
		}
	}
	return false
}

func (pd *EditProxyDialog) switchType() {
	// Default ui settings
	pd.viewModel = proxyViewModel{
		LocalAddrVisible: true, LocalPortVisible: true,
		PluginEnable: true, HealthCheckEnable: true,
		HealthCheckVisible: pd.binder.HealthCheckType != "",
	}
	// All types: tcp, udp, xtcp, stcp, sudp, http, https, tcpmux
	switch pd.typeView.Text() {
	case consts.ProxyTypeTCP, consts.ProxyTypeUDP:
		pd.viewModel.RemotePortVisible = true
	case consts.ProxyTypeXTCP, consts.ProxyTypeSTCP, consts.ProxyTypeSUDP:
		pd.viewModel.RoleVisible = true
		pd.viewModel.SKVisible = true
		// For visitor
		if pd.roleView.Checked() {
			pd.viewModel.ServerNameVisible = true
			pd.viewModel.BindAddrVisible = true
			pd.viewModel.BindPortVisible = true
			// Visitor doesn't have these options, so it should be hided
			pd.viewModel.LocalAddrVisible = false
			pd.viewModel.LocalPortVisible = false
			// Plugin, bandwidth, group options should be hided too.
			// Thus, it can share the same control flag
			pd.viewModel.PluginEnable = false
			pd.viewModel.HealthCheckVisible = false
			pd.viewModel.HealthCheckEnable = false
		}
	case consts.ProxyTypeHTTP:
		pd.viewModel.DomainVisible = true
		pd.viewModel.HTTPVisible = true
	case consts.ProxyTypeHTTPS:
		pd.viewModel.DomainVisible = true
	case consts.ProxyTypeTCPMUX:
		pd.viewModel.DomainVisible = true
		pd.viewModel.MuxVisible = true
	}
	pd.viewModel.HealthCheckURLVisible = pd.viewModel.HealthCheckVisible && pd.binder.HealthCheckType == "http"
	if pd.pluginView.CurrentIndex() > 0 {
		pd.viewModel.LocalAddrVisible = false
		pd.viewModel.LocalPortVisible = false
		if pd.viewModel.PluginEnable {
			switch pd.pluginView.Text() {
			case consts.PluginUnixDomain:
				pd.viewModel.PluginUnixVisible = true
			case consts.PluginHttpProxy:
				pd.viewModel.PluginHTTPAuthVisible = true
			case consts.PluginSocks5:
				pd.viewModel.PluginAuthVisible = true
			case consts.PluginStaticFile:
				pd.viewModel.PluginStaticVisible = true
				pd.viewModel.PluginHTTPAuthVisible = true
			case consts.PluginHttps2Http, consts.PluginHttps2Https:
				pd.viewModel.PluginHTTPFwdVisible = true
				pd.viewModel.PluginCertVisible = true
			case consts.PluginHttp2Https:
				pd.viewModel.PluginHTTPFwdVisible = true
			}
		}
	}
	pd.watchRangePort()
	pd.vmDB.Reset()
}

func (pd *EditProxyDialog) watchRangePort() {
	var isRange bool
	// The "range:" function requires both local port and remote port are set
	if pd.viewModel.LocalPortVisible && pd.viewModel.RemotePortVisible {
		for _, portView := range []*walk.LineEdit{pd.localPortView, pd.remotePortView} {
			portText := portView.Text()
			isRange = strings.Contains(portText, "-") || strings.Contains(portText, ",")
			if isRange {
				break
			}
		}
	}
	proxyName := pd.nameView.Text()
	hasPrefix := strings.HasPrefix(proxyName, consts.RangePrefix)
	if isRange {
		if !hasPrefix {
			pd.nameView.SetText(consts.RangePrefix + proxyName)
		}
	} else {
		if hasPrefix {
			pd.nameView.SetText(strings.TrimPrefix(proxyName, consts.RangePrefix))
		}
	}
}

func splitBandwidth(s string) (string, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "MB"
	}
	if strings.HasSuffix(s, "MB") {
		return strings.TrimSuffix(s, "MB"), "MB"
	} else if strings.HasSuffix(s, "KB") {
		return strings.TrimSuffix(s, "KB"), "KB"
	}
	return "", "MB"
}
