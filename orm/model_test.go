package orm

import (
	"log"
	"reflect"
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
	info := getModelInfo(&Unit{T1: &test{val: 123}})
	if info == nil {
		t.Errorf("getModelInfo failed,")
		return
	}

	info.Dump()

	t1 := test{val: 10}
	var t2 Test
	elm1 := reflect.Indirect(reflect.ValueOf(t1))
	log.Print(elm1.Kind())

	t2 = &t1
	elm2 := reflect.Indirect(reflect.ValueOf(t2))
	log.Print(elm2.Kind())
}
