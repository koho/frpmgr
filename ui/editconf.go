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

func NewEditClientDialog(conf *Conf) *EditClientDialog {
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
	return v
}

func (cd *EditClientDialog) View() Dialog {
	dlg := NewBasicDialog(&cd.Dialog, "编辑配置", loadSysIcon("imageres", consts.IconEditDialog, 32), DataBinder{
		AssignTo:   &cd.db,
		Name:       "common",
		DataSource: cd.binder,
	}, cd.onSave,
		TabWidget{
			Pages: []TabPage{
				cd.baseConfPage(),
				cd.authConfPage(),
				cd.logConfPage(),
				cd.adminConfPage(),
				cd.connectionConfPage(),
				cd.tlsConfPage(),
				cd.advancedConfPage(),
			},
		},
	)
	dlg.Layout = VBox{Margins: Margins{7, 9, 7, 9}}
	return dlg
}

func (cd *EditClientDialog) baseConfPage() TabPage {
	return TabPage{
		Title:  "基本",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "名称:"},
			LineEdit{AssignTo: &cd.nameView, Text: Bind("Name", consts.ValidateNonEmpty), OnTextChanged: func() {
				if name := cd.nameView.Text(); name != "" {
					curLog := strings.TrimSpace(cd.logFileView.Text())
					// Automatically change the log file if it's empty or using the default log directory
					if curLog == "" || strings.HasPrefix(curLog, "logs/") {
						cd.logFileView.SetText("logs" + "/" + name + ".log")
					}
				}
			}},
			Label{Text: "服务器地址:"},
			LineEdit{Text: Bind("ServerAddress", consts.ValidateNonEmpty)},
			Label{Text: "服务器端口:"},
			LineEdit{Text: Bind("ServerPort", consts.ValidateRequireInteger)},
			Label{Text: "用户:"},
			LineEdit{Text: Bind("User")},
			VSpacer{ColumnSpan: 2},
		},
	}
}

func (cd *EditClientDialog) authConfPage() TabPage {
	return TabPage{
		Title:  "认证",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "认证方式:", MinSize: Size{Width: 55}},
			NewRadioButtonGroup("AuthMethod", nil, []RadioButton{
				{Name: "tokenCheck", Text: "Token", Value: consts.AuthToken},
				{Name: "oidcCheck", Text: "OIDC", Value: consts.AuthOIDC},
				{Name: "noAuthCheck", Text: "无", Value: ""},
			}),
			Label{Visible: Bind("tokenCheck.Checked"), Text: "令牌:"},
			LineEdit{Visible: Bind("tokenCheck.Checked"), Text: Bind("Token")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: "ID:"},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCClientId")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: "密钥:"},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCClientSecret")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: "接受者:"},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCAudience")},
			Label{Visible: Bind("oidcCheck.Checked"), Text: "令牌地址:"},
			LineEdit{Visible: Bind("oidcCheck.Checked"), Text: Bind("OIDCTokenEndpoint")},
			Label{Visible: Bind("!noAuthCheck.Checked"), Text: "鉴权:"},
			Composite{
				Visible: Bind("!noAuthCheck.Checked"),
				Layout:  HBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					CheckBox{Text: "心跳消息", Checked: Bind("AuthenticateHeartBeats")},
					CheckBox{Text: "工作连接", Checked: Bind("AuthenticateNewWorkConns")},
				},
			},
		},
	}
}

func (cd *EditClientDialog) logConfPage() TabPage {
	return TabPage{
		Title:  "日志",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "*留空则不记录日志，且删除原来的日志文件", ColumnSpan: 2},
			Label{Text: "日志文件:"},
			NewBrowseLineEdit(&cd.logFileView, true, true, Bind("LogFile"),
				"选择日志文件", "日志文件 (*.log, *.txt)|*.log;*.txt|", true),
			Label{Text: "级别:"},
			ComboBox{
				Value: Bind("LogLevel"),
				Model: []string{"trace", "debug", "info", "warn", "error"},
			},
			Label{Text: "最大天数:"},
			NumberEdit{Value: Bind("LogMaxDays")},
		},
	}
}

func (cd *EditClientDialog) adminConfPage() TabPage {
	return TabPage{
		Title:  "管理",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "管理地址:"},
			LineEdit{Text: Bind("AdminAddr")},
			Label{Text: "管理端口:"},
			LineEdit{Name: "adminPort", Text: Bind("AdminPort", consts.ValidateInteger)},
			Label{Enabled: Bind("adminPort.Text != ''"), Text: "用户名:"},
			LineEdit{Enabled: Bind("adminPort.Text != ''"), Text: Bind("AdminUser")},
			Label{Enabled: Bind("adminPort.Text != ''"), Text: "密码:"},
			LineEdit{Enabled: Bind("adminPort.Text != ''"), Text: Bind("AdminPwd")},
			Label{Enabled: Bind("adminPort.Text != ''"), Text: "静态资源:"},
			NewBrowseLineEdit(nil, true, Bind("adminPort.Text != ''"), Bind("AssetsDir"),
				"选择静态资源目录", "", false),
			Label{Enabled: Bind("adminPort.Text != ''"), Text: "调试:"},
			CheckBox{Text: "pprof", Checked: Bind("PprofEnable"), Enabled: Bind("adminPort.Text != ''")},
		},
	}
}

