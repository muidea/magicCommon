package util

import (
	"fmt"
	"log"
	"testing"
)

type Car struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	Age  int    `json:"age"`
	Car  *Car   `json:"car"`
}

func TestIntArray2Str(t *testing.T) {
	tempArray := []int{1, 2}
	str := IntArray2Str(tempArray)
	if str != "1,2" {
		t.Errorf("IntArray2Str failed, %s", str)
	}

	tempArray = []int{}
	str = IntArray2Str(tempArray)
	if str != "" {
		t.Errorf("IntArray2Str failed, %s", str)
	}
}

func TestStr2IntArray(t *testing.T) {
	str := ""
	tempArray, ok := Str2IntArray(str)
	if !ok || len(tempArray) > 0 {
		t.Errorf("Str2IntArray failed, ok=%v, len(tempArray)=%d", ok, len(tempArray))
	}

	str = "1"
	tempArray, ok = Str2IntArray(str)
	if !ok || len(tempArray) != 1 || tempArray[0] != 1 {
		t.Errorf("Str2IntArray failed, ok=%v, len(tempArray)=%d", ok, len(tempArray))
	}

	str = ",1"
	tempArray, ok = Str2IntArray(str)
	if !ok || len(tempArray) != 1 || tempArray[0] != 1 {
		t.Errorf("Str2IntArray failed, ok=%v, len(tempArray)=%d", ok, len(tempArray))
	}
	str = "1,"
	tempArray, ok = Str2IntArray(str)
	if !ok || len(tempArray) != 1 || tempArray[0] != 1 {
		t.Errorf("Str2IntArray failed, ok=%v, len(tempArray)=%d", ok, len(tempArray))
	}
	str = ",1,"
	tempArray, ok = Str2IntArray(str)
	if !ok || len(tempArray) != 1 || tempArray[0] != 1 {
		t.Errorf("Str2IntArray failed, ok=%v, len(tempArray)=%d", ok, len(tempArray))
	}

	str = ",1,2,3,4"
	tempArray, ok = Str2IntArray(str)
	if !ok || len(tempArray) != 4 || tempArray[0] != 1 {
		t.Errorf("Str2IntArray failed, ok=%v, len(tempArray)=%d", ok, len(tempArray))
	}
}

func TestMarshalString(t *testing.T) {
	iVal := 1234
	marshalVal := MarshalString(iVal)
	if marshalVal != "1234" {
		t.Errorf("marshal int failed")
	}

	strVal := "1234"
	marshalVal = MarshalString(strVal)
	if marshalVal != "1234" {
		t.Errorf("marshal string failed")
	}

	fVal := 12.34
	marshalVal = MarshalString(fVal)
	if marshalVal != "12.34" {
		t.Errorf("marshal float failed")
	}

	bVal := false
	marshalVal = MarshalString(bVal)
	if marshalVal != "false" {
		t.Errorf("marshal bool failed")
	}

	bVal = true
	marshalVal = MarshalString(bVal)
	if marshalVal != "true" {
		t.Errorf("marshal bool failed")
	}

	strVal = "61d383cb134f4db6a367046ffac3051d"
	marshalVal = MarshalString(strVal)
	if marshalVal != "61d383cb134f4db6a367046ffac3051d" {
		t.Errorf("marshal string failed")
	}

	strVal = "[61d383cb134f4db]6a367046ffac3051d"
	marshalVal = MarshalString(strVal)
	if marshalVal != "[61d383cb134f4db]6a367046ffac3051d" {
		t.Errorf("marshal string failed")
	}

	strVal = "{61d383cb134f4db6a367046ffac3051d"
	marshalVal = MarshalString(strVal)
	if marshalVal != "{61d383cb134f4db6a367046ffac3051d" {
		t.Errorf("marshal string failed")
	}

	strVal = "%61d383cb134f4db6a367046ffac3051d"
	marshalVal = MarshalString(strVal)
	if marshalVal != "%61d383cb134f4db6a367046ffac3051d" {
		t.Errorf("marshal string failed")
	}

	strVal = "-61d383cb134f4db6a367046ffac3051d"
	marshalVal = MarshalString(strVal)
	if marshalVal != "-61d383cb134f4db6a367046ffac3051d" {
		t.Errorf("marshal string failed")
	}

	obj1 := &User{
		ID:   110,
		Name: "Hello",
		Desc: "hey boy",
		Age:  123,
		Car:  &Car{ID: 100, Name: "Car"},
	}

	marshalVal = MarshalString(obj1)
	if marshalVal == "" {
		t.Errorf("marshal user failed")
	}
	log.Print(marshalVal)

	obj2 := &User{
		ID:   110,
		Name: "Hello",
		Desc: "hey boy",
		Age:  123,
	}

	marshalVal = MarshalString(obj2)
	if marshalVal == "" {
		t.Errorf("marshal user failed")
	}
	log.Print(marshalVal)
}

