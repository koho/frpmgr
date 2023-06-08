package ui

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"

	frpConfig "github.com/fatedier/frp/pkg/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/thoas/go-funk"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/layout"
	"github.com/koho/frpmgr/pkg/util"
)

type ConfView struct {
	*walk.Composite

	// List view
	listView     *walk.TableView
	lsEditAction *walk.Action
	model        *ConfListModel

	// Toolbar view
	toolbar        *walk.ToolBar
	tbAddAction    *walk.Action
	tbDeleteAction *walk.Action
	tbExportAction *walk.Action
}

var cachedListViewIconsForWidthAndState = make(map[widthAndState]*walk.Bitmap)

func NewConfView() *ConfView {
	v := new(ConfView)
	v.model = NewConfListModel(confList)
	return v
}

func (cv *ConfView) View() Widget {
	return Composite{
		AssignTo: &cv.Composite,
		Layout:   VBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			TableView{
				AssignTo:            &cv.listView,
				LastColumnStretched: true,
				HeaderHidden:        true,
				Columns:             []TableViewColumn{{DataMember: "DisplayName"}},
				Model:               cv.model,
				ContextMenuItems: []MenuItem{
					Action{
						AssignTo:    &cv.lsEditAction,
						Text:        i18n.Sprintf("Edit"),
						Enabled:     Bind("conf.Selected"),
						OnTriggered: cv.editCurrent,
					},
					Action{Text: i18n.Sprintf("Open File"),
						Enabled:     Bind("conf.Selected"),
						OnTriggered: func() { cv.onOpen(false) },
					},
					Action{
						Text:        i18n.Sprintf("Show in Folder"),
						Enabled:     Bind("conf.Selected"),
						OnTriggered: func() { cv.onOpen(true) },
					},
					Separator{},
					Action{Text: i18n.Sprintf("New Configuration"), OnTriggered: cv.editNew},
					Menu{
						Text:    i18n.Sprintf("Create a Copy"),
						Enabled: Bind("conf.Selected"),
						Items: []MenuItem{
							Action{Text: i18n.Sprintf("All"), OnTriggered: func() { cv.editCopy(true) }},
							Action{Text: i18n.Sprintf("Common Only"), OnTriggered: func() { cv.editCopy(false) }},
						},
					},
					Menu{
						Text: i18n.Sprintf("Import Config"),
						Items: []MenuItem{
							Action{Text: i18n.SprintfEllipsis("Import from File"), OnTriggered: cv.onFileImport},
							Action{Text: i18n.SprintfEllipsis("Import from URL"), OnTriggered: cv.onURLImport},
							Action{Text: i18n.Sprintf("Import from Clipboard"), OnTriggered: cv.onClipboardImport},
						},
					},
					Separator{},
					Action{
						Text:        i18n.Sprintf("NAT Discovery"),
						OnTriggered: cv.onNATDiscovery,
					},
					Separator{},
					Action{
						Text:        i18n.Sprintf("Copy Share Link"),
						Enabled:     Bind("conf.Selected"),
						OnTriggered: cv.onCopyShareLink,
					},
					Action{
						Text:        i18n.Sprintf("Export All Configs to ZIP"),
						Enabled:     Bind("conf.Selected"),
						OnTriggered: cv.onExport,
					},
					Separator{},
					Action{
						Text:        i18n.Sprintf("Delete"),
						Enabled:     Bind("conf.Selected"),
						OnTriggered: cv.onDelete,
					},
				},
				StyleCell: func(style *walk.CellStyle) {
					row := style.Row()
					if row < 0 || row >= len(cv.model.items) {
						return
					}
					conf := cv.model.items[row]
					if !conf.Data.AutoStart() {
						style.TextColor = consts.ColorBlue
					}
					margin := cv.listView.IntFrom96DPI(1)
					bitmapWidth := cv.listView.IntFrom96DPI(16)
					cacheKey := widthAndState{bitmapWidth, conf.State}
					if cacheValue, ok := cachedListViewIconsForWidthAndState[cacheKey]; ok {
						style.Image = cacheValue
						return
					}
					bitmap, err := walk.NewBitmapWithTransparentPixelsForDPI(walk.Size{Width: bitmapWidth, Height: bitmapWidth}, cv.listView.DPI())
					if err != nil {
						return
					}
					canvas, err := walk.NewCanvasFromImage(bitmap)
					if err != nil {
						return
					}
					bounds := walk.Rectangle{X: margin, Y: margin, Height: bitmapWidth - 2*margin, Width: bitmapWidth - 2*margin}
					err = canvas.DrawImageStretchedPixels(iconForState(conf.State, 14), bounds)
					canvas.Dispose()
					if err != nil {
						return
					}
					cachedListViewIconsForWidthAndState[cacheKey] = bitmap
					style.Image = bitmap
				},
				LayoutItem: func() walk.LayoutItem {
					return layout.NewGreedyLayoutItem(walk.Vertical)
				},
			},
			Composite{
				Layout:          HBox{MarginsZero: true, SpacingZero: true},
				DoubleBuffering: true,
				Children: []Widget{
					ToolBar{
						AssignTo:    &cv.toolbar,
						ButtonStyle: ToolBarButtonImageBeforeText,
						Orientation: Horizontal,
						Items: []MenuItem{
							Menu{
								OnTriggered: cv.editNew,
								Text:        i18n.Sprintf("New Config"),
								Image:       loadSysIcon("shell32", consts.IconNewConf, 16),
								Items: []MenuItem{
									Action{
										AssignTo:    &cv.tbAddAction,
										Text:        i18n.Sprintf("Manual Settings"),
										Image:       loadSysIcon("shell32", consts.IconCreate, 16),
										OnTriggered: cv.editNew,
									},
									Action{
										Text:        i18n.SprintfEllipsis("Import from File"),
										Image:       loadSysIcon("shell32", consts.IconFileImport, 16),
										OnTriggered: cv.onFileImport,
									},
									Action{
										Text:        i18n.SprintfEllipsis("Import from URL"),
										Image:       loadSysIcon("imageres", consts.IconURLImport, 16),
										OnTriggered: cv.onURLImport,
									},
									Action{
										Text:        i18n.Sprintf("Import from Clipboard"),
										Image:       loadSysIcon("shell32", consts.IconClipboard, 16),
										OnTriggered: cv.onClipboardImport,
									},
								},
							},
							Separator{},
							Action{
								Enabled:     Bind("conf.Selected"),
								AssignTo:    &cv.tbDeleteAction,
								Image:       loadSysIcon("shell32", consts.IconDelete, 16),
								OnTriggered: cv.onDelete,
							},
							Separator{},
							Action{
								Enabled:     Bind("conf.Selected"),
								AssignTo:    &cv.tbExportAction,
								Image:       loadSysIcon("imageres", consts.IconExport, 16),
								OnTriggered: cv.onExport,
							},
						},
					},
				},
			},
		},
	}
}

