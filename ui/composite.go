package ui

import (
	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// NewBrowseLineEdit places a tool button at the tail of a LineEdit, and opens a file dialog when the button is clicked
func NewBrowseLineEdit(assignTo **walk.LineEdit, visible, enable, text Property, title, filter string, file bool) Composite {
	var editView *walk.LineEdit
	if assignTo == nil {
		assignTo = &editView
	}
	return Composite{
		Visible: visible,
		Layout:  HBox{MarginsZero: true, SpacingZero: false, Spacing: 3},
		Children: []Widget{
			LineEdit{Enabled: enable, AssignTo: assignTo, Text: text},
			ToolButton{Enabled: enable, Text: "...", MaxSize: Size{Width: 24}, OnClicked: func() {
				openFileDialog(*assignTo, title, filter, file)
			}},
		}}
}

// NewBasicDialog returns a dialog with given widgets and default buttons
func NewBasicDialog(assignTo **walk.Dialog, title string, icon Property, db DataBinder, yes func(), widgets ...Widget) Dialog {
	var w *walk.Dialog
	if assignTo == nil {
		assignTo = &w
	}
	if yes == nil {
		// Default handler for "yes" button
		yes = func() {
			if err := (*assignTo).DataBinder().Submit(); err == nil {
				(*assignTo).Accept()
			}
		}
	}
	var acceptPB, cancelPB *walk.PushButton
	dlg := Dialog{
		AssignTo:      assignTo,
		Icon:          icon,
		Title:         title,
		Layout:        VBox{},
		Font:          consts.TextRegular,
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder:    db,
		Children:      make([]Widget, 0),
	}
	dlg.Children = append(dlg.Children, widgets...)
	dlg.Children = append(dlg.Children, Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			HSpacer{},
			PushButton{Text: i18n.Sprintf("OK"), AssignTo: &acceptPB, OnClicked: yes},
			PushButton{Text: i18n.Sprintf("Cancel"), AssignTo: &cancelPB, OnClicked: func() { (*assignTo).Cancel() }},
		},
	})
	return dlg
}

// NewRadioButtonGroup returns a simple radio button group
func NewRadioButtonGroup(dataMember string, db *DataBinder, buttons []RadioButton) Composite {
	v := Composite{
		Layout: HBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			RadioButtonGroup{
				DataMember: dataMember,
				Buttons:    buttons,
			},
			HSpacer{},
		},
	}
	if db != nil {
		v.DataBinder = *db
	}
	return v
}

// AlignGrid resizes the first child of a grid to the width of the first column.
// After that, we keep a fixed width column regardless of whether the row is hidden or not.
func AlignGrid(page TabPage, n int) TabPage {
	widgets := page.Children
	if n > 0 {
		widgets = page.Children[:n]
	}
	head := page.Children[0].(Label)
	head.MinSize = Size{Width: calculateHeadColumnTextWidth(widgets, page.Layout.(Grid).Columns)}
	page.Children[0] = head
	return page
}
