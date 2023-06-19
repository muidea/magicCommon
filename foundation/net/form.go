package net

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

// MultipartFormFile 接受文件参数
// string 文件名
// error 错误码
func MultipartFormFile(r *http.Request, field, dstPath string) (ret string, err error) {
	fileContent, fileHead, fileErr := r.FormFile(field)
	if fileErr != nil {
		err = fileErr
		log.Printf("get file field failed, field:%s, err:%s", field, err.Error())
		return
	}
	defer fileContent.Close()

	_, fileErr = os.Stat(dstPath)
	if fileErr != nil {
		if os.IsNotExist(fileErr) {
			fileErr = os.MkdirAll(dstPath, os.ModePerm)
		}
	}
	if fileErr != nil {
		err = fileErr
		log.Printf("destination path is invalid, err:%s", err.Error())
		return
	}

	dstFilePath := path.Join(dstPath, fileHead.Filename)
	dstFile, dstErr := os.Create(dstFilePath)
	if dstErr != nil {
		err = dstErr
		log.Printf("create destination file failed, err:%s", err.Error())
		return
	}

	defer dstFile.Close()
	_, err = io.Copy(dstFile, fileContent)
	if err != nil {
		log.Printf("copy destination file failed, err%s", err.Error())
		return
	}
	ret = dstFilePath
	return
}
