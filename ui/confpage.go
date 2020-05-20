package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type ConfPage struct {
	*ConfView
	*DetailView
	confContainer   *walk.Composite
	detailContainer *walk.Composite
}

func NewConfPage() *ConfPage {
	v := new(ConfPage)
	v.ConfView = NewConfView(&v.confContainer)
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
				},
			},
		},
	}
}

func (t *ConfPage) Initialize() {
	t.ConfView.Initialize()
	t.ConfView.ConfListView.view.CurrentIndexChanged().Attach(t.UpdateDetailView)
	t.DetailView.Initialize()
	t.ConfView.ConfListView.view.SetCurrentIndex(0)
}

func (t *ConfPage) UpdateDetailView() {
	conf := t.ConfView.ConfListView.CurrentConf()
	t.DetailView.SetConf(conf)
}
