package ui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/services"
)

type EditClientDialog struct {
	*walk.Dialog

	// Config data
	Conf          *Conf
	data          *config.ClientConfig
	ShouldRestart bool
	Added         bool

	// Views
	logFileView *walk.LineEdit
	nameView    *walk.LineEdit

	// View models
	binder *editClientBinder
	db     *walk.DataBinder
}

// Data binder contains a copy of config
type editClientBinder struct {
	// Name of this config
	Name string
	// Common settings
	config.ClientCommon
}

func NewEditClientDialog(conf *Conf, name string) *EditClientDialog {
	v := new(EditClientDialog)
	if conf == nil {
		newConf := newDefaultClientConfig()
		v.Conf = &Conf{Data: newConf}
	} else {
		v.Conf = conf
	}
	data, ok := v.Conf.Data.(*config.ClientConfig)
	if !ok {
		return nil
	}
	v.data = data
	v.binder = &editClientBinder{
		Name:         v.Conf.Name,
		ClientCommon: v.data.ClientCommon,
	}
	if name != "" {
		v.binder.Name = name
	}
	if v.binder.DeleteAfterDate.IsZero() {
		v.binder.DeleteAfterDate = time.Now().AddDate(0, 0, 1)
	}
	return v
}

func (cd *EditClientDialog) View() Dialog {
	pages := []TabPage{
		cd.basicConfPage(),
		cd.authConfPage(),
		cd.logConfPage(),
		cd.adminConfPage(),
		cd.connectionConfPage(),
		cd.tlsConfPage(),
		cd.advancedConfPage(),
	}
	title := i18n.Sprintf("New Client")
	if cd.Conf.Name != "" {
		title = i18n.Sprintf("Edit Client - %s", cd.Conf.Name)
	}
	dlg := NewBasicDialog(&cd.Dialog, title, loadIcon(res.IconEditDialog, 32), DataBinder{
		AssignTo:   &cd.db,
		Name:       "common",
		DataSource: cd.binder,
	}, cd.onSave,
		TabWidget{
			Pages: pages,
		},
	)
	dlg.Layout = VBox{Margins: Margins{Left: 7, Top: 9, Right: 7, Bottom: 9}}
	minWidth := lo.Sum(lo.Map(pages, func(page TabPage, i int) int {
		return calculateStringWidth(page.Title.(string)) + 19
	})) + 70
	dlg.MinSize = Size{Width: minWidth, Height: 380}
	return dlg
}

func (cd *EditClientDialog) basicConfPage() TabPage {
	return TabPage{
		Title:  i18n.Sprintf("Basic"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Name")},
			LineEdit{AssignTo: &cd.nameView, Text: Bind("Name", res.ValidateNonEmpty), OnTextChanged: func() {
				if name := cd.nameView.Text(); name != "" {
					curLog := strings.TrimSpace(cd.logFileView.Text())
					// Automatically change the log file if it's empty or using the default log directory
					if curLog == "" || strings.HasPrefix(curLog, "logs/") {
						cd.logFileView.SetText("logs" + "/" + name + ".log")
					}
				}
			}},
			Label{Text: i18n.SprintfColon("Server Address")},
			LineEdit{Text: Bind("ServerAddress", res.ValidateNonEmpty)},
			Label{Text: i18n.SprintfColon("Server Port")},
			NewNumberInput(NIOption{Value: Bind("ServerPort"), Max: 65535, Width: 90}),
			Label{Text: i18n.SprintfColon("User")},
			LineEdit{Text: Bind("User")},
			Label{Text: i18n.SprintfColon("STUN Server")},
			LineEdit{Text: Bind("NatHoleSTUNServer")},
			VSpacer{ColumnSpan: 2},
		},
	}
}

