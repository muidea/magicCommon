package util

import (
	"testing"

	"log/slog"
)

func TestRandom(t *testing.T) {
	//t1 := RandomString(100)
	//slog.Info("RandomString:t1", "field", t1)
	//if len(t1) != 100 {
	//	t.Error("create RandomString failed")
	//}

	val := RandomIdentifyCode()
	slog.Info(val)

	t2 := RandomAscII(32)
	slog.Info("RandomAscII:t2", "field", t2)
	if len(t2) != 32 {
		t.Error("create RandomAscII failed")
	}

	t3 := RandomAlphabetic(32)
	slog.Info("RandomAlphabetic:t3", "field", t3)
	if len(t3) != 32 {
		t.Error("create RandomAlphabetic failed")
	}

	t4 := RandomAlphanumeric(32)
	slog.Info("RandomAlphanumeric:t4", "field", t4)
	if len(t4) != 32 {
		t.Error("create RandomAlphanumeric failed")
	}

	t5 := RandomNumeric(32)
	slog.Info("RandomNumeric:t5", "field", t5)
	if len(t5) != 32 {
		t.Error("create RandomNumeric failed")
	}

}
