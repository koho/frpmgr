package ui

import (
	"github.com/koho/frpmgr/config"
	"github.com/koho/frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type EditSectionDialog struct {
	view     *walk.Dialog
	section  *config.Section
	baseInfo BaseInfo
}

type BaseInfo struct {
	LocalAddrVisible  bool
	LocalPortVisible  bool
	RemotePortVisible bool
	RoleVisible       bool
	SKVisible         bool
	ServerNameVisible bool
	BindAddrVisible   bool
	BindPortVisible   bool

	config.Section
}

func NewEditSectionDialog(sect *config.Section) *EditSectionDialog {
	v := new(EditSectionDialog)
	if sect == nil {
		sect = &config.Section{}
		sect.Type = "tcp"
	}
	v.section = sect
	v.baseInfo.Section = *v.section
	return v
}

func (t *EditSectionDialog) View() Dialog {
	var db, baseDB *walk.DataBinder
	var acceptPB, cancelPB *walk.PushButton
	var nameView *walk.LineEdit
	var typeBox *walk.ComboBox
	var roleCheck *walk.CheckBox
	var customText *walk.TextEdit
	icon, _ := loadLogoIcon(32)
	return Dialog{
		Icon:          icon,
		AssignTo:      &t.view,
		Title:         "编辑项目",
		Layout:        VBox{},
		Font:          Font{Family: "微软雅黑", PointSize: 9},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder: DataBinder{
			AssignTo:   &db,
			Name:       "section",
			DataSource: t.section,
		},
		Children: []Widget{
			Composite{
				Layout: VBox{MarginsZero: true},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2, SpacingZero: false, Margins: Margins{0, 4, 0, 4}},
						Children: []Widget{
							Label{Text: "名称:", Alignment: AlignHNearVCenter},
							LineEdit{AssignTo: &nameView, Text: Bind("Name", Regexp{".+"})},
							Label{Text: "类型:", Alignment: AlignHNearVCenter},
							ComboBox{
								AssignTo: &typeBox,
								Editable: true,
								Model:    []string{"tcp", "udp", "xtcp", "stcp"},
								Value:    Bind("Type"),
								OnTextChanged: func() {
									baseDB.Submit()
									t.switchType(typeBox.Text(), roleCheck.Checked())
									baseDB.Reset()
								},
							},
						},
					},
					TabWidget{
						MinSize: Size{320, 230},
						Pages: []TabPage{
							{
								Title:      "基本信息",
								Layout:     Grid{Columns: 2},
								DataBinder: DataBinder{AssignTo: &baseDB, DataSource: &t.baseInfo, Name: "base"},
								Children: []Widget{
									Label{Visible: Bind("RoleVisible"), Text: "角色:"},
									CheckBox{Visible: Bind("RoleVisible"), AssignTo: &roleCheck, Text: "访问者", Checked: t.section.Role == "visitor", OnCheckedChanged: func() {
										baseDB.Submit()
										t.switchType(typeBox.Text(), roleCheck.Checked())
										baseDB.Reset()
									}},
									Label{Visible: Bind("SKVisible"), Text: "私钥:"},
									LineEdit{Visible: Bind("SKVisible"), Text: Bind("SK")},
									Label{Visible: Bind("LocalAddrVisible"), Text: "本地地址:"},
									LineEdit{Visible: Bind("LocalAddrVisible"), Text: Bind("LocalIP")},
									Label{Visible: Bind("LocalPortVisible"), Text: "本地端口:"},
									LineEdit{Visible: Bind("LocalPortVisible"), Text: Bind("LocalPort")},
									Label{Visible: Bind("RemotePortVisible"), Text: "远程端口:"},
									LineEdit{Visible: Bind("RemotePortVisible"), Text: Bind("RemotePort")},
									Label{Visible: Bind("BindAddrVisible"), Text: "绑定地址:"},
									LineEdit{Visible: Bind("BindAddrVisible"), Text: Bind("BindAddr")},
									Label{Visible: Bind("BindPortVisible"), Text: "绑定端口:"},
									LineEdit{Visible: Bind("BindPortVisible"), Text: Bind("BindPort")},
									Label{Visible: Bind("ServerNameVisible"), Text: "服务名称:"},
									LineEdit{Visible: Bind("ServerNameVisible"), Text: Bind("ServerName")},
								},
							},
							{
								Title:  "高级",
								Layout: HBox{Alignment: AlignHNearVNear},
								Children: []Widget{
									CheckBox{Text: "加密传输", Checked: Bind("UseEncryption")},
									CheckBox{Text: "压缩传输", Checked: Bind("UseCompression")},
								},
							},
							{
								Title:  "自定义",
								Layout: VBox{},
								Children: []Widget{
									Label{Text: "*参考 FRP 支持的参数"},
									TextEdit{AssignTo: &customText, Text: utils.Map2String(t.section.Custom), VScroll: true},
								},
							},
						},
					},
				},
			},
			VSpacer{},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: "确定", AssignTo: &acceptPB, OnClicked: func() {
						if nameView.Text() == "" {
							return
						}
						baseDB.Submit()
						db.Submit()
						t.refineConfig(roleCheck.Checked(), customText.Text())
						t.view.Accept()
					}},
					PushButton{Text: "取消", AssignTo: &cancelPB, OnClicked: func() { t.view.Cancel() }},
				},
			},
		},
	}
}

