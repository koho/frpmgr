//go:build ignore

// generates .ico files from images.

package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const iconExt = ".ico"

var dir string

func init() {
	flag.StringVar(&dir, "d", "icon", "The directory of image files.")
	flag.Parse()
}

func main() {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path := filepath.Join(dir, f.Name())
		ext := filepath.Ext(path)
		if ext == iconExt {
			continue
		}
		output := strings.TrimSuffix(path, ext) + iconExt
		res, err := exec.Command(".deps/convert.exe", "-background", "none", path,
			"-define", "icon:auto-resize=256,192,128,96,64,48,40,32,24,20,16", "-compress", "zip", output).CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to convert icon: %s", string(res))
		}
	}
}
