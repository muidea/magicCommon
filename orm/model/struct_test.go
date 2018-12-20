package model

import (
	"log"
	"reflect"
	"testing"
	"time"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int64 `json:"id" orm:"id key"`
	// Name 名称
	Name      string    `json:"name" orm:"name"`
	Value     float32   `json:"value" orm:"value"`
	TimeStamp time.Time `json:"timeStamp" orm:"timeStamp"`
	T1        test      `orm:"t1"`
}

type bT struct {
	id  int `orm:"id key"`
	val int `orm:"val"`
}

type base struct {
	id  int `orm:"id key"`
	val int `orm:"val"`
	bt  bT  `orm:"bt"`
}

type test struct {
	id    int  `orm:"id key"`
	val   int  `orm:"val"`
	base  base `orm:"b1"`
	base2 bT   `orm:"b2"`
}

func TestStruct(t *testing.T) {
	now := time.Now()
	info, _, _ := GetStructInfo(&Unit{T1: test{val: 123}, TimeStamp: now})
	if info == nil {
		t.Errorf("GetStructInfo failed,")
		return
	}

	err := info.Verify()
	if err != nil {
		t.Errorf("Verify failed, err:%s", err.Error())
	}

	info.Dump()
}

func TestStructValue(t *testing.T) {
	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	unit := &Unit{Name: "AA", T1: test{val: 123}, TimeStamp: now}
	info, _, _ := GetStructInfo(unit)
	if info == nil {
		t.Errorf("GetStructInfo failed,")
		return
	}

	err := info.Verify()
	if err != nil {
		t.Errorf("Verify failed, err:%s", err.Error())
	}

	log.Print(*unit)

	id := 123320
	info.primaryKey.SetFieldValue(reflect.ValueOf(id))

	name := "abcdfrfe"
	info.UpdateFieldValue("Name", reflect.ValueOf(name))

	now = time.Now()
	tsVal := reflect.ValueOf(now)
	info.UpdateFieldValue("TimeStamp", tsVal)

	log.Print(*unit)
}

func TestReference(t *testing.T) {
	type AB struct {
		F32 float32 `orm:"f32"`
	}

	type CD struct {
		AB
		I64 int `orm:"i64"`
	}

	type Demo struct {
		II int   `orm:"ii"`
		AB *AB   `orm:"ab"`
		CD []int `orm:"cd"`
		EF []*AB `orm:"ef"`
	}

	f32Info, _, _ := GetStructInfo(&Demo{AB: &AB{}})

	f32Info.Dump()

	i64Info, _, _ := GetStructInfo(&CD{})
	i64Info.Dump()
}
