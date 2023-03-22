package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type DetailView struct {
	*walk.Composite

	panelView *PanelView
	proxyView *ProxyView
}

func NewDetailView() *DetailView {
	v := new(DetailView)
	v.panelView = NewPanelView()
	v.proxyView = NewProxyView()
	return v
}

func (dv *DetailView) View() Widget {
	return Composite{
		Visible:  Bind("conf.Selected"),
		AssignTo: &dv.Composite,
		Layout:   VBox{Margins: Margins{Left: 5}, SpacingZero: true},
		Children: []Widget{
			dv.panelView.View(),
			VSpacer{Size: 6},
			dv.proxyView.View(),
		},
	}
}

func (dv *DetailView) OnCreate() {
	// Create all child views
	dv.panelView.OnCreate()
	dv.proxyView.OnCreate()
	dv.proxyView.toolbar.ApplyDPI(dv.DPI())
	confDB.ResetFinished().Attach(dv.Invalidate)
}

func (dv *DetailView) Invalidate() {
	dv.panelView.Invalidate()
	dv.proxyView.Invalidate()
}
