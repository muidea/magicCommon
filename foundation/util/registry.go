package util

import (
	"sort"
)

// ObjectList object list
type ObjectList []interface{}

// ObjectFilter object filter
type ObjectFilter interface {
	Filter(obj interface{}) bool
}

// ObjectRegistry object 仓库
type ObjectRegistry interface {
	Put(id string, object interface{})
	Get(id string) interface{}
	Sort(sorter ObjectSorter) *ObjectList
	FetchList() *ObjectList
	Filter(filter ObjectFilter) *ObjectList
	Remove(id string)
}

// NewRegistry create new Registry
func NewRegistry() ObjectRegistry {
	impl := &registry{actionChannel: make(actionChannel)}
	go impl.run()

	return impl
}

const (
	putObj = iota
	getObj
	fetchObjList
	filterObj
	sortObj
	removeObj
)

type actionObj struct {
	action int
	id     string
	data   interface{}
	reply  chan interface{}
}

type actionChannel chan *actionObj

type registry struct {
	actionChannel actionChannel
}

type objItem struct {
	id  string
	obj interface{}
}

func (s *registry) run() {
	objectItemList := []*objItem{}
	for {
		item := <-s.actionChannel
		switch item.action {
		case putObj:
			objectItemList = append(objectItemList, &objItem{id: item.id, obj: item.data})
		case getObj:
			found := false
			for _, val := range objectItemList {
				if val.id == item.id {
					item.reply <- val.obj
					found = true
				}
			}
			if !found {
				item.reply <- nil
			}
		case sortObj:
			retList := ObjectList{}
			sorter := item.data.(ObjectSorter)
			objectItemList = s.sortObjectList(objectItemList, sorter)
			for _, val := range objectItemList {
				retList = append(retList, val.obj)
			}

			item.reply <- &retList
		case fetchObjList:
			retList := ObjectList{}
			for _, val := range objectItemList {
				retList = append(retList, val.obj)
			}
			item.reply <- &retList
		case filterObj:
			retList := ObjectList{}
			filter := item.data.(ObjectFilter)
			for _, val := range objectItemList {
				if filter.Filter(val.obj) {
					retList = append(retList, val)
				}
			}
			item.reply <- &retList
		case removeObj:
			newList := []*objItem{}
			for idx, val := range objectItemList {
				if val.id == item.id {
					newList = append(newList, objectItemList[:idx]...)
					if idx < len(objectItemList)-1 {
						newList = append(newList, objectItemList[idx+1:]...)
						break
					}
				}
			}
			objectItemList = newList
		}
	}
}

func (s *registry) Put(id string, object interface{}) {
	item := &actionObj{action: putObj, id: id, data: object}
	s.actionChannel <- item
}

func (s *registry) Get(id string) interface{} {
	reply := make(chan interface{})

	item := &actionObj{action: getObj, id: id, reply: reply}
	s.actionChannel <- item

	val := <-reply
	return val
}

func (s *registry) Sort(sorter ObjectSorter) *ObjectList {
	reply := make(chan interface{})

	item := &actionObj{action: sortObj, data: sorter, reply: reply}
	s.actionChannel <- item

	val := <-reply
	if val == nil {
		return nil
	}

	return val.(*ObjectList)
}

func (s *registry) FetchList() *ObjectList {
	reply := make(chan interface{})

	item := &actionObj{action: fetchObjList, reply: reply}
	s.actionChannel <- item

	val := <-reply
	if val == nil {
		return nil
	}

	return val.(*ObjectList)
}

func (s *registry) Filter(filter ObjectFilter) *ObjectList {
	reply := make(chan interface{})

	item := &actionObj{action: filterObj, data: filter, reply: reply}
	s.actionChannel <- item

	val := <-reply
	if val == nil {
		return nil
	}

	return val.(*ObjectList)
}

func (s *registry) Remove(id string) {
	item := &actionObj{action: removeObj, id: id}
	s.actionChannel <- item
}

type sortHelper struct {
	objList []*objItem
	sorter  ObjectSorter
}

func (s sortHelper) Len() int {
	return len(s.objList)
}

func (s sortHelper) Less(i, j int) bool {
	return s.sorter.Less(s.objList[i].obj, s.objList[j].obj)
}

func (s sortHelper) Swap(i, j int) {
	s.objList[i], s.objList[j] = s.objList[j], s.objList[i]
}

func (s *registry) sortObjectList(objList []*objItem, sorter ObjectSorter) []*objItem {
	sortHelper := sortHelper{objList: objList, sorter: sorter}

	sort.Sort(sortHelper)
	return objList
}
