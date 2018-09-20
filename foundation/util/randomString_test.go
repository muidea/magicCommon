package util

import (
	"log"
	"testing"
)

func TestRandom(t *testing.T) {
	//t1 := RandomString(100)
	//log.Printf("RandomString:%s", t1)
	//if len(t1) != 100 {
	//	t.Error("create RandomString failed")
	//}

	val := RandomIdentifyCode()
	log.Print(val)

	t2 := RandomAscII(32)
	log.Printf("RandomAscII:%s", t2)
	if len(t2) != 32 {
		t.Error("create RandomAscII failed")
	}

	t3 := RandomAlphabetic(32)
	log.Printf("RandomAlphabetic:%s", t3)
	if len(t3) != 32 {
		t.Error("create RandomAlphabetic failed")
	}

	t4 := RandomAlphanumeric(32)
	log.Printf("RandomAlphanumeric:%s", t4)
	if len(t4) != 32 {
		t.Error("create RandomAlphanumeric failed")
	}

	t5 := RandomNumeric(32)
	log.Printf("RandomNumeric:%s", t5)
	if len(t5) != 32 {
		t.Error("create RandomNumeric failed")
	}

}
