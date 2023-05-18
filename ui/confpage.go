package ui

import (
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
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
								if err := services.ReloadService(conf.Name); err != nil {
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
		},
	}
}

func (cp *ConfPage) welcomeView() Composite {
	return Composite{
		Visible: Bind("!conf.Selected"),
		Layout:  VBox{Margins: Margins{Left: 100, Right: 100}, Spacing: 20},
		Children: []Widget{
			HSpacer{},
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
		configRemoved := false
		list := getConfListSafe()
		for _, conf := range list {
			conf.Lock()
			lastState := conf.State
			lastInstall := conf.Install
			running, err := services.QueryService(conf.Name)
			if running {
				conf.State = consts.StateStarted
			} else {
				conf.State = consts.StateStopped
			}
			conf.Install = err == nil
			if conf.State != lastState || conf.Install != lastInstall {
				stateChanged = true
				// Check whether the config file was deleted after a service uninstallation
				if !conf.Install && !util.FileExists(conf.Path) && deleteConf(conf) {
					configRemoved = true
				}
			}
			conf.Unlock()
		}
		// Only update views on state changes
		if stateChanged {
			cp.Synchronize(func() {
				if configRemoved {
					if conf := getCurrentConf(); conf != nil {
						cp.confView.reset(conf.Name)
					} else {
						cp.confView.Invalidate()
					}
				} else {
					cp.confView.listView.Invalidate()
					cp.detailView.panelView.Invalidate()
				}
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

func ensureExistingConfig(name string, owner walk.Form) bool {
	if hasConf(name) {
		return true
	}
	warnConfigRemoved(owner, name)
	return false
}

func warnConfigRemoved(owner walk.Form, name string) {
	showWarningMessage(owner, i18n.Sprintf("Config already removed"), i18n.Sprintf("The config \"%s\" already removed.", name))
}
