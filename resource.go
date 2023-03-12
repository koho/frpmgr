//go:build ignore

// generates resource files.

package main

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/koho/frpmgr/pkg/version"
)

var versionArray = strings.ReplaceAll(version.Number, ".", ",") + ",0"

func main() {
	rcFiles, err := filepath.Glob("cmd/*/*.rc")
	if err != nil {
		log.Fatal(err)
	}
	for _, rc := range rcFiles {
		output := strings.TrimSuffix(rc, filepath.Ext(rc)) + ".syso"
		res, err := exec.Command("windres", "-DVERSION_ARRAY="+versionArray, "-DVERSION_STR="+version.Number,
			"-i", rc, "-o", output, "-O", "coff", "-c", "65001").CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to compile resource: %s", string(res))
		}
	}
}
