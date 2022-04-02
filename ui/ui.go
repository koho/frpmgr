package ui

import (
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/services"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"math/rand"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// View is the interface that must be implemented to build a Widget.
type View interface {
	// View should define widget in declarative way, and will
	// be called by the parent widget.
	View() Widget
	// OnCreate will be called after the creation of views. The
	// view reference should be available now.
	OnCreate()
	// Invalidate should be called if data that view relying on
	// is changed. The view should be updated with new data.
	Invalidate()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

type FRPManager struct {
	*walk.MainWindow

	tabs      *walk.TabWidget
	confPage  *ConfPage
	logPage   *LogPage
	aboutPage *AboutPage
}

func RunUI() error {
	if err := loadAllConfs(); err != nil {
		return err
	}
	fm := new(FRPManager)
	fm.confPage = NewConfPage()
	fm.logPage = NewLogPage()
	fm.aboutPage = NewAboutPage()
	mw := MainWindow{
		Icon:       loadLogoIcon(32),
		AssignTo:   &fm.MainWindow,
		Title:      "FRP 管理器",
		Persistent: true,
		MinSize:    Size{640, 400},
		Size:       Size{986, 525},
		Layout:     VBox{Margins: Margins{5, 5, 5, 5}},
		Font:       consts.TextRegular,
		Children: []Widget{
			TabWidget{
				AssignTo: &fm.tabs,
				Pages: []TabPage{
					fm.confPage.Page(),
					fm.logPage.Page(),
					fm.aboutPage.Page(),
				},
			},
		},
	}
	if err := mw.Create(); err != nil {
		return err
	}
	// Initialize child pages
	fm.confPage.OnCreate()
	fm.logPage.OnCreate()
	fm.aboutPage.OnCreate()
	fm.Run()
	services.Cleanup()
	return nil
}

func showError(err error, owner walk.Form) bool {
	if err == nil {
		return false
	}
	showErrorMessage(owner, "错误", err.Error())
	return true
}

func showErrorMessage(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconError)
}

func showWarningMessage(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconWarning)
}

func showInfoMessage(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconInformation)
}

// openPath opens a file or url with default application
func openPath(path string) {
	if path == "" {
		return
	}
	openCmd := exec.Command("cmd", "/c", "start", "", path)
	openCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	openCmd.Start()
}

// openFolder opens the explorer and select the given file
func openFolder(path string) {
	if path == "" {
		return
	}
	if absPath, err := filepath.Abs(path); err == nil {
		exec.Command(`explorer`, `/select,`, absPath).Run()
	}
}

// openFileDialog shows a file dialog to choose file or directory and sends the selected path to the LineEdit view
func openFileDialog(receiver *walk.LineEdit, title string, filter string, file bool) error {
	dlg := walk.FileDialog{
		Filter: filter + consts.FilterAllFiles,
		Title:  title,
	}
	var ok bool
	var err error
	if file {
		ok, err = dlg.ShowOpen(receiver.Form())
	} else {
		ok, err = dlg.ShowBrowseFolder(receiver.Form())
	}
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return receiver.SetText(strings.ReplaceAll(dlg.FilePath, "\\", "/"))
}
