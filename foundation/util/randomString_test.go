package util

import (
	"testing"

	"github.com/muidea/magicCommon/foundation/log"
)

func TestRandom(t *testing.T) {
	//t1 := RandomString(100)
	//log.Infof("RandomString:%s", t1)
	//if len(t1) != 100 {
	//	t.Error("create RandomString failed")
	//}

	val := RandomIdentifyCode()
	log.Infof(val)

	t2 := RandomAscII(32)
	log.Infof("RandomAscII:%s", t2)
	if len(t2) != 32 {
		t.Error("create RandomAscII failed")
	}

	t3 := RandomAlphabetic(32)
	log.Infof("RandomAlphabetic:%s", t3)
	if len(t3) != 32 {
		t.Error("create RandomAlphabetic failed")
	}

	t4 := RandomAlphanumeric(32)
	log.Infof("RandomAlphanumeric:%s", t4)
	if len(t4) != 32 {
		t.Error("create RandomAlphanumeric failed")
	}

	t5 := RandomNumeric(32)
	log.Infof("RandomNumeric:%s", t5)
	if len(t5) != 32 {
		t.Error("create RandomNumeric failed")
	}

}