func (cd *EditClientDialog) authConfPage() TabPage {
	token := Bind("tokenCheck.Checked")
	oidc := Bind("oidcCheck.Checked")
	auth := Bind("!noAuthCheck.Checked")
	return AlignGrid(TabPage{
		Title:  i18n.Sprintf("Auth"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Auth Method")},
			NewRadioButtonGroup("AuthMethod", nil, nil, []RadioButton{
				{Name: "tokenCheck", Text: "Token", Value: consts.AuthToken},
				{Name: "oidcCheck", Text: "OIDC", Value: consts.AuthOIDC},
				{Name: "noAuthCheck", Text: i18n.Sprintf("None"), Value: ""},
			}),
			Label{Visible: token, Text: i18n.SprintfColon("Token")},
			LineEdit{Visible: token, Text: Bind("Token"), PasswordMode: true},
			Label{Visible: oidc, Text: "ID:"},
			LineEdit{Visible: oidc, Text: Bind("OIDCClientId")},
			Label{Visible: oidc, Text: i18n.SprintfColon("Secret")},
			LineEdit{Visible: oidc, Text: Bind("OIDCClientSecret"), PasswordMode: true},
			Label{Visible: oidc, Text: i18n.SprintfColon("Audience")},
			LineEdit{Visible: oidc, Text: Bind("OIDCAudience")},
			Label{Visible: oidc, Text: i18n.SprintfColon("Scope")},
			LineEdit{Visible: oidc, Text: Bind("OIDCScope")},
			Label{Visible: oidc, Text: i18n.SprintfColon("Token Endpoint")},
			Composite{
				Visible: oidc,
				Layout:  HBox{MarginsZero: true},
				Children: []Widget{
					LineEdit{Text: Bind("OIDCTokenEndpoint")},
					ToolButton{Text: "#", ToolTipText: i18n.Sprintf("Parameters"), OnClicked: func() {
						NewAttributeDialog(i18n.Sprintf("Parameters"), &cd.binder.OIDCAdditionalEndpointParams).Run(cd.Form())
					}},
				},
			},
			Label{Visible: auth, Text: i18n.SprintfColon("Scope")},
			Composite{
				Visible: auth,
				Layout:  HBox{MarginsZero: true},
				Children: []Widget{
					CheckBox{Text: i18n.Sprintf("Heart Beats"), Checked: Bind("AuthenticateHeartBeats")},
					CheckBox{Text: i18n.Sprintf("Work Conns"), Checked: Bind("AuthenticateNewWorkConns")},
				},
			},
		},
	}, 0)
}

func (cd *EditClientDialog) logConfPage() TabPage {
	return TabPage{
		Title:  i18n.Sprintf("Log"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			TextLabel{Text: i18n.Sprintf("* Leave blank to record no log and delete the original log file."), ColumnSpan: 2},
			VSpacer{Size: 2, ColumnSpan: 2},
			Label{Text: i18n.SprintfColon("Log File")},
			NewBrowseLineEdit(&cd.logFileView, true, true, Bind("LogFile"),
				i18n.Sprintf("Select Log File"), res.FilterLog, true),
			Label{Text: i18n.SprintfColon("Level")},
			ComboBox{
				Value: Bind("LogLevel"),
				Model: consts.LogLevels,
			},
			Label{Text: i18n.SprintfColon("Max Days")},
			NewNumberInput(NIOption{Value: Bind("LogMaxDays"), Suffix: i18n.Sprintf("Days"), Max: math.MaxFloat64, Width: 90}),
			VSpacer{ColumnSpan: 2},
		},
	}
}