func (cd *EditClientDialog) connectionConfPage() TabPage {
	return TabPage{
		Title:  "连接",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "协议:"},
			ComboBox{
				Value: Bind("Protocol"),
				Model: []string{"tcp", "kcp", "websocket"},
			},
			Label{Text: "HTTP 代理:"},
			LineEdit{Text: Bind("HTTPProxy")},
			Label{Text: "连接池数量:"},
			NumberEdit{Value: Bind("PoolCount")},
			Label{Text: "连接超时:"},
			NumberEdit{Value: Bind("DialServerTimeout"), Suffix: " 秒"},
			Label{Text: "TCP 保活周期:"},
			NumberEdit{Value: Bind("DialServerKeepAlive"), Suffix: " 秒"},
			Label{Text: "心跳间隔:"},
			NumberEdit{Value: Bind("HeartbeatInterval"), Suffix: " 秒"},
			Label{Text: "心跳超时:"},
			NumberEdit{Value: Bind("HeartbeatTimeout"), Suffix: " 秒"},
		},
	}
}

func (cd *EditClientDialog) tlsConfPage() TabPage {
	return TabPage{
		Title:  "TLS",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "TLS:", MinSize: Size{Width: 65}},
			NewRadioButtonGroup("TLSEnable", nil, []RadioButton{
				{Name: "tlsCheck", Text: "开启", Value: true},
				{Text: "关闭", Value: false},
			}),
			Label{Visible: Bind("tlsCheck.Checked"), Text: "主机名称:"},
			LineEdit{Visible: Bind("tlsCheck.Checked"), Text: Bind("TLSServerName")},
			Label{Visible: Bind("tlsCheck.Checked"), Text: "证书文件:"},
			NewBrowseLineEdit(nil, Bind("tlsCheck.Checked"), true, Bind("TLSCertFile"),
				"选择证书文件", consts.FilterCert, true),
			Label{Visible: Bind("tlsCheck.Checked"), Text: "密钥文件:"},
			NewBrowseLineEdit(nil, Bind("tlsCheck.Checked"), true, Bind("TLSKeyFile"),
				"选择密钥文件", consts.FilterKey, true),
			Label{Visible: Bind("tlsCheck.Checked"), Text: "受信任证书:"},
			NewBrowseLineEdit(nil, Bind("tlsCheck.Checked"), true, Bind("TLSTrustedCaFile"),
				"选择受信任的证书", consts.FilterCert, true),
		},
	}
}

func (cd *EditClientDialog) advancedConfPage() TabPage {
	return TabPage{
		Title:  "高级",
		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "多路复用:", MinSize: Size{Width: 65}},
			NewRadioButtonGroup("TCPMux", nil, []RadioButton{
				{Name: "muxCheck", Text: "开启", Value: true},
				{Text: "关闭", Value: false},
			}),
			Label{Enabled: Bind("muxCheck.Checked"), Text: "复用器心跳:"},
			NumberEdit{Enabled: Bind("muxCheck.Checked"), Value: Bind("TCPMuxKeepaliveInterval"), Suffix: " 秒"},
			Label{Text: "DNS:"},
			LineEdit{Text: Bind("DNSServer")},
			Label{Text: "使用源地址:"},
			LineEdit{Text: Bind("ConnectServerLocalIP")},
			Composite{
				Layout: VBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					VSpacer{Size: 6},
					Label{Text: "其他选项:", Alignment: AlignHNearVNear},
				},
			},
			Composite{
				Layout: VBox{MarginsZero: true, SpacingZero: true, Alignment: AlignHNearVNear},
				Children: []Widget{
					CheckBox{Text: "初次登录失败后退出", Checked: Bind("LoginFailExit")},
					CheckBox{Text: "禁用开机自启动", Checked: Bind("ManualStart")},
					VSpacer{Size: 4},
					LinkLabel{Text: "<a>自定义...</a>", OnLinkActivated: func(link *walk.LinkLabelLink) {
						cd.customConfDialog().Run(cd.Form())
					}},
				},
			},
		},
	}
}

func (cd *EditClientDialog) customConfDialog() Dialog {
	customDialog := NewBasicDialog(nil, "自定义参数", cd.Icon(), DataBinder{DataSource: cd.binder}, nil,
		Label{Text: "*参考 FRP 配置文件的 [common] 部分，每行格式为 a = b"},
		TextEdit{Text: Bind("CustomText"), VScroll: true},
	)
	customDialog.MinSize = Size{380, 280}
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
		showWarningMessage(cd.Form(), "配置已存在", fmt.Sprintf("配置名「%s」已存在。", name))
		return true
	}
	return false
}

func (cd *EditClientDialog) Run(owner walk.Form) (int, error) {
	return cd.View().Run(owner)
}
