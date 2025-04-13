package ui

import (
	"math"
	"strconv"
	"strings"

	"github.com/fatedier/frp/pkg/config/v1/validation"
	frputil "github.com/fatedier/frp/pkg/util/util"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
)

type EditProxyDialog struct {
	*walk.Dialog

	Proxy        *config.Proxy
	visitors     []string
	create       bool
	legacyFormat bool
	nameChecker  func(string) bool

	// View models
	binder    *editProxyBinder
	dbs       [2]*walk.DataBinder
	vmDB      *walk.DataBinder
	viewModel proxyViewModel
	metaModel *AttributeModel

	// Views
	nameView   *walk.LineEdit
	typeView   *walk.ComboBox
	pluginView *walk.ComboBox
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
	PluginAddrVisible     bool
	HealthCheckEnable     bool
	HealthCheckVisible    bool
	HealthCheckURLVisible bool
}

// Data binder contains a copy of proxy
type editProxyBinder struct {
	config.Proxy

	// Extra fields needed for ui display
	Visitor       bool
	BandwidthNum  int64
	BandwidthUnit string
}

func NewEditProxyDialog(proxy *config.Proxy, visitors []string, create, legacyFormat bool, nameChecker func(string) bool) *EditProxyDialog {
	v := &EditProxyDialog{
		Proxy:        proxy,
		visitors:     visitors,
		create:       create,
		legacyFormat: legacyFormat,
		nameChecker:  nameChecker,
	}
	if v.Proxy == nil {
		v.Proxy = config.NewDefaultProxyConfig("")
	}
	v.binder = &editProxyBinder{
		Proxy:   *v.Proxy,
		Visitor: v.Proxy.IsVisitor(),
	}
	v.binder.BandwidthNum, v.binder.BandwidthUnit = splitBandwidth(v.Proxy.BandwidthLimit)
	if v.Proxy.BandwidthLimitMode == "" {
		v.binder.BandwidthLimitMode = consts.BandwidthMode[0]
	}
	v.metaModel = NewAttributeModel(v.binder.Metas)
	// HTTP/2 should be enabled by default.
	if v.binder.Plugin != consts.PluginHttps2Http && v.binder.Plugin != consts.PluginHttps2Https {
		v.binder.PluginEnableHTTP2 = true
	}
	return v
}

func (pd *EditProxyDialog) View() Dialog {
	pages := []TabPage{
		pd.basicProxyPage(),
		pd.advancedProxyPage(),
		pd.pluginProxyPage(),
		pd.loadBalanceProxyPage(),
		pd.healthCheckProxyPage(),
		pd.metadataProxyPage(),
	}
	title := i18n.Sprintf("New Proxy")
	if !pd.create {
		title = i18n.Sprintf("Edit Proxy - %s", pd.Proxy.Name)
	}
	var header Widget = ComboBox{
		Name:                  "proxyType",
		AssignTo:              &pd.typeView,
		Model:                 consts.ProxyTypes,
		Value:                 Bind("Type"),
		Greedy:                true,
		OnCurrentIndexChanged: pd.switchType,
	}
	if !pd.legacyFormat {
		header = Composite{
			Layout: HBox{MarginsZero: true},
			Children: []Widget{
				header,
				HSpacer{},
				ToolButton{
					Image:       loadIcon(res.IconInfo, 16),
					ToolTipText: i18n.Sprintf("Annotations"),
					OnClicked: func() {
						NewAttributeDialog(i18n.Sprintf("Annotations"), &pd.binder.Annotations).Run(pd.Form())
					},
				},
			},
		}
	}
	dlg := NewBasicDialog(&pd.Dialog, title, loadIcon(res.IconEditDialog, 32), DataBinder{
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
			Layout: Grid{Columns: 2, Spacing: 12, Margins: Margins{Top: 4, Bottom: 4}},
			Children: []Widget{
				Label{Text: i18n.SprintfColon("Name"), Alignment: AlignHNearVCenter},
				Composite{
					Layout: HBox{MarginsZero: true},
					Children: []Widget{
						LineEdit{AssignTo: &pd.nameView, Text: Bind("Name", res.ValidateNonEmpty)},
						PushButton{Text: i18n.SprintfLSpace("Random"), Image: loadIcon(res.IconRandom, 16), OnClicked: func() {
							pd.nameView.SetText(lo.RandomString(8, lo.AlphanumericCharset))
						}},
					},
				},
				Label{Text: i18n.SprintfColon("Type"), Alignment: AlignHNearVCenter},
				header,
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
					MinSize: Size{Height: 240},
					Pages:   pages,
				},
			},
		},
	)
	dlg.Layout = VBox{Margins: Margins{Left: 7, Top: 9, Right: 7, Bottom: 9}}
	minWidth := lo.Sum(lo.Map(pages, func(page TabPage, i int) int {
		return calculateStringWidth(page.Title.(string)) + 20
	})) + 20
	// Keep a better aspect ratio
	if minWidth < 350 {
		minWidth += 30
	}
	dlg.MinSize = Size{Width: minWidth, Height: 420}
	return dlg
}

