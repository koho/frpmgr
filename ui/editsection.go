package ui

import (
	"frpmgr/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type EditSectionDialog struct {
	view    *walk.Dialog
	section *config.Section
}

func NewEditSectionDialog(sect *config.Section) *EditSectionDialog {
	v := new(EditSectionDialog)
	if sect == nil {
		sect = &config.Section{}
	}
	v.section = sect
	return v
}

func (t *EditSectionDialog) View() Dialog {
	var acceptPB, cancelPB *walk.PushButton
	return Dialog{
		AssignTo:      &t.view,
		Title:         "编辑项目",
		Layout:        VBox{},
		Font:          Font{Family: "微软雅黑", PointSize: 9},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder: DataBinder{
			Name:       "common",
			DataSource: t.section,
		},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "名称:"},
					LineEdit{Text: Bind("Name")},
					Label{Text: "类型:"},
					ComboBox{
						Model: []string{"tcp", "udp"},
						Value: Bind("Type"),
					},
					Label{Text: "本地地址:"},
					LineEdit{Text: Bind("LocalIP")},
					Label{Text: "本地端口:"},
					LineEdit{Text: Bind("LocalPort")},
					Label{Text: "远程端口:"},
					LineEdit{Text: Bind("RemotePort")},
					CheckBox{Text: "加密传输", Checked: Bind("UseEncryption")},
					CheckBox{Text: "压缩传输", Checked: Bind("UseCompression")},
				},
			},
			VSpacer{},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: "确定", AssignTo: &acceptPB, OnClicked: func() {
						t.view.DataBinder().Submit()
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
