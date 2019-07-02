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
	Sort(sorter ObjectSorter)
	FetchList(pageFilter *PageFilter) (*ObjectList, int)
	Filter(filter ObjectFilter, pageFilter *PageFilter) (*ObjectList, int)
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
	sortObj
	fetchObjList
	filterObj
	removeObj
)

type result struct {
	result bool
}

type putParam struct {
	id   string
	data interface{}
}

type putResult struct {
	result
}

type getParam struct {
	id string
}

type getResult struct {
	result
	obj interface{}
}

type sortParam struct {
	sorter ObjectSorter
}

type sortResult struct {
	result
}

type fetchParam struct {
	pageFilter *PageFilter
}

type fetchResult struct {
	result
	objList   *ObjectList
	totalSize int
}

type filterParam struct {
	filter     ObjectFilter
	pageFilter *PageFilter
}

type filterResult struct {
	result
	objList   *ObjectList
	totalSize int
}

type removeParam struct {
	id string
}

type removeResult struct {
	result
}

type actionObj struct {
	action int
	param  interface{}
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
			found := false
			param := item.param.(*putParam)
			result := &putResult{result: result{result: true}}
			for idx := range objectItemList {
				val := objectItemList[idx]
				if val.id == param.id {
					objectItemList[idx] = &objItem{id: param.id, obj: param.data}
					found = true
					break
				}
			}
			if !found {
				objectItemList = append(objectItemList, &objItem{id: param.id, obj: param.data})
			}
			item.reply <- result
		case getObj:
			param := item.param.(*getParam)
			result := &getResult{result: result{result: false}}
			for _, val := range objectItemList {
				if val.id == param.id {
					result.obj = val.obj
					result.result.result = true
					break
				}
			}
			item.reply <- result
		case sortObj:
			retList := ObjectList{}
			param := item.param.(*sortParam)
			result := &sortResult{result: result{result: true}}
			sorter := param.sorter
			objectItemList = s.sortObjectList(objectItemList, sorter)
			for _, val := range objectItemList {
				retList = append(retList, val.obj)
			}

			item.reply <- result
		case fetchObjList:
			param := item.param.(*fetchParam)
			result := &fetchResult{result: result{result: true}}
			result.totalSize = len(objectItemList)

			retList := ObjectList{}
			if param.pageFilter == nil {
				for _, val := range objectItemList {
					retList = append(retList, val.obj)
				}
			} else {
				lIdx := param.pageFilter.PageNum * param.pageFilter.PageSize
				hIdx := lIdx + param.pageFilter.PageSize
				subList := objectItemList[lIdx:hIdx]
				for _, val := range subList {
					retList = append(retList, val.obj)
				}
			}
			result.objList = &retList

			item.reply <- result
		case filterObj:
			param := item.param.(*filterParam)
			result := &filterResult{result: result{result: true}}

			retList := ObjectList{}
			filter := param.filter
			for _, val := range objectItemList {
				if filter.Filter(val.obj) {
					retList = append(retList, val.obj)
				}
			}

			result.totalSize = len(retList)
			if param.pageFilter == nil {
				result.objList = &retList
			} else {
				lIdx := param.pageFilter.PageNum * param.pageFilter.PageSize
				hIdx := lIdx + param.pageFilter.PageSize
				subList := retList[lIdx:hIdx]
				result.objList = &subList
			}

			item.reply <- result
		case removeObj:
			param := item.param.(*removeParam)
			result := &removeResult{result: result{result: true}}

			newList := []*objItem{}
			for idx, val := range objectItemList {
				if val.id == param.id {
					newList = append(newList, objectItemList[:idx]...)
					if idx < len(objectItemList)-1 {
						newList = append(newList, objectItemList[idx+1:]...)
						break
					}
				}
			}

			objectItemList = newList
			item.reply <- result
		}
	}
}

func (s *registry) Put(id string, object interface{}) {
	reply := make(chan interface{})
	item := &actionObj{action: putObj, param: &putParam{id: id, data: object}, reply: reply}
	s.actionChannel <- item

	<-reply
}

func (s *registry) Get(id string) interface{} {
	reply := make(chan interface{})

	item := &actionObj{action: getObj, param: &getParam{id: id}, reply: reply}
	s.actionChannel <- item

	val := <-reply
	return val.(*getResult).obj
}

func (s *registry) Sort(sorter ObjectSorter) {
	reply := make(chan interface{})

	item := &actionObj{action: sortObj, param: &sortParam{sorter: sorter}, reply: reply}
	s.actionChannel <- item

	<-reply
}

func (s *registry) FetchList(pageFilter *PageFilter) (*ObjectList, int) {
	reply := make(chan interface{})

	item := &actionObj{action: fetchObjList, param: &fetchParam{pageFilter: pageFilter}, reply: reply}
	s.actionChannel <- item

	val := <-reply
	result := val.(*fetchResult)
	return result.objList, result.totalSize
}

func (s *registry) Filter(filter ObjectFilter, pageFilter *PageFilter) (*ObjectList, int) {
	reply := make(chan interface{})

	item := &actionObj{action: filterObj, param: &filterParam{filter: filter, pageFilter: pageFilter}, reply: reply}
	s.actionChannel <- item

	val := <-reply
	result := val.(*filterResult)
	return result.objList, result.totalSize
}

func (s *registry) Remove(id string) {
	reply := make(chan interface{})

	item := &actionObj{action: removeObj, param: &removeParam{id: id}, reply: reply}
	s.actionChannel <- item

	<-reply
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
