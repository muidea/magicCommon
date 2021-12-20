package util

import (
	"sort"
	"sync"
)

// ObjectList object list
type ObjectList []interface{}

// ObjectRegistry object 仓库
type ObjectRegistry interface {
	Put(id string, object interface{})
	Get(id string) interface{}
	Sort(sorter ObjectSorter)
	FetchList(pageFilter *Pagination) *ObjectList
	Filter(filter Filter, pageFilter *Pagination) *ObjectList
	Remove(id string)
}

// CatalogObjectRegistry catalog object registry
type CatalogObjectRegistry interface {
	Put(id, catalog string, object interface{})
	Get(id, catalog string) interface{}
	Sort(catalog string, sorter ObjectSorter)
	FetchList(catalog string, pageFilter *Pagination) *ObjectList
	Filter(catalog string, filter Filter, pageFilter *Pagination) *ObjectList
	Remove(id, catalog string)
}

// NewRegistry create new Registry
func NewRegistry() ObjectRegistry {
	impl := &registry{actionChannel: make(actionChannel)}
	go impl.run()

	return impl
}

// NewCatalogRegistry create new catalog registry
func NewCatalogRegistry() CatalogObjectRegistry {
	impl := &catalogRegistry{catalogRegistry: map[string]ObjectRegistry{}}
	return impl
}

type catalogRegistry struct {
	catalogRegistry map[string]ObjectRegistry
	registryLock    sync.RWMutex
}

func (s *catalogRegistry) Put(id, catalog string, object interface{}) {
	s.registryLock.Lock()
	defer s.registryLock.Unlock()

	registry, ok := s.catalogRegistry[catalog]
	if !ok {
		registry = NewRegistry()
		s.catalogRegistry[catalog] = registry
	}

	registry.Put(id, object)
}

func (s *catalogRegistry) Get(id, catalog string) interface{} {
	s.registryLock.RLock()
	defer s.registryLock.RUnlock()

	registry, ok := s.catalogRegistry[catalog]
	if !ok {
		return nil
	}

	return registry.Get(id)
}

func (s *catalogRegistry) Sort(catalog string, sorter ObjectSorter) {
	s.registryLock.RLock()
	defer s.registryLock.RUnlock()

	registry, ok := s.catalogRegistry[catalog]
	if !ok {
		return
	}

	registry.Sort(sorter)
}

func (s *catalogRegistry) FetchList(catalog string, pageFilter *Pagination) *ObjectList {
	s.registryLock.RLock()
	defer s.registryLock.RUnlock()

	registry, ok := s.catalogRegistry[catalog]
	if !ok {
		return nil
	}

	return registry.FetchList(pageFilter)
}

func (s *catalogRegistry) Filter(catalog string, filter Filter, pageFilter *Pagination) *ObjectList {
	s.registryLock.RLock()
	defer s.registryLock.RUnlock()

	registry, ok := s.catalogRegistry[catalog]
	if !ok {
		return nil
	}

	return registry.Filter(filter, pageFilter)
}

func (s *catalogRegistry) Remove(id, catalog string) {
	s.registryLock.RLock()
	defer s.registryLock.RUnlock()

	registry, ok := s.catalogRegistry[catalog]
	if !ok {
		return
	}

	registry.Remove(id)
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
	pageFilter *Pagination
}

type fetchResult struct {
	result
	objList *ObjectList
}

type filterParam struct {
	filter     Filter
	pageFilter *Pagination
}

type filterResult struct {
	result
	objList *ObjectList
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

func (s *registry) FetchList(pageFilter *Pagination) *ObjectList {
	reply := make(chan interface{})

	item := &actionObj{action: fetchObjList, param: &fetchParam{pageFilter: pageFilter}, reply: reply}
	s.actionChannel <- item

	val := <-reply
	result := val.(*fetchResult)
	return result.objList
}

func (s *registry) Filter(filter Filter, pageFilter *Pagination) *ObjectList {
	reply := make(chan interface{})

	item := &actionObj{action: filterObj, param: &filterParam{filter: filter, pageFilter: pageFilter}, reply: reply}
	s.actionChannel <- item

	val := <-reply
	result := val.(*filterResult)
	return result.objList
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
