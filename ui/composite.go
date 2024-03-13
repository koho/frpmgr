package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/res"
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
			if binder := (*assignTo).DataBinder(); binder != nil {
				if err := binder.Submit(); err == nil {
					(*assignTo).Accept()
				}
			} else {
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
		Font:          res.TextRegular,
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
func NewRadioButtonGroup(dataMember string, db *DataBinder, visible Property, buttons []RadioButton) Composite {
	v := Composite{
		Visible: visible,
		Layout:  HBox{MarginsZero: true, SpacingZero: true},
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

// NewAttributeTable returns a two-column table view. The first column is name and the second column is value.
// It provides the ability to edit cells by double-clicking.
func NewAttributeTable(m *AttributeModel, nameWidth, valueWidth int) Composite {
	var tv *walk.TableView
	fc := func(value interface{}) string {
		return *value.(*string)
	}
	return Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			TableView{
				AssignTo: &tv,
				Columns: []TableViewColumn{
					{Title: i18n.Sprintf("Name"), Width: nameWidth, FormatFunc: fc},
					{Title: i18n.Sprintf("Value"), Width: valueWidth, FormatFunc: fc},
				},
				Model:    m,
				Editable: true,
			},
			Composite{
				Layout: VBox{MarginsZero: true},
				Children: []Widget{
					PushButton{Text: i18n.Sprintf("Add"), OnClicked: func() {
						m.Add("", "")
					}},
					PushButton{Text: i18n.Sprintf("Delete"), OnClicked: func() {
						if i := tv.CurrentIndex(); i >= 0 {
							m.Delete(i)
						}
					}},
					VSpacer{Size: 16},
					PushButton{Text: i18n.Sprintf("Clear All"), OnClicked: func() {
						m.Clear()
					}},
					VSpacer{},
				},
			},
		},
	}
}

// NewAttributeDialog returns a dialog box with data displayed in the attribute table.
func NewAttributeDialog(title string, data *map[string]string) Dialog {
	var p *walk.Dialog
	m := NewAttributeModel(*data)
	dlg := NewBasicDialog(&p, title, loadIcon(res.IconFile, 32), DataBinder{}, func() {
		*data = m.AsMap()
		p.Accept()
	},
		NewAttributeTable(m, 120, 120),
	)
	dlg.MinSize = Size{Width: 420, Height: 280}
	return dlg
}

type NIOption struct {
	Title        string
	Value        Property
	Suffix       string
	Min          float64
	Max          float64
	Width        int
	Style        uint32
	Greedy       bool
	NoSpinButton bool
	NoSpacer     bool
	Visible      Property
	Enabled      Property
}

// NewNumberInput returns a number edit with custom prefix and suffix.
func NewNumberInput(opt NIOption) Composite {
	var widgets []Widget
	if opt.Title != "" {
		widgets = append(widgets, Label{Text: opt.Title})
	}
	ne := NumberEdit{
		Value:              opt.Value,
		SpinButtonsVisible: !opt.NoSpinButton,
		MinSize:            Size{Width: opt.Width},
		MinValue:           opt.Min,
		MaxValue:           opt.Max,
		Style:              opt.Style,
		Greedy:             opt.Greedy,
	}
	if ne.MinSize.Width == 0 {
		ne.MinSize.Width = 70
	}
	widgets = append(widgets, ne)
	if opt.Suffix != "" {
		widgets = append(widgets, Label{Text: opt.Suffix})
	}
	if !opt.NoSpacer {
		widgets = append(widgets, HSpacer{})
	}
	return Composite{
		Layout:   HBox{MarginsZero: true},
		Visible:  opt.Visible,
		Enabled:  opt.Enabled,
		Children: widgets,
	}
}
