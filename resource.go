//go:build ignore

// generates resource files.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/koho/frpmgr/pkg/version"
)

var versionArray = strings.ReplaceAll(version.Number, ".", ",")

func main() {
	rcFiles, err := filepath.Glob("cmd/*/*.rc")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	arch := os.Getenv("FRPMGR_TARGET")
	if arch == "" {
		arch = os.Getenv("GOARCH")
	}
	for _, arch := range strings.Split(arch, " ") {
		var args []string
		var goArch string
		switch strings.TrimSpace(arch) {
		case "x64", "amd64":
			goArch = "amd64"
			args = append(args, "windres", "-F", "pe-x86-64")
		case "x86", "386":
			goArch = "386"
			args = append(args, "windres", "-F", "pe-i386")
		case "arm64":
			goArch = "arm64"
			args = append(args, "aarch64-w64-mingw32-windres")
		default:
			continue
		}
		for _, rc := range rcFiles {
			output := strings.TrimSuffix(rc, filepath.Ext(rc)) + fmt.Sprintf("_windows_%s.syso", goArch)
			res, err := exec.Command(args[0], append([]string{
				"-DVERSION_ARRAY=" + versionArray, "-DVERSION_STR=" + version.Number,
				"-i", rc, "-o", output, "-O", "coff", "-c", "65001",
			}, args[1:]...)...).CombinedOutput()
			if err != nil {
				println(err.Error(), string(res))
				os.Exit(1)
			}
		}
	}
	fmt.Println("VERSION=" + version.Number)
	fmt.Println("BUILD_DATE=" + time.Now().Format(time.DateOnly))
}
