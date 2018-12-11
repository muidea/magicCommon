package orm

import (
	"testing"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int `json:"id" orm:"id key"`
	// Name 名称
	Name      string  `json:"name" orm:"name"`
	Value     float32 `json:"value" orm:"value"`
	TimeStamp *int    `json:"timeStamp" orm:"timeStamp"`
	T1        *test   `orm:"t1"`
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

func TestModel(t *testing.T) {
	intVal := 10
	info := getModelInfo(&Unit{T1: &test{val: 123}, TimeStamp: &intVal})
	if info == nil {
		t.Errorf("getModelInfo failed,")
		return
	}

	err := info.verify()
	if err != nil {
		t.Errorf("verify failed")
	}

	info.Dump()
}
