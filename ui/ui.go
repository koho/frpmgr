package ui

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/samber/lo"
	"golang.org/x/sys/windows"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
)

const AppName = "FRP Manager"

var AppLocalName = i18n.Sprintf(AppName)

func init() {
	walk.SetTranslationFunc(func(source string, context ...string) string {
		translation := i18n.Sprintf(source)
		s1 := strings.ReplaceAll(translation, "%!f(MISSING)", "%.f")
		return strings.ReplaceAll(s1, "%!f(BADINDEX)", "%.f")
	})
}

type FRPManager struct {
	*walk.MainWindow

	tabs      *walk.TabWidget
	confPage  *ConfPage
	logPage   *LogPage
	prefPage  *PrefPage
	aboutPage *AboutPage
}

func RunUI() error {
	var err error
	// Make sure the config directory exists.
	if err = os.MkdirAll(PathOfConf(""), os.ModePerm); err != nil {
		return err
	}
	cfgList, err := loadAllConfs()
	if err != nil {
		return err
	}
	if appConf.Password != "" {
		if r, err := NewValidateDialog().Run(); err != nil || r != win.IDOK {
			return err
		}
	}
	fm := new(FRPManager)
	fm.confPage = NewConfPage(cfgList)
	fm.logPage, err = NewLogPage()
	if err != nil {
		return err
	}
	fm.prefPage = NewPrefPage()
	fm.aboutPage = NewAboutPage()
	mw := MainWindow{
		Icon:       loadLogoIcon(32),
		AssignTo:   &fm.MainWindow,
		Title:      AppLocalName,
		Persistent: true,
		Visible:    false,
		Layout:     VBox{Margins: Margins{Left: 5, Top: 5, Right: 5, Bottom: 5}},
		Font:       res.TextRegular,
		Children: []Widget{
			TabWidget{
				AssignTo: &fm.tabs,
				Pages: []TabPage{
					fm.confPage.Page(),
					fm.logPage.Page(),
					fm.prefPage.Page(),
					fm.aboutPage.Page(),
				},
			},
		},
		OnDropFiles: fm.confPage.confView.ImportFiles,
	}
	if err = mw.Create(); err != nil {
		return err
	}
	// Initialize child pages
	fm.confPage.OnCreate()
	fm.logPage.OnCreate()
	fm.prefPage.OnCreate()
	fm.aboutPage.OnCreate()
	// Minimum window height.
	confPageHeight := fm.confPage.SizeHint().Height
	maxPageHeight := max(
		confPageHeight,
		fm.logPage.SizeHint().Height,
		fm.prefPage.SizeHint().Height,
		fm.aboutPage.SizeHint().Height,
	)
	// Resize window
	margins := fm.Layout().Margins()
	size := fm.tabs.SizeHint()
	bias := fm.confPage.detailView.sizeBias()
	size.Width += bias.Width + walk.IntFrom96DPI(margins.HNear+margins.HFar, fm.DPI())
	size.Height += bias.Height + walk.IntFrom96DPI(margins.VNear+margins.VFar, fm.DPI()) - maxPageHeight + confPageHeight
	fm.SetClientSizePixels(size)
	fm.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		// Save window state.
		var wp win.WINDOWPLACEMENT
		wp.Length = uint32(unsafe.Sizeof(wp))
		if win.GetWindowPlacement(fm.Handle(), &wp) {
			appConf.Position = []int32{wp.RcNormalPosition.Left, wp.RcNormalPosition.Top}
			saveAppConfig()
		}
	})
	// Restore window state.
	if len(appConf.Position) > 1 {
		bounds := fm.BoundsPixels()
		wp := win.WINDOWPLACEMENT{
			Flags:         0,
			ShowCmd:       win.SW_SHOWNORMAL,
			PtMinPosition: win.POINT{X: -1, Y: -1},
			PtMaxPosition: win.POINT{X: -1, Y: -1},
			RcNormalPosition: win.RECT{
				Left:   appConf.Position[0],
				Top:    appConf.Position[1],
				Right:  appConf.Position[0] + int32(bounds.Width),
				Bottom: appConf.Position[1] + int32(bounds.Height),
			},
		}
		wp.Length = uint32(unsafe.Sizeof(wp))
		win.SetWindowPlacement(fm.Handle(), &wp)
	}
	fm.Show()
	fm.Run()
	fm.confPage.Close()
	fm.logPage.Close()
	return nil
}

func showError(err error, owner walk.Form) bool {
	if err == nil {
		return false
	}
	showErrorMessage(owner, "", err.Error())
	return true
}

func showErrorMessage(owner walk.Form, title, message string) {
	if title == "" {
		title = AppLocalName
	}
	walk.MsgBox(owner, title, message, walk.MsgBoxIconError)
}

func showWarningMessage(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconWarning)
}

func showInfoMessage(owner walk.Form, title, message string) {
	if title == "" {
		title = AppLocalName
	}
	walk.MsgBox(owner, title, message, walk.MsgBoxIconInformation)
}

// openPath opens a file or url with default application
func openPath(path string) {
	if path == "" {
		return
	}
	win.ShellExecute(0, nil, windows.StringToUTF16Ptr(path), nil, nil, win.SW_SHOWNORMAL)
}

// openFolder opens the explorer and select the given file
func openFolder(path string) {
	if path == "" {
		return
	}
	if absPath, err := filepath.Abs(path); err == nil {
		win.ShellExecute(0, nil, windows.StringToUTF16Ptr(`explorer`),
			windows.StringToUTF16Ptr(`/select,`+absPath), nil, win.SW_SHOWNORMAL)
	}
}

// openFileDialog shows a file dialog to choose file or directory and sends the selected path to the LineEdit view
func openFileDialog(receiver *walk.LineEdit, title string, filter string, file bool) error {
	dlg := walk.FileDialog{
		Filter: filter + res.FilterAllFiles,
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

// calculateHeadColumnTextWidth returns the estimated display width of the first column
func calculateHeadColumnTextWidth(widgets []Widget, columns int) int {
	maxLen := 0
	for i := range widgets {
		if label, ok := widgets[i].(Label); ok && i%columns == 0 {
			if textLen := calculateStringWidth(label.Text.(string)); textLen > maxLen {
				maxLen = textLen
			}
		}
	}
	return maxLen + 5
}

// calculateStringWidth returns the estimated display width of the given string
func calculateStringWidth(str string) int {
	return lo.Sum(lo.Map(util.RuneSizeInString(str), func(s int, i int) int {
		// For better estimation, reduce size for non-ascii character
		if s > 1 {
			return s - 1
		}
		return s
	})) * 6
}
