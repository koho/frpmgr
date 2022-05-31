package ui

import (
	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/services"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
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
				Commit: func(conf *Conf, forceStart bool) {
					if conf != nil {
						if err := conf.Save(); err != nil {
							showError(err, cp.Form())
							return
						}
						if forceStart {
							// The service of config is stopped by other code, but it should be restarted
						} else if conf.State == consts.StateStarted {
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
			HSplitter{
				Children: []Widget{
					cp.confView.View(),
					Composite{
						StretchFactor: 10,
						Layout:        HBox{MarginsZero: true, SpacingZero: true},
						Children: []Widget{
							cp.detailView.View(),
							cp.welcomeView(),
						},
					},
				},
			},
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
				Text:    i18n.Sprintf("New Configuration"),
				MinSize: Size{200, 0},
				MaxSize: Size{200, 0},
				OnClicked: func() {
					cp.confView.onEditConf(nil)
				},
			},
			PushButton{
				Text:      i18n.Sprintf("Import Config from File"),
				MinSize:   Size{200, 0},
				MaxSize:   Size{200, 0},
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
			}
			conf.Unlock()
		}
		// Only update views on state changes
		if stateChanged {
			cp.confView.listView.Invalidate()
			cp.detailView.panelView.Invalidate()
		}
	}
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		// Trigger a state query first
		query()
		for {
			select {
			case <-ticker.C:
				query()
			}
		}
	}()
}
