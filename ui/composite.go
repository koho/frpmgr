package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// NewBrowseLineEdit places a tool button at the tail of a LineEdit, and opens a file dialog when the button is clicked
func NewBrowseLineEdit(assignTo **walk.LineEdit, visible Property, text Property, title, filter string, file bool) Composite {
	var editView *walk.LineEdit
	if assignTo == nil {
		assignTo = &editView
	}
	return Composite{
		Visible: visible,
		Layout:  HBox{MarginsZero: true, SpacingZero: false, Spacing: 3},
		Children: []Widget{
			LineEdit{AssignTo: assignTo, Text: text},
			ToolButton{Text: "...", MaxSize: Size{Width: 24}, OnClicked: func() {
				openFileDialog(*assignTo, title, filter, file)
			}},
		}}
}
