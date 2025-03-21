package ui

import (
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/services"
)

type ConfPage struct {
	*walk.TabPage

	// Views
	confView   *ConfView
	detailView *DetailView
}

func NewConfPage() *ConfPage {
	v := new(ConfPage)
	v.confView = NewConfView()
	v.detailView = NewDetailView()
	return v
}

func (cp *ConfPage) Page() TabPage {
	return TabPage{
		AssignTo: &cp.TabPage,
		Title:    i18n.Sprintf("Configuration"),
		Layout:   HBox{},
		DataBinder: DataBinder{
			AssignTo: &confDB,
			DataSource: &ConfBinder{
				Current: nil,
				Commit: func(conf *Conf, flag runFlag) {
					if conf != nil {
						if err := conf.Save(); err != nil {
							showError(err, cp.Form())
							return
						}
						if flag == runFlagForceStart {
							// The service of config is stopped by other code, but it should be restarted
						} else if conf.State == consts.StateStarted {
							// Hot-Reloading frp configuration
							if flag == runFlagReload {
								if err := services.ReloadService(conf.Path); err != nil {
									showError(err, cp.Form())
								}
								return
							}
							// The service is running, we should stop it and restart it later
							if err := cp.detailView.panelView.StopService(conf); err != nil {
								showError(err, cp.Form())
								return
							}
						} else {
							// The service is stopped all the time, there's nothing to do about it
							return
						}
						if err := cp.detailView.panelView.StartService(conf); err != nil {
							showError(err, cp.Form())
							return
						}
					}
				},
			},
			Name: "conf",
		},
		Children: []Widget{
			cp.confView.View(),
			cp.detailView.View(),
			cp.welcomeView(),
			cp.multiSelectionView(),
		},
	}
}

func (cp *ConfPage) welcomeView() Composite {
	return Composite{
		Visible: Bind("confView.SelectedCount == 0"),
		Layout:  HBox{},
		Children: []Widget{
			HSpacer{},
			Composite{
				Layout: VBox{Spacing: 20},
				Children: []Widget{
					VSpacer{},
					PushButton{
						Text:      i18n.Sprintf("New Configuration"),
						MinSize:   Size{Width: 200},
						MaxSize:   Size{Width: 200},
						OnClicked: cp.confView.editNew,
					},
					PushButton{
						Text:      i18n.Sprintf("Import Config from File"),
						MinSize:   Size{Width: 200},
						MaxSize:   Size{Width: 200},
						OnClicked: cp.confView.onFileImport,
					},
					VSpacer{},
				},
			},
			HSpacer{},
		},
	}
}

func (cp *ConfPage) multiSelectionView() Composite {
	count := "{Count}"
	text := i18n.Sprintf("Delete %s configs", count)
	expr := "confView.SelectedCount"
	if i := strings.Index(text, count); i >= 0 {
		if left := text[:i]; left != "" {
			expr = "'" + left + "' + " + expr
		}
		if right := text[i+len(count):]; right != "" {
			expr += " + '" + right + "'"
		}
	}
	return Composite{
		Visible: Bind("confView.SelectedCount > 1"),
		Layout:  HBox{},
		Children: []Widget{
			HSpacer{},
			PushButton{
				Text:      Bind(expr),
				MinSize:   Size{Width: 200},
				MaxSize:   Size{Width: 200},
				OnClicked: cp.confView.onDelete,
			},
			HSpacer{},
		},
	}
}

func (cp *ConfPage) OnCreate() {
	// Create all child views
	cp.confView.OnCreate()
	cp.detailView.OnCreate()
	// Select the first config
	if len(confList) > 0 {
		cp.confView.listView.SetCurrentIndex(0)
	}
	// Query service state of configs
	cp.startQueryService()
}

func (cp *ConfPage) startQueryService() {
	query := func() {
		stateChanged := false
		list := getConfListSafe()
		for _, conf := range list {
			conf.Lock()
			lastState := conf.State
			lastInstall := conf.Install
			running, err := services.QueryService(conf.Path)
			if running {
				conf.State = consts.StateStarted
			} else {
				conf.State = consts.StateStopped
			}
			conf.Install = err == nil
			if conf.State != lastState || conf.Install != lastInstall {
				stateChanged = true
			}
			conf.Unlock()
		}
		// Only update views on state changes
		if stateChanged {
			cp.Synchronize(func() {
				cp.confView.listView.Invalidate()
				cp.detailView.panelView.Invalidate()
			})
		}
	}
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		// Trigger a state query first
		query()
		for range ticker.C {
			query()
		}
	}()
}

func warnConfigRemoved(owner walk.Form, name string) {
	showWarningMessage(owner, i18n.Sprintf("Config already removed"), i18n.Sprintf("The config \"%s\" already removed.", name))
}
