package ui

import (
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/services"
)

type ConfPage struct {
	*walk.TabPage

	// Views
	confView    *ConfView
	detailView  *DetailView
	welcomeView *walk.Composite

	svcCleanup func() error
}

func NewConfPage(cfgList []*Conf) *ConfPage {
	v := new(ConfPage)
	v.confView = NewConfView(cfgList)
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
				List:     cp.confView.model.List,
				SetState: cp.confView.model.SetStateByConf,
				Commit: func(conf *Conf, flag runFlag) {
					if conf != nil {
						if err := conf.Save(); err != nil {
							showError(err, cp.Form())
							return
						}
						if flag == runFlagForceStart {
							// The service of config is stopped by other code, but it should be restarted
						} else if conf.State == consts.ConfigStateStarted {
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
			cp.createWelcomeView(),
			cp.createMultiSelectionView(),
		},
	}
}

func (cp *ConfPage) createWelcomeView() Composite {
	return Composite{
		AssignTo: &cp.welcomeView,
		Visible:  Bind("confView.SelectedCount == 0"),
		Layout:   HBox{},
		Children: []Widget{
			HSpacer{},
			Composite{
				Layout: VBox{Spacing: 20},
				Children: []Widget{
					VSpacer{},
					PushButton{
						Text:      i18n.Sprintf("New Configuration"),
						MinSize:   Size{Width: 200},
						OnClicked: cp.confView.editNew,
					},
					PushButton{
						Text:      i18n.Sprintf("Import from File"),
						MinSize:   Size{Width: 200},
						OnClicked: cp.confView.onFileImport,
					},
					VSpacer{},
				},
			},
			HSpacer{},
		},
	}
}

func (cp *ConfPage) createMultiSelectionView() Composite {
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
	if cp.confView.model.RowCount() > 0 {
		cp.confView.listView.SetCurrentIndex(0)
	}
	cp.confView.model.RowEdited().Attach(func(i int) {
		cp.detailView.panelView.Invalidate(false)
	})
	cp.addVisibleChangedListener()
	cleanup, err := services.WatchConfigServices(func() []string {
		return lo.Map(getConfList(), func(item *Conf, index int) string {
			return item.Path
		})
	}, func(path string, state consts.ConfigState) {
		cp.Synchronize(func() {
			if cp.confView.model.SetStateByPath(path, state) {
				if conf := getCurrentConf(); conf != nil && conf.Path == path {
					cp.detailView.panelView.setState(state)
					if !cp.Visible() {
						return
					}
					if state == consts.ConfigStateStarted {
						cp.detailView.proxyView.startTracker(true)
					} else {
						if cp.detailView.proxyView.stopTracker() {
							cp.detailView.proxyView.resetProxyState(-1)
						}
					}
				}
			}
		})
	})
	if err != nil {
		showError(err, cp.Form())
		return
	}
	cp.svcCleanup = cleanup
}

func (cp *ConfPage) addVisibleChangedListener() {
	var oldState consts.ConfigState
	cp.VisibleChanged().Attach(func() {
		if cp.Visible() {
			defer func() {
				oldState = consts.ConfigStateUnknown
			}()
			if conf := getCurrentConf(); conf != nil {
				if conf.State == consts.ConfigStateStarted {
					cp.detailView.proxyView.startTracker(true)
				} else if oldState == consts.ConfigStateStarted {
					cp.detailView.proxyView.resetProxyState(-1)
				}
			}
		} else {
			cp.detailView.proxyView.stopTracker()
			if conf := getCurrentConf(); conf != nil {
				oldState = conf.State
			}
		}
	})
}

func (cp *ConfPage) Close() error {
	if cp.svcCleanup != nil {
		return cp.svcCleanup()
	}
	cp.detailView.proxyView.stopTracker()
	return nil
}

func warnConfigRemoved(owner walk.Form, name string) {
	showWarningMessage(owner, i18n.Sprintf("Config already removed"), i18n.Sprintf("The config \"%s\" already removed.", name))
}