func (pd *EditProxyDialog) basicProxyPage() TabPage {
	showRequestHeaders := func() {
		NewAttributeDialog(i18n.Sprintf("Request headers"), &pd.binder.Headers).Run(pd.Form())
	}
	var headerBtn Widget
	if pd.legacyFormat {
		headerBtn = ToolButton{Text: "H", ToolTipText: i18n.Sprintf("Request headers"), OnClicked: showRequestHeaders}
	} else {
		headerBtn = SplitButton{
			Text:        "H",
			ToolTipText: i18n.Sprintf("Request headers"),
			MaxSize:     Size{Width: 40},
			OnClicked:   showRequestHeaders,
			MenuItems: []MenuItem{
				Action{Text: i18n.Sprintf("Request headers"), OnTriggered: showRequestHeaders},
				Action{Text: i18n.Sprintf("Response headers"), OnTriggered: func() {
					NewAttributeDialog(i18n.Sprintf("Response headers"), &pd.binder.ResponseHeaders).Run(pd.Form())
				}},
			},
		}
	}
	return AlignGrid(TabPage{
		Title:  i18n.Sprintf("Basic"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Visible: Bind("vm.RoleVisible"), Text: i18n.SprintfColon("Role")},
			NewRadioButtonGroup("Visitor", &DataBinder{DataSource: pd.binder, AutoSubmit: true},
				Bind("vm.RoleVisible"), []RadioButton{
					{Text: i18n.Sprintf("Server"), Value: false, OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
					{Text: i18n.Sprintf("Visitor"), Value: true, OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
				}),
			Label{Visible: Bind("vm.SKVisible"), Text: i18n.SprintfColon("Secret Key")},
			LineEdit{Visible: Bind("vm.SKVisible"), Text: Bind("SK"), PasswordMode: true},
			Label{Visible: Bind("vm.LocalAddrVisible"), Text: i18n.SprintfColon("Local Address")},
			LineEdit{Visible: Bind("vm.LocalAddrVisible"), Text: Bind("LocalIP"), CueBanner: "127.0.0.1"},
			Label{Visible: Bind("vm.LocalPortVisible"), Text: i18n.SprintfColon("Local Port")},
			LineEdit{Visible: Bind("vm.LocalPortVisible"), Text: Bind("LocalPort")},
			Label{Visible: Bind("vm.RemotePortVisible"), Text: i18n.SprintfColon("Remote Port")},
			LineEdit{Visible: Bind("vm.RemotePortVisible"), Text: Bind("RemotePort")},
			Label{Visible: Bind("vm.RoleVisible && !vm.ServerNameVisible"), Text: i18n.SprintfColon("Allow Users")},
			NewListEdit(pd, Bind("vm.RoleVisible && !vm.ServerNameVisible"), Bind("AllowUsers"), i18n.Sprintf("Allow Users")),
			Label{Visible: Bind("vm.BindAddrVisible"), Text: i18n.SprintfColon("Bind Address")},
			LineEdit{Visible: Bind("vm.BindAddrVisible"), Text: Bind("BindAddr"), CueBanner: "127.0.0.1"},
			Label{Visible: Bind("vm.BindPortVisible"), Text: i18n.SprintfColon("Bind Port")},
			NumberEdit{Visible: Bind("vm.BindPortVisible"), Value: Bind("BindPort"), MinValue: -math.MaxFloat64, MaxValue: 65535},
			Label{Visible: Bind("vm.ServerNameVisible"), Text: i18n.SprintfColon("Server Name")},
			LineEdit{Visible: Bind("vm.ServerNameVisible"), Text: Bind("ServerName")},
			Label{Visible: Bind("vm.ServerNameVisible"), Text: i18n.SprintfColon("Server User")},
			LineEdit{Visible: Bind("vm.ServerNameVisible"), Text: Bind("ServerUser")},
			Label{Visible: Bind("vm.DomainVisible"), Text: i18n.SprintfColon("Subdomain")},
			LineEdit{Visible: Bind("vm.DomainVisible"), Text: Bind("SubDomain")},
			Label{Visible: Bind("vm.DomainVisible"), Text: i18n.SprintfColon("Custom Domains")},
			NewListEdit(pd, Bind("vm.DomainVisible"), Bind("CustomDomains"), i18n.Sprintf("Custom Domains")),
			Label{Visible: Bind("vm.HTTPVisible"), Text: i18n.SprintfColon("Locations")},
			NewListEdit(pd, Bind("vm.HTTPVisible"), Bind("Locations"), i18n.Sprintf("Locations"), headerBtn),
			Label{Visible: Bind("vm.MuxVisible"), Text: i18n.SprintfColon("Multiplexer")},
			ComboBox{
				Visible: Bind("vm.MuxVisible"),
				Model:   []string{consts.HTTPConnectTCPMultiplexer},
				Value:   Bind("Multiplexer"),
			},
			Label{Visible: Bind("vm.MuxVisible || vm.HTTPVisible"), Text: i18n.SprintfColon("Route User")},
			LineEdit{Visible: Bind("vm.MuxVisible || vm.HTTPVisible"), Text: Bind("RouteByHTTPUser")},
		},
	}, 16) // We only calculate children related to the first widget "role"
}

func (pd *EditProxyDialog) advancedProxyPage() TabPage {
	bandwidthMode := NewListModel(consts.BandwidthMode, i18n.Sprintf("Client"), i18n.Sprintf("Server"))
	var xtcpVisitor = Bind("proxyType.Value == 'xtcp' && vm.ServerNameVisible")
	return TabPage{
		Title:  i18n.Sprintf("Advanced"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Visible: Bind("vm.PluginEnable"), Text: i18n.SprintfColon("Bandwidth")},
			Composite{
				Visible: Bind("vm.PluginEnable"),
				Layout:  HBox{MarginsZero: true},
				Children: []Widget{
					NumberEdit{
						Value:              Bind("BandwidthNum"),
						MinValue:           0,
						MaxValue:           math.MaxFloat64,
						SpinButtonsVisible: true,
						Style:              win.ES_RIGHT,
						Greedy:             true,
					},
					ComboBox{Model: consts.Bandwidth, Value: Bind("BandwidthUnit")},
					Label{Text: "@"},
					ComboBox{
						Model:         bandwidthMode,
						BindingMember: "Value",
						DisplayMember: "Title",
						Value:         Bind("BandwidthLimitMode"),
					},
				},
			},
			Label{Visible: Bind("vm.PluginEnable"), Text: i18n.SprintfColon("Proxy Protocol")},
			ComboBox{
				Visible:       Bind("vm.PluginEnable"),
				Model:         NewListModel([]string{"", "v1", "v2"}, i18n.Sprintf("auto")),
				BindingMember: "Value",
				DisplayMember: "Title",
				Value:         Bind("ProxyProtocolVersion"),
			},
			Label{Visible: xtcpVisitor, Text: i18n.SprintfColon("Protocol")},
			ComboBox{
				Visible:       xtcpVisitor,
				Model:         NewListModel([]string{"", consts.ProtoQUIC, consts.ProtoKCP}, i18n.Sprintf("default")),
				BindingMember: "Value",
				DisplayMember: "Title",
				Value:         Bind("Protocol"),
			},
			Composite{
				Layout:     HBox{MarginsZero: true},
				ColumnSpan: 2,
				Children: []Widget{
					CheckBox{Name: "keepTunnel", Visible: xtcpVisitor, Text: i18n.Sprintf("Keep Tunnel"), Checked: Bind("KeepTunnelOpen")},
					CheckBox{Text: i18n.Sprintf("Encryption"), Checked: Bind("UseEncryption")},
					CheckBox{Text: i18n.Sprintf("Compression"), Checked: Bind("UseCompression")},
					CheckBox{Text: "HTTP/2", Visible: Bind("vm.PluginHTTPFwdVisible && vm.PluginCertVisible"), Checked: Bind("PluginEnableHTTP2")},
				},
			},
			Label{Visible: xtcpVisitor, Text: i18n.SprintfColon("Fallback")},
			ComboBox{
				Name:     "fallback",
				Editable: true,
				Visible:  xtcpVisitor,
				Model:    pd.visitors,
				Value:    Bind("FallbackTo"),
			},
			Label{Visible: xtcpVisitor, Enabled: Bind("fallback.Value != ''"), Text: i18n.SprintfColon("Fallback Timeout")},
			NewNumberInput(NIOption{
				Visible: xtcpVisitor,
				Enabled: Bind("fallback.Value != ''"),
				Value:   Bind("FallbackTimeoutMs"),
				Suffix:  i18n.Sprintf("ms"),
				Max:     math.MaxFloat64,
			}),
			Label{Visible: xtcpVisitor, Enabled: Bind("keepTunnel.Checked"), Text: i18n.SprintfColon("Retry Count")},
			NewNumberInput(NIOption{
				Visible: xtcpVisitor,
				Enabled: Bind("keepTunnel.Checked"),
				Value:   Bind("MaxRetriesAnHour"),
				Suffix:  i18n.Sprintf("Times/Hour"),
				Max:     math.MaxFloat64,
			}),
			Label{Visible: xtcpVisitor, Enabled: Bind("keepTunnel.Checked"), Text: i18n.SprintfColon("Retry Interval")},
			NewNumberInput(NIOption{
				Visible: xtcpVisitor,
				Enabled: Bind("keepTunnel.Checked"),
				Value:   Bind("MinRetryInterval"),
				Suffix:  i18n.Sprintf("s"),
				Max:     math.MaxFloat64,
			}),
			Label{Visible: Bind("vm.MuxVisible || vm.HTTPVisible"), Text: i18n.SprintfColon("HTTP User")},
			LineEdit{Visible: Bind("vm.MuxVisible || vm.HTTPVisible"), Text: Bind("HTTPUser")},
			Label{Visible: Bind("vm.MuxVisible || vm.HTTPVisible"), Text: i18n.SprintfColon("HTTP Password")},
			LineEdit{Visible: Bind("vm.MuxVisible || vm.HTTPVisible"), Text: Bind("HTTPPwd"), PasswordMode: true},
			Label{Visible: Bind("vm.HTTPVisible"), Text: i18n.SprintfColon("Host Rewrite")},
			LineEdit{Visible: Bind("vm.HTTPVisible"), Text: Bind("HostHeaderRewrite")},
		},
	}
}

