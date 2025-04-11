package path

import (
	"os"
	"path"
	"testing"
)

func TestIsDirEmpty(t *testing.T) {
	isEmpty, isErr := IsDirEmpty("./")
	if isErr != nil {
		t.Errorf("check is dir empty failed, error:%s", isErr.Error())
		return
	}
	if isEmpty {
		t.Errorf("check is dir empty failed")
	}

	curWD, curErr := os.Getwd()
	if curErr != nil {
		t.Errorf("check is dir empty failed, error:%s", curErr.Error())
		return
	}

	emptyDir := path.Join(curWD, "emptydir")
	t.Logf("emptyDir:%s", emptyDir)
	os.Mkdir(emptyDir, os.ModePerm)
	defer os.Remove(emptyDir)
	isEmpty, isErr = IsDirEmpty(emptyDir)
	if isErr != nil {
		t.Errorf("check is dir empty failed, error:%s", isErr.Error())
		return
	}
	if !isEmpty {
		t.Errorf("check is dir empty failed")
	}
}

func TestSplitParentDir(t *testing.T) {
	dirPath := "/var/test/abc.jpg"
	parentPath := SplitParentDir(dirPath)
	if parentPath != "/var/test" {
		t.Errorf("SplitParentDir failed, parentPath:%s,expect:%s", parentPath, "/var/test")
		return
	}
	parentPath = SplitParentDir(parentPath)
	if parentPath != "/var" {
		t.Errorf("SplitParentDir failed, parentPath:%s,expect:%s", parentPath, "/var")
		return
	}
	parentPath = SplitParentDir(parentPath)
	if parentPath != "" {
		t.Errorf("SplitParentDir failed, parentPath:%s,expect:%s", parentPath, "")
		return
	}

	dirPath = "/var/test/"
	parentPath = SplitParentDir(dirPath)
	if parentPath != "/var/test" {
		t.Errorf("SplitParentDir failed, parentPath:%s,expect:%s", parentPath, "/var/test")
		return
	}
}
