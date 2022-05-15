package event

import (
	"strings"
	"sync"

	log "github.com/cihub/seelog"
)

type Values map[string]interface{}

func (s Values) Set(key string, value interface{}) {
	s[key] = value
}

func (s Values) Get(key string) interface{} {
	val, ok := s[key]
	if ok {
		return val
	}

	return nil
}

func (s Values) GetString(key string) string {
	val := s.Get(key)
	if val == nil {
		return ""
	}

	switch val.(type) {
	case string:
		return val.(string)
	}

	return ""
}

func (s Values) GetInt(key string) int {
	val := s.Get(key)
	if val == nil {
		return 0
	}

	switch val.(type) {
	case int:
		return val.(int)
	}

	return 0
}

func (s Values) GetBool(key string) bool {
	val := s.Get(key)
	if val == nil {
		return false
	}

	switch val.(type) {
	case bool:
		return val.(bool)
	}

	return false
}

type Event interface {
	ID() string
	Source() string
	Destination() string
	Header() Values
	Data() interface{}
	Match(pattern string) bool
}

type Result interface {
	Error() error
	Set(data interface{}, err error)
	Get() (interface{}, error)
}

type Observer interface {
	ID() string
	Notify(event Event, result Result)
}

type SimpleObserver interface {
	Observer
	Subscribe(eventID string, observerFunc ObserverFunc)
	Unsubscribe(eventID string)
}

type Hub interface {
	Subscribe(eventID string, observer Observer)
	Unsubscribe(eventID string, observer Observer)
	Post(event Event)
	Send(event Event) Result
	Call(event Event) Result
	Terminate()
}

type ObserverList []Observer
type ID2ObserverMap map[string]ObserverList
type ObserverFunc func(Event, Result)
type ID2ObserverFuncMap map[string]ObserverFunc

func NewHub() Hub {
	hub := &hImpl{event2Observer: ID2ObserverMap{}, actionChannel: make(chan action)}
	go hub.run()
	return hub
}

func NewSimpleObserver(id string, hub Hub) SimpleObserver {
	return &simpleObserver{id: id, eventHub: hub, id2ObserverFunc: ID2ObserverFuncMap{}}
}

func matchID(pattern, id string) bool {
	pIdx := 0
	pOffset := 0
	pItems := strings.Split(pattern, "/")

	iIdx := 0
	iOffset := 0
	iItems := strings.Split(id, "/")
	for iIdx < len(iItems) {
		iv := iItems[iIdx]
		if pIdx >= len(pItems) {
			return false
		}

		pv := pItems[pIdx]
		if pv == iv {
			pIdx++
			iIdx++
			continue
		}

		if (pv == "+" || pv == ":id") && iv != "" {
			pIdx++
			iIdx++
			continue
		}

		if pv == "#" && iv != "" {
			pOffset++
			if pIdx+pOffset >= len(pItems) {
				return true
			}

			iOffset++
			if iIdx+iOffset >= len(iItems) {
				return false
			}

			for iIdx+iOffset < len(iItems) {
				if pIdx+pOffset >= len(pItems) {
					return false
				}

				pn := pItems[pIdx+pOffset]
				in := iItems[iIdx+iOffset]
				if pn == in {
					pIdx += pOffset + 1
					pOffset = 0
					break
				}
				if pn == "+" || pn == ":id" {
					if pIdx+pOffset+1 >= len(pItems) {
						return true
					}

					pnn := pItems[pIdx+pOffset+1]
					if pnn == in {
						pIdx += pOffset + 2
						pOffset = 0
						break
					}

					if pv != "#" {
						pOffset++
					}
				}

				iOffset++
				continue
			}

			iIdx += iOffset + 1
			iOffset = 0
			if pIdx > iIdx {
				return false
			}

			continue
		}

		return false
	}

	return pIdx == len(pItems)
}

const (
	subscribe   = 1
	unsubscribe = 2
	post        = 3
	send        = 4
	terminate   = 5
)

type action interface {
	Code() int
}

type subscribeData struct {
	eventID  string
	observer Observer
	result   chan bool
}

func (s *subscribeData) Code() int {
	return subscribe
}

type unsubscribeData subscribeData

func (s *unsubscribeData) Code() int {
	return unsubscribe
}

type postData struct {
	event Event
}

func (s *postData) Code() int {
	return post
}

type sendData struct {
	event  Event
	result chan Result
}

func (s *sendData) Code() int {
	return send
}

type terminateData struct {
	result chan bool
}

func (s *terminateData) Code() int {
	return terminate
}

type hImpl struct {
	event2Lock     sync.RWMutex
	event2Observer ID2ObserverMap

	actionChannel chan action
	terminateFlag bool
}

func (s *hImpl) Subscribe(eventID string, observer Observer) {
	if s.terminateFlag {
		return
	}

	replay := make(chan bool)
	go func() {
		action := &subscribeData{eventID: eventID, observer: observer, result: replay}

		s.actionChannel <- action
	}()
	<-replay

	return
}

func (s *hImpl) Unsubscribe(eventID string, observer Observer) {
	if s.terminateFlag {
		return
	}

	replay := make(chan bool)
	go func() {
		action := &unsubscribeData{eventID: eventID, observer: observer, result: replay}

		s.actionChannel <- action
	}()
	<-replay

	return
}

