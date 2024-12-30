package util

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SplitExt splits the path into base name and file extension
func SplitExt(path string) (string, string) {
	if path == "" {
		return "", ""
	}
	fileName := filepath.Base(path)
	ext := filepath.Ext(path)
	return strings.TrimSuffix(fileName, ext), ext
}

// FileExists checks whether the given path is a file
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// FindLogFiles returns the files and dates archived by date
func FindLogFiles(path string) ([]string, []time.Time, error) {
	if path == "" || path == "console" {
		return nil, nil, os.ErrInvalid
	}
	fileDir, fileName := filepath.Split(path)
	baseName, ext := SplitExt(fileName)
	pattern := regexp.MustCompile(`^\.\d{4}(0[1-9]|1[0-2])(0[1-9]|[12][0-9]|3[01])-([0-1][0-9]|2[0-3])([0-5][0-9])([0-5][0-9])$`)
	if fileDir == "" {
		fileDir = "."
	}
	files, err := os.ReadDir(fileDir)
	if err != nil {
		return nil, nil, err
	}
	logs := []string{filepath.Clean(path)}
	dates := []time.Time{{}}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), baseName) && strings.HasSuffix(file.Name(), ext) {
			tailPart := strings.TrimPrefix(file.Name(), baseName)
			datePart := strings.TrimSuffix(tailPart, ext)
			if pattern.MatchString(datePart) {
				if date, err := time.Parse("20060102-150405", datePart[1:]); err == nil {
					logs = append(logs, filepath.Join(fileDir, file.Name()))
					dates = append(dates, date)
				}
			}
		}
	}
	return logs, dates, nil
}

// DeleteFiles removes the given file list ignoring errors
func DeleteFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

// RenameFiles renames the given file list with new file list ignoring errors
func RenameFiles(old []string, new []string) {
	for i := range old {
		// Create directory if the target directory does not exist
		if err := os.MkdirAll(filepath.Dir(new[i]), os.ModePerm); err == nil {
			os.Rename(old[i], new[i])
		}
	}
}

// ReadFileLines reads the last n lines in a file starting at a given offset
func ReadFileLines(path string, offset int64, n int) ([]string, int, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, -1, 0, err
	}
	defer file.Close()
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, -1, 0, err
	}
	reader := bufio.NewReader(file)

	var line string
	lines := make([]string, 0)
	i := -1
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		if n < 0 || len(lines) < n {
			lines = append(lines, line)
		} else {
			i = (i + 1) % n
			lines[i] = line
		}
	}
	offset, err = file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, -1, 0, err
	}
	if i >= 0 {
		i = (i + 1) % n
	}
	return lines, i, offset, nil
}

// ZipFiles compresses the given file list to a zip file
func ZipFiles(filename string, files map[string]string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for src, dst := range files {
		if err = addFileToZip(zipWriter, src, dst); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, src, dst string) error {
	fileToZip, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(dst)

	// Change to deflate to gain better compression
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

// CopyFile copy a file from src path to dest path
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
			return 0, err
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

// IsDirectory determines if a file represented by `path` is a directory or not
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

// AddFileSuffix adds a suffix between the file name and the file extension
func AddFileSuffix(filename, suffix string) string {
	baseName, ext := SplitExt(filename)
	return baseName + suffix + ext
}
