package ui

import (
	"archive/zip"
	"fmt"
	"frpmgr/config"
	"frpmgr/services"
	"frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	config.ConfMutex.Lock()
	config.Configurations = confList
	config.ConfMutex.Unlock()
	config.StatusChan <- true
	if t.ConfigChanged != nil {
		t.ConfigChanged(len(confList))
	}
	t.ConfListView.resetModel()
	if idx, found := utils.Find(config.GetConfigNames(), lastEditName); found {
		t.ConfListView.view.SetCurrentIndex(idx)
	}
	if t.toolbarDB != nil {
		t.toolbarDB.Reset()
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
		Filter: "配置文件 (*.zip, *.ini)|*.zip;*.ini|All Files (*.*)|*.*",
		Title:  "从文件导入配置",
	}

	if ok, _ := dlg.ShowOpenMultiple(t.ConfListView.view.Form()); !ok {
		return
	}
	os.Chdir(curDir)
	for _, path := range dlg.FilePaths {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".ini":
			newPath := filepath.Base(path)
			if !t.askForOverride(path) {
				continue
			}
			_, err := utils.CopyFile(path, newPath)
			if err != nil {
				walk.MsgBox(t.ConfListView.view.Form(), "错误", "复制文件时出现错误", walk.MsgBoxOK|walk.MsgBoxIconError)
			} else {
				lastEditName = config.NameFromPath(path)
			}
		case ".zip":
			t.unzipFiles(path)
		}
	}
	t.reloadConf()
}

func (t *ConfView) askForOverride(path string) bool {
	newPath := filepath.Base(path)
	if _, err := os.Stat(newPath); err == nil {
		if walk.MsgBox(t.ConfListView.view.Form(), "覆盖文件", fmt.Sprintf("文件 %s 已存在，是否覆盖?", newPath), walk.MsgBoxYesNo|walk.MsgBoxIconWarning) == walk.DlgCmdNo {
			return false
		}
	}
	return true
}

func (t *ConfView) unzipFiles(path string) {
	showError := func() {
		walk.MsgBox(t.ConfListView.view.Form(), "错误", "读取压缩文件时出现错误", walk.MsgBoxOK|walk.MsgBoxIconError)
	}
	unzip := func(file *zip.File) error {
		fr, err := file.Open()
		if err != nil {
			return err
		}
		defer fr.Close()
		if !t.askForOverride(file.Name) {
			return nil
		}
		fw, err := os.OpenFile(file.Name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer fw.Close()
		_, err = io.Copy(fw, fr)
		if err != nil {
			return err
		}
		return nil
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		showError()
		return
	}
	defer zr.Close()
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if err := unzip(file); err != nil {
			showError()
		} else {
			lastEditName = config.NameFromPath(file.Name)
		}
	}
}

func (t *ConfView) onDelete() {
	c := t.CurrentConf()
	if c != nil {
		if walk.MsgBox(t.ConfListView.view.Form(), fmt.Sprintf("删除配置「%s」", c.Name), fmt.Sprintf("确定要删除配置「%s」吗? 此操作无法撤销。", c.Name), walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
			return
		}
		c.Delete()
		services.UninstallService(c.Name)
		if c.LogFile != "" {
			related, _ := utils.FindRelatedFiles(c.LogFile, "")
			utils.TryAlterFile(c.LogFile, "", false)
			for _, f := range related {
				utils.TryAlterFile(f, "", false)
			}
		}
		t.reloadConf()
		t.ConfListView.view.SetCurrentIndex(0)
	}
}

func (t *ConfView) onExport() {
	dlg := walk.FileDialog{
		Filter: "配置文件 (*.zip)|*.zip",
		Title:  "导出配置文件 (ZIP 压缩包)",
	}

	if ok, _ := dlg.ShowSave(t.ConfListView.view.Form()); !ok {
		return
	}

	if !strings.HasSuffix(dlg.FilePath, ".zip") {
		dlg.FilePath += ".zip"
	}

	allConfPath := make([]string, 0)
	for _, conf := range config.Configurations {
		allConfPath = append(allConfPath, conf.Path)
	}
	utils.ZipFiles(dlg.FilePath, allConfPath)
}

