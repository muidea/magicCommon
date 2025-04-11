package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/muidea/magicCommon/foundation/log"
)

func UnZipFile(zipFile, destDir string) (ret []string, err error) {
	zipHandler, zipErr := zip.OpenReader(zipFile)
	if zipErr != nil {
		err = zipErr
		log.Errorf("UnZipFile failed, zip.OpenReader error:%s", zipErr.Error())
		return
	}
	defer zipHandler.Close()

	for _, f := range zipHandler.File {
		zfHandle, zfErr := f.Open()
		if zfErr != nil {
			err = zfErr
			log.Errorf("getFieldReferenceValue failed,Open %s error:%s", f.Name, zfErr.Error())
			return
		}
		defer zfHandle.Close()

		path := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			os.MkdirAll(filepath.Dir(path), os.ModePerm)
			fHandle, fErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if fErr != nil {
				err = fErr
				return
			}
			defer fHandle.Close()

			_, err = io.Copy(fHandle, zfHandle)
			if err != nil {
				log.Errorf("UnZipFile failed, copy file error:%s", err.Error())
				return
			}
		}
		ret = append(ret, path)
	}

	return
}

func ZipDir(sourceDir, outputFile string) error {
	newZipFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = relPath
		header.Method = zip.Deflate

		if info.IsDir() {
			header.Name += "/"
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func ZipFiles(files []string, outputFile string) error {
	newZipFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		err = addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(zw *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
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

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	if err != nil {
		return err
	}

	return nil
}

func ZipFile(file, outputFile string) (err error) {
	return ZipFiles([]string{file}, outputFile)
}
