package ui

import (
	"github.com/koho/frpmgr/config"
	"github.com/koho/frpmgr/services"
	"github.com/koho/frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
)

var lastEditName string
var lastRunningState bool

type EditConfDialog struct {
	view *walk.Dialog

	conf            *config.Config
	nameList        []string
	originalName    string
	originalLogFile string
	authInfo        config.AuthInfo
}

func NewEditConfDialog(conf *config.Config, nameList []string) *EditConfDialog {
	v := new(EditConfDialog)
	v.nameList = nameList
	if conf == nil {
		conf = new(config.Config)
		conf.ServerPort = "7000"
		conf.LogLevel = "info"
		conf.LogMaxDays = 3
	}
	if conf.Token != "" {
		conf.AuthMethod = "token"
	}
	v.conf = conf
	v.originalName = conf.Name
	v.originalLogFile = conf.LogFile
	v.authInfo = conf.Common.AuthInfo
	return v
}

func (t *EditConfDialog) View() Dialog {
	var acceptPB, cancelPB *walk.PushButton
	var logFileView *walk.LineEdit
	var customText *walk.TextEdit
	var nameView, serverAddrView, serverPortView *walk.LineEdit
	var db *walk.DataBinder
	var authDB *walk.DataBinder
	changeAuthMethod := func() {
		authDB.Submit()
		authDB.Reset()
	}
	icon, _ := loadLogoIcon(32)
	return Dialog{
		Icon:          icon,
		AssignTo:      &t.view,
		Title:         "编辑配置",
		MinSize:       Size{400, 300},
		Size:          Size{400, 400},
		Layout:        VBox{Margins: Margins{7, 9, 7, 9}},
		Font:          Font{Family: "微软雅黑", PointSize: 9},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder: DataBinder{
			AssignTo:   &db,
			Name:       "common",
			DataSource: t.conf,
		},
		Children: []Widget{
			TabWidget{
				Pages: []TabPage{
					{
						Title:  "基本",
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Label{Text: "名称:"},
							LineEdit{AssignTo: &nameView, Text: Bind("Name", Regexp{".+"}), OnTextChanged: func() {
								logFileView.SetText("logs" + "/" + nameView.Text() + ".log")
							}},
							Label{Text: "服务器地址:"},
							LineEdit{AssignTo: &serverAddrView, Text: Bind("ServerAddress", Regexp{".+"})},
							Label{Text: "服务器端口:"},
							LineEdit{AssignTo: &serverPortView, Text: Bind("ServerPort", Regexp{"^\\d+$"})},
							Label{Text: "用户:"},
							LineEdit{Text: Bind("User")},
							VSpacer{ColumnSpan: 2},
						},
					},
					{
						Title:  "认证",
						Layout: Grid{Columns: 2},
						DataBinder: DataBinder{
							AssignTo:   &authDB,
							Name:       "auth",
							DataSource: &t.authInfo,
						},
						Children: []Widget{
							Label{Text: "认证方式:"},
							Composite{
								Layout: HBox{MarginsZero: true, SpacingZero: true},
								Children: []Widget{
									RadioButtonGroup{
										DataMember: "AuthMethod",
										Buttons: []RadioButton{
											{Text: "Token", Value: "token", OnClicked: changeAuthMethod},
											{Text: "OIDC", Value: "oidc", OnClicked: changeAuthMethod},
											{Text: "无", Value: "", OnClicked: changeAuthMethod},
										},
									},
									HSpacer{},
								},
							},
							Label{Text: "令牌:", Visible: Bind("auth.AuthMethod == 'token'")},
							LineEdit{Text: Bind("Token"), Visible: Bind("auth.AuthMethod == 'token'")},
							Composite{
								Visible:    Bind("auth.AuthMethod == 'oidc'"),
								Layout:     Grid{Columns: 2, MarginsZero: true},
								RowSpan:    4,
								ColumnSpan: 2,
								Children: []Widget{
									Label{Text: "ID:"},
									LineEdit{Text: Bind("OIDCClientId")},
									Label{Text: "密钥:"},
									LineEdit{Text: Bind("OIDCClientSecret")},
									Label{Text: "接受者:"},
									LineEdit{Text: Bind("OIDCAudience")},
									Label{Text: "令牌地址:"},
									LineEdit{Text: Bind("OIDCTokenEndpoint")},
								},
							},
						},
					},
					{
						Title:  "日志",
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Label{Text: "日志文件:"},
							LineEdit{AssignTo: &logFileView, Text: Bind("LogFile")},
							Label{Text: "级别:"},
							ComboBox{
								Value: Bind("LogLevel"),
								Model: []string{"trace", "debug", "info", "warn", "error"},
							},
							Label{Text: "最大天数:"},
							NumberEdit{Value: Bind("LogMaxDays")},
						},
					},
					{
						Title:  "管理",
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Label{Text: "管理地址:"},
							LineEdit{Text: Bind("AdminAddr")},
							Label{Text: "管理端口:"},
							LineEdit{Text: Bind("AdminPort", Regexp{"^\\d*$"})},
							Label{Text: "用户名:"},
							LineEdit{Text: Bind("AdminUser")},
							Label{Text: "密码:"},
							LineEdit{Text: Bind("AdminPwd")},
						},
					},
					{
						Title:  "高级",
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Label{Text: "协议:"},
							ComboBox{
								Value: Bind("Protocol"),
								Model: []string{"tcp", "kcp", "websocket"},
							},
							Label{Text: "网络:", Alignment: AlignHNearVNear},
							Composite{
								Layout: VBox{MarginsZero: true, SpacingZero: true},
								Children: []Widget{
									CheckBox{Text: "TLS", Checked: Bind("TLSEnable")},
									CheckBox{Text: "多路复用", Checked: Bind("TcpMux")},
								},
							},
							Label{Text: "HTTP 代理:"},
							LineEdit{Text: Bind("HTTPProxy")},
							Label{Text: "使用源地址:"},
							LineEdit{Text: Bind("ConnectServerLocalIP")},
							Label{Text: "连接池数量:"},
							NumberEdit{Value: Bind("PoolCount")},
							Label{Text: "DNS:"},
							LineEdit{Text: Bind("DNSServer")},
							Label{Text: "其他:", Alignment: AlignHNearVNear},
							Composite{
								Layout: VBox{MarginsZero: true, SpacingZero: true, Alignment: AlignHNearVNear},
								Children: []Widget{
									CheckBox{Text: "初次登录失败后退出", Checked: Bind("LoginFailExit")},
									CheckBox{Text: "禁用开机自启动", Checked: Bind("ManualStart")},
								},
							},
						},
					},
					{
						Title:  "自定义",
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "*参考 FRP 配置文件的 [common] 部分"},
							TextEdit{AssignTo: &customText, Text: utils.Map2String(t.conf.Custom), VScroll: true},
						},
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: "确定", AssignTo: &acceptPB, OnClicked: func() {
						if nameView.Text() == "" || serverAddrView.Text() == "" || serverPortView.Text() == "" {
							return
						}
						if _, found := utils.Find(t.nameList, nameView.Text()); found && nameView.Text() != t.originalName {
							if walk.MsgBox(t.view.Form(), "覆盖文件", "已存在同名称的配置文件，继续保存将覆盖文件。", walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
								return
							}
						}
						db.Submit()
						authDB.Submit()
						t.syncAuthInfo()
						t.conf.Custom = utils.String2Map(customText.Text())
						t.conf.Save()

						lastEditName = t.conf.Name
						lastRunningState, _ = services.QueryService(t.originalName)
						if t.conf.Name != t.originalName && t.originalName != "" {
							services.UninstallService(t.originalName)
							os.Remove(t.originalName + ".ini")
						}
						if t.originalName != "" && t.originalLogFile != "" && t.originalLogFile != t.conf.LogFile {
							services.UninstallService(t.originalName)
							related, target := utils.FindRelatedFiles(t.originalLogFile, t.conf.LogFile)
							if t.conf.LogFile == "" {
								utils.TryAlterFile(t.originalLogFile, "", false)
								for _, file := range related {
									utils.TryAlterFile(file, "", false)
								}
							} else {
								utils.TryAlterFile(t.originalLogFile, t.conf.LogFile, true)
								for i := 0; i < len(related); i++ {
									utils.TryAlterFile(related[i], target[i], true)
								}
							}
						}
						t.view.Accept()
					}},
					PushButton{Text: "取消", AssignTo: &cancelPB, OnClicked: func() { t.view.Cancel() }},
				},
			},
		},
	}
}

func (t *EditConfDialog) Run(owner walk.Form) (int, error) {
	return t.View().Run(owner)
}

func (t *EditConfDialog) syncAuthInfo() {
	t.conf.AuthInfo = config.AuthInfo{AuthMethod: t.authInfo.AuthMethod}
	switch t.authInfo.AuthMethod {
	case "token":
		t.conf.Token = t.authInfo.Token
	case "oidc":
		t.conf.OIDCClientId = t.authInfo.OIDCClientId
		t.conf.OIDCAudience = t.authInfo.OIDCAudience
		t.conf.OIDCClientSecret = t.authInfo.OIDCClientSecret
		t.conf.OIDCTokenEndpoint = t.authInfo.OIDCTokenEndpoint
	}
}