func (cv *ConfView) OnCreate() {
	// Setup config list view
	cv.listView.ItemActivated().Attach(cv.editCurrent)
	cv.lsEditAction.SetDefault(true)
	cv.listView.CurrentIndexChanged().Attach(func() {
		if idx := cv.listView.CurrentIndex(); idx >= 0 && idx < len(cv.model.items) {
			setCurrentConf(cv.model.items[idx])
		} else {
			setCurrentConf(nil)
		}
	})
	// Setup toolbar
	cv.tbAddAction.SetDefault(true)
	cv.tbDeleteAction.SetToolTip(i18n.Sprintf("Delete"))
	cv.tbExportAction.SetToolTip(i18n.Sprintf("Export All Configs to ZIP"))
	cv.toolbar.ApplyDPI(cv.DPI())
	cv.fixWidthToToolbarWidth()
	cv.toolbar.SizeChanged().Attach(cv.fixWidthToToolbarWidth)
}

func (cv *ConfView) editCurrent() {
	cv.onEditConf(getCurrentConf(), "")
}

func (cv *ConfView) editNew() {
	cv.onEditConf(nil, "")
}

func (cv *ConfView) editCopy(all bool) {
	if conf := getCurrentConf(); conf != nil {
		cv.onEditConf(NewConf("", conf.Data.Copy(all)), "")
	}
}

func (cv *ConfView) fixWidthToToolbarWidth() {
	toolbarWidth := cv.toolbar.SizeHint().Width
	cv.SetMinMaxSizePixels(walk.Size{Width: toolbarWidth}, walk.Size{Width: toolbarWidth})
}

func (cv *ConfView) onEditConf(conf *Conf, name string) {
	dlg := NewEditClientDialog(conf, name)
	if dlg == nil {
		return
	}
	if res, _ := dlg.Run(cv.Form()); res == walk.DlgCmdOK {
		if dlg.Added {
			// Created new config
			// The list is resorted, we should select by name
			cv.reset(dlg.Conf.Name)
		} else {
			cv.model.Items()
			cv.listView.Invalidate()
			// Reset current conf
			confDB.Reset()
		}
		// Commit the config
		flag := runFlagAuto
		if dlg.ShouldRestart {
			flag = runFlagForceStart
		}
		commitConf(dlg.Conf, flag)
	}
}

