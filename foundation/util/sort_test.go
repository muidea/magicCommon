package util

import (
	"fmt"
	"log"
	"sort"
	"testing"
)

type testItem struct {
	val int
}

func (s *testItem) String() {
	fmt.Printf("val:%d", s.val)
}

type testItemSorter struct {
}

func (s *testItemSorter) Less(left, right interface{}) bool {
	lVal := left.(*testItem)
	rVal := right.(*testItem)
	return lVal.val > rVal.val
}

func TestSort(t *testing.T) {
	objList := []*objItem{}

	obj3 := &objItem{id: "3", obj: &testItem{val: 3}}
	objList = append(objList, obj3)

	obj0 := &objItem{id: "0", obj: &testItem{val: 0}}
	objList = append(objList, obj0)

	obj2 := &objItem{id: "2", obj: &testItem{val: 2}}
	objList = append(objList, obj2)

	obj1 := &objItem{id: "1", obj: &testItem{val: 1}}
	objList = append(objList, obj1)

	obj4 := &objItem{id: "2", obj: &testItem{val: 2}}
	objList = append(objList, obj4)

	helper := sortHelper{objList: objList, sorter: &testItemSorter{}}

	sort.Sort(helper)

	for _, val := range objList {
		log.Printf("id:%s, val:%v", val.id, val.obj)
	}
}
