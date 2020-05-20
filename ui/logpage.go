package ui

import (
	"frpmgr/config"
	"frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type LogPage struct {
	view       *walk.TabPage
	logView    *walk.TableView
	nameSelect *walk.ComboBox
	model      *LogModel
	lastName   string
}

func NewLogPage() *LogPage {
	v := new(LogPage)
	return v
}

func (t *LogPage) View() TabPage {
	return TabPage{
		AssignTo: &t.view,
		Title:    "日志",
		Layout:   VBox{},
		Children: []Widget{
			ComboBox{
				AssignTo:      &t.nameSelect,
				DisplayMember: "Name",
				OnCurrentIndexChanged: func() {
					index := t.nameSelect.CurrentIndex()
					if index < 0 {
						t.logView.SetModel(nil)
						return
					}
					conf := config.Configurations[index]
					t.lastName = conf.Name
					mdl := NewLogModel(conf.LogFile)
					if mdl == nil {
						t.logView.SetModel(nil)
					} else {
						t.logView.SetModel(mdl)
					}
				},
			},
			TableView{
				AssignTo:            &t.logView,
				AlternatingRowBG:    true,
				LastColumnStretched: true,
				HeaderHidden:        true,
				Columns:             []TableViewColumn{{DataMember: "Text"}},
			},
		},
	}
}

func (t *LogPage) Initialize() {
	t.view.VisibleChanged().Attach(func() {
		if t.view.Visible() {
			t.nameSelect.SetModel(NewConfListModel(config.Configurations))
			if i, found := utils.Find(config.GetConfigNames(), t.lastName); found && t.lastName != "" && i >= 0 {
				t.nameSelect.SetCurrentIndex(i)
			}
		}
	})
}
