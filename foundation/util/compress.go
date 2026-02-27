package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"log/slog"
)

func UnZipFile(zipFile, destDir string) (ret []string, err error) {
	zipHandler, zipErr := zip.OpenReader(zipFile)
	if zipErr != nil {
		err = zipErr
		slog.Error("UnZipFile failed", "error", zipErr.Error())
		return
	}
	defer func() {
		if closeErr := zipHandler.Close(); closeErr != nil {
			slog.Warn("Failed to close zip handler", "error", closeErr)
		}
	}()

	for _, f := range zipHandler.File {
		zfHandle, zfErr := f.Open()
		if zfErr != nil {
			err = zfErr
			slog.Error("getFieldReferenceValue failed", "file", f.Name, "error", zfErr.Error())
			return
		}
		defer func() {
			if closeErr := zfHandle.Close(); closeErr != nil {
				slog.Warn("Failed to close zip file handle", "file", f.Name, "error", closeErr)
			}
		}()

		path := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path, 0755)
			if err != nil {
				slog.Error("getFieldReferenceValue failed", "path", path, "error", err.Error())
				return
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				slog.Error("getFieldReferenceValue failed", "path", filepath.Dir(path), "error", err.Error())
				return
			}
			fHandle, fErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if fErr != nil {
				err = fErr
				return
			}
			defer func() {
				if closeErr := fHandle.Close(); closeErr != nil {
					slog.Warn("Failed to close file handle", "path", path, "error", closeErr)
				}
			}()

			_, err = io.Copy(fHandle, zfHandle)
			if err != nil {
				slog.Error("UnZipFile failed", "error", err.Error())
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
	defer func() { _ = newZipFile.Close() }()

	zipWriter := zip.NewWriter(newZipFile)
	defer func() { _ = zipWriter.Close() }()

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
			defer func() { _ = file.Close() }()

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
	defer func() { _ = newZipFile.Close() }()

	zipWriter := zip.NewWriter(newZipFile)
	defer func() { _ = zipWriter.Close() }()

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
	defer func() { _ = fileToZip.Close() }()

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