func (pd *EditProxyDialog) pluginProxyPage() TabPage {
	return AlignGrid(TabPage{
		Title:  i18n.Sprintf("Plugin"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Plugin Name"), Enabled: Bind("vm.PluginEnable")},
			ComboBox{
				AssignTo:              &pd.pluginView,
				Enabled:               Bind("vm.PluginEnable"),
				Model:                 NewListModel(append([]string{""}, consts.PluginTypes...), i18n.Sprintf("None")),
				Value:                 Bind("Plugin"),
				BindingMember:         "Value",
				DisplayMember:         "Title",
				OnCurrentIndexChanged: pd.switchType,
				Greedy:                true,
			},
			Label{Visible: Bind("vm.PluginUnixVisible"), Text: i18n.SprintfColon("Unix Path")},
			NewBrowseLineEdit(nil, Bind("vm.PluginUnixVisible"), true, Bind("PluginUnixPath"),
				i18n.Sprintf("Select Unix Path"), "", true),
			Label{Visible: Bind("vm.PluginStaticVisible"), Text: i18n.SprintfColon("Local Path")},
			NewBrowseLineEdit(nil, Bind("vm.PluginStaticVisible"), true, Bind("PluginLocalPath"),
				i18n.Sprintf("Select a folder for directory listing."), "", false),
			Label{Visible: Bind("vm.PluginStaticVisible"), Text: i18n.SprintfColon("Strip Prefix")},
			LineEdit{Visible: Bind("vm.PluginStaticVisible"), Text: Bind("PluginStripPrefix")},
			Label{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: i18n.SprintfColon("HTTP User")},
			LineEdit{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: Bind("PluginHttpUser")},
			Label{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: i18n.SprintfColon("HTTP Password")},
			LineEdit{Visible: Bind("vm.PluginHTTPAuthVisible"), Text: Bind("PluginHttpPasswd"), PasswordMode: true},
			Label{Visible: Bind("vm.PluginAuthVisible"), Text: i18n.SprintfColon("User")},
			LineEdit{Visible: Bind("vm.PluginAuthVisible"), Text: Bind("PluginUser")},
			Label{Visible: Bind("vm.PluginAuthVisible"), Text: i18n.SprintfColon("Password")},
			LineEdit{Visible: Bind("vm.PluginAuthVisible"), Text: Bind("PluginPasswd"), PasswordMode: true},
			Label{Visible: Bind("vm.PluginAddrVisible"), Text: i18n.SprintfColon("Local Address")},
			Composite{
				Visible: Bind("vm.PluginAddrVisible"),
				Layout:  HBox{MarginsZero: true},
				Children: []Widget{
					LineEdit{Text: Bind("PluginLocalAddr")},
					ToolButton{
						Visible:     Bind("vm.PluginHTTPFwdVisible"),
						Text:        "H",
						ToolTipText: i18n.Sprintf("Request headers"),
						OnClicked: func() {
							NewAttributeDialog(i18n.Sprintf("Request headers"), &pd.binder.PluginHeaders).Run(pd.Form())
						},
					},
				},
			},
			Label{Visible: Bind("vm.PluginCertVisible"), Text: i18n.SprintfColon("Certificate")},
			NewBrowseLineEdit(nil, Bind("vm.PluginCertVisible"), true, Bind("PluginCrtPath"),
				i18n.Sprintf("Select Certificate File"), res.FilterCert, true),
			Label{Visible: Bind("vm.PluginCertVisible"), Text: i18n.SprintfColon("Certificate Key")},
			NewBrowseLineEdit(nil, Bind("vm.PluginCertVisible"), true, Bind("PluginKeyPath"),
				i18n.Sprintf("Select Certificate Key File"), res.FilterKey, true),
			Label{Visible: Bind("vm.PluginHTTPFwdVisible"), Text: i18n.SprintfColon("Host Rewrite")},
			LineEdit{Visible: Bind("vm.PluginHTTPFwdVisible"), Text: Bind("PluginHostHeaderRewrite")},
		},
	}, 0)
}

