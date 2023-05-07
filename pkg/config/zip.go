package config

import (
	"archive/zip"
	"os"

	"github.com/koho/frpmgr/pkg/util"
)

// Zip compresses the given config file list to a zip file.
func Zip(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	baseName, _ := util.SplitExt(filename)
	header.Name = baseName + ".ini"

	// Change to deflate to gain better compression
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	data, err := ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}