func (cd *EditClientDialog) adminConfPage() TabPage {
	adminEnabled := Bind("adminPort.Value > 0")
	absChecked := Bind("absCheck.Checked")
	relChecked := Bind("relCheck.Checked")
	return AlignGrid(TabPage{
		Title:  i18n.Sprintf("Admin"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Admin Address")},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					LineEdit{Text: Bind("AdminAddr"), StretchFactor: 2},
					Label{Text: ":"},
					NumberEdit{
						Name:     "adminPort",
						Value:    Bind("AdminPort"),
						MinValue: 0,
						MaxValue: 65535,
						MinSize:  Size{Width: 70},
					},
					ToolButton{
						Enabled:     Bind("adminPort.Value > 0 && !legacyFormat.Checked"),
						Image:       loadIcon(res.IconLock, 16),
						ToolTipText: "TLS", OnClicked: func() {
							cd.adminTLSDialog().Run(cd.Form())
						},
					},
				},
			},
			Label{Enabled: adminEnabled, Text: i18n.SprintfColon("User")},
			LineEdit{Enabled: adminEnabled, Text: Bind("AdminUser")},
			Label{Enabled: adminEnabled, Text: i18n.SprintfColon("Password")},
			LineEdit{Enabled: adminEnabled, Text: Bind("AdminPwd"), PasswordMode: true},
			Label{Enabled: adminEnabled, Text: i18n.SprintfColon("Assets")},
			NewBrowseLineEdit(nil, true, adminEnabled, Bind("AssetsDir"),
				i18n.Sprintf("Select a local directory that the admin server will load resources from."), "", false),
			Label{Enabled: adminEnabled, Text: i18n.SprintfColon("Other Options")},
			CheckBox{Enabled: adminEnabled, Text: "Pprof", Checked: Bind("PprofEnable")},
			Label{Text: i18n.SprintfColon("Auto Delete")},
			NewRadioButtonGroup("DeleteMethod", nil, nil, []RadioButton{
				{Name: "absCheck", Text: i18n.Sprintf("Absolute"), Value: consts.DeleteAbsolute},
				{Name: "relCheck", Text: i18n.Sprintf("Relative"), Value: consts.DeleteRelative},
				{Name: "noDelCheck", Text: i18n.Sprintf("None"), Value: ""},
			}),
			Label{Visible: absChecked, Text: i18n.SprintfColon("Delete Date")},
			DateEdit{Visible: absChecked, Date: Bind("DeleteAfterDate")},
			Label{Visible: relChecked, Text: i18n.SprintfColon("Delete Days")},
			NewNumberInput(NIOption{
				Visible: relChecked,
				Value:   Bind("DeleteAfterDays"),
				Suffix:  i18n.Sprintf("Days"),
				Max:     math.MaxFloat64,
			}),
		},
	}, 0)
}

func (cd *EditClientDialog) connectionConfPage() TabPage {
	expr := func(op, proto string) string {
		return fmt.Sprintf("proto.Value %s '%s'", op, proto)
	}
	quic := Bind(expr("==", consts.ProtoQUIC))
	tcp := Bind(expr("!=", consts.ProtoQUIC))
	second := i18n.Sprintf("s")
	groupMargins := Margins{Left: 9, Top: 9, Right: 9, Bottom: 16}
	return TabPage{
		Title:  i18n.Sprintf("Connection"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Composite{
				Layout:     HBox{MarginsZero: true},
				ColumnSpan: 2,
				Children: []Widget{
					Label{Text: i18n.SprintfColon("Protocol")},
					HSpacer{Size: 8},
					ComboBox{
						Name:    "proto",
						Value:   Bind("Protocol"),
						Model:   consts.Protocols,
						MinSize: Size{Width: 150},
					},
					HSpacer{},
					LinkLabel{Text: "<a>" + i18n.SprintfEllipsis("Advanced Options") + "</a>", OnLinkActivated: func(link *walk.LinkLabelLink) {
						cd.advancedConnDialog().Run(cd.Form())
					}},
				},
			},
			GroupBox{
				Title:      i18n.Sprintf("Parameters"),
				Layout:     Grid{Columns: 2, Spacing: 16, Margins: groupMargins},
				ColumnSpan: 2,
				MaxSize:    Size{Height: 105},
				Children: []Widget{
					NewNumberInput(NIOption{
						Title:   i18n.SprintfColon("Dial Timeout"),
						Value:   Bind("DialServerTimeout"),
						Suffix:  second,
						Visible: tcp,
						Max:     math.MaxFloat64,
					}),
					NewNumberInput(NIOption{
						Title:   i18n.SprintfColon("Keepalive"),
						Value:   Bind("DialServerKeepAlive"),
						Suffix:  second,
						Visible: tcp,
					}),
					NewNumberInput(NIOption{
						Title:   i18n.SprintfColon("Idle Timeout"),
						Value:   Bind("QUICMaxIdleTimeout"),
						Suffix:  second,
						Visible: quic,
						Max:     math.MaxFloat64,
					}),
					NewNumberInput(NIOption{
						Title:   i18n.SprintfColon("Keepalive"),
						Value:   Bind("QUICKeepalivePeriod"),
						Suffix:  second,
						Visible: quic,
						Max:     math.MaxFloat64,
					}),
					NewNumberInput(NIOption{
						Title: i18n.SprintfColon("Pool Count"),
						Value: Bind("PoolCount"),
						Max:   math.MaxFloat64,
					}),
					NewNumberInput(NIOption{
						Title:   i18n.SprintfColon("Max Streams"),
						Value:   Bind("QUICMaxIncomingStreams"),
						Visible: quic,
					}),
				},
			},
			GroupBox{
				Title:      i18n.Sprintf("Heartbeat"),
				Layout:     Grid{Columns: 2, Margins: groupMargins},
				ColumnSpan: 2,
				Children: []Widget{
					NewNumberInput(NIOption{
						Title:  i18n.SprintfColon("Interval"),
						Value:  Bind("HeartbeatInterval"),
						Suffix: second,
					}),
					NewNumberInput(NIOption{
						Title:  i18n.SprintfColon("Timeout"),
						Value:  Bind("HeartbeatTimeout"),
						Suffix: second,
					}),
				},
			},
		},
	}
}