func (t *ConfView) Initialize() {
	t.ToolbarView.Initialize()
	t.ToolbarView.addAction.Triggered().Attach(func() {
		t.onEditConf(nil)
	})
	t.ToolbarView.addMenuAction.Triggered().Attach(func() {
		t.onEditConf(nil)
	})
	t.ToolbarView.importAction.Triggered().Attach(t.onImport)
	t.ToolbarView.deleteAction.Triggered().Attach(t.onDelete)
	t.ToolbarView.exportAction.Triggered().Attach(t.onExport)
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

var cachedListViewIconsForWidthAndState = make(map[widthAndState]*walk.Bitmap)

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
			Action{AssignTo: &t.editAction, Text: "编辑配置"},
			Action{AssignTo: &t.newAction, Text: "创建新配置"},
			Action{AssignTo: &t.importAction, Text: "从文件导入配置"},
			Action{AssignTo: &t.deleteAction, Text: "删除配置"},
		},
		StyleCell: func(style *walk.CellStyle) {
			row := style.Row()
			if row < 0 || row >= len(config.Configurations) {
				return
			}
			conf := config.Configurations[row]
			margin := t.view.IntFrom96DPI(1)
			bitmapWidth := t.view.IntFrom96DPI(16)
			cacheKey := widthAndState{bitmapWidth, conf.Status}
			if cacheValue, ok := cachedListViewIconsForWidthAndState[cacheKey]; ok {
				style.Image = cacheValue
				return
			}
			bitmap, err := walk.NewBitmapWithTransparentPixelsForDPI(walk.Size{bitmapWidth, bitmapWidth}, t.view.DPI())
			if err != nil {
				return
			}
			canvas, err := walk.NewCanvasFromImage(bitmap)
			if err != nil {
				return
			}
			bounds := walk.Rectangle{X: margin, Y: margin, Height: bitmapWidth - 2*margin, Width: bitmapWidth - 2*margin}
			err = canvas.DrawImageStretchedPixels(iconForState(conf.Status, 14), bounds)
			canvas.Dispose()
			if err != nil {
				return
			}
			cachedListViewIconsForWidthAndState[cacheKey] = bitmap
			style.Image = bitmap
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
	exportAction  *walk.Action

	toolbarDB *walk.DataBinder
}

func NewToolbarView(parent **walk.Composite) *ToolbarView {
	v := new(ToolbarView)
	v.parent = parent
	return v
}

func (t *ToolbarView) View() Widget {
	return Composite{
		DataBinder: DataBinder{AssignTo: &t.toolbarDB, DataSource: &struct {
			ConfSize func() int
		}{func() int {
			return len(config.Configurations)
		}}, Name: "conf"},
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
						OnTriggered:    func() {},
						Text:           "新建配置",
						Image:          loadSysIcon("shell32", 149, 16),
						Items: []MenuItem{
							Action{
								AssignTo: &t.addAction,
								Text:     "创建新配置",
								Image:    loadSysIcon("shell32", 205, 16),
							},
							Action{
								AssignTo: &t.importAction,
								Text:     "从文件导入",
								Image:    loadSysIcon("shell32", 132, 16),
							},
						},
					},
					Separator{},
					Action{
						Enabled:  Bind("conf.ConfSize != 0"),
						AssignTo: &t.deleteAction,
						Image:    loadSysIcon("shell32", 131, 16),
					},
					Separator{},
					Action{
						Enabled:  Bind("conf.ConfSize != 0"),
						AssignTo: &t.exportAction,
						Image:    loadSysIcon("imageres", -174, 16),
					},
				},
			},
		},
	}
}

func (t *ToolbarView) Initialize() {
	t.addAction.SetDefault(true)
	t.deleteAction.SetToolTip("删除配置")
	t.exportAction.SetToolTip("导出所有配置 (ZIP 压缩包)")
	t.view.ApplyDPI((*t.parent).DPI())
}

func (t *ToolbarView) fixWidth() {
	toolbarWidth := t.view.SizeHint().Width
	(*t.parent).SetMinMaxSizePixels(walk.Size{toolbarWidth, 0}, walk.Size{toolbarWidth, 0})
}
