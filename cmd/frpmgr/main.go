package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/version"
	"github.com/koho/frpmgr/services"
	"github.com/koho/frpmgr/ui"
)

func fatal(v ...interface{}) {
	windows.MessageBox(0, windows.StringToUTF16Ptr(fmt.Sprint(v...)), windows.StringToUTF16Ptr(ui.AppLocalName), windows.MB_ICONERROR)
	os.Exit(1)
}

func info(title string, format string, v ...interface{}) {
	windows.MessageBox(0, windows.StringToUTF16Ptr(fmt.Sprintf(format, v...)), windows.StringToUTF16Ptr(title), windows.MB_ICONINFORMATION)
}

var (
	confPath    string
	showVersion bool
	showHelp    bool
	flagOutput  strings.Builder
)

func init() {
	flag.StringVar(&confPath, "c", "", "The path to config `file` (Service-only).")
	flag.BoolVar(&showVersion, "v", false, "Display version information.")
	flag.BoolVar(&showHelp, "h", false, "Show help information.")
	flag.CommandLine.SetOutput(&flagOutput)
	flag.Parse()
}

func main() {
	if showHelp {
		flag.Usage()
		info(ui.AppLocalName, flagOutput.String())
		return
	}
	if showVersion {
		info(ui.AppLocalName, strings.Join([]string{
			i18n.Sprintf("Version: %s", version.Number),
			i18n.Sprintf("FRP version: %s", version.FRPVersion),
			i18n.Sprintf("Built on: %s", version.BuildDate),
		}, "\n"))
		return
	}
	inService, err := svc.IsWindowsService()
	if err != nil {
		fatal(err)
	}
	if inService {
		if confPath == "" {
			os.Exit(1)
			return
		}
		if err = services.Run(confPath); err != nil {
			fatal(err)
		}
	} else {
		h, err := checkSingleton()
		defer windows.CloseHandle(h)
		if errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
			showMainWindow()
			return
		}
		if err = ui.RunUI(); err != nil {
			fatal(err)
		}
	}
}