func (s *hImpl) Post(event Event) {
	if s.terminateFlag {
		return
	}

	go func() {
		action := &postData{event: event}

		s.actionChannel <- action
	}()
	return
}

func (s *hImpl) Send(event Event) (ret Result) {
	if s.terminateFlag {
		return
	}

	replay := make(chan Result)

	go func() {
		action := &sendData{event: event, result: replay}

		s.actionChannel <- action
	}()

	ret = <-replay
	return
}

func (s *hImpl) Call(event Event) Result {
	if s.terminateFlag {
		return nil
	}

	result := NewResult(event.ID(), event.Source(), event.Destination())
	s.sendInternal(event, result)
	return result
}

func (s *hImpl) Terminate() {
	if s.terminateFlag {
		return
	}

	replay := make(chan bool)
	go func() {
		action := &terminateData{result: replay}

		s.actionChannel <- action
	}()

	<-replay

	s.event2Observer = nil
	return
}

func (s *hImpl) run() {
	s.terminateFlag = false
	for action := range s.actionChannel {
		switch action.Code() {
		case subscribe:
			data := action.(*subscribeData)
			s.subscribeInternal(data.eventID, data.observer)
			data.result <- true
		case unsubscribe:
			data := action.(*unsubscribeData)
			s.unsubscribeInternal(data.eventID, data.observer)
			data.result <- true
		case post:
			data := action.(*postData)
			s.postInternal(data.event)
		case send:
			data := action.(*sendData)
			result := NewResult(data.event.ID(), data.event.Source(), data.event.Destination())
			s.sendInternal(data.event, result)
			data.result <- result
		case terminate:
			data := action.(*terminateData)
			data.result <- true
		default:
		}
	}

	s.terminateFlag = true
	close(s.actionChannel)
}

func (s *hImpl) subscribeInternal(eventID string, observer Observer) {
	s.event2Lock.Lock()
	defer s.event2Lock.Unlock()

	observerList, observerOK := s.event2Observer[eventID]
	if !observerOK {
		observerList = ObserverList{}
	}
	existFlag := false
	for _, val := range observerList {
		if val.ID() == observer.ID() {
			existFlag = true
			break
		}
	}
	if !existFlag {
		observerList = append(observerList, observer)
	}
	s.event2Observer[eventID] = observerList
}

func (s *hImpl) unsubscribeInternal(eventID string, observer Observer) {
	s.event2Lock.Lock()
	defer s.event2Lock.Unlock()

	observerList, observerOK := s.event2Observer[eventID]
	if !observerOK {
		return
	}

	newList := ObserverList{}
	for _, val := range observerList {
		if val.ID() == observer.ID() {
			continue
		}

		newList = append(newList, val)
	}
	if len(newList) > 0 {
		s.event2Observer[eventID] = newList
		return
	}

	delete(s.event2Observer, eventID)
}

func (s *hImpl) postInternal(event Event) {
	s.event2Lock.RLock()
	defer s.event2Lock.RUnlock()

	for key, value := range s.event2Observer {
		if matchID(key, event.ID()) {
			for _, sv := range value {
				if matchID(event.Destination(), sv.ID()) {
					sv.Notify(event, nil)
				}
			}
		}
	}
}

func (s *hImpl) sendInternal(event Event, result Result) {
	s.event2Lock.RLock()
	defer s.event2Lock.RUnlock()

	for key, value := range s.event2Observer {
		if matchID(key, event.ID()) {
			for _, sv := range value {
				if matchID(event.Destination(), sv.ID()) {
					sv.Notify(event, result)
				}
			}
		}
	}
}

type simpleObserver struct {
	id              string
	eventHub        Hub
	id2ObserverFunc ID2ObserverFuncMap
	idLock          sync.RWMutex
}

func (s *simpleObserver) ID() string {
	return s.id
}

func (s *simpleObserver) Notify(event Event, result Result) {
	var funcVal ObserverFunc
	func() {
		s.idLock.RLock()
		defer s.idLock.RUnlock()

		for k, v := range s.id2ObserverFunc {
			if event.Match(k) {
				funcVal = v
				break
			}
		}
	}()

	if funcVal != nil {
		funcVal(event, result)
	}

	return
}

func (s *simpleObserver) Subscribe(eventID string, observerFunc ObserverFunc) {
	s.idLock.Lock()
	defer s.idLock.Unlock()

	_, ok := s.id2ObserverFunc[eventID]
	if ok {
		log.Errorf("duplicate eventID:%v", eventID)
		return
	}

	s.id2ObserverFunc[eventID] = observerFunc
	s.eventHub.Subscribe(eventID, s)
}

func (s *simpleObserver) Unsubscribe(eventID string) {
	s.idLock.Lock()
	defer s.idLock.Unlock()

	_, ok := s.id2ObserverFunc[eventID]
	if !ok {
		log.Errorf("not exist eventID:%v", eventID)
		return
	}

	delete(s.id2ObserverFunc, eventID)
	s.eventHub.Unsubscribe(eventID, s)
}
