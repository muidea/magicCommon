package generator

import (
	"fmt"
	"testing"
	"time"
)

func TestGenImpl_GenCode(t *testing.T) {
	datTime := time.Now().Local().Format("20060102150405")
	generator1, generatorErr := New("prefix-{YYYYMMDDHHmmSS}-{num}")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect := fmt.Sprintf("prefix-%s-%04d", datTime, 1)
	result := generator1.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	_, generatorErr = New("{YYYYMMDDHHmmSS}-{num}")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}
}
