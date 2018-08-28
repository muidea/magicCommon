package util

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

func (s *registry) run() {
	objectInfo := map[string]interface{}{}
	for {
		item := <-s.actionChannel
		switch item.action {
		case putObj:
			objectInfo[item.id] = item.data
		case getObj:
			obj, ok := objectInfo[item.id]
			if ok {
				item.reply <- obj
			} else {
				item.reply <- nil
			}
		case fetchObjList:
			retList := ObjectList{}
			for _, val := range objectInfo {
				retList = append(retList, val)
			}
			item.reply <- &retList
		case filterObj:
			retList := ObjectList{}
			filter := item.data.(ObjectFilter)
			for _, val := range objectInfo {
				if filter.Filter(val) {
					retList = append(retList, val)
				}
			}
			item.reply <- &retList
		case removeObj:
			delete(objectInfo, item.id)
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
