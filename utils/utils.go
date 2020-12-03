package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
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

func TryAlterFile(f1 string, f2 string, rename bool) {
	for i := 0; i < 5; i++ {
		var err error
		if rename {
			err = os.Rename(f1, f2)
		} else {
			err = os.Remove(f1)
		}
		if err == nil {
			break
		}
		if err, ok := err.(*os.LinkError); ok && (err.Err == syscall.ENOTDIR || err.Err == syscall.ERROR_FILE_NOT_FOUND) {
			break
		}
		if err, ok := err.(*os.PathError); ok && err.Err == syscall.ERROR_FILE_NOT_FOUND {
			break
		}
		time.Sleep(time.Second * 1)
	}
}

func FindRelatedFiles(path string, replace string) (relatedFiles []string, newFiles []string) {
	baseDir := filepath.Dir(path)
	fileName := filepath.Base(path)
	extName := filepath.Ext(fileName)
	fileName = strings.TrimSuffix(fileName, extName)
	if extName == "" {
		extName = `.log`
	}
	pattern := fileName + `\.((\d+)-(\d+)-(\d+))\` + extName
	relatedFiles = make([]string, 0)
	newFiles = make([]string, 0)
	p := regexp.MustCompile(pattern)
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return
	}
	for _, file := range files {
		if p.MatchString(file.Name()) {
			relatedFiles = append(relatedFiles, filepath.Join(baseDir, file.Name()))
			if replace != "" {
				dateStr := p.FindStringSubmatch(file.Name())[1]
				newFileName := filepath.Base(replace)
				newExt := filepath.Ext(newFileName)
				if newExt == "" {
					newExt = ".log"
				}
				newName := strings.TrimSuffix(newFileName, newExt)
				replaced := filepath.Join(filepath.Dir(replace), newName+"."+dateStr+newExt)
				newFiles = append(newFiles, replaced)
			}
		}
	}
	return
}
