package util

import (
	"io"
	"os"

	"github.com/muidea/magicCommon/foundation/log"
)

func CopyFile(srcFile, dstFile string, delSource bool) (err error) {
	srcFileHandle, err := os.Open(srcFile)
	if err != nil {
		log.Errorf("CopyFile failed, rawFile:%s, open source file error:%s", srcFile, err.Error())
		return
	}
	defer srcFileHandle.Close()

	dstFileHandle, err := os.Create(dstFile)
	if err != nil {
		log.Errorf("CopyFile failed, filePath:%s, create destination file error:%s", dstFile, err.Error())
		return
	}
	defer dstFileHandle.Close()

	_, err = io.Copy(dstFileHandle, srcFileHandle)
	if err != nil {
		log.Errorf("CopyFile failed, filePath:%s, copy file content error:%s", dstFile, err.Error())
		return
	}

	err = dstFileHandle.Sync()
	if err != nil {
		log.Errorf("CopyFile failed, filePath:%s, sync file error:%s", dstFile, err.Error())
		return
	}

	if delSource {
		err = os.Remove(srcFile)
		if err != nil {
			log.Errorf("CopyFile failed, rawFile:%s, remove source file error:%s", srcFile, err.Error())
			return
		}
	}

	return
}