func (cd *EditClientDialog) tlsConfPage() TabPage {
	tlsChecked := Bind("tlsCheck.Checked")
	return TabPage{
		Title:  "TLS",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "TLS:"},
			NewRadioButtonGroup("TLSEnable", nil, nil, []RadioButton{
				{Name: "tlsCheck", Text: i18n.Sprintf("On"), Value: true},
				{Text: i18n.Sprintf("Off"), Value: false},
			}),
			Label{Visible: tlsChecked, Text: i18n.SprintfColon("Host Name"), AlwaysConsumeSpace: true},
			LineEdit{Visible: tlsChecked, Text: Bind("TLSServerName")},
			Label{Visible: tlsChecked, Text: i18n.SprintfColon("Certificate")},
			NewBrowseLineEdit(nil, tlsChecked, true, Bind("TLSCertFile"),
				i18n.Sprintf("Select Certificate File"), res.FilterCert, true),
			Label{Visible: tlsChecked, Text: i18n.SprintfColon("Certificate Key"), AlwaysConsumeSpace: true},
			NewBrowseLineEdit(nil, tlsChecked, true, Bind("TLSKeyFile"),
				i18n.Sprintf("Select Certificate Key File"), res.FilterKey, true),
			Label{Visible: tlsChecked, Text: i18n.SprintfColon("Trusted CA"), AlwaysConsumeSpace: true},
			NewBrowseLineEdit(nil, tlsChecked, true, Bind("TLSTrustedCaFile"),
				i18n.Sprintf("Select Trusted CA File"), res.FilterCert, true),
			Label{Visible: tlsChecked, Text: i18n.SprintfColon("Other Options")},
			CheckBox{Visible: tlsChecked, Text: i18n.Sprintf("Disable custom first byte"), Checked: Bind("DisableCustomTLSFirstByte")},
		},
	}
}

