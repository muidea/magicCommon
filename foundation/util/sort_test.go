package util

import (
	"fmt"
	"sort"
	"testing"

	"github.com/muidea/magicCommon/foundation/log"
)

type testItem struct {
	val  int
	val2 int
}

func (s *testItem) String() {
	fmt.Printf("val:%d, val2:%d", s.val, s.val2)
}

type testItemSorter1 struct {
}

func (s *testItemSorter1) Less(left, right interface{}) bool {
	lVal := left.(*testItem)
	rVal := right.(*testItem)

	if lVal.val <= rVal.val {
		return true
	}

	return false
}

type testItemSorter2 struct {
}

func (s *testItemSorter2) Less(left, right interface{}) bool {
	lVal := left.(*testItem)
	rVal := right.(*testItem)

	if lVal.val <= rVal.val && lVal.val2 <= rVal.val2 {
		return true
	}

	return false
}

func TestSort(t *testing.T) {
	objList := []*objItem{}

	obj3 := &objItem{id: "3", obj: &testItem{val: 3, val2: 1}}
	objList = append(objList, obj3)

	obj0 := &objItem{id: "0", obj: &testItem{val: 0, val2: 3}}
	objList = append(objList, obj0)

	obj2 := &objItem{id: "2", obj: &testItem{val: 2, val2: 6}}
	objList = append(objList, obj2)

	obj1 := &objItem{id: "1", obj: &testItem{val: 1, val2: 2}}
	objList = append(objList, obj1)

	obj4 := &objItem{id: "2", obj: &testItem{val: 2, val2: 0}}
	objList = append(objList, obj4)

	obj5 := &objItem{id: "2", obj: &testItem{val: 2, val2: 3}}
	objList = append(objList, obj5)

	helper1 := sortHelper{objList: objList, sorter: &testItemSorter1{}}

	sort.Sort(helper1)

	helper2 := sortHelper{objList: objList, sorter: &testItemSorter2{}}

	sort.Sort(helper2)

	for _, val := range objList {
		log.Infof("id:%s, val:%v", val.id, val.obj)
	}
}