func TestUnMarshalString(t *testing.T) {
	rawVal := "1234"
	iVal := 1234
	uVal := UnmarshalString(rawVal)
	switch uVal.(type) {
	case float64:
		if int(uVal.(float64)) != iVal {
			t.Errorf("unmarshal int failed")
		}
	default:
		t.Errorf("unmarshal int failed")
	}

	rawVal = "a1234"
	strVal := "a1234"
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case string:
		if uVal.(string) != strVal {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal int failed")
	}

	rawVal = "12.34"
	fVal := 12.34
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case float64:
		if uVal.(float64) != fVal {
			t.Errorf("unmarshal float failed")
		}
	default:
		t.Errorf("unmarshal float failed")
	}

	rawVal = "false"
	bVal := false
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case bool:
		if uVal.(bool) != bVal {
			t.Errorf("unmarshal bool failed")
		}
	default:
		t.Errorf("unmarshal bool failed")
	}

	rawVal = "true"
	bVal = true
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case bool:
		if uVal.(bool) != bVal {
			t.Errorf("unmarshal bool failed")
		}
	default:
		t.Errorf("unmarshal bool failed")
	}

	rawVal = "61d383cb134f4db6a367046ffac3051d"
	strVal = "61d383cb134f4db6a367046ffac3051d"
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case string:
		if uVal.(string) != strVal {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal string failed")
	}

	rawVal = "[61d383cb134f4db6a367046ffac3051d"
	strVal = "[61d383cb134f4db6a367046ffac3051d"
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case string:
		if uVal.(string) != strVal {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal string failed")
	}

	rawVal = "[61d383cb13]4f4db6a367046ffac3051d"
	strVal = "[61d383cb13]4f4db6a367046ffac3051d"
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case string:
		if uVal.(string) != strVal {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal string failed")
	}

	rawVal = "{61d383cb134f4db6a367046ffac3051d"
	strVal = "{61d383cb134f4db6a367046ffac3051d"
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case string:
		if uVal.(string) != strVal {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal string failed")
	}

	rawVal = "%61d383cb134f4db6a367046ffac3051d"
	strVal = "%61d383cb134f4db6a367046ffac3051d"
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case string:
		if uVal.(string) != strVal {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal string failed")
	}

	rawVal = "-61d383cb134f4db6a367046ffac3051d"
	strVal = "-61d383cb134f4db6a367046ffac3051d"
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case string:
		if uVal.(string) != strVal {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal string failed")
	}

	rawVal = "{\"id\":110,\"name\":\"Hello\",\"desc\":\"hey boy\",\"age\":123,\"car\":{\"id\":100,\"name\":\"Car\"}}"
	/*
		objVal := &User{
			ID: 110,
			Name: "Hello",
			Desc: "hey boy",
			Age: 123,
		}
	*/

	uVal = UnmarshalString(rawVal)
	if uVal == nil {
		t.Errorf("unmarshal object failed")
	}

	rawVal = "{\"id\":110,\"name\":\"Hello\",\"desc\":\"hey boy\",\"age\":123,\"car\":null}"
	uVal = UnmarshalString(rawVal)
	if uVal == nil {
		t.Errorf("unmarshal object failed")
	}

	rawVal = "{\"id\":110,\"name\":\"Hello\",\"desc\":\"hey boy\",\"age\":123}"
	uVal = UnmarshalString(rawVal)
	if uVal == nil {
		t.Errorf("unmarshal object failed")
	}
}

func TestArrayMarshalString(t *testing.T) {
	iVal := []int{1234}
	sVal := fmt.Sprintf("%v", iVal)
	log.Print(sVal)
	marshalVal := MarshalString(iVal)
	if marshalVal != "[1234]" {
		t.Errorf("marshal int failed")
	}

	strVal := []string{"1234"}
	marshalVal = MarshalString(strVal)
	if marshalVal != "[\"1234\"]" {
		t.Errorf("marshal string failed")
	}

	fVal := []float64{12.34}
	marshalVal = MarshalString(fVal)
	if marshalVal != "[12.34]" {
		t.Errorf("marshal float failed")
	}

	bVal := []bool{false}
	marshalVal = MarshalString(bVal)
	if marshalVal != "[false]" {
		t.Errorf("marshal bool failed")
	}

	bVal = []bool{true}
	marshalVal = MarshalString(bVal)
	if marshalVal != "[true]" {
		t.Errorf("marshal bool failed")
	}
}

func TestArrayUnMarshalString(t *testing.T) {
	rawVal := "[1234]"
	iVal := []int{1234}
	uVal := UnmarshalString(rawVal)
	switch uVal.(type) {
	case []float64:
		tVal := uVal.([]float64)
		if len(tVal) != len(iVal) {
			t.Errorf("unmarshal int failed")
		}
	default:
		t.Errorf("unmarshal int failed")
	}

	rawVal = "[\"1234\"]"
	strVal := []string{"1234"}
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case []string:
		tVal := uVal.([]string)
		if len(tVal) != len(strVal) {
			t.Errorf("unmarshal string failed")
		}
	default:
		t.Errorf("unmarshal string failed")
	}

	rawVal = "[12.34]"
	fVal := []float64{12.34}
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case []float64:
		tVal := uVal.([]float64)
		if len(tVal) != len(fVal) {
			t.Errorf("unmarshal float failed")
		}
	default:
		t.Errorf("unmarshal float failed")
	}

	rawVal = "[false,true,false,true,true,false,false]"
	bVal := []bool{false, true, false, true, true, false, false}
	uVal = UnmarshalString(rawVal)
	switch uVal.(type) {
	case []bool:
		tVal := uVal.([]bool)
		if len(tVal) != len(bVal) {
			t.Errorf("unmarshal bool failed")
		}
	default:
		t.Errorf("unmarshal bool failed")
	}
}
