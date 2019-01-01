package test

import (
	"testing"

	"muidea.com/magicCommon/orm"
)

func TestGroup(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb")
	defer orm.Uninitialize()

	gorup1 := &Group{Name: "testGroup1", User: []*User{}, SubGroup: []*Group{}}
	gorup2 := &Group{Name: "testGroup2", User: []*User{}, SubGroup: []*Group{}}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
	}

	err = o1.Insert(gorup1)
	if err != nil {
		t.Errorf("insert Group failed, err:%s", err.Error())
	}

	err = o1.Drop(gorup1)
	if err != nil {
		t.Errorf("drop Group failed, err:%s", err.Error())
	}

	err = o1.Insert(gorup1)
	if err != nil {
		t.Errorf("insert Group failed, err:%s", err.Error())
	}

	gorup2.ParentGroup = gorup1
	err = o1.Insert(gorup2)
	if err != nil {
		t.Errorf("insert Group failed, err:%s", err.Error())
	}
}