func (cv *ConfView) onURLImport() {
	dlg := NewURLImportDialog()
	if result, err := dlg.Run(cv.Form()); err != nil || result != walk.DlgCmdOK {
		return
	}
	cv.importConfig(func() (total, imported int) {
		for _, item := range dlg.Items {
			if item.Zip {
				subTotal, subImported := cv.importZip(item.Filename, item.Data, item.Rename)
				total += subTotal
				imported += subImported
			} else {
				total++
				if newPath, ok := cv.checkConfName(item.Filename, item.Rename); ok {
					conf, err := config.UnmarshalClientConfFromIni(item.Data)
					if err != nil {
						showError(err, cv.Form())
						continue
					}
					if err = os.WriteFile(newPath, item.Data, 0666); err != nil {
						showError(err, cv.Form())
						continue
					}
					addConf(NewConf(newPath, conf))
					imported++
				}
			}
		}
		return
	})
}

func (cv *ConfView) checkConfName(filename string, rename bool) (string, bool) {
	suffix := ""
checkName:
	newPath := PathOfConf(util.AddFileSuffix(filename, suffix))
	if _, err := os.Stat(newPath); err == nil {
		if rename {
			suffix = "_" + funk.RandomString(4)
			goto checkName
		}
		return newPath, false
	}
	return newPath, true
}

func (cv *ConfView) onFileImport() {
	dlg := walk.FileDialog{
		Filter: consts.FilterConfig + consts.FilterAllFiles,
		Title:  i18n.Sprintf("Import from File"),
	}

	if ok, _ := dlg.ShowOpenMultiple(cv.Form()); !ok {
		return
	}
	cv.ImportFiles(dlg.FilePaths)
}

func (cv *ConfView) importConfig(f func() (int, int)) {
	if total, imported := f(); imported > 0 {
		showInfoMessage(cv.Form(),
			i18n.Sprintf("Import Config"),
			i18n.Sprintf("Imported %d of %d configs.", imported, total))
		// Reselect the current config after refreshing list view
		if conf := getCurrentConf(); conf != nil {
			cv.reset(conf.Name)
		} else {
			cv.Invalidate()
		}
	}
}

func (cv *ConfView) ImportFiles(files []string) {
	cv.importConfig(func() (total, imported int) {
		for _, path := range files {
			if dir, err := util.IsDirectory(path); err != nil || dir {
				continue
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".ini":
				total++
				newPath, ok := cv.checkConfName(path, false)
				if !ok {
					baseName, _ := util.SplitExt(newPath)
					showWarningMessage(cv.Form(),
						i18n.Sprintf("Import Config"),
						i18n.Sprintf("Another config already exists with the name \"%s\".", baseName))
					continue
				}
				// Verify config before copying file
				conf, err := config.UnmarshalClientConfFromIni(path)
				if err != nil {
					showError(err, cv.Form())
					continue
				}
				if _, err = util.CopyFile(path, newPath); err != nil {
					showError(err, cv.Form())
					continue
				}
				addConf(NewConf(newPath, conf))
				imported++
			case ".zip":
				subTotal, subImported := cv.importZip(path, nil, false)
				total += subTotal
				imported += subImported
			}
		}
		return
	})
}

func (cv *ConfView) importZip(path string, data []byte, rename bool) (total, imported int) {
	importFile := func(file *zip.File, dst string) error {
		fr, err := file.Open()
		if err != nil {
			return err
		}
		defer fr.Close()
		src, err := io.ReadAll(fr)
		if err != nil {
			return err
		}
		conf, err := config.UnmarshalClientConfFromIni(src)
		if err != nil {
			return err
		}
		fw, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer fw.Close()
		if _, err = fw.Write(src); err != nil {
			return err
		}
		addConf(NewConf(dst, conf))
		return nil
	}
	var zr *zip.Reader
	var err error
	if data == nil {
		// Read from the given file path
		var fr *zip.ReadCloser
		if fr, err = zip.OpenReader(path); err == nil {
			zr = &fr.Reader
			defer fr.Close()
		}
	} else {
		// Read from the memory buffer
		zr, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	}
	if err != nil {
		showErrorMessage(cv.Form(), "", i18n.Sprintf("The file \"%s\" is not a valid ZIP file.", path))
		return
	}
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(file.Name)) != ".ini" {
			continue
		}
		total++
		if dstPath, ok := cv.checkConfName(file.Name, rename); ok {
			if err = importFile(file, dstPath); err == nil {
				imported++
			}
		}
	}
	return
}

