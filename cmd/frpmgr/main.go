package main

import (
	"fmt"
	"github.com/koho/frpmgr/services"
	"github.com/koho/frpmgr/ui"
	"golang.org/x/sys/windows"
	"os"
	"strings"
)

func fatal(v ...interface{}) {
	windows.MessageBox(0, windows.StringToUTF16Ptr(fmt.Sprint(v...)), windows.StringToUTF16Ptr("错误"), windows.MB_ICONERROR)
	os.Exit(1)
}

func info(title string, format string, v ...interface{}) {
	windows.MessageBox(0, windows.StringToUTF16Ptr(fmt.Sprintf(format, v...)), windows.StringToUTF16Ptr(title), windows.MB_ICONINFORMATION)
}

func usage() {
	var flags = [...]string{
		fmt.Sprintf("(no argument): run the frp manager"),
		"/service CONFIG_PATH",
	}
	builder := strings.Builder{}
	for _, flag := range flags {
		builder.WriteString(fmt.Sprintf("    %s\n", flag))
	}
	info(fmt.Sprintf("Command Line Options"), "Usage: %s [\n%s]", os.Args[0], builder.String())
	os.Exit(1)
}

func main() {
	if len(os.Args) <= 1 {
		ui.RunUI()
		return
	}
	switch os.Args[1] {
	case "/service":
		if len(os.Args) != 3 {
			usage()
		}
		err := services.Run(os.Args[2])
		if err != nil {
			fatal(err)
		}
		return
	}
	usage()
}
