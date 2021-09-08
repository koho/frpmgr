package services

import (
	"fmt"
	"io"
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

var setupCmd = `$mmc = New-Object -ComObject MMC20.Application;`

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
		fmt.Fprintln(mmc.stdin, "$mmc.Document.Close(0);sleep 2;$mmc.Quit();exit;")
	}
}

func ShowPropertyDialog(displayName string) {
	if mmc.handle == nil {
		handle, stdin, stdout, stderr, err := StartProcess("powershell.exe", "-NoExit", "-Command", "-")
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
