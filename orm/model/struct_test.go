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
	T1        Test      `orm:"t1"`
}

type BT struct {
	ID  int `orm:"id key"`
	Val int `orm:"val"`
}

type Base struct {
	ID  int `orm:"id key"`
	Val int `orm:"val"`
	Bt  BT  `orm:"bt"`
}

type Test struct {
	ID    int  `orm:"id key"`
	Val   int  `orm:"val"`
	Base  Base `orm:"b1"`
	Base2 BT   `orm:"b2"`
}

func TestStruct(t *testing.T) {
	cache := NewCache()
	now := time.Now()
	info, err := GetObjectStructInfo(&Unit{T1: Test{Val: 123}, TimeStamp: now}, true, cache)
	if info == nil || err != nil {
		t.Errorf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	info.Dump()
}

func TestStructValue(t *testing.T) {
	cache := NewCache()
	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	unit := &Unit{Name: "AA", T1: Test{Val: 123}, TimeStamp: now}
	info, err := GetObjectStructInfo(unit, true, cache)
	if err != nil {
		t.Errorf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	log.Print(*unit)

	id := 123320
	pk := info.GetPrimaryField()
	if pk == nil {
		t.Errorf("GetPrimaryField faield")
		return
	}
	pk.SetFieldValue(reflect.ValueOf(id))

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
		AB  AB  `orm:"ab"`
		I64 int `orm:"i64"`
	}

	type Demo struct {
		II int   `orm:"ii"`
		AB *AB   `orm:"ab"`
		CD []int `orm:"cd"`
		EF []*AB `orm:"ef"`
	}

	cache := NewCache()
	f32Info, err := GetObjectStructInfo(&Demo{AB: &AB{}}, true, cache)
	if err != nil {
		t.Errorf("GetObjectStructInfo failed, err:%s", err.Error())
	}

	f32Info.Dump()

	i64Info, err := GetObjectStructInfo(&CD{}, true, cache)
	if err != nil {
		t.Errorf("GetObjectStructInfo failed, err:%s", err.Error())
	}

	i64Info.Dump()
}