func (t *EditSectionDialog) Run(owner walk.Form) (int, error) {
	return t.View().Run(owner)
}

func (t *EditSectionDialog) refineConfig(visitor bool, custom string) {
	t.section.LocalIP = t.baseInfo.LocalIP
	t.section.LocalPort = t.baseInfo.LocalPort
	t.section.RemotePort = t.baseInfo.RemotePort
	t.section.BindAddr = t.baseInfo.BindAddr
	t.section.BindPort = t.baseInfo.BindPort
	t.section.ServerName = t.baseInfo.ServerName
	t.section.SK = t.baseInfo.SK
	t.section.Custom = utils.String2Map(custom)
	switch t.section.Type {
	case "tcp", "udp":
		t.section.Role = ""
		t.section.SK = ""
		t.section.ServerName = ""
		t.section.BindAddr = ""
		t.section.BindPort = ""
	case "xtcp", "stcp":
		if visitor {
			t.section.Role = "visitor"
			t.section.LocalIP = ""
			t.section.LocalPort = ""
			t.section.RemotePort = ""
		} else {
			t.section.Role = ""
			t.section.RemotePort = ""
			t.section.ServerName = ""
			t.section.BindAddr = ""
			t.section.BindPort = ""
		}
	}
}

func (t *EditSectionDialog) switchType(name string, visitor bool) {
	switch name {
	case "tcp", "udp":
		t.baseInfo.LocalAddrVisible = true
		t.baseInfo.LocalPortVisible = true
		t.baseInfo.RemotePortVisible = true
		t.baseInfo.RoleVisible = false
		t.baseInfo.SKVisible = false
		t.baseInfo.ServerNameVisible = false
		t.baseInfo.BindAddrVisible = false
		t.baseInfo.BindPortVisible = false
	case "xtcp", "stcp":
		t.baseInfo.RemotePortVisible = false
		t.baseInfo.RoleVisible = true
		t.baseInfo.SKVisible = true
		if visitor {
			t.baseInfo.ServerNameVisible = true
			t.baseInfo.BindAddrVisible = true
			t.baseInfo.BindPortVisible = true
			t.baseInfo.LocalAddrVisible = false
			t.baseInfo.LocalPortVisible = false
		} else {
			t.baseInfo.ServerNameVisible = false
			t.baseInfo.BindAddrVisible = false
			t.baseInfo.BindPortVisible = false
			t.baseInfo.LocalAddrVisible = true
			t.baseInfo.LocalPortVisible = true
		}
	}
}
