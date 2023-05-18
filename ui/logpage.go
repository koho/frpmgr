package ui

import (
	"sort"
	"sync"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/thoas/go-funk"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/util"
)

type LogPage struct {
	*walk.TabPage
	sync.Mutex

	nameModel   *ListModel
	dateModel   []*StringPair
	logModel    *LogModel
	db          *walk.DataBinder
	logFileChan chan logSelect

	// Views
	logView  *walk.TableView
	nameView *walk.ComboBox
	dateView *walk.ComboBox
}

type logSelect struct {
	path string
	// main defines whether the log file is used by config now.
	main bool
}

func NewLogPage() *LogPage {
	v := new(LogPage)
	v.logFileChan = make(chan logSelect)
	v.logModel = NewLogModel("")
	return v
}

func (lp *LogPage) Page() TabPage {
	return TabPage{
		AssignTo: &lp.TabPage,
		Title:    i18n.Sprintf("Log"),
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					ComboBox{
						AssignTo:              &lp.nameView,
						StretchFactor:         2,
						OnCurrentIndexChanged: lp.switchLogName,
					},
					ComboBox{
						AssignTo:              &lp.dateView,
						StretchFactor:         1,
						DisplayMember:         "DisplayName",
						OnCurrentIndexChanged: lp.switchLogDate,
					},
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
						MinSize: Size{Width: 150},
						Enabled: Bind("HasLogFile"),
						Text:    i18n.Sprintf("Open Log Folder"),
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
		var lastLog string
		for {
			select {
			case logFile := <-lp.logFileChan:
				// CurrentIndexChanged event may be triggered multiple times.
				// Try to avoid duplicate operations.
				if lastLog != "" && logFile.path == lastLog {
					continue
				}
				lastLog = logFile.path

				lp.Lock()
				lp.logModel = NewLogModel(logFile.path)
				lp.logModel.Reset()
				// A change of main log name.
				// The available date list need to be updated.
				if logFile.main {
					fl, dl, _ := util.FindLogFiles(logFile.path)
					lp.dateModel = NewStringPairModel(fl, dl, i18n.Sprintf("Latest"))
					sort.SliceStable(lp.dateModel, func(i, j int) bool {
						t1, err := time.Parse("2006-01-02", lp.dateModel[i].DisplayName)
						if err != nil {
							// Put non-date string at top.
							return true
						}
						t2, err := time.Parse("2006-01-02", lp.dateModel[j].DisplayName)
						if err != nil {
							// Put non-date string at top.
							return false
						}
						return t1.After(t2)
					})
				}
				lp.Unlock()

				lp.Synchronize(func() {
					lp.Lock()
					defer lp.Unlock()
					lp.db.Reset()
					lp.logView.SetModel(lp.logModel)
					if logFile.main {
						// A change of main log name always reads the latest log file.
						// So there's no need to trigger another same operation.
						// Moreover, the update of model in date view will trigger
						// CurrentIndexChanged event which may cause a deadlock.
						// Thus, we must disable this event first.
						lp.dateView.CurrentIndexChanged().Detach(0)
						lp.dateView.SetModel(lp.dateModel)
						if len(lp.dateModel) > 0 {
							lp.dateView.SetCurrentIndex(0)
						}
						// We are safe to restore the event now.
						lp.dateView.CurrentIndexChanged().Attach(lp.switchLogDate)
					}
					lp.scrollToBottom()
				})
			case <-ticker.C:
				// We should only read log when the log page is visible.
				// Also, there's no need to reload the backup log.
				if !lp.Visible() || lp.dateView.CurrentIndex() > 0 {
					continue
				}
				lp.Lock()
				err := lp.logModel.Reset()
				lp.Unlock()
				if err == nil {
					lp.Synchronize(func() {
						lp.Lock()
						defer lp.Unlock()
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

func (lp *LogPage) switchLogName() {
	index := lp.nameView.CurrentIndex()
	if index < 0 || lp.nameModel == nil {
		// No config selected, the log page should be empty
		lp.logFileChan <- logSelect{"", true}
		return
	}
	lp.logFileChan <- logSelect{lp.nameModel.items[index].Data.GetLogFile(), true}
}

func (lp *LogPage) switchLogDate() {
	index := lp.dateView.CurrentIndex()
	if index < 0 || lp.dateModel == nil {
		return
	}
	lp.logFileChan <- logSelect{lp.dateModel[index].Name, false}
}
