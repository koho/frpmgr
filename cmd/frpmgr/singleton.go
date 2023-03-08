package main

import (
	"crypto/md5"
	"encoding/hex"
	"io/fs"
	"os"
	"syscall"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

// checkSingleton returns an error when another program is running.
// This function should be only called in gui mode before the window is created.
func checkSingleton() (windows.Handle, error) {
	path, err := os.Executable()
	if err != nil {
		return 0, err
	}
	hashName := md5.Sum([]byte(path))
	name, err := syscall.UTF16PtrFromString("Local\\" + hex.EncodeToString(hashName[:]))
	if err != nil {
		return 0, err
	}
	return windows.CreateMutex(nil, false, name)
}

// showMainWindow activates and brings the window of running process to the foreground.
func showMainWindow() error {
	var windowToShow win.HWND
	path, err := os.Executable()
	if err != nil {
		return err
	}
	execFileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	syscall.MustLoadDLL("user32.dll").MustFindProc("EnumWindows").Call(
		syscall.NewCallback(func(hwnd syscall.Handle, lparam uintptr) uintptr {
			className := make([]uint16, windows.MAX_PATH)
			if _, err = win.GetClassName(win.HWND(hwnd), &className[0], len(className)); err != nil {
				return 1
			}
			if windows.UTF16ToString(className) == "\\o/ Walk_MainWindow_Class \\o/" {
				var pid uint32
				var imageName string
				var imageFileInfo fs.FileInfo
				if _, err = windows.GetWindowThreadProcessId(windows.HWND(hwnd), &pid); err != nil {
					return 1
				}
				imageName, err = getImageName(pid)
				if err != nil {
					return 1
				}
				imageFileInfo, err = os.Stat(imageName)
				if err != nil {
					return 1
				}
				if os.SameFile(execFileInfo, imageFileInfo) {
					windowToShow = win.HWND(hwnd)
					return 0
				}
			}
			return 1
		}), 0)
	if windowToShow != 0 {
		if win.IsIconic(windowToShow) {
			win.ShowWindow(windowToShow, win.SW_RESTORE)
		} else {
			win.SetForegroundWindow(windowToShow)
		}
	}
	return nil
}

// getImageName returns the full process image name of the given process id.
func getImageName(pid uint32) (string, error) {
	proc, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(proc)
	var exeNameBuf [261]uint16
	exeNameLen := uint32(len(exeNameBuf) - 1)
	err = windows.QueryFullProcessImageName(proc, 0, &exeNameBuf[0], &exeNameLen)
	if err != nil {
		return "", err
	}
	return windows.UTF16ToString(exeNameBuf[:exeNameLen]), nil
}
