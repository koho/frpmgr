//go:build ignore

// generates resource files.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/koho/frpmgr/pkg/version"
)

var (
	versionArray = strings.ReplaceAll(version.Number, ".", ",")
	archMap      = map[string]string{"amd64": "pe-x86-64", "386": "pe-i386"}
)

func main() {
	rcFiles, err := filepath.Glob("cmd/*/*.rc")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	for _, rc := range rcFiles {
		for goArch, resArch := range archMap {
			output := strings.TrimSuffix(rc, filepath.Ext(rc)) + fmt.Sprintf("_windows_%s.syso", goArch)
			res, err := exec.Command("windres", "-DVERSION_ARRAY="+versionArray, "-DVERSION_STR="+version.Number,
				"-i", rc, "-o", output, "-O", "coff", "-c", "65001", "-F", resArch).CombinedOutput()
			if err != nil {
				println(err.Error(), string(res))
			}
		}
	}
}
