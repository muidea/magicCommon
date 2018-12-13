package executor

import (
	"log"
	"testing"

	"muidea.com/magicCommon/orm"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int64 `json:"id" orm:"id key auto"`
	// Name 名称
	Name  string  `json:"name" orm:"name"`
	Value float32 `json:"value" orm:"value"`
	//TimeStamp *time.Time `json:"timeStamp" orm:"ts"`
}

func TestExecutor(t *testing.T) {

	orm.Initialize("root", "rootkit", "localhost:3306", "testdb")
	defer orm.Uninitialize()

	//now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	//obj := &Unit{ID: 10, Name: "Hello world", Value: 12.3456, TimeStamp: &now}
	obj := &Unit{ID: 10, Name: "Hello world", Value: 12.3456}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
	}

	err = o1.Insert(obj)
	if err != nil {
		t.Errorf("insert obj failed, err:%s", err.Error())
	}
	log.Print(obj)

	obj.Name = "abababa"
	obj.Value = 100.000
	err = o1.Update(obj)
	if err != nil {
		t.Errorf("update obj failed, err:%s", err.Error())
	}
	log.Print(obj)

	obj2 := &Unit{ID: 1, Name: "", Value: 0.0}
	err = o1.Query(obj2)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}
	log.Print(obj2)

	defer o1.Drop(obj)
}
