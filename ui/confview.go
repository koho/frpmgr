package ui

import (
	"fmt"
	"frpmgr/config"
	"frpmgr/services"
	"frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
	"path/filepath"
)

type ConfView struct {
	*ConfListView
	*ToolbarView
	ConfigChanged func(int)
}

func NewConfView(parent **walk.Composite) *ConfView {
	v := new(ConfView)
	v.ConfListView = NewConfListView()
	v.ToolbarView = NewToolbarView(parent)
	return v
}

func (t *ConfView) reloadConf() {
	confList, err := config.LoadConfig()
	if err != nil {
		walk.MsgBox(t.ConfListView.view.Form(), "错误", "读取配置文件失败", walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}
	config.Configurations = confList
	if t.ConfigChanged != nil {
		t.ConfigChanged(len(confList))
	}
	t.ConfListView.resetModel()
	if idx, found := utils.Find(config.GetConfigNames(), lastEditName); found {
		t.ConfListView.view.SetCurrentIndex(idx)
	}
}

func (t *ConfView) onEditConf(conf *config.Config) {
	res, _ := NewEditConfDialog(conf, config.GetConfigNames()).Run(t.ConfListView.view.Form())
	if res == walk.DlgCmdOK {
		t.reloadConf()
	}
}

func (t *ConfView) onImport() {
	dlg := walk.FileDialog{
		Filter: "配置文件 (*.ini)|*.ini|All Files (*.*)|*.*",
		Title:  "从文件导入配置",
	}

	if ok, _ := dlg.ShowOpenMultiple(t.ConfListView.view.Form()); !ok {
		return
	}
	os.Chdir(curDir)
	for _, path := range dlg.FilePaths {
		newPath := filepath.Base(path)
		if _, err := os.Stat(newPath); err == nil {
			if walk.MsgBox(t.ConfListView.view.Form(), "提示", fmt.Sprintf("文件 %s 已存在，是否覆盖？", newPath), walk.MsgBoxOKCancel|walk.MsgBoxIconQuestion) == walk.DlgCmdCancel {
				continue
			}
		}
		_, err := utils.CopyFile(path, newPath)
		if err != nil {
			walk.MsgBox(t.ConfListView.view.Form(), "错误", "复制文件时出现错误", walk.MsgBoxOK|walk.MsgBoxIconError)
		} else {
			lastEditName = config.NameFromPath(path)
		}
	}
	t.reloadConf()
}

func (t *ConfView) onDelete() {
	c := t.CurrentConf()
	if c != nil {
		if walk.MsgBox(t.ConfListView.view.Form(), "提示", fmt.Sprintf("确定删除配置 %s ？", c.Name), walk.MsgBoxOKCancel|walk.MsgBoxIconQuestion) == walk.DlgCmdCancel {
			return
		}
		c.Delete()
		services.UninstallService(c.Name)
		if c.LogFile != "" {
			os.Remove(c.LogFile)
		}
		t.reloadConf()
		t.ConfListView.view.SetCurrentIndex(0)
	}
}

func (t *ConfView) Initialize() {
	t.ToolbarView.Initialize()
	t.ToolbarView.addAction.Triggered().Attach(func() {
		t.onEditConf(nil)
	})
	t.ToolbarView.importAction.Triggered().Attach(t.onImport)
	t.ToolbarView.deleteAction.Triggered().Attach(t.onDelete)
	t.ConfListView.editAction.Triggered().Attach(func() {
		t.onEditConf(t.ConfListView.CurrentConf())
	})
	t.ConfListView.newAction.Triggered().Attach(func() {
		t.onEditConf(nil)
	})
	t.ConfListView.importAction.Triggered().Attach(t.onImport)
	t.ConfListView.deleteAction.Triggered().Attach(t.onDelete)
}

type ConfListView struct {
	model        *ConfListModel
	view         *walk.TableView
	editAction   *walk.Action
	newAction    *walk.Action
	importAction *walk.Action
	deleteAction *walk.Action
}

func NewConfListView() *ConfListView {
	clv := new(ConfListView)
	clv.model = NewConfListModel(config.Configurations)
	return clv
}

func (t *ConfListView) View() Widget {
	return TableView{
		AssignTo:            &t.view,
		LastColumnStretched: true,
		HeaderHidden:        true,
		Columns:             []TableViewColumn{{DataMember: "Name"}},
		Model:               t.model,
		ContextMenuItems: []MenuItem{
			Action{AssignTo: &t.editAction, Text: "编辑"},
			Action{AssignTo: &t.newAction, Text: "新建配置"},
			Action{AssignTo: &t.importAction, Text: "导入配置"},
			Action{AssignTo: &t.deleteAction, Text: "删除"},
		},
	}
}

func (t *ConfListView) resetModel() {
	t.model = NewConfListModel(config.Configurations)
	t.view.SetModel(t.model)
}

func (t *ConfListView) CurrentConf() *config.Config {
	index := t.view.CurrentIndex()
	if len(t.model.items) > 0 && index >= 0 {
		return t.model.items[index]
	}
	return nil
}

type ToolbarView struct {
	view   *walk.ToolBar
	parent **walk.Composite

	addMenuAction *walk.Action
	importAction  *walk.Action
	addAction     *walk.Action
	deleteAction  *walk.Action
}

func NewToolbarView(parent **walk.Composite) *ToolbarView {
	v := new(ToolbarView)
	v.parent = parent
	return v
}

func (t *ToolbarView) View() Widget {
	return Composite{
		Layout: HBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			ToolBar{
				AssignTo:      &t.view,
				OnSizeChanged: t.fixWidth,
				ButtonStyle:   ToolBarButtonImageBeforeText,
				Orientation:   Horizontal,
				Items: []MenuItem{
					Menu{
						AssignActionTo: &t.addMenuAction,
						Text:           "添加配置",
						Image:          loadSysIcon("shell32", 149, 16),
						Items: []MenuItem{
							Action{
								AssignTo: &t.importAction,
								Text:     "导入配置",
								Image:    loadSysIcon("imageres", 3, 16),
							},
							Action{
								AssignTo: &t.addAction,
								Text:     "手动添加",
								Image:    loadSysIcon("imageres", 2, 16),
							},
						},
					},
					Separator{},
					Action{
						AssignTo: &t.deleteAction,
						Image:    loadSysIcon("shell32", 131, 16),
					},
				},
			},
		},
	}
}

func (t *ToolbarView) Initialize() {
	t.importAction.SetDefault(true)
	t.deleteAction.SetToolTip("删除配置")
}

func (t *ToolbarView) fixWidth() {
	toolbarWidth := t.view.SizeHint().Width
	(*t.parent).SetMinMaxSizePixels(walk.Size{toolbarWidth, 0}, walk.Size{toolbarWidth, 0})
}
