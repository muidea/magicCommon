package test

import (
	"log"
	"testing"

	"muidea.com/magicCommon/orm"
)

func TestGroup(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb")
	defer orm.Uninitialize()

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

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

	group2.Parent = group1
	err = o1.Insert(group2)
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
	}

	group3.Parent = group1
	err = o1.Insert(group3)
	if err != nil {
		t.Errorf("insert Group3 failed, err:%s", err.Error())
	}

	err = o1.Delete(group3)
	if err != nil {
		t.Errorf("delete Group3 failed, err:%s", err.Error())
	}

	group4 := &Group{ID: group2.ID, Parent: &Group{}}
	err = o1.Query(group4)
	if err != nil {
		t.Errorf("query Group4 failed, err:%s", err.Error())
	}
	log.Print(*group4)
	log.Print(*(group4.Parent))

	group5 := &Group{ID: group2.ID, Parent: &Group{}}
	err = o1.Query(group5)
	if err != nil {
		t.Errorf("query Group5 failed, err:%s", err.Error())
	}
	log.Print(*group5)
	log.Print(*(group5.Parent))

	if !group2.Equle(group5) {
		t.Errorf("query Group5 failed")
	}
}

func TestUser(t *testing.T) {
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}

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

	err = o1.Insert(group2)
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
	}

	user1 := &User{Name: "demo", EMail: "123@demo.com", Group: []*Group{}}
	err = o1.Drop(user1)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
	}

	err = o1.Create(user1)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
	}

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)
	err = o1.Insert(user1)
	if err != nil {
		t.Errorf("insert user1 failed, err:%s", err.Error())
	}

	user2 := &User{ID: user1.ID}
	err = o1.Query(user2)
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
	}

	log.Print(*user2)
	if !user2.Equle(user1) {
		t.Errorf("query user2 failed")
	}
}
