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
func MultipartFormFile(r *http.Request, field, dstPath string) (string, error) {
	dstFilePath := ""
	var retErr error

	for true {
		fileContent, fileHead, fileErr := r.FormFile(field)
		if fileErr != nil {
			log.Printf("get file field failed, field:%s, err:%s", field, fileErr.Error())
			retErr = fileErr
			break
		}
		defer fileContent.Close()

		_, fileErr = os.Stat(dstPath)
		if fileErr != nil {
			if os.IsNotExist(fileErr) {
				fileErr = os.MkdirAll(dstPath, os.ModePerm)
			}
		}

		if fileErr != nil {
			log.Printf("destination path is invalid, err:%s", fileErr.Error())
			retErr = fileErr
			break
		}
		dstFilePath = path.Join(dstPath, fileHead.Filename)
		dstFile, err := os.Create(dstFilePath)
		if err != nil {
			log.Printf("create destination file failed, err:%s", err.Error())
			retErr = err
			break
		}

		defer dstFile.Close()
		_, err = io.Copy(dstFile, fileContent)
		if err != nil {
			log.Printf("copy destination file failed, err%s", err.Error())
		}

		retErr = err
		break
	}

	return dstFilePath, retErr
}
