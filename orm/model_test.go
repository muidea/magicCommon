package orm

import (
	"testing"

	dd1 "muidea.com/magicCommon/orm/test1/demo"
	dd2 "muidea.com/magicCommon/orm/test2/demo"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int `json:"id" orm:"id key"`
	// Name 名称
	Name      string    `json:"name" orm:"name"`
	Value     float32   `json:"value" orm:"value"`
	TimeStamp *int      `json:"timeStamp" orm:"timeStamp"`
	T1        *test     `orm:"t1"`
	Demo1     *dd1.Demo `orm:"demo"`
	Demo2     *dd2.Demo `orm:"demo"`
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
	info := getModelInfo(&Unit{T1: &test{val: 123}, TimeStamp: &intVal, Demo1: &dd1.Demo{}, Demo2: &dd2.Demo{}})
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
