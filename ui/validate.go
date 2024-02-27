package ui

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/sec"
)

const stmSetIcon = 0x0170

// ValidateDialog validates the administration password.
type ValidateDialog struct {
	hIcon win.HICON
}

func NewValidateDialog() *ValidateDialog {
	return new(ValidateDialog)
}

func (vd *ValidateDialog) Run() (int, error) {
	name, err := syscall.UTF16PtrFromString(consts.DialogValidate)
	if err != nil {
		return -1, err
	}
	defer func() {
		if vd.hIcon != 0 {
			win.DestroyIcon(vd.hIcon)
			vd.hIcon = 0
		}
	}()
	return win.DialogBoxParam(win.GetModuleHandle(nil), name, 0, syscall.NewCallback(vd.proc), 0), nil
}

func (vd *ValidateDialog) proc(h win.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case win.WM_INITDIALOG:
		SetWindowText(h, fmt.Sprintf("%s - %s", i18n.Sprintf("Enter Password"), AppLocalName))
		SetWindowText(win.GetDlgItem(h, consts.DialogTitle), i18n.Sprintf("You must enter an administration password to operate the %s.", AppLocalName))
		SetWindowText(win.GetDlgItem(h, consts.DialogStatic1), i18n.Sprintf("Enter Administration Password"))
		SetWindowText(win.GetDlgItem(h, consts.DialogStatic2), i18n.SprintfColon("Password"))
		SetWindowText(win.GetDlgItem(h, win.IDOK), i18n.Sprintf("OK"))
		SetWindowText(win.GetDlgItem(h, win.IDCANCEL), i18n.Sprintf("Cancel"))
		vd.setIcon(h, int(win.GetDpiForWindow(h)))
		return win.TRUE
	case win.WM_COMMAND:
		switch win.LOWORD(uint32(wp)) {
		case win.IDOK:
			passwd := GetWindowText(win.GetDlgItem(h, consts.DialogEdit))
			if sec.EncryptPassword(passwd) != appConf.Password {
				win.MessageBox(h, windows.StringToUTF16Ptr(i18n.Sprintf("The password is incorrect. Re-enter password.")),
					windows.StringToUTF16Ptr(AppLocalName), windows.MB_ICONERROR)
				win.SetFocus(win.GetDlgItem(h, consts.DialogEdit))
			} else {
				win.EndDialog(h, win.IDOK)
			}
		case win.IDCANCEL:
			win.SendMessage(h, win.WM_CLOSE, 0, 0)
		}
	case win.WM_CTLCOLORBTN, win.WM_CTLCOLORDLG, win.WM_CTLCOLOREDIT, win.WM_CTLCOLORMSGBOX, win.WM_CTLCOLORSTATIC:
		return uintptr(win.GetStockObject(win.WHITE_BRUSH))
	case win.WM_DPICHANGED:
		vd.setIcon(h, int(win.HIWORD(uint32(wp))))
	case win.WM_CLOSE:
		win.EndDialog(h, win.IDCANCEL)
	}
	return win.FALSE
}

func (vd *ValidateDialog) setIcon(h win.HWND, dpi int) error {
	system32, err := windows.GetSystemDirectory()
	if err != nil {
		return err
	}
	iconFile, err := syscall.UTF16PtrFromString(filepath.Join(system32, consts.IconKey.Dll+".dll"))
	if err != nil {
		return err
	}
	if vd.hIcon != 0 {
		win.DestroyIcon(vd.hIcon)
		vd.hIcon = 0
	}
	size := walk.SizeFrom96DPI(walk.Size{Width: 32, Height: 32}, dpi)
	win.SHDefExtractIcon(iconFile, int32(consts.IconKey.Index),
		0, nil, &vd.hIcon, win.MAKELONG(0, uint16(size.Width)))
	if vd.hIcon != 0 {
		win.SendDlgItemMessage(h, consts.DialogIcon, stmSetIcon, uintptr(vd.hIcon), 0)
	}
	return nil
}

func SetWindowText(hWnd win.HWND, text string) bool {
	txt, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return false
	}
	if win.TRUE != win.SendMessage(hWnd, win.WM_SETTEXT, 0, uintptr(unsafe.Pointer(txt))) {
		return false
	}
	return true
}

func GetWindowText(hWnd win.HWND) string {
	textLength := win.SendMessage(hWnd, win.WM_GETTEXTLENGTH, 0, 0)
	buf := make([]uint16, textLength+1)
	win.SendMessage(hWnd, win.WM_GETTEXT, textLength+1, uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}
