package log

import (
	"fmt"
	"testing"
)

func TestCriticalf(t *testing.T) {
	name := "test"
	errInfo := fmt.Errorf("test error")
	Criticalf("test log,name:%s, err:%v", name, errInfo.Error())

	Infof("abcdefg")
}
