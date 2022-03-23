package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type DetailView struct {
	*walk.Composite

	panelView   *PanelView
	sectionView *SectionView
}

func NewDetailView() *DetailView {
	v := new(DetailView)
	v.panelView = NewPanelView()
	v.sectionView = NewSectionView()
	return v
}

func (dv *DetailView) View() Widget {
	return Composite{
		Visible:  Bind("conf.Selected"),
		AssignTo: &dv.Composite,
		Layout:   VBox{Margins: Margins{5, 0, 0, 0}, SpacingZero: true},
		Children: []Widget{
			dv.panelView.View(),
			VSpacer{Size: 6},
			dv.sectionView.View(),
		},
	}
}

func (dv *DetailView) OnCreate() {
	// Create all child views
	dv.panelView.OnCreate()
	dv.sectionView.OnCreate()
	dv.sectionView.toolbar.ApplyDPI(dv.DPI())
	confDB.ResetFinished().Attach(dv.Invalidate)
}

func (dv *DetailView) Invalidate() {
	dv.panelView.Invalidate()
	dv.sectionView.Invalidate()
}
