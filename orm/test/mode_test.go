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

	group5 := &Group{ID: group2.ID, Parent: &Group{}}
	err = o1.Query(group5)
	if err != nil {
		t.Errorf("query Group5 failed, err:%s", err.Error())
	}

	if !group2.Equle(group5) {
		t.Errorf("query Group5 failed")
	}
}

func TestUser(t *testing.T) {
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

	err = o1.Insert(group2)
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
	}

	err = o1.Insert(group3)
	if err != nil {
		t.Errorf("insert group3 failed, err:%s", err.Error())
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

	user1.Group = append(user1.Group, group3)
	err = o1.Update(user1)
	if err != nil {
		t.Errorf("update user1 failed, err:%s", err.Error())
	}

	err = o1.Query(user2)
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
	}
	if len(user2.Group) != 3 {
		t.Errorf("query user2 failed")
	}
	if !user2.Equle(user1) {
		t.Errorf("query user2 failed")
	}

	log.Print(*user2)

	err = o1.Delete(group1)
	if err != nil {
		t.Errorf("delete group1 failed, err:%s", err.Error())
	}
	err = o1.Delete(group2)
	if err != nil {
		t.Errorf("delete group2 failed, err:%s", err.Error())
	}
	err = o1.Delete(group3)
	if err != nil {
		t.Errorf("delete group3 failed, err:%s", err.Error())
	}
	err = o1.Delete(user2)
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}

}

func TestSystem(t *testing.T) {
	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
	}

	err = o1.Drop(user1)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
	}

	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}
	err = o1.Drop(sys1)
	if err != nil {
		t.Errorf("drop system failed, err:%s", err.Error())
	}

	err = o1.Create(user1)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
	}

	err = o1.Insert(user1)
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
	}
	err = o1.Insert(user2)
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
	}

	users := []User{*user1, *user2}
	sys1.Users = &users

	err = o1.Create(sys1)
	if err != nil {
		t.Errorf("create system failed, err:%s", err.Error())
	}

	err = o1.Insert(sys1)
	if err != nil {
		t.Errorf("insert system failed, err:%s", err.Error())
	}

	users = append(users, *user1)
	users = append(users, *user2)
	sys1.Users = &users
	err = o1.Update(sys1)
	if err != nil {
		t.Errorf("update system failed, err:%s", err.Error())
	}

	sys2 := &System{ID: sys1.ID, Users: &[]User{}}
	err = o1.Query(sys2)
	if err != nil {
		t.Errorf("query system failed, err:%s", err.Error())
	}

	log.Print(*sys2)
	log.Print(*(sys2.Users))

	if !sys1.Equle(sys2) {
		t.Error("query sys2 faield")
	}

	err = o1.Delete(sys2)
	if err != nil {
		t.Errorf("delete system failed, err:%s", err.Error())
	}

	err = o1.Delete(user1)
	if err != nil {
		t.Errorf("delete user1 failed, err:%s", err.Error())
	}
	err = o1.Delete(user2)
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}
}