func (cv *ConfView) onClipboardImport() {
	text, err := walk.Clipboard().Text()
	if err != nil {
		return
	}
	if text = strings.TrimSpace(text); text == "" {
		return
	}
	var name string
	// Check for a share link
	if strings.HasPrefix(text, consts.ShareLinkScheme) {
		text = strings.TrimPrefix(text, consts.ShareLinkScheme)
		content, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			showError(err, cv.Form())
			return
		}
		text = string(content)
		// Extract the config name in the first line
		if i := bytes.IndexByte(content, '\n'); i > 0 && content[0] == '#' {
			name = string(bytes.TrimSpace(content[1:i]))
		}
	}
	conf, err := config.UnmarshalClientConfFromIni([]byte(text))
	if err != nil {
		showError(err, cv.Form())
		return
	}
	cv.onEditConf(NewConf("", conf), name)
}

func (cv *ConfView) onCopyShareLink() {
	if conf := getCurrentConf(); conf != nil {
		content, err := os.ReadFile(conf.Path)
		if err != nil {
			showError(err, cv.Form())
			return
		}
		// Insert the config name in the first line
		content = append([]byte("# "+conf.Name+"\n"), content...)
		walk.Clipboard().SetText(consts.ShareLinkScheme + base64.StdEncoding.EncodeToString(content))
	}
}

func (cv *ConfView) onOpen(folder bool) {
	if conf := getCurrentConf(); conf != nil {
		if path, err := filepath.Abs(conf.Path); err == nil {
			if folder {
				openFolder(path)
			} else {
				openPath(path)
			}
		}
	}
}

func (cv *ConfView) onDelete() {
	if conf := getCurrentConf(); conf != nil {
		if walk.MsgBox(cv.Form(), i18n.Sprintf("Delete config \"%s\"", conf.Name),
			i18n.Sprintf("Are you sure you would like to delete config \"%s\"?", conf.Name),
			walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
			return
		}
		if !hasConf(conf.Name) {
			return
		}
		// Fully delete config
		if removed, err := conf.Delete(); err != nil {
			showError(err, cv.Form())
			return
		} else if removed {
			cv.Invalidate()
		}
	}
}

func (cv *ConfView) onExport() {
	dlg := walk.FileDialog{
		Filter: consts.FilterZip,
		Title:  i18n.Sprintf("Export All Configs to ZIP"),
	}

	if ok, _ := dlg.ShowSave(cv.Form()); !ok {
		return
	}

	if !strings.HasSuffix(dlg.FilePath, ".zip") {
		dlg.FilePath += ".zip"
	}

	files := funk.Map(confList, func(conf *Conf) string {
		return conf.Path
	})
	if err := util.ZipFiles(dlg.FilePath, files.([]string)); err != nil {
		showError(err, cv.Form())
	}
}

func (cv *ConfView) onNATDiscovery() {
	var stunServer string
	// Try to use the address in config first
	if conf := getCurrentConf(); conf != nil {
		stunServer = conf.Data.GetSTUNServer()
	}
	if stunServer == "" {
		if appConf.Defaults.NatHoleSTUNServer != "" {
			stunServer = appConf.Defaults.NatHoleSTUNServer
		} else {
			stunServer = frpConfig.GetDefaultClientConf().NatHoleSTUNServer
		}
	}
	if _, err := NewNATDiscoveryDialog(stunServer).Run(cv.Form()); err != nil {
		showError(err, cv.Form())
	}
}

// reset config listview with selected name
func (cv *ConfView) reset(selectName string) {
	// Make sure `sel` is a valid index
	sel := funk.MaxInt([]int{cv.listView.CurrentIndex(), 0})
	// Refresh the whole config list
	// The confList will be sorted
	cv.model = NewConfListModel(confList)
	cv.listView.SetModel(cv.model)
	if selectName != "" {
		if idx := funk.IndexOf(cv.model.items, func(conf *Conf) bool { return conf.Name == selectName }); idx >= 0 {
			sel = idx
		}
	}
	// Make sure the final selected index is valid
	if selectIdx := funk.MinInt([]int{sel, len(cv.model.items) - 1}); selectIdx >= 0 {
		cv.listView.SetCurrentIndex(selectIdx)
	}
}

// Invalidate conf view with last selected index
func (cv *ConfView) Invalidate() {
	cv.reset("")
}
