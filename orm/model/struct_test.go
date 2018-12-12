package model

import (
	"testing"
	"time"

	dd1 "muidea.com/magicCommon/orm/test1/demo"
	dd2 "muidea.com/magicCommon/orm/test2/demo"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int `json:"id" orm:"id key"`
	// Name 名称
	Name      string     `json:"name" orm:"name"`
	Value     float32    `json:"value" orm:"value"`
	TimeStamp *time.Time `json:"timeStamp" orm:"timeStamp"`
	T1        *test      `orm:"t1"`
	Demo1     *dd1.Demo  `orm:"demo"`
	Demo2     *dd2.Demo  `orm:"demo"`
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
	info := GetStructInfo(&Unit{T1: &test{val: 123}, TimeStamp: &now, Demo1: &dd1.Demo{}, Demo2: &dd2.Demo{}})
	if info == nil {
		t.Errorf("GetStructInfo failed,")
		return
	}

	err := info.Verify()
	if err != nil {
		t.Errorf("Verify failed")
	}

	info.Dump()
}
