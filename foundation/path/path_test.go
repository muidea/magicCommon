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
		t.Errorf("check is dir empty failed, error:%s", isErr.Error())
		return
	}

	emptyDir := path.Join(curWD, "emptydir")
	t.Logf("emptyDir:%s", emptyDir)
	isEmpty, isErr = IsDirEmpty(emptyDir)
	if isErr != nil {
		t.Errorf("check is dir empty failed, error:%s", isErr.Error())
		return
	}
	if !isEmpty {
		t.Errorf("check is dir empty failed")
	}
}
