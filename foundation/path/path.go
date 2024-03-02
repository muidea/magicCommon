package path

import "os"

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
	_, err = dir.Readdirnames(1)
	if err == nil {
		// 目录为空
		return true, nil
	} else if len(err.Error()) > 0 {
		// 目录不为空
		return false, nil
	}

	return false, err
}
