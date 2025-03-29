package ui

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/layout"
	"github.com/koho/frpmgr/pkg/res"
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

func NewConfView(cfgList []*Conf) *ConfView {
	v := new(ConfView)
	v.model = NewConfListModel(cfgList)
	return v
}

func (cv *ConfView) View() Widget {
	moveUpCond := Bind("confView.SelectedCount == 1 && confView.CurrentIndex > 0")
	moveDownCond := Bind("confView.SelectedCount == 1 && confView.CurrentIndex < confView.ItemCount - 1")
	return Composite{
		AssignTo: &cv.Composite,
		Layout:   VBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			TableView{
				Name:                "confView",
				AssignTo:            &cv.listView,
				LastColumnStretched: true,
				HeaderHidden:        true,
				Columns:             []TableViewColumn{{}},
				Model:               cv.model,
				MultiSelection:      true,
				ContextMenuItems: []MenuItem{
					Action{
						AssignTo:    &cv.lsEditAction,
						Text:        i18n.Sprintf("Edit"),
						Enabled:     Bind("confView.SelectedCount == 1"),
						OnTriggered: cv.editCurrent,
					},
					Menu{
						Text:    i18n.Sprintf("Move"),
						Enabled: Bind("confView.SelectedCount == 1 && confView.ItemCount > 1"),
						Items: []MenuItem{
							Action{
								Text:    i18n.Sprintf("Up"),
								Enabled: moveUpCond,
								OnTriggered: func() {
									cv.onMove(-1)
								},
							},
							Action{
								Text:    i18n.Sprintf("Down"),
								Enabled: moveDownCond,
								OnTriggered: func() {
									cv.onMove(1)
								},
							},
							Action{
								Text:    i18n.Sprintf("To Top"),
								Enabled: moveUpCond,
								OnTriggered: func() {
									cv.onMove(-cv.listView.CurrentIndex())
								},
							},
							Action{
								Text:    i18n.Sprintf("To Bottom"),
								Enabled: moveDownCond,
								OnTriggered: func() {
									cv.onMove(len(cv.model.items) - cv.listView.CurrentIndex() - 1)
								},
							},
						},
					},
					Action{Text: i18n.Sprintf("Open File"),
						Enabled:     Bind("confView.SelectedCount == 1"),
						OnTriggered: func() { cv.onOpen(false) },
					},
					Action{
						Text:        i18n.Sprintf("Show in Folder"),
						Enabled:     Bind("confView.SelectedCount == 1"),
						OnTriggered: func() { cv.onOpen(true) },
					},
					Separator{},
					Action{Text: i18n.Sprintf("New Configuration"), OnTriggered: cv.editNew},
					Menu{
						Text:    i18n.Sprintf("Create a Copy"),
						Enabled: Bind("confView.SelectedCount == 1"),
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
						Enabled:     Bind("confView.SelectedCount == 1"),
						OnTriggered: cv.onCopyShareLink,
					},
					Action{
						Text:        i18n.Sprintf("Export All Configs to ZIP"),
						Enabled:     Bind("confView.ItemCount > 0"),
						OnTriggered: cv.onExport,
					},
					Separator{},
					Action{
						Enabled: Bind("confView.SelectedCount < confView.ItemCount"),
						Text:    i18n.Sprintf("Select all"),
						OnTriggered: func() {
							cv.listView.SetSelectedIndexes([]int{-1})
						},
					},
					Action{
						Text:        i18n.Sprintf("Delete"),
						Enabled:     Bind("confView.SelectedCount > 0"),
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
						style.TextColor = res.ColorBlue
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
								Image:       loadIcon(res.IconNewConf, 16),
								Items: []MenuItem{
									Action{
										AssignTo:    &cv.tbAddAction,
										Text:        i18n.Sprintf("Manual Settings"),
										Image:       loadIcon(res.IconCreate, 16),
										OnTriggered: cv.editNew,
									},
									Action{
										Text:        i18n.SprintfEllipsis("Import from File"),
										Image:       loadIcon(res.IconFileImport, 16),
										OnTriggered: cv.onFileImport,
									},
									Action{
										Text:        i18n.SprintfEllipsis("Import from URL"),
										Image:       loadIcon(res.IconURLImport, 16),
										OnTriggered: cv.onURLImport,
									},
									Action{
										Text:        i18n.Sprintf("Import from Clipboard"),
										Image:       loadIcon(res.IconClipboard, 16),
										OnTriggered: cv.onClipboardImport,
									},
								},
							},
							Separator{},
							Action{
								Enabled:     Bind("confView.SelectedCount > 0"),
								AssignTo:    &cv.tbDeleteAction,
								Image:       loadIcon(res.IconDelete, 16),
								OnTriggered: cv.onDelete,
							},
							Separator{},
							Action{
								Enabled:     Bind("confView.ItemCount > 0"),
								AssignTo:    &cv.tbExportAction,
								Image:       loadIcon(res.IconExport, 16),
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
	cv.listView.SetIgnoreNowhere(true)
	cv.listView.SetScrollbarOrientation(walk.Vertical)
	cv.listView.ItemActivated().Attach(cv.editCurrent)
	cv.lsEditAction.SetDefault(true)
	cv.listView.CurrentIndexChanged().Attach(func() {
		if idx := cv.listView.CurrentIndex(); idx >= 0 && idx < len(cv.model.items) {
			setCurrentConf(cv.model.items[idx])
		} else {
			setCurrentConf(nil)
		}
	})
	cv.listView.SelectedIndexesChanged().Attach(func() {
		if indexes := cv.listView.SelectedIndexes(); len(indexes) == 1 {
			cv.listView.SetCurrentIndex(indexes[0])
		} else if len(indexes) == 0 {
			cv.listView.SetCurrentIndex(-1)
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
	cv.onEditConf(getCurrentConf(), false)
}

func (cv *ConfView) editNew() {
	cv.onEditConf(NewConf("", newDefaultClientConfig()), true)
}

func (cv *ConfView) editCopy(all bool) {
	if conf := getCurrentConf(); conf != nil {
		cv.onEditConf(NewConf("", conf.Data.Copy(all)), true)
	}
}

func (cv *ConfView) fixWidthToToolbarWidth() {
	toolbarWidth := cv.toolbar.SizeHint().Width
	cv.SetMinMaxSizePixels(walk.Size{Width: toolbarWidth}, walk.Size{Width: toolbarWidth})
}

func (cv *ConfView) onEditConf(conf *Conf, create bool) {
	dlg := NewEditClientDialog(conf.Data.(*config.ClientConfig), create)
	if result, _ := dlg.Run(cv.Form()); result == walk.DlgCmdOK {
		if create {
			cv.model.Add(conf)
			cv.listView.SetCurrentIndex(cv.model.RowCount() - 1)
		} else {
			if i := cv.listView.CurrentIndex(); i >= 0 {
				cv.model.PublishRowsChanged(i, i)
			}
			// Reset current conf
			confDB.Reset()
		}
		// Commit the config
		commitConf(conf, runFlagAuto)
	}
}

func (cv *ConfView) onURLImport() {
	dlg := NewURLImportDialog()
	if result, err := dlg.Run(cv.Form()); err != nil || result != walk.DlgCmdOK {
		return
	}
	var cfgList []*Conf
	cv.importConfig(func() (total, imported int) {
		for _, item := range dlg.Items {
			if item.Zip {
				subList, subTotal, subImported := cv.importZip(item.Filename, item.Data)
				total += subTotal
				imported += subImported
				cfgList = append(cfgList, subList...)
			} else {
				total++
				conf, err := config.UnmarshalClientConf(item.Data)
				if err != nil {
					showError(err, cv.Form())
					continue
				}
				if conf.Name() == "" {
					conf.ClientCommon.Name = util.FileNameWithoutExt(item.Filename)
				}
				cfg := NewConf("", conf)
				if err = cfg.Save(); err != nil {
					showError(err, cv.Form())
					continue
				}
				cfgList = append(cfgList, cfg)
				imported++
			}
		}
		return
	})
	cv.model.Add(cfgList...)
}

func (cv *ConfView) onFileImport() {
	dlg := walk.FileDialog{
		Filter: res.FilterConfig + res.FilterAllFiles,
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
	}
}

func (cv *ConfView) ImportFiles(files []string) {
	var cfgList []*Conf
	cv.importConfig(func() (total, imported int) {
		for _, path := range files {
			if dir, err := util.IsDirectory(path); err != nil || dir {
				continue
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".zip" {
				subList, subTotal, subImported := cv.importZip(path, nil)
				total += subTotal
				imported += subImported
				cfgList = append(cfgList, subList...)
			} else if slices.Contains(res.SupportedConfigFormats, ext) {
				total++
				conf, err := config.UnmarshalClientConf(path)
				if err != nil {
					showError(err, cv.Form())
					continue
				}
				if conf.Name() == "" {
					conf.ClientCommon.Name = util.FileNameWithoutExt(path)
				}
				cfg := NewConf("", conf)
				if err = cfg.Save(); err != nil {
					showError(err, cv.Form())
					continue
				}
				cfgList = append(cfgList, cfg)
				imported++
			}
		}
		return
	})
	cv.model.Add(cfgList...)
}

func (cv *ConfView) importZip(path string, data []byte) (cfgList []*Conf, total, imported int) {
	importFile := func(file *zip.File) (*Conf, error) {
		fr, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer fr.Close()
		src, err := io.ReadAll(fr)
		if err != nil {
			return nil, err
		}
		conf, err := config.UnmarshalClientConf(src)
		if err != nil {
			return nil, err
		}
		if conf.Name() == "" {
			conf.ClientCommon.Name = util.FileNameWithoutExt(file.Name)
		}
		cfg := NewConf("", conf)
		if err = cfg.Save(); err != nil {
			return nil, err
		}
		return cfg, nil
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
		if !slices.Contains(res.SupportedConfigFormats, strings.ToLower(filepath.Ext(file.Name))) {
			continue
		}
		total++
		if cfg, err := importFile(file); err == nil {
			imported++
			cfgList = append(cfgList, cfg)
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
	// Check for a share link
	if strings.HasPrefix(text, res.ShareLinkScheme) {
		text = strings.TrimPrefix(text, res.ShareLinkScheme)
		content, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			showError(err, cv.Form())
			return
		}
		text = string(content)
	}
	conf, err := config.UnmarshalClientConf([]byte(text))
	if err != nil {
		showError(err, cv.Form())
		return
	}
	cv.onEditConf(NewConf("", conf), true)
}

func (cv *ConfView) onCopyShareLink() {
	if conf := getCurrentConf(); conf != nil {
		content, err := os.ReadFile(conf.Path)
		if err != nil {
			showError(err, cv.Form())
			return
		}
		walk.Clipboard().SetText(res.ShareLinkScheme + base64.StdEncoding.EncodeToString(content))
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
	indexes := cv.listView.SelectedIndexes()
	count := len(indexes)
	if count == 0 {
		return
	}
	if count == 1 {
		if conf := cv.model.items[indexes[0]]; conf != nil {
			if walk.MsgBox(cv.Form(), i18n.Sprintf("Delete config \"%s\"", conf.Name()),
				i18n.Sprintf("Are you sure you would like to delete config \"%s\"?", conf.Name()),
				walk.MsgBoxYesNo|walk.MsgBoxIconWarning) == walk.DlgCmdNo {
				return
			}
			if conf.State == consts.StateStarting {
				showErrorMessage(cv.Form(), i18n.Sprintf("Delete config \"%s\"", conf.Name()),
					i18n.Sprintf("The config is currently locked."))
				return
			}
			if err := conf.Delete(); err != nil {
				showError(err, cv.Form())
				return
			}
			cv.model.Remove(indexes[0])
		}
	} else {
		if walk.MsgBox(cv.Form(), i18n.Sprintf("Delete %d configs", count),
			i18n.Sprintf("Are you sure that you want to delete these %d configs?", count),
			walk.MsgBoxYesNo|walk.MsgBoxIconWarning) == walk.DlgCmdNo {
			return
		}
		var succeeded []int
		for _, idx := range indexes {
			conf := cv.model.items[idx]
			if conf.State == consts.StateStarting {
				continue
			}
			if err := conf.Delete(); err == nil {
				succeeded = append(succeeded, idx)
			}
		}
		succeededCount := len(succeeded)
		if count != succeededCount {
			failedCount := count - succeededCount
			walk.MsgBox(cv.Form(), i18n.Sprintf("Delete %d configs", count),
				i18n.Sprintf("%d succeeded, %d failed.", succeededCount, failedCount),
				walk.MsgBoxOK|walk.MsgBoxIconInformation)
		}
		if succeededCount > 0 {
			cv.model.Remove(succeeded...)
		}
	}
	// Restore the first selected index.
	if i := min(indexes[0], cv.model.RowCount()-1); i >= 0 {
		cv.listView.SetCurrentIndex(i)
	}
}

func (cv *ConfView) onExport() {
	dlg := walk.FileDialog{
		Filter: res.FilterZip,
		Title:  i18n.Sprintf("Export All Configs to ZIP"),
	}

	if ok, _ := dlg.ShowSave(cv.Form()); !ok {
		return
	}

	if !strings.HasSuffix(dlg.FilePath, ".zip") {
		dlg.FilePath += ".zip"
	}
	files := lo.SliceToMap(cv.model.List(), func(conf *Conf) (string, string) {
		return conf.Path, conf.Name() + conf.Data.Ext()
	})
	if err := util.ZipFiles(dlg.FilePath, files); err != nil {
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
			stunServer = consts.DefaultSTUNServer
		}
	}
	if _, err := NewNATDiscoveryDialog(stunServer).Run(cv.Form()); err != nil {
		showError(err, cv.Form())
	}
}

func (cv *ConfView) onMove(delta int) {
	curIdx := cv.listView.CurrentIndex()
	if curIdx < 0 || curIdx >= len(cv.model.items) {
		return
	}
	targetIdx := curIdx + delta
	if targetIdx < 0 || targetIdx >= len(cv.model.items) {
		return
	}
	cv.model.Move(curIdx, targetIdx)
	cv.listView.SetCurrentIndex(targetIdx)
}