func (pd *EditProxyDialog) loadBalanceProxyPage() TabPage {
	return TabPage{
		Title:  i18n.Sprintf("Load Balance"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Enabled: Bind("vm.PluginEnable"), Text: i18n.SprintfColon("Group")},
			LineEdit{Enabled: Bind("vm.PluginEnable"), Text: Bind("Group")},
			Label{Enabled: Bind("vm.PluginEnable"), Text: i18n.SprintfColon("Group Key")},
			LineEdit{Enabled: Bind("vm.PluginEnable"), Text: Bind("GroupKey")},
		},
	}
}

func (pd *EditProxyDialog) healthCheckProxyPage() TabPage {
	var url Widget = LineEdit{Visible: Bind("vm.HealthCheckURLVisible"), Text: Bind("HealthCheckURL")}
	if !pd.legacyFormat {
		url = Composite{
			Visible: Bind("vm.HealthCheckURLVisible"),
			Layout:  HBox{MarginsZero: true},
			Children: []Widget{
				url,
				ToolButton{Text: "H", ToolTipText: i18n.Sprintf("Request headers"), OnClicked: func() {
					NewAttributeDialog(i18n.Sprintf("Request headers"), &pd.binder.HealthCheckHTTPHeaders).Run(pd.Form())
				}},
			},
		}
	}
	return AlignGrid(TabPage{
		Title:  i18n.Sprintf("Health Check"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Check Type"), Enabled: Bind("vm.HealthCheckEnable")},
			NewRadioButtonGroup("HealthCheckType", &DataBinder{DataSource: pd.binder, AutoSubmit: true}, nil, []RadioButton{
				{Text: "TCP", Value: "tcp", Enabled: Bind("vm.HealthCheckEnable"), OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
				{Text: "HTTP", Value: "http", Enabled: Bind("vm.HealthCheckEnable"), OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
				{Text: i18n.Sprintf("None"), Value: "", Enabled: Bind("vm.HealthCheckEnable"), OnClicked: pd.switchType, MaxSize: Size{Width: 80}},
			}),
			Label{Visible: Bind("vm.HealthCheckURLVisible"), Text: "URL:"},
			url,
			Label{Visible: Bind("vm.HealthCheckVisible"), Text: i18n.SprintfColon("Check Timeout")},
			NewNumberInput(NIOption{Visible: Bind("vm.HealthCheckVisible"), Value: Bind("HealthCheckTimeoutS"), Suffix: i18n.Sprintf("s")}),
			Label{Visible: Bind("vm.HealthCheckVisible"), Text: i18n.SprintfColon("Check Interval")},
			NewNumberInput(NIOption{Visible: Bind("vm.HealthCheckVisible"), Value: Bind("HealthCheckIntervalS"), Suffix: i18n.Sprintf("s")}),
			Label{Visible: Bind("vm.HealthCheckVisible"), Text: i18n.SprintfColon("Failure Count")},
			NewNumberInput(NIOption{Visible: Bind("vm.HealthCheckVisible"), Value: Bind("HealthCheckMaxFailed")}),
		},
	}, 0)
}

func (pd *EditProxyDialog) metadataProxyPage() TabPage {
	return TabPage{
		Title:  i18n.Sprintf("Metadata"),
		Layout: VBox{},
		Children: []Widget{
			NewAttributeTable(pd.metaModel, 0, 0),
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
	if pd.create || pd.binder.Name != pd.Proxy.Name {
		if pd.hasProxy(pd.binder.Name) {
			return
		}
	}
	// Update metadata
	pd.binder.Proxy.Metas = pd.metaModel.AsMap()
	// Update role
	if pd.binder.Visitor {
		pd.binder.Proxy.Role = "visitor"
	} else {
		pd.binder.Proxy.Role = ""
	}
	// Update bandwidth
	if pd.binder.BandwidthNum > 0 {
		pd.binder.Proxy.BandwidthLimit = strconv.FormatInt(pd.binder.BandwidthNum, 10) + pd.binder.BandwidthUnit
		if pd.binder.Proxy.BandwidthLimitMode == consts.BandwidthMode[0] {
			pd.binder.Proxy.BandwidthLimitMode = ""
		}
	} else {
		pd.binder.Proxy.BandwidthLimit = ""
		pd.binder.Proxy.BandwidthLimitMode = ""
	}
	pd.binder.Proxy.LocalPort = strings.TrimSpace(pd.binder.Proxy.LocalPort)
	pd.binder.Proxy.RemotePort = strings.TrimSpace(pd.binder.Proxy.RemotePort)
	if ok := pd.validateProxy(pd.binder.Proxy); !ok {
		return
	}
	*pd.Proxy = pd.binder.Proxy
	pd.Proxy.Complete()
	pd.Accept()
}

func (pd *EditProxyDialog) hasProxy(name string) bool {
	if pd.nameChecker(name) {
		showWarningMessage(pd.Form(), i18n.Sprintf("Proxy already exists"), i18n.Sprintf("The proxy name \"%s\" already exists.", name))
		return true
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
		if pd.binder.Visitor {
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
				pd.viewModel.PluginAddrVisible = true
				pd.viewModel.PluginHTTPFwdVisible = true
				pd.viewModel.PluginCertVisible = true
			case consts.PluginHttp2Https, consts.PluginHttp2Http:
				pd.viewModel.PluginAddrVisible = true
				pd.viewModel.PluginHTTPFwdVisible = true
			case consts.PluginTLS2Raw:
				pd.viewModel.PluginAddrVisible = true
				pd.viewModel.PluginCertVisible = true
			}
		}
	}
	pd.vmDB.Reset()
}

func splitBandwidth(s string) (int64, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, "MB"
	}

	if strings.HasSuffix(s, "MB") || strings.HasSuffix(s, "KB") {
		unit := s[len(s)-2:]
		num, _ := strconv.ParseInt(strings.TrimSuffix(s, unit), 10, 64)
		return num, unit
	}
	return 0, "MB"
}

func (pd *EditProxyDialog) validateProxy(p config.Proxy) bool {
	if p.IsVisitor() {
		if p.ServerName == "" {
			showErrorMessage(pd.Form(), "", i18n.Sprintf("Server name is required."))
			return false
		}
		if p.BindPort == 0 {
			showErrorMessage(pd.Form(), "", i18n.Sprintf("Bind port is required."))
			return false
		}
		return true
	}
	if !pd.legacyFormat {
		if err := validation.ValidateAnnotations(p.Annotations); err != nil {
			showError(err, pd.Form())
			return false
		}
	}
	if p.Plugin == "" && p.LocalPort == "" {
		showErrorMessage(pd.Form(), "", i18n.Sprintf("Requires local port or plugin."))
		return false
	}
	if p.Plugin != "" {
		p.LocalIP = ""
		p.LocalPort = ""
		switch p.Plugin {
		case consts.PluginHttp2Https, consts.PluginHttps2Http, consts.PluginHttps2Https, consts.PluginTLS2Raw:
			if p.PluginLocalAddr == "" {
				showErrorMessage(pd.Form(), "", i18n.Sprintf("Local address is required."))
				return false
			}
		case consts.PluginStaticFile:
			if p.PluginLocalPath == "" {
				showErrorMessage(pd.Form(), "", i18n.Sprintf("Local path is required."))
				return false
			}
		case consts.PluginUnixDomain:
			if p.PluginUnixPath == "" {
				showErrorMessage(pd.Form(), "", i18n.Sprintf("Unix path is required."))
				return false
			}
		}
	} else if p.Type != consts.ProxyTypeTCP && p.Type != consts.ProxyTypeUDP {
		if port, err := strconv.ParseInt(p.LocalPort, 10, 64); err != nil || port <= 0 || port > 65535 {
			showErrorMessage(pd.Form(), "", i18n.Sprintf("Invalid local port."))
			return false
		}
	}
	if p.HealthCheckType == "http" && p.HealthCheckURL == "" {
		showErrorMessage(pd.Form(), "", i18n.Sprintf("Health check url is required."))
		return false
	}

	switch p.Type {
	case consts.ProxyTypeTCP, consts.ProxyTypeUDP:
		if p.RemotePort == "" {
			p.RemotePort = "0"
		}
		if p.Plugin != "" {
			if p.IsRange() {
				showErrorMessage(pd.Form(), "", i18n.Sprintf("The plugin does not support range ports."))
			} else if port, err := strconv.ParseInt(p.RemotePort, 10, 64); err != nil || port < 0 {
				showErrorMessage(pd.Form(), "", i18n.Sprintf("Invalid remote port."))
			} else {
				break
			}
			return false
		} else {
			localPorts, err := frputil.ParseRangeNumbers(p.LocalPort)
			if err != nil {
				showError(err, pd.Form())
				return false
			}
			remotePorts, err := frputil.ParseRangeNumbers(p.RemotePort)
			if err != nil {
				showError(err, pd.Form())
				return false
			}
			if p.IsRange() && len(localPorts) != len(remotePorts) {
				showErrorMessage(pd.Form(), "", i18n.Sprintf("The number of local ports should be the same as the number of remote ports."))
				return false
			}
		}
	case consts.ProxyTypeTCPMUX, consts.ProxyTypeHTTP, consts.ProxyTypeHTTPS:
		if p.CustomDomains == "" && p.SubDomain == "" {
			showErrorMessage(pd.Form(), "", i18n.Sprintf("Custom domains and subdomain should have at least one of these set."))
			return false
		}
	}
	return true
}
