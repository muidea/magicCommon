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
	ID int `json:"id" orm:"id key"`
	// Name 名称
	Name      string    `json:"name" orm:"name"`
	Value     float32   `json:"value" orm:"value"`
	TimeStamp time.Time `json:"timeStamp" orm:"timeStamp"`
	T1        test      `orm:"t1"`
}

type Test interface {
	Demo() string
}

type test struct {
	val int
}

func (s *test) Demo() string {
	return "test demo"
}

func TestStruct(t *testing.T) {
	now := time.Now()
	info := GetStructInfo(&Unit{T1: test{val: 123}, TimeStamp: now})
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
	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	unit := &Unit{T1: test{val: 123}, TimeStamp: now}
	info := GetStructInfo(unit)
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
