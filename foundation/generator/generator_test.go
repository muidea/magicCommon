package generator

import (
	"fmt"
	"testing"
	"time"
)

func TestNewWithVal(t *testing.T) {
	datTime := time.Now().Local().Format("20060102150405")

	patternVal := "prefix-{YYYYMMDDHHmmSS}-{num}"
	initVal := fmt.Sprintf("prefix-%s-%04d", datTime, 1)
	generator, generatorErr := NewWithVal(patternVal, initVal)
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}
	expect := fmt.Sprintf("prefix-%s-%04d", datTime, 2)
	result := generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	patternVal = "A{YYYYMMDDHHmmSS}-abcde{num}"
	initVal = fmt.Sprintf("A%s-abcde%04d", datTime, 1)
	generator, generatorErr = NewWithVal(patternVal, initVal)
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}
	expect = fmt.Sprintf("A%s-abcde%04d", datTime, 2)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	patternVal = "A-abcde{num}"
	initVal = fmt.Sprintf("A-abcde%04d", 1)
	generator, generatorErr = NewWithVal(patternVal, initVal)
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}
	expect = fmt.Sprintf("A-abcde%04d", 2)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = NewWithVal("prefix-{YYYYMMDDHHmmSS}-{fixed(5):123}", "")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = fmt.Sprintf("prefix-%s-%05d", datTime, 124)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = NewWithVal("prefix-{YYYYMMDDHHmmSS}-{fixed(5):123}", "")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = fmt.Sprintf("prefix-%s-%05d", datTime, 124)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = NewWithVal("Abcd{fixed(6):100000}", "100020")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = "Abcd100021"
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = NewWithVal("STORE-{fixed(4):1000}", "STORE-1002")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = "STORE-1003"
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = NewWithVal("STORE-{fixed(4):1000}", "")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = "STORE-1001"
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = NewWithVal("SO-{YYYYMMDDHHmmSS}-{fixed(5):num}", "")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = fmt.Sprintf("SO-%s-00001", datTime)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

}

func TestGenImpl_GenCode(t *testing.T) {
	datTime := time.Now().Local().Format("20060102150405")
	generator, generatorErr := New("prefix-{YYYYMMDDHHmmSS}-{num}")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect := fmt.Sprintf("prefix-%s-%04d", datTime, 1)
	result := generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	_, generatorErr = New("{YYYYMMDDHHmmSS}-{num}")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	generator, generatorErr = New("prefix-{YYYYMMDDHHmmSS}-{fixed(5):123}")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = fmt.Sprintf("prefix-%s-%05d", datTime, 124)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = New("prefix-{fixed(5):123}")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = fmt.Sprintf("prefix-%05d", 124)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}

	generator, generatorErr = New("prefix-{fixed(5):123}")
	if generatorErr != nil {
		t.Errorf("illgel gengerator")
		return
	}

	expect = fmt.Sprintf("prefix-%05d", 124)
	result = generator.GenCode()
	if expect != result {
		t.Errorf("genCode failed, expect:%s, result:%s", expect, result)
		return
	}
}

func TestSplitNum(t *testing.T) {
	numStr := SplitNum("prefix-{num}")
	if numStr != "" {
		t.Error("SplitNum failed")
		return
	}

	numStr = SplitNum("prefix-123")
	if numStr != "123" {
		t.Error("SplitNum failed")
		return
	}

	numStr = SplitNum("123")
	if numStr != "123" {
		t.Error("SplitNum failed")
		return
	}
}
