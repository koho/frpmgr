package ui

import (
	"fmt"
	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/services"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/thoas/go-funk"
	"os"
	"path/filepath"
	"strings"
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
	// CustomText contains the user-defined parameters
	CustomText string
	// Common settings
	config.ClientCommon
}

func NewEditClientDialog(conf *Conf, name string) *EditClientDialog {
	v := new(EditClientDialog)
	if conf == nil {
		newConf := config.NewDefaultClientConfig()
		newConf.AuthMethod = ""
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
		CustomText:   util.Map2String(data.Custom),
		ClientCommon: v.data.ClientCommon,
	}
	if name != "" {
		v.binder.Name = name
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
	dlg := NewBasicDialog(&cd.Dialog, i18n.Sprintf("Edit Client"), loadSysIcon("imageres", consts.IconEditDialog, 32), DataBinder{
		AssignTo:   &cd.db,
		Name:       "common",
		DataSource: cd.binder,
	}, cd.onSave,
		TabWidget{
			Pages: pages,
		},
	)
	dlg.Layout = VBox{Margins: Margins{7, 9, 7, 9}}
	minWidth := int(funk.Sum(funk.Map(pages, func(page TabPage) int {
		return calculateStringWidth(page.Title.(string)) + 19
	})) + 70)
	dlg.MinSize = Size{Width: minWidth, Height: 380}
	return dlg
}

func (cd *EditClientDialog) basicConfPage() TabPage {
	return TabPage{
		Title:  i18n.Sprintf("Basic"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Name")},
			LineEdit{AssignTo: &cd.nameView, Text: Bind("Name", consts.ValidateNonEmpty), OnTextChanged: func() {
				if name := cd.nameView.Text(); name != "" {
					curLog := strings.TrimSpace(cd.logFileView.Text())
					// Automatically change the log file if it's empty or using the default log directory
					if curLog == "" || strings.HasPrefix(curLog, "logs/") {
						cd.logFileView.SetText("logs" + "/" + name + ".log")
					}
				}
			}},
			Label{Text: i18n.SprintfColon("Server Address")},
			LineEdit{Text: Bind("ServerAddress", consts.ValidateNonEmpty)},
			Label{Text: i18n.SprintfColon("Server Port")},
			LineEdit{Text: Bind("ServerPort", consts.ValidateRequireInteger)},
			Label{Text: i18n.SprintfColon("User")},
			LineEdit{Text: Bind("User")},
			VSpacer{ColumnSpan: 2},
		},
	}
}

func (cd *EditClientDialog) authConfPage() TabPage {
	return AlignGrid(TabPage{
		Title:  i18n.Sprintf("Auth"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Auth Method")},
			NewRadioButtonGroup("AuthMethod", nil, []RadioButton{
				{Name: "tokenCheck", Text: "Token", Value: consts.AuthToken},
				{Name: "oidcCheck", Text: "OIDC", Value: consts.AuthOIDC},
				{Name: "noAuthCheck", Text: i18n.Sprintf("None"), Value: ""},
			}),
			Label{Visible: Bind("tokenCheck.Checked"), Text: i18n.SprintfColon("Token")},
			LineEdit{Visible: Bind("tokenCheck.Checked"), Text: Bind("Token")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: "ID:"},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCClientId")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: i18n.SprintfColon("Secret")},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCClientSecret")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: i18n.SprintfColon("Audience")},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCAudience")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: i18n.SprintfColon("Scope")},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCScope")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: i18n.SprintfColon("Token Endpoint")},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCTokenEndpoint")},
			Label{Visible: Bind("!noAuthCheck.Checked"), Text: i18n.SprintfColon("Authentication")},
			Composite{
				Visible: Bind("!noAuthCheck.Checked"),
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
				i18n.Sprintf("Select Log File"), consts.FilterLog, true),
			Label{Text: i18n.SprintfColon("Level")},
			ComboBox{
				Value: Bind("LogLevel"),
				Model: []string{"trace", "debug", "info", "warn", "error"},
			},
			Label{Text: i18n.SprintfColon("Max Days")},
			NumberEdit{Value: Bind("LogMaxDays")},
			VSpacer{ColumnSpan: 2},
		},
	}
}