func (cd *EditClientDialog) advancedConfPage() TabPage {
	muxChecked := Bind("muxCheck.Checked")
	var legacy *walk.CheckBox
	return TabPage{
		Title:  i18n.Sprintf("Advanced"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "DNS:"},
			LineEdit{Text: Bind("DNSServer")},
			Label{Text: i18n.SprintfColon("Source Address")},
			LineEdit{Text: Bind("ConnectServerLocalIP")},
			Composite{
				Layout: VBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					VSpacer{Size: 6},
					Label{Text: i18n.SprintfColon("Other Options"), Alignment: AlignHNearVNear},
				},
			},
			Composite{
				Layout: VBox{MarginsZero: true, SpacingZero: true, Alignment: AlignHNearVNear},
				Children: []Widget{
					Composite{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							CheckBox{Name: "muxCheck", Text: i18n.Sprintf("TCP Mux"), Checked: Bind("TCPMux")},
							HSpacer{},
							Label{Enabled: muxChecked, Text: i18n.SprintfColon("Heartbeat")},
							NumberEdit{
								Enabled:            muxChecked,
								Value:              Bind("TCPMuxKeepaliveInterval"),
								MinValue:           0,
								MaxValue:           math.MaxFloat64,
								SpinButtonsVisible: true,
								MinSize:            Size{Width: 85},
								Style:              win.ES_RIGHT,
							},
							Label{Enabled: muxChecked, Text: i18n.Sprintf("s")},
						},
					},
					CheckBox{Text: i18n.Sprintf("Exit after login failure"), Checked: Bind("LoginFailExit")},
					CheckBox{Text: i18n.Sprintf("Disable auto-start at boot"), Checked: Bind("ManualStart")},
					CheckBox{
						AssignTo: &legacy,
						Name:     "legacyFormat",
						Text:     i18n.Sprintf("Use legacy format config file"),
						Checked:  Bind("LegacyFormat"),
						OnCheckedChanged: func() {
							if !legacy.Checked() && !cd.canUpgradeFormat() {
								legacy.SetChecked(true)
							}
						},
					},
					VSpacer{Size: 4},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 18},
						Children: []Widget{
							LinkLabel{
								Text: fmt.Sprintf("<a>%s</a>", i18n.SprintfEllipsis("Metadata")),
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									NewAttributeDialog(i18n.Sprintf("Metadata"), &cd.binder.Metas).Run(cd.Form())
								},
							},
							LinkLabel{
								Text: fmt.Sprintf("<a>%s</a>", i18n.SprintfEllipsis("Experimental Features")),
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									cd.experimentDialog().Run(cd.Form())
								},
							},
						},
					},
				},
			},
		},
	}
}

func (cd *EditClientDialog) experimentDialog() Dialog {
	dlg := NewBasicDialog(nil, i18n.Sprintf("Experimental Features"),
		loadIcon(res.IconExperiment, 32), DataBinder{DataSource: cd.binder}, nil,
		Label{Text: i18n.Sprintf("* The following features may affect the stability of the service.")},
		CheckBox{Checked: Bind("SVCBEnable"), Text: i18n.Sprintf("Use server SVCB records"), Alignment: AlignHNearVNear},
		VSpacer{},
	)
	dlg.MinSize = Size{Width: 300, Height: 180}
	dlg.FixedSize = true
	return dlg
}

func (cd *EditClientDialog) adminTLSDialog() Dialog {
	var widgets [4]*walk.LineEdit
	dlg := NewBasicDialog(nil, "TLS",
		loadIcon(res.IconLock, 32),
		DataBinder{DataSource: &cd.binder.AdminTLS}, nil,
		Label{Text: i18n.SprintfColon("Host Name")},
		LineEdit{AssignTo: &widgets[0], Text: Bind("ServerName")},
		Label{Text: i18n.SprintfColon("Certificate")},
		NewBrowseLineEdit(&widgets[1], true, true, Bind("CertFile"),
			i18n.Sprintf("Select Certificate File"), res.FilterCert, true),
		Label{Text: i18n.SprintfColon("Certificate Key")},
		NewBrowseLineEdit(&widgets[2], true, true, Bind("KeyFile"),
			i18n.Sprintf("Select Certificate Key File"), res.FilterKey, true),
		Label{Text: i18n.SprintfColon("Trusted CA")},
		NewBrowseLineEdit(&widgets[3], true, true, Bind("TrustedCaFile"),
			i18n.Sprintf("Select Trusted CA File"), res.FilterCert, true),
		VSpacer{Size: 4},
	)
	dlg.MinSize = Size{Width: 350}
	dlg.FixedSize = true
	buttons := dlg.Children[len(dlg.Children)-1].(Composite)
	buttons.Children = append([]Widget{PushButton{Text: i18n.Sprintf("Clear All"), OnClicked: func() {
		for _, widget := range widgets {
			widget.SetText("")
		}
	}}}, buttons.Children...)
	dlg.Children[len(dlg.Children)-1] = buttons
	return dlg
}

