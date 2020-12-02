package ui

import (
	"frpmgr/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type ConfPage struct {
	*ConfView
	*DetailView
	confContainer   *walk.Composite
	detailContainer *walk.Composite
	fillerContainer *walk.Composite
}

func NewConfPage() *ConfPage {
	v := new(ConfPage)
	v.ConfView = NewConfView(&v.confContainer)
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
						Children: []Widget{
							t.ConfListView.View(),
							t.ToolbarView.View(),
						},
					},
					Composite{
						StretchFactor: 10,
						AssignTo:      &t.detailContainer,
						Layout:        VBox{MarginsZero: true, SpacingZero: true},
						Children: []Widget{
							t.DetailView.ConfStatusView.View(),
							t.DetailView.ConfSectionView.View(),
						},
					},
					Composite{
						StretchFactor: 10,
						AssignTo:      &t.fillerContainer,
						Layout:        VBox{Margins: Margins{Left: 100, Right: 100}, Spacing: 20},
						Children: []Widget{
							HSpacer{},
							VSpacer{},
							PushButton{Text: "新建配置", MinSize: Size{200, 0}, MaxSize: Size{200, 0}, OnClicked: func() {
								t.ConfView.onEditConf(nil)
							}},
							PushButton{Text: "导入配置", MinSize: Size{200, 0}, MaxSize: Size{200, 0}, OnClicked: t.ConfView.onImport},
							VSpacer{},
						},
					},
				},
			},
		},
	}
}

func (t *ConfPage) Initialize() {
	t.ConfView.Initialize()
	t.ConfView.ConfListView.view.CurrentIndexChanged().Attach(t.UpdateDetailView)
	t.DetailView.Initialize()
	t.onConfigChanged(len(config.Configurations))
	t.ConfView.ConfListView.view.SetCurrentIndex(0)
}

func (t *ConfPage) UpdateDetailView() {
	conf := t.ConfView.ConfListView.CurrentConf()
	t.DetailView.SetConf(conf)
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