func (cd *EditClientDialog) adminConfPage() TabPage {
	return TabPage{
		Title:  i18n.Sprintf("Admin"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Admin Address")},
			LineEdit{Text: Bind("AdminAddr")},
			Label{Text: i18n.SprintfColon("Admin Port")},
			LineEdit{Name: "adminPort", Text: Bind("AdminPort", consts.ValidateInteger)},
			Label{Enabled: Bind("adminPort.Text != ''"), Text: i18n.SprintfColon("User")},
			LineEdit{Enabled: Bind("adminPort.Text != ''"), Text: Bind("AdminUser")},
			Label{Enabled: Bind("adminPort.Text != ''"), Text: i18n.SprintfColon("Password")},
			LineEdit{Enabled: Bind("adminPort.Text != ''"), Text: Bind("AdminPwd")},
			Label{Enabled: Bind("adminPort.Text != ''"), Text: i18n.SprintfColon("Assets")},
			NewBrowseLineEdit(nil, true, Bind("adminPort.Text != ''"), Bind("AssetsDir"),
				i18n.Sprintf("Select a local directory that the admin server will load resources from."), "", false),
			Label{Enabled: Bind("adminPort.Text != ''"), Text: i18n.SprintfColon("Debug")},
			CheckBox{Text: "pprof", Checked: Bind("PprofEnable"), Enabled: Bind("adminPort.Text != ''")},
		},
	}
}

func (cd *EditClientDialog) connectionConfPage() TabPage {
	expr := func(op, proto string) string {
		return fmt.Sprintf("proto.Value %s '%s'", op, proto)
	}
	quic := Bind(expr("==", consts.ProtoQUIC))
	tcp := Bind(expr("!=", consts.ProtoQUIC))
	return AlignGrid(TabPage{
		Title:  i18n.Sprintf("Connection"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Protocol")},
			ComboBox{
				Name:  "proto",
				Value: Bind("Protocol"),
				Model: []string{consts.ProtoTCP, consts.ProtoKCP, consts.ProtoQUIC, consts.ProtoWebsocket},
			},
			Label{Text: i18n.SprintfColon("HTTP Proxy")},
			LineEdit{Text: Bind("HTTPProxy")},
			Label{Text: i18n.SprintfColon("Pool Count")},
			NumberEdit{Value: Bind("PoolCount")},
			Label{Text: i18n.SprintfColon("Heartbeat")},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					NumberEdit{
						Value:  Bind("HeartbeatInterval"),
						Prefix: i18n.SprintfRSpace("Interval"),
						Suffix: i18n.SprintfLSpace("s"),
					},
					NumberEdit{
						Value:  Bind("HeartbeatTimeout"),
						Prefix: i18n.SprintfRSpace("Timeout"),
						Suffix: i18n.SprintfLSpace("s"),
					},
				},
			},
			Label{Visible: tcp, Text: i18n.SprintfColon("Dial Timeout")},
			NumberEdit{Visible: tcp, Value: Bind("DialServerTimeout"), Suffix: i18n.SprintfLSpace("s")},
			Label{Visible: tcp, Text: i18n.SprintfColon("Keepalive")},
			NumberEdit{Visible: tcp, Value: Bind("DialServerKeepAlive"), Suffix: i18n.SprintfLSpace("s")},
			Label{Visible: quic, Text: i18n.SprintfColon("Idle Timeout")},
			NumberEdit{Visible: quic, Value: Bind("QUICMaxIdleTimeout"), Suffix: i18n.SprintfLSpace("s")},
			Label{Visible: quic, Text: i18n.SprintfColon("Keepalive")},
			NumberEdit{Visible: quic, Value: Bind("QUICKeepalivePeriod"), Suffix: i18n.SprintfLSpace("s")},
			Label{Visible: quic, Text: i18n.SprintfColon("Max Streams")},
			NumberEdit{Visible: quic, Value: Bind("QUICMaxIncomingStreams")},
		},
	}, 0)
}

