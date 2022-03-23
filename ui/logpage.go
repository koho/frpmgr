package ui

import (
	"github.com/koho/frpmgr/pkg/util"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/thoas/go-funk"
	"time"
)

type LogPage struct {
	*walk.TabPage

	nameModel   *ListModel
	logModel    *LogModel
	db          *walk.DataBinder
	logFileChan chan string

	// Views
	logView  *walk.TableView
	nameView *walk.ComboBox
}

func NewLogPage() *LogPage {
	v := new(LogPage)
	v.logFileChan = make(chan string)
	v.logModel = NewLogModel("")
	return v
}

func (lp *LogPage) Page() TabPage {
	return TabPage{
		AssignTo: &lp.TabPage,
		Title:    "日志",
		Layout:   VBox{},
		Children: []Widget{
			ComboBox{
				AssignTo: &lp.nameView,
				OnCurrentIndexChanged: func() {
					index := lp.nameView.CurrentIndex()
					if index < 0 || lp.nameModel == nil {
						// No config selected, the log page should be empty
						lp.logFileChan <- ""
						return
					}
					lp.logFileChan <- lp.nameModel.items[index].Data.GetLogFile()
				},
			},
			TableView{
				AssignTo:            &lp.logView,
				AlternatingRowBG:    true,
				LastColumnStretched: true,
				HeaderHidden:        true,
				Columns:             []TableViewColumn{{DataMember: "Text"}},
			},
			Composite{
				DataBinder: DataBinder{
					AssignTo: &lp.db,
					DataSource: &struct {
						HasLogFile func() bool
					}{lp.hasLogFile},
				},
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						MinSize: Size{150, 0},
						Enabled: Bind("HasLogFile"),
						Text:    "打开日志文件夹",
						OnClicked: func() {
							openFolder(lp.logModel.path)
						},
					},
				},
			},
		},
	}
}

func (lp *LogPage) OnCreate() {
	lp.VisibleChanged().Attach(lp.onVisibleChanged)
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case logFile := <-lp.logFileChan:
				lp.logModel = NewLogModel(logFile)
				lp.logModel.Reset()
				lp.Synchronize(func() {
					lp.db.Reset()
					lp.logView.SetModel(lp.logModel)
					lp.scrollToBottom()
				})
			case <-ticker.C:
				// We should only read log when the log page is visible
				if !lp.Visible() {
					continue
				}
				if err := lp.logModel.Reset(); err == nil {
					lp.Synchronize(func() {
						lp.logView.SetModel(lp.logModel)
						lp.scrollToBottom()
					})
				}
			}
		}
	}()
}

func (lp *LogPage) onVisibleChanged() {
	if lp.Visible() {
		// Remember the previous selected name
		var preName string
		if idx := lp.nameView.CurrentIndex(); idx >= 0 && lp.nameModel != nil && idx < len(lp.nameModel.items) {
			preName = lp.nameModel.items[idx].Name
		}
		// Refresh config name list
		lp.nameModel = NewListModel(confList)
		lp.nameView.SetModel(lp.nameModel)
		if len(lp.nameModel.items) == 0 {
			return
		}
		// Switch to current config log first
		if conf := getCurrentConf(); conf != nil {
			if i := funk.IndexOf(lp.nameModel.items, func(c *Conf) bool { return c.Name == conf.Name }); i >= 0 {
				lp.nameView.SetCurrentIndex(i)
				return
			}
		}
		// Select previous config log
		if preName != "" {
			if i := funk.IndexOf(lp.nameModel.items, func(c *Conf) bool { return c.Name == preName }); i >= 0 {
				lp.nameView.SetCurrentIndex(i)
				return
			}
		}
		// Fallback to the first config log
		lp.nameView.SetCurrentIndex(0)
	}
}

func (lp *LogPage) scrollToBottom() {
	if len(lp.logModel.lines) > 0 {
		lp.logView.EnsureItemVisible(len(lp.logModel.lines) - 1)
	}
}

func (lp *LogPage) hasLogFile() bool {
	if lp.logModel.path != "" {
		return util.FileExists(lp.logModel.path)
	}
	return false
}
