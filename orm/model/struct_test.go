package model

import (
	"testing"
	"time"
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
	info := GetStructInfo(&Unit{T1: &test{val: 123}, TimeStamp: &now})
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
