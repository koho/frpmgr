package ui

import (
	"frpmgr/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
)

type ConfPage struct {
	*ConfView
	*DetailView
	confContainer   *walk.Composite
	detailContainer *walk.Composite
	fillerContainer *walk.Composite
	confDB          *walk.DataBinder
}

func NewConfPage() *ConfPage {
	v := new(ConfPage)
	v.ConfView = NewConfView(&v.confContainer, &v.confDB)
	v.ConfView.ConfigChanged = v.onConfigChanged
	v.DetailView = NewDetailView()
	return v
}

func (t *ConfPage) View() TabPage {
	return TabPage{
		Title:  "配置",
		Layout: HBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					Composite{
						StretchFactor: 1,
						AssignTo:      &t.confContainer,
						Layout:        VBox{MarginsZero: true, SpacingZero: true},
						DataBinder: DataBinder{AssignTo: &t.confDB, DataSource: &struct {
							ConfSize      func() int
							SelectedIndex func() int
						}{func() int {
							return len(config.Configurations)
						}, func() int {
							return t.ConfListView.view.CurrentIndex()
						}}, Name: "conf"},
						Children: []Widget{
							t.ConfListView.View(),
							t.ToolbarView.View(),
						},
					},
					Composite{
						StretchFactor: 10,
						Layout:        HBox{MarginsZero: true, SpacingZero: true},
						Children: []Widget{
							Composite{
								AssignTo: &t.detailContainer,
								Layout:   VBox{Margins: Margins{5, 0, 0, 0}, SpacingZero: true},
								Children: []Widget{
									t.DetailView.ConfStatusView.View(),
									VSpacer{Size: 6},
									t.DetailView.ConfSectionView.View(),
								},
							},
							Composite{
								AssignTo: &t.fillerContainer,
								Layout:   VBox{Margins: Margins{Left: 100, Right: 100}, Spacing: 20},
								Children: []Widget{
									HSpacer{},
									VSpacer{},
									PushButton{Text: "创建新配置", MinSize: Size{200, 0}, MaxSize: Size{200, 0}, OnClicked: func() {
										t.ConfView.onEditConf(nil)
									}},
									PushButton{Text: "从文件导入配置", MinSize: Size{200, 0}, MaxSize: Size{200, 0}, OnClicked: t.ConfView.onImport},
									VSpacer{},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (t *ConfPage) Initialize() {
	t.ConfView.Initialize()
	t.ConfView.ConfListView.view.CurrentIndexChanged().Attach(t.UpdateView)
	t.DetailView.Initialize()
	t.onConfigChanged(len(config.Configurations))
	t.ConfView.ConfListView.view.SetCurrentIndex(0)
	t.startQueryStatus()
}

func (t *ConfPage) startQueryStatus() {
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-config.StatusChan:
			case <-ticker.C:
			}
			config.ConfMutex.Lock()
			for _, conf := range config.Configurations {
				if s, _ := t.queryState(conf.Name); s {
					conf.Status = config.StateStarted
				} else {
					conf.Status = config.StateStopped
				}
			}
			config.ConfMutex.Unlock()
			t.ConfListView.view.Invalidate()
		}
	}()
}

func (t *ConfPage) UpdateView() {
	conf := t.ConfView.ConfListView.CurrentConf()
	t.DetailView.SetConf(conf)
	if conf != nil {
		lastEditName = conf.Name
	}
	if *(t.db) != nil {
		(*t.db).Reset()
	}
}

func (t *ConfPage) SwapFiller(filler bool) {
	if filler {
		t.fillerContainer.SetVisible(true)
		t.detailContainer.SetVisible(false)
	} else {
		t.fillerContainer.SetVisible(false)
		t.detailContainer.SetVisible(true)
	}
}

func (t *ConfPage) onConfigChanged(size int) {
	if size > 0 {
		t.SwapFiller(false)
	} else {
		t.SwapFiller(true)
	}
}
