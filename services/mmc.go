package services

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type shell struct {
	handle *exec.Cmd
	stdin  io.Writer
	stdout io.Reader
	stderr io.Reader
}

var mmc shell

// Start powershell via cmd
// The console code page must be set to UTF8 before starting powershell
var setupCmd = `chcp 65001
powershell
$mmc = New-Object -ComObject MMC20.Application;`

// Powershell command to find and show a service property dialog
var propCmd = `$mmc.Document.Close(0);
$mmc.load("services.msc");
$view = $mmc.document.ActiveView;
foreach ($x in $view.ListItems) {
  if ($x.Name -eq "%s") {
    $view.Select($x);
    $view.DisplaySelectionPropertySheet();
    break
  }
}
`

func CloseMMC() {
	if mmc.handle != nil {
		// This line should exit the powershell
		fmt.Fprintln(mmc.stdin, "$mmc.Document.Close(0);sleep 2;$mmc.Quit();exit;")
		// Exit cmd
		mmc.handle.Process.Kill()
	}
}

// ShowPropertyDialog shows up a service property dialog with given service
func ShowPropertyDialog(displayName string) {
	if mmc.handle == nil {
		handle, stdin, stdout, stderr, err := StartProcess("cmd.exe")
		if err != nil {
			return
		}
		mmc.handle = handle
		mmc.stdin = stdin
		mmc.stdout = stdout
		mmc.stderr = stderr
		fmt.Fprintln(mmc.stdin, setupCmd)
	}
	fmt.Fprintln(mmc.stdin, fmt.Sprintf(propCmd, displayName))
}

func StartProcess(cmd string, args ...string) (*exec.Cmd, io.Writer, io.Reader, io.Reader, error) {
	command := exec.Command(cmd, args...)
	command.Dir = os.TempDir()
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = command.Start()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return command, stdin, stdout, stderr, nil
}