func (cd *EditClientDialog) tlsConfPage() TabPage {
	return TabPage{
		Title:  "TLS",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "TLS:"},
			NewRadioButtonGroup("TLSEnable", nil, []RadioButton{
				{Name: "tlsCheck", Text: i18n.Sprintf("On"), Value: true},
				{Text: i18n.Sprintf("Off"), Value: false},
			}),
			Label{Visible: Bind("tlsCheck.Checked"), Text: i18n.SprintfColon("Host Name"), AlwaysConsumeSpace: true},
			LineEdit{Visible: Bind("tlsCheck.Checked"), Text: Bind("TLSServerName")},
			Label{Visible: Bind("tlsCheck.Checked"), Text: i18n.SprintfColon("Certificate")},
			NewBrowseLineEdit(nil, Bind("tlsCheck.Checked"), true, Bind("TLSCertFile"),
				i18n.Sprintf("Select Certificate File"), consts.FilterCert, true),
			Label{Visible: Bind("tlsCheck.Checked"), Text: i18n.SprintfColon("Certificate Key"), AlwaysConsumeSpace: true},
			NewBrowseLineEdit(nil, Bind("tlsCheck.Checked"), true, Bind("TLSKeyFile"),
				i18n.Sprintf("Select Certificate Key File"), consts.FilterKey, true),
			Label{Visible: Bind("tlsCheck.Checked"), Text: i18n.SprintfColon("Trusted CA"), AlwaysConsumeSpace: true},
			NewBrowseLineEdit(nil, Bind("tlsCheck.Checked"), true, Bind("TLSTrustedCaFile"),
				i18n.Sprintf("Select Trusted CA File"), consts.FilterCert, true),
		},
	}
}

func (cd *EditClientDialog) advancedConfPage() TabPage {
	return TabPage{
		Title:  i18n.Sprintf("Advanced"),
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("TCP Mux")},
			NewRadioButtonGroup("TCPMux", nil, []RadioButton{
				{Name: "muxCheck", Text: i18n.Sprintf("On"), Value: true},
				{Text: i18n.Sprintf("Off"), Value: false},
			}),
			Label{Enabled: Bind("muxCheck.Checked"), Text: i18n.SprintfColon("Mux Keepalive")},
			NumberEdit{Enabled: Bind("muxCheck.Checked"), Value: Bind("TCPMuxKeepaliveInterval"), Suffix: i18n.SprintfLSpace("s")},
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
					CheckBox{Text: i18n.Sprintf("Exit after login failure"), Checked: Bind("LoginFailExit")},
					CheckBox{Text: i18n.Sprintf("Disable auto-start at boot"), Checked: Bind("ManualStart")},
					VSpacer{Size: 4},
					LinkLabel{Text: fmt.Sprintf("<a>%s</a>", i18n.SprintfEllipsis("Custom")), OnLinkActivated: func(link *walk.LinkLabelLink) {
						cd.customConfDialog().Run(cd.Form())
					}},
				},
			},
		},
	}
}

func (cd *EditClientDialog) customConfDialog() Dialog {
	customDialog := NewBasicDialog(nil, i18n.Sprintf("Custom Options"), cd.Icon(), DataBinder{DataSource: cd.binder}, nil,
		Label{Text: i18n.Sprintf("* Refer to the [common] section of the FRP configuration file.")},
		TextEdit{Text: Bind("CustomText"), VScroll: true},
	)
	customDialog.MinSize = Size{420, 280}
	return customDialog
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
				util.RenameFiles(logs, funk.Map(funk.Zip(logs, dates), func(t funk.Tuple) string {
					if t.Element2 == "" {
						return newConf.LogFile
					} else {
						return filepath.Join(filepath.Dir(newConf.LogFile), baseName+"."+t.Element2.(string)+ext)
					}
				}).([]string))
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
	// The order matters
	cd.data.ClientCommon = newConf.ClientCommon
	cd.data.Custom = util.String2Map(newConf.CustomText)
	cd.Accept()
}

func (cd *EditClientDialog) hasConf(name string) bool {
	if funk.Contains(confList, func(e *Conf) bool { return e.Name == name }) {
		showWarningMessage(cd.Form(), i18n.Sprintf("Config already exists"), i18n.Sprintf("The config name \"%s\" already exists.", name))
		return true
	}
	return false
}

func (cd *EditClientDialog) Run(owner walk.Form) (int, error) {
	return cd.View().Run(owner)
}
