package ui

import (
	conf2 "frpmgr/config"
	"frpmgr/services"
	"frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
	"syscall"
	"time"
)

var lastEditName string
var lastRunningState bool

type EditConfDialog struct {
	view *walk.Dialog

	conf            *conf2.Config
	nameList        []string
	originalName    string
	originalLogFile string
	authInfo        conf2.AuthInfo
}

func NewEditConfDialog(conf *conf2.Config, nameList []string) *EditConfDialog {
	v := new(EditConfDialog)
	v.nameList = nameList
	if conf == nil {
		conf = new(conf2.Config)
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
	var nameView *walk.LineEdit
	var db *walk.DataBinder
	var authDB *walk.DataBinder
	changeAuthMethod := func() {
		authDB.Submit()
		authDB.Reset()
	}
	return Dialog{
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
								logFileView.SetText("logs" + "/" + nameView.Text())
							}},
							Label{Text: "服务器地址:"},
							LineEdit{Text: Bind("ServerAddress", Regexp{".+"})},
							Label{Text: "服务器端口:"},
							LineEdit{Text: Bind("ServerPort", Regexp{"^\\d+$"})},
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
							Label{Text: "连接池数量:"},
							NumberEdit{Value: Bind("PoolCount")},
							Label{Text: "其他:"},
							CheckBox{Text: "初次登录失败后退出", Checked: Bind("LoginFailExit")},
						},
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: "确定", AssignTo: &acceptPB, OnClicked: func() {
						if _, found := utils.Find(t.nameList, nameView.Text()); found && nameView.Text() != t.originalName {
							if walk.MsgBox(t.view.Form(), "提示", "已存在同名称的配置文件，是否覆盖？", walk.MsgBoxOKCancel|walk.MsgBoxIconQuestion) == walk.DlgCmdCancel {
								return
							}
						}
						db.Submit()
						authDB.Submit()
						t.syncAuthInfo()
						t.conf.Save()

						lastEditName = t.conf.Name
						lastRunningState, _ = services.QueryService(t.originalName)
						if t.conf.Name != t.originalName && t.originalName != "" {
							services.UninstallService(t.originalName)
							os.Remove(t.originalName + ".ini")
						}
						if t.originalName != "" && t.originalLogFile != "" && t.originalLogFile != t.conf.LogFile {
							services.UninstallService(t.originalName)
							if t.conf.LogFile == "" {
								go tryAlterFile(t.originalLogFile, "", false)
							} else {
								tryAlterFile(t.originalLogFile, t.conf.LogFile, true)
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
	t.conf.AuthInfo = conf2.AuthInfo{AuthMethod: t.authInfo.AuthMethod}
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

func tryAlterFile(f1 string, f2 string, rename bool) {
	for i := 0; i < 5; i++ {
		var err error
		if rename {
			err = os.Rename(f1, f2)
		} else {
			err = os.Remove(f1)
		}
		if err == nil {
			break
		}
		if err, ok := err.(*os.LinkError); ok && err.Err == syscall.ERROR_FILE_NOT_FOUND {
			break
		}
		if err, ok := err.(*os.PathError); ok && err.Err == syscall.ERROR_FILE_NOT_FOUND {
			break
		}
		time.Sleep(time.Second * 1)
	}
}