func (cd *EditClientDialog) advancedConnDialog() Dialog {
	dlg := NewBasicDialog(nil, i18n.Sprintf("Advanced Options"),
		loadIcon(res.IconEditDialog, 32),
		DataBinder{DataSource: cd.binder}, nil,
		Label{Text: i18n.SprintfColon("HTTP Proxy")},
		LineEdit{Text: Bind("HTTPProxy")},
		Label{Text: i18n.SprintfColon("UDP Packet Size")},
		NewNumberInput(NIOption{Value: Bind("UDPPacketSize"), Max: math.MaxFloat64, Width: 90}),
		VSpacer{Size: 4},
	)
	dlg.MinSize = Size{Width: 350}
	dlg.FixedSize = true
	return dlg
}

func (cd *EditClientDialog) shutdownService(wait bool) error {
	if !cd.ShouldRestart {
		cd.ShouldRestart = cd.Conf.State == consts.StateStarted
	}
	return services.UninstallService(cd.Conf.Name, wait)
}

func (cd *EditClientDialog) onSave() {
	if err := cd.db.Submit(); err != nil {
		return
	}
	newConf := cd.binder
	cd.ShouldRestart = false
	// Edit existing config
	if cd.Conf.Name != "" {
		if !ensureExistingConfig(cd.Conf.Name, cd.Form()) {
			cd.Cancel()
			return
		}
		// Change config name
		if newConf.Name != cd.Conf.Name {
			if cd.hasConf(newConf.Name) {
				return
			}
			// Delete old service
			// We should start the new config if the old one is already started
			if err := cd.shutdownService(false); err != nil && cd.ShouldRestart {
				showError(err, cd.Form())
				return
			}
			// Delete old config file
			if err := os.Remove(cd.Conf.Path); err != nil {
				showError(err, cd.Form())
				return
			}
		}
		// Change log files
		if newConf.LogFile != cd.data.LogFile &&
			!(newConf.LogFile == "console" && cd.data.LogFile == "") &&
			!(newConf.LogFile == "" && cd.data.LogFile == "console") {
			// Rename or remove log files
			logs, dates, err := util.FindLogFiles(cd.data.LogFile)
			if newConf.LogFile == "" || newConf.LogFile == "console" {
				// Remove old log files
				// The service should be stopped first
				cd.shutdownService(true)
				util.DeleteFiles(logs)
			} else if cd.data.LogFile != "" && cd.data.LogFile != "console" && err == nil {
				baseName, ext := util.SplitExt(newConf.LogFile)
				// Rename old log files
				// The service should be stopped first
				cd.shutdownService(true)
				util.RenameFiles(logs, lo.Map(dates, func(item time.Time, i int) string {
					if item.IsZero() {
						return newConf.LogFile
					} else {
						return filepath.Join(filepath.Dir(newConf.LogFile), baseName+"."+item.Format("20060102-150405")+ext)
					}
				}))
			}
		}
	} else if cd.hasConf(newConf.Name) {
		return
	} else {
		// For new config
		addConf(cd.Conf)
		cd.Added = true
	}
	cd.Conf.Name = newConf.Name
	cd.data.ClientCommon = newConf.ClientCommon
	cd.Accept()
}

func (cd *EditClientDialog) hasConf(name string) bool {
	if hasConf(name) {
		showWarningMessage(cd.Form(), i18n.Sprintf("Config already exists"), i18n.Sprintf("The config name \"%s\" already exists.", name))
		return true
	}
	return false
}

func (cd *EditClientDialog) Run(owner walk.Form) (int, error) {
	return cd.View().Run(owner)
}

func (cd *EditClientDialog) canUpgradeFormat() bool {
	for _, v := range cd.data.Proxies {
		if !v.IsVisitor() {
			if _, err := config.ClientProxyToV1(v); err != nil {
				showErrorMessage(cd.Form(), "", i18n.Sprintf("Unable to upgrade your config file due to proxy conversion failure, "+
					"please check the proxy config and try again.\n\nBad proxy: %s", v.Name))
				return false
			}
		}
	}
	return true
}
