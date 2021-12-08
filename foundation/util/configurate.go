package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func LoadConfig(filePath string, ptr interface{}) (err error) {
	fileHandle, fileErr := os.OpenFile(filePath, os.O_RDONLY, os.ModeType)
	if fileErr != nil {
		err = fileErr
		return
	}
	byteContent, byteErr := ioutil.ReadAll(fileHandle)
	if byteErr != nil {
		err = byteErr
		return
	}

	err = json.Unmarshal(byteContent, ptr)
	return
}

func SaveConfig(filePath string, ptr interface{}) (err error) {
	if ptr == nil {
		return
	}

	byteContent, byteErr := json.Marshal(ptr)
	if byteErr != nil {
		err = byteErr
		return
	}

	err = ioutil.WriteFile(filePath, byteContent, os.ModeType)
	return
}
