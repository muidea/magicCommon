package path

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/muidea/magicCommon/foundation/log"
)

// Exist 路径是否存在
func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// IsDirEmpty 判断目录是否为空
func IsDirEmpty(dirPath string) (bool, error) {
	// 打开目录
	dir, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// 读取目录中的文件和子目录
	nameSlice, nameErr := dir.Readdirnames(1)
	if nameErr != nil && !errors.Is(nameErr, io.EOF) {
		return false, nameErr
	}

	return len(nameSlice) == 0, nil
}

func SplitParentDir(dirPath string) string {
	parentPath, _ := path.Split(dirPath)
	return strings.TrimRight(parentPath, "/")
}

func CleanPathContent(dirPath string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Errorf("clean path content failed, dirPath:%s, error:%s", dirPath, err.Error())
		return
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if err := os.RemoveAll(entryPath); err != nil {
				log.Errorf("clean path content failed, dirPath:%s, error:%s", entryPath, err.Error())
			}
		} else {
			if err := os.Remove(entryPath); err != nil {
				log.Errorf("clean path content failed, dirPath:%s, error:%s", entryPath, err.Error())
			}
		}
	}
}
