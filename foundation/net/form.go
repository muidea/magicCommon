package net

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"log/slog"
)

// MultipartFormFile 从 HTTP 请求中提取指定的文件，并将其保存到指定的路径。
// req 是 HTTP 请求。
// fieldName 是表单中文件字段的名称。
// dstFilePath 是文件将被保存的目录路径。
// fileName 是指定保存的文件名，如果值为空，则使用原始文件名。
// 返回值 ret 是上传文件的名称，err 是错误信息（如果有）。
func MultipartFormFile(req *http.Request, fieldName, dstFilePath, fileName string) (ret string, err error) {
	// 从请求中获取文件内容和文件头信息。
	fileContent, fileHead, fileErr := req.FormFile(fieldName)
	if fileErr != nil {
		err = fileErr
		slog.Error("get file field failed, field: fieldName, err: err.Error(", "field", fieldName, "error", err.Error())
		return
	}
	defer func() { _ = fileContent.Close() }()

	if fileName == "" {
		fileName = fileHead.Filename
	}

	if err = WriteFileAtomic(req.Context(), fileContent, dstFilePath, fileName, fileHead.Size); err != nil {
		return "", err
	}

	// 设置返回值为文件名
	ret = fileName
	return
}

// isValidDirectory 验证路径是否为合法的目录
func isValidDirectory(path string) bool {
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return false
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(cleanPath, 0755)
			return err == nil
		}
		return false
	}
	return info.IsDir()
}

// isValidFileName 验证文件名是否合法
func isValidFileName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsRune(name, 0) && !strings.ContainsAny(name, `\/:*?"<>|`)
}
