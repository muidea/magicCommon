package log

import (
	"fmt"
	log "github.com/cihub/seelog"
	"testing"
)

func TestCriticalf(t *testing.T) {
	name := "test"
	errInfo := fmt.Errorf("test error")
	log.Criticalf("test log,name:%s, err:%v", name, errInfo)
}
