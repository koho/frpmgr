package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(src string, dest string) (int64, error) {
	srcStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !srcStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	srcFile.Seek(0, 0)
	defer srcFile.Close()

	dstStat, err := os.Stat(dest)
	if err != nil {
		if !os.IsNotExist(err) {
			return 0, nil
		}
	} else {
		if os.SameFile(srcStat, dstStat) {
			return 0, nil
		}
	}
	destFile, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	return io.Copy(destFile, srcFile)
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func EnsurePath(path string) {
	if path == "" {
		return
	}
	dirPath := filepath.Dir(path)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, os.ModePerm)
	}
}

func ReadFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var line string
	lines := make([]string, 0)
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, line)
	}
	return lines, nil
}
