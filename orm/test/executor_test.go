package executor

import (
	"log"
	"testing"
	"time"

	"muidea.com/magicCommon/orm"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID  int    `json:"id" orm:"id key auto"`
	I8  int8   `orm:"i8"`
	I16 int16  `orm:"i16"`
	I32 int32  `orm:"i32"`
	I64 uint64 `orm:"i64"`
	// Name 名称
	Name      string    `json:"name" orm:"name"`
	Value     float32   `json:"value" orm:"value"`
	F64       float64   `orm:"f64"`
	TimeStamp time.Time `json:"timeStamp" orm:"ts"`
	Flag      bool      `orm:"flag"`
}

func TestExecutor(t *testing.T) {

	orm.Initialize("root", "rootkit", "localhost:3306", "testdb")
	defer orm.Uninitialize()

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	obj := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}
	//obj := &Unit{ID: 10, Name: "Hello world", Value: 12.3456}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
	}

	err = o1.Insert(obj)
	if err != nil {
		t.Errorf("insert obj failed, err:%s", err.Error())
	}

	obj.Name = "abababa"
	obj.Value = 100.000
	err = o1.Update(obj)
	if err != nil {
		t.Errorf("update obj failed, err:%s", err.Error())
	}

	obj2 := &Unit{ID: 1, Name: "", Value: 0.0}
	err = o1.Query(obj2)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}
	if obj.Name != obj2.Name || obj.Value != obj2.Value {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

	err = o1.Delete(obj)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

	defer o1.Drop(obj)

	log.Print(*obj2)
}
