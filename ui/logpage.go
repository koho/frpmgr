package ui

import (
	"github.com/koho/frpmgr/config"
	"github.com/koho/frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type LogPage struct {
	view        *walk.TabPage
	logView     *walk.TableView
	nameSelect  *walk.ComboBox
	model       *LogModel
	logFileChan chan string
	logDB       *walk.DataBinder
}

func NewLogPage() *LogPage {
	v := new(LogPage)
	v.logFileChan = make(chan string)
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
						t.logFileChan <- ""
						return
					}
					conf := config.Configurations[index]
					t.logFileChan <- conf.LogFile
				},
			},
			TableView{
				AssignTo:            &t.logView,
				AlternatingRowBG:    true,
				LastColumnStretched: true,
				HeaderHidden:        true,
				Columns:             []TableViewColumn{{DataMember: "Text"}},
			},
			Composite{
				DataBinder: DataBinder{AssignTo: &t.logDB, DataSource: &struct {
					LogPathValid func() bool
				}{t.isLogPathValid}, Name: "logData"},
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						MinSize:   Size{150, 0},
						Enabled:   Bind("logData.LogPathValid"),
						Text:      "打开日志文件夹",
						OnClicked: t.openLogFolder,
					},
				},
			},
		},
	}
}

func (t *LogPage) Initialize() {
	t.view.VisibleChanged().Attach(func() {
		if t.view.Visible() {
			t.nameSelect.SetModel(NewConfListModel(config.Configurations))
			if i, found := utils.Find(config.GetConfigNames(), lastEditName); found && lastEditName != "" && i >= 0 {
				t.nameSelect.SetCurrentIndex(i)
				t.logFileChan <- config.Configurations[i].LogFile
			} else if len(config.Configurations) > 0 {
				t.nameSelect.SetCurrentIndex(0)
				t.logFileChan <- config.Configurations[0].LogFile
			}
		}
	})
	ticker := time.NewTicker(time.Second * 3)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case f := <-t.logFileChan:
				t.model = NewLogModel(f)
				t.view.Synchronize(func() {
					if t.model == nil {
						t.logView.SetModel(nil)
					} else {
						t.logView.SetModel(t.model)
					}
					t.logDB.Reset()
					t.scrollToBottom()
				})
			case <-ticker.C:
				t.view.Synchronize(func() {
					if t.view.Visible() && t.model != nil {
						t.model.Reset()
						t.logView.SetModel(t.model)
						t.scrollToBottom()
					}
				})
			}
		}
	}()
}

func (t *LogPage) scrollToBottom() {
	if t.model != nil && len(t.model.items) > 0 {
		t.logView.EnsureItemVisible(len(t.model.items) - 1)
	}
}

func (t *LogPage) isLogPathValid() bool {
	if t.model != nil && t.model.path != "" {
		if _, err := os.Stat(t.model.path); err == nil {
			if _, err := filepath.Abs(t.model.path); err == nil {
				return true
			}
		}
	}
	return false
}

func (t *LogPage) openLogFolder() {
	if t.isLogPathValid() {
		if absPath, err := filepath.Abs(t.model.path); err == nil {
			exec.Command(`explorer`, `/select,`, absPath).Run()
		}
	}
}
