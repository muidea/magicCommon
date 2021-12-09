package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

func LoadConfig(filePath string, ptr interface{}) (err error) {
	fileHandle, fileErr := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if fileErr != nil {
		err = fileErr
		return
	}
	defer fileHandle.Close()

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

	fileHandle, fileErr := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if fileErr != nil {
		err = fileErr
		return
	}
	defer fileHandle.Close()

	var byteBuffer bytes.Buffer
	err = json.Indent(&byteBuffer, byteContent, "", "\t")
	if err != nil {
		return
	}

	_, err = byteBuffer.WriteTo(fileHandle)
	return
}
