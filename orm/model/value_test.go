package model

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestGetValueStr(t *testing.T) {
	iVal := int(123)
	fiVal := newFieldValue(reflect.ValueOf(&iVal))
	ret, _ := fiVal.GetValueStr()
	if ret != "123" {
		t.Errorf("GetValueStr failed, iVal:%d", iVal)
	}

	fVal := 12.34
	ffVal := newFieldValue(reflect.ValueOf(&fVal))
	ret, _ = ffVal.GetValueStr()
	if ret != "12.340000" {
		t.Errorf("GetValueStr failed, fVal:%f", fVal)
	}

	strVal := "abc"
	fstrVal := newFieldValue(reflect.ValueOf(&strVal))
	ret, _ = fstrVal.GetValueStr()
	if ret != "'abc'" {
		t.Errorf("GetValueStr failed, ret:%s, strVal:%s", ret, strVal)
	}

	bVal := true
	fbVal := newFieldValue(reflect.ValueOf(&bVal))
	ret, _ = fbVal.GetValueStr()
	if ret != "1" {
		t.Errorf("GetValueStr failed, ret:%s, bVal:%v", ret, bVal)
	}

	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	ftimeVal := newFieldValue(reflect.ValueOf(&now))
	ret, _ = ftimeVal.GetValueStr()
	if ret != "'2018-01-02 15:04:05'" {
		t.Errorf("GetValueStr failed, ret:%s, ftimeVal:%v", ret, now)
	}

	ii := 123
	var iiVal *int
	iiVal = &ii
	fiVal = newFieldValue(reflect.ValueOf(&iiVal))
	ret, _ = fiVal.GetValueStr()
	if ret != "123" {
		t.Errorf("GetValueStr failed, iVal:%d", iVal)
	}
}

func TestSetValue(t *testing.T) {
	var iVal int
	fiVal := newFieldValue(reflect.ValueOf(&iVal))

	intVal := 123
	fiVal.SetValue(reflect.ValueOf(&intVal))
	ret, _ := fiVal.GetValueStr()
	if ret != "123" {
		t.Errorf("GetValueStr failed, iVal:%d", iVal)
	}
	if iVal != 123 {
		t.Errorf("SetValue failed, iVal:%d", iVal)
	}

	var fVal float32
	ffVal := newFieldValue(reflect.ValueOf(&fVal))
	fltVal := 12.34
	ffVal.SetValue(reflect.ValueOf(&fltVal))
	ret, _ = ffVal.GetValueStr()
	if ret != "12.340000" {
		t.Errorf("GetValueStr failed, fVal:%f", fVal)
	}
	if fVal != 12.34 {
		t.Errorf("SetValue failed, fVal:%f", fVal)
	}

	var strVal string
	fstrVal := newFieldValue(reflect.ValueOf(&strVal))
	stringVal := "abc"
	fstrVal.SetValue(reflect.ValueOf(&stringVal))
	ret, _ = fstrVal.GetValueStr()
	if ret != "'abc'" {
		t.Errorf("GetValueStr failed, ret:%s, strVal:%s", ret, strVal)
	}
	if strVal != "abc" {
		t.Errorf("SetValue failed, strVal:%s", strVal)
	}

	var bVal bool
	fbVal := newFieldValue(reflect.ValueOf(&bVal))
	boolVal := true
	fbVal.SetValue(reflect.ValueOf(&boolVal))
	ret, _ = fbVal.GetValueStr()
	if ret != "1" {
		t.Errorf("GetValueStr failed, ret:%s, bVal:%v", ret, bVal)
	}
	if !bVal {
		t.Errorf("SetValue failed, bVal:%v", bVal)
	}
	bIntVal := 0
	fbVal.SetValue(reflect.ValueOf(&bIntVal))
	ret, _ = fbVal.GetValueStr()
	if ret != "0" {
		t.Errorf("GetValueStr failed, ret:%s, bVal:%v", ret, bVal)
	}
	if bVal {
		t.Errorf("SetValue failed, bVal:%v", bVal)
	}

	var now time.Time
	ftimeVal := newFieldValue(reflect.ValueOf(&now))
	timeVal := "2018-01-02 15:04:05"
	ftimeVal.SetValue(reflect.ValueOf(&timeVal))
	ret, _ = ftimeVal.GetValueStr()
	if ret != "'2018-01-02 15:04:05'" {
		t.Errorf("GetValueStr failed, ret:%s, ftimeVal:%v", ret, now)
	}

	ret = now.Format("2006-01-02 15:04:05")
	if ret != "2018-01-02 15:04:05" {
		t.Errorf("SetValue failed, ret:%v", ret)
	}

	curTime := time.Now()
	ftimeVal.SetValue(reflect.ValueOf(&curTime))
	ret, _ = ftimeVal.GetValueStr()
	if ret != fmt.Sprintf("'%s'", curTime.Format("2006-01-02 15:04:05")) {
		t.Errorf("GetValueStr failed, ret:%s, ftimeVal:%v", ret, now)
	}
	if now.Sub(curTime) != 0 {
		t.Errorf("SetValue failed, ret:%v", ret)
	}
}

func TestDepend(t *testing.T) {
	type AA struct {
		ii int
		jj int
		kk *int
	}
	structVal := []*AA{&AA{ii: 12, jj: 23}, &AA{ii: 23, jj: 34}}
	structSlicefv := newFieldValue(reflect.ValueOf(structVal))

	structFds, _ := structSlicefv.GetDepend()
	if len(structFds) != 2 {
		t.Errorf("fv.GetDepend failed. fds size:%d", len(structFds))
	}

	strSliceVal := []string{"10", "20", "30"}
	strSliceValfv := newFieldValue(reflect.ValueOf(&strSliceVal))
	strFds, _ := strSliceValfv.GetDepend()
	if len(strFds) != 0 {
		t.Errorf("fv.GetDepend failed. fds size:%d", len(strFds))
	}

	log.Print(strSliceValfv.GetValueStr())
}

func TestPtr(t *testing.T) {
	ii := 10
	var iVal *int
	fiVal := newFieldValue(reflect.ValueOf(&iVal))
	ret, _ := fiVal.GetValueStr()
	if ret != "" {
		t.Errorf("GetValueStr failed, iVal:%d", iVal)
	}

	iVal = &ii
	fiVal = newFieldValue(reflect.ValueOf(&iVal))

	intVal := 123
	fiVal.SetValue(reflect.ValueOf(&intVal))
	ret, _ = fiVal.GetValueStr()
	if ret != "123" {
		t.Errorf("GetValueStr failed, iVal:%d", iVal)
	}
	if *iVal != 123 {
		t.Errorf("SetValue failed, iVal:%d", iVal)
	}

	bVal := false
	log.Print(reflect.ValueOf(bVal).Type().String())
}
