package path

import "os"

// Exist 路径是否存在
func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
