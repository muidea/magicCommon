package test

import (
	"log"
	"testing"

	"muidea.com/magicCommon/orm"
)

func TestGroup(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb")
	defer orm.Uninitialize()

	group1 := &Group{Name: "testGroup1", Users: &[]*User{}, Children: &[]*Group{}}
	group2 := &Group{Name: "testGroup2", Users: &[]*User{}, Children: &[]*Group{}}
	group3 := &Group{Name: "testGroup3", Users: &[]*User{}, Children: &[]*Group{}}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
	}

	err = o1.Drop(group1)
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
	}

	err = o1.Create(group1)
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
	}

	err = o1.Insert(group1)
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
	}

	log.Print(*group1)
	group2.Parent = group1
	log.Print(*group2)
	err = o1.Insert(group2)
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
	}
	log.Print(*group1)

	group3.Parent = group1
	err = o1.Insert(group3)
	if err != nil {
		t.Errorf("insert Group3 failed, err:%s", err.Error())
	}

	err = o1.Delete(group3)
	if err != nil {
		t.Errorf("delete Group3 failed, err:%s", err.Error())
	}
}
