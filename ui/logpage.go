package ui

import (
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/util"
)

type LogPage struct {
	*walk.TabPage

	nameModel []*Conf
	dateModel ListModel
	logModel  *LogModel
	ch        chan logSelect
	watcher   *fsnotify.Watcher

	// Views
	logView  *walk.TableView
	nameView *walk.ComboBox
	dateView *walk.ComboBox
	openView *walk.PushButton
}

type logSelect struct {
	paths    []string
	maxLines int
}

func NewLogPage() (*LogPage, error) {
	lp := &LogPage{
		ch: make(chan logSelect),
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	lp.watcher = watcher
	return lp, nil
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
						DisplayMember:         "Name",
						OnCurrentIndexChanged: lp.switchLogName,
					},
					ComboBox{
						AssignTo:              &lp.dateView,
						StretchFactor:         1,
						DisplayMember:         "Title",
						Format:                time.DateOnly,
						OnCurrentIndexChanged: lp.switchLogDate,
					},
				},
			},
			TableView{
				Name:                "log",
				AssignTo:            &lp.logView,
				AlternatingRowBG:    true,
				LastColumnStretched: true,
				HeaderHidden:        true,
				Columns:             []TableViewColumn{{}},
				ContextMenuItems: []MenuItem{
					Action{
						Text:    i18n.Sprintf("Copy"),
						Visible: Bind("log.CurrentIndex >= 0"),
						OnTriggered: func() {
							if i := lp.logView.CurrentIndex(); i >= 0 && lp.logModel != nil {
								walk.Clipboard().SetText(lp.logModel.Value(i, 0).(string))
							}
						},
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &lp.openView,
						MinSize:  Size{Width: 150},
						Text:     i18n.Sprintf("Open Log Folder"),
						OnClicked: func() {
							if i := lp.dateView.CurrentIndex(); i >= 0 && i < len(lp.dateModel) {
								paths := lp.dateModel[i : i+1]
								if i == 0 {
									paths = lp.dateModel
								}
								for _, path := range paths {
									if util.FileExists(path.Value) {
										openFolder(path.Value)
										break
									}
								}
							}
						},
					},
				},
			},
		},
	}
}

func (lp *LogPage) OnCreate() {
	lp.VisibleChanged().Attach(lp.onVisibleChanged)
	go func() {
		// Due to the file caching mechanism, new logs may not be written to
		// the disk immediately, and therefore no write events will be received.
		// It is still necessary to read files regularly.
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		var path string
		var watch bool
		for {
			select {
			case event, ok := <-lp.watcher.Events:
				if !ok {
					return
				}
				if path != event.Name {
					continue
				}
				if event.Has(fsnotify.Write) {
					lp.refreshLog()
				} else if event.Has(fsnotify.Create) {
					lp.logView.Synchronize(func() {
						if lp.logModel != nil {
							lp.logModel.Reset()
						}
						if !lp.openView.Enabled() {
							lp.openView.SetEnabled(true)
						}
					})
				}
			case logs := <-lp.ch:
				// Try to avoid duplicate operations
				if path != "" && len(logs.paths) > 0 && logs.paths[0] == path {
					continue
				}
				if path != "" {
					if watch {
						lp.watcher.Remove(filepath.Dir(path))
					}
					path = ""
					watch = false
				}
				var model *LogModel
				var ok bool
				if len(logs.paths) > 0 {
					path = logs.paths[0]
					watch = logs.maxLines > 0
					if watch {
						lp.watcher.Add(filepath.Dir(path))
					}
					model, ok = NewLogModel(logs.paths, logs.maxLines)
				}
				lp.Synchronize(func() {
					lp.openView.SetEnabled(ok)
					lp.logModel = model
					if model != nil {
						lp.logView.SetModel(model)
						lp.scrollToBottom()
					} else {
						lp.logView.SetModel(nil)
					}
				})
			case <-ticker.C:
				if path != "" && watch {
					lp.refreshLog()
				}
			}
		}
	}()
}

func (lp *LogPage) refreshLog() {
	lp.logView.Synchronize(func() {
		if lp.logModel != nil {
			scroll := lp.logModel.RowCount() == 0 || lp.logView.ItemVisible(lp.logModel.RowCount()-1)
			if n, err := lp.logModel.ReadMore(); err == nil && n > 0 && scroll {
				lp.scrollToBottom()
			}
		}
	})
}

func (lp *LogPage) onVisibleChanged() {
	if lp.Visible() {
		// Try to avoid duplicate operations
		if lp.nameView.CurrentIndex() >= 0 {
			return
		}
		// Refresh config name list
		lp.nameModel = getConfListSafe()
		lp.nameView.SetModel(lp.nameModel)
		if len(lp.nameModel) == 0 {
			return
		}
		// Switch to current config log first
		if conf := getCurrentConf(); conf != nil {
			if i := slices.Index(lp.nameModel, conf); i >= 0 {
				lp.nameView.SetCurrentIndex(i)
				return
			}
		}
		// Fallback to the first config log
		lp.nameView.SetCurrentIndex(0)
	} else {
		lp.nameView.SetCurrentIndex(-1)
	}
}

func (lp *LogPage) scrollToBottom() {
	if count := lp.logModel.RowCount(); count > 0 {
		lp.logView.EnsureItemVisible(count - 1)
	}
}

func (lp *LogPage) switchLogName() {
	index := lp.nameView.CurrentIndex()
	cleanup := func() {
		lp.dateModel = nil
		lp.dateView.SetModel(nil)
		lp.ch <- logSelect{}
	}
	if index < 0 || lp.nameModel == nil {
		cleanup()
		return
	}
	files, dates, err := util.FindLogFiles(lp.nameModel[index].Data.GetLogFile())
	if err != nil {
		cleanup()
		return
	}
	pairs := lo.Zip2(files, dates)
	sort.SliceStable(pairs[1:], func(i, j int) bool {
		return pairs[i+1].B.After(pairs[j+1].B)
	})
	files, dates = lo.Unzip2(pairs)
	titles := lo.ToAnySlice(dates)
	titles[0] = i18n.Sprintf("Latest")
	lp.dateModel = NewListModel(files, titles...)
	lp.dateView.SetCurrentIndex(-1)
	lp.dateView.SetModel(lp.dateModel)
	lp.dateView.SetCurrentIndex(0)
}

func (lp *LogPage) switchLogDate() {
	index := lp.dateView.CurrentIndex()
	if index < 0 || lp.dateModel == nil {
		return
	}
	if index == 0 {
		lp.ch <- logSelect{
			paths: lo.Map(lp.dateModel, func(item *ListItem, index int) string {
				return item.Value
			}),
			maxLines: 2000,
		}
	} else {
		lp.ch <- logSelect{paths: []string{lp.dateModel[index].Value}, maxLines: -1}
	}
}

func (lp *LogPage) Close() error {
	return lp.watcher.Close()
}
