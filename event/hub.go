package event

import (
	"context"
	"fmt"
	"strings"
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/execute"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/util"
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
	default:
		log.Warnf("illegal value, not string, value:%v", val)
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
	default:
		log.Warnf("illegal value, not int, value:%v", val)
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
	default:
		log.Warnf("illegal value, not bool, value:%v", val)
	}

	return false
}

type Event interface {
	ID() string
	Source() string
	Destination() string
	Header() Values
	Context() context.Context
	BindContext(ctx context.Context)
	Data() interface{}
	SetData(key string, val interface{})
	GetData(key string) interface{}
	Match(pattern string) bool
}

type Result interface {
	Error() *cd.Result
	Set(data interface{}, err *cd.Result)
	Get() (interface{}, *cd.Result)
	SetVal(key string, val interface{})
	GetVal(key string) interface{}
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
type actionChannel chan action
type ID2ActionChanelMap map[string]actionChannel

func NewHub(capacitySize int) Hub {
	hub := &hubImpl{
		Execute:             execute.NewExecute(capacitySize),
		event2Observer:      ID2ObserverMap{},
		actionChannel:       make(chan action),
		event2ActionChannel: ID2ActionChanelMap{},
		terminateFlag:       false,
	}
	go hub.run()
	return hub
}

func NewSimpleObserver(id string, hub Hub) SimpleObserver {
	return &simpleObserver{id: id, eventHub: hub, id2ObserverFunc: ID2ObserverFuncMap{}}
}

func MatchValue(pattern, val string) bool {
	pIdx := 0
	pOffset := 0
	pItems := strings.Split(pattern, "/")

	iIdx := 0
	iOffset := 0
	iItems := strings.Split(val, "/")
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
	waitGroup *sync.WaitGroup
	result    chan bool
}

func (s *terminateData) Code() int {
	return terminate
}

func (s actionChannel) run(hubPtr *hubImpl) {
	terminateFlag := false
	for {
		actionData, actionOK := <-s
		if !actionOK {
			log.Criticalf("eventHub actionChannel unexpect!")
			break
		}

		actionFunc := func() {
			switch actionData.Code() {
			case subscribe:
				data := actionData.(*subscribeData)
				hubPtr.subscribeInternal(data.eventID, data.observer)
				data.result <- true
			case unsubscribe:
				data := actionData.(*unsubscribeData)
				hubPtr.unsubscribeInternal(data.eventID, data.observer)
				data.result <- true
			case post:
				data := actionData.(*postData)
				hubPtr.postInternal(data.event)
			case send:
				data := actionData.(*sendData)
				result := NewResult(data.event.ID(), data.event.Source(), data.event.Destination())
				hubPtr.sendInternal(data.event, result)
				data.result <- result
			case terminate:
				data := actionData.(*terminateData)
				data.result <- true
				data.waitGroup.Done()
				terminateFlag = true
			default:
			}
		}

		hubPtr.Execute.Run(actionFunc)

		if terminateFlag {
			break
		}
	}
}

type hubImpl struct {
	execute.Execute
	event2Lock     sync.RWMutex
	event2Observer ID2ObserverMap

	actionChannel       actionChannel
	event2ActionChannel ID2ActionChanelMap

	terminateFlag bool
}

func (s *hubImpl) Subscribe(eventID string, observer Observer) {
	if s.terminateFlag {
		return
	}

	replay := make(chan bool)
	s.Execute.Run(func() {
		actionData := &subscribeData{eventID: eventID, observer: observer, result: replay}

		s.actionChannel <- actionData
	})
	<-replay

	return
}

func (s *hubImpl) Unsubscribe(eventID string, observer Observer) {
	if s.terminateFlag {
		return
	}

	replay := make(chan bool)
	s.Execute.Run(func() {
		actionData := &unsubscribeData{eventID: eventID, observer: observer, result: replay}

		s.actionChannel <- actionData
	})
	<-replay

	return
}

func (s *hubImpl) Post(event Event) {
	if s.terminateFlag {
		return
	}

	actionData := &postData{event: event}
	var eventChannel actionChannel
	func() {
		s.event2Lock.Lock()
		defer s.event2Lock.Unlock()
		channelVal, channelOK := s.event2ActionChannel[event.Destination()]
		if !channelOK {
			channelVal = make(actionChannel)
			go channelVal.run(s)

			s.event2ActionChannel[event.Destination()] = channelVal
		}

		eventChannel = channelVal
	}()

	if event.Source() == event.Destination() {
		s.Execute.Run(func() {
			eventChannel <- actionData
		})
	} else {
		eventChannel <- actionData
	}

	return
}

func (s *hubImpl) Send(event Event) (ret Result) {
	if s.terminateFlag {
		return
	}

	replay := make(chan Result)
	actionData := &sendData{event: event, result: replay}

	var eventChannel actionChannel
	func() {
		s.event2Lock.Lock()
		defer s.event2Lock.Unlock()
		channelVal, channelOK := s.event2ActionChannel[event.Destination()]
		if !channelOK {
			channelVal = make(actionChannel)
			go channelVal.run(s)

			s.event2ActionChannel[event.Destination()] = channelVal
		}

		eventChannel = channelVal
	}()

	if event.Source() == event.Destination() {
		s.Execute.Run(func() {
			eventChannel <- actionData
		})
	} else {
		eventChannel <- actionData
	}

	ret = <-replay
	return
}

func (s *hubImpl) Call(event Event) Result {
	if s.terminateFlag {
		return nil
	}

	result := NewResult(event.ID(), event.Source(), event.Destination())
	s.sendInternal(event, result)
	return result
}

func (s *hubImpl) Terminate() {
	if s.terminateFlag {
		return
	}

	s.terminateFlag = true
	var waitGroup sync.WaitGroup
	replay := make(chan bool)
	actionData := &terminateData{result: replay, waitGroup: &waitGroup}
	go func() {
		s.event2Lock.Lock()
		defer s.event2Lock.Unlock()

		for _, val := range s.event2ActionChannel {
			waitGroup.Add(1)
			val <- actionData
		}
	}()
	waitGroup.Add(1)
	s.actionChannel <- actionData

	<-replay

	waitGroup.Wait()

	s.event2Lock.Lock()
	defer s.event2Lock.Unlock()
	for _, val := range s.event2ActionChannel {
		close(val)
	}
	s.event2Observer = ID2ObserverMap{}
	s.event2ActionChannel = ID2ActionChanelMap{}
	return
}

func (s *hubImpl) run() {
	s.actionChannel.run(s)

	close(s.actionChannel)
}

func (s *hubImpl) subscribeInternal(eventID string, observer Observer) {
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

func (s *hubImpl) unsubscribeInternal(eventID string, observer Observer) {
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

func (s *hubImpl) postInternal(event Event) {
	matchList := ObserverList{}

	func() {
		s.event2Lock.RLock()
		defer s.event2Lock.RUnlock()
		for key, value := range s.event2Observer {
			if MatchValue(key, event.ID()) {
				for _, sv := range value {
					if MatchValue(event.Destination(), sv.ID()) {
						matchList = append(matchList, sv)
					}
				}
			}
		}
	}()

	for _, sv := range matchList {
		func() {
			defer func() {
				if err := recover(); err != nil {
					stackInfo := util.GetStack(3)
					log.Warnf("notify event exception, event:%v \nPANIC:%v \nstack:%s", event.ID(), err, stackInfo)
				}
			}()
			sv.Notify(event, nil)
		}()
	}
}

func (s *hubImpl) sendInternal(event Event, result Result) {
	matchList := ObserverList{}
	finalFlag := false

	func() {
		s.event2Lock.RLock()
		defer s.event2Lock.RUnlock()
		for key, value := range s.event2Observer {
			if MatchValue(key, event.ID()) {
				for _, sv := range value {
					if MatchValue(event.Destination(), sv.ID()) {
						matchList = append(matchList, sv)
						finalFlag = true
					}
				}
			}
		}
	}()

	for _, sv := range matchList {
		func() {
			defer func() {
				if err := recover(); err != nil {
					stackInfo := util.GetStack(3)
					log.Warnf("notify event exception, event:%v \nPANIC:%v \nstack:%s", event.ID(), err, stackInfo)

					if result != nil {
						result.Set(nil, cd.NewError(cd.UnExpected, fmt.Sprintf("%v", err)))
					}
				}
			}()

			sv.Notify(event, result)
		}()
	}

	if !finalFlag && result != nil {
		result.Set(nil, cd.NewWarn(cd.Warned, fmt.Sprintf("missing observer, event id:%s", event.ID())))
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
		func() {
			defer func() {
				if err := recover(); err != nil {
					stackInfo := util.GetStack(3)
					log.Warnf("notify event exception, event:%v \nPANIC:%v \nstack:%s", event.ID(), err, stackInfo)

					if result != nil {
						result.Set(nil, cd.NewError(cd.UnExpected, fmt.Sprintf("%v", err)))
					}
				}
			}()

			funcVal(event, result)
		}()
	}
}

func (s *simpleObserver) Subscribe(eventID string, observerFunc ObserverFunc) {
	okFlag := false
	func() {
		s.idLock.Lock()
		defer s.idLock.Unlock()

		_, ok := s.id2ObserverFunc[eventID]
		if ok {
			log.Warnf("duplicate eventID:%v", eventID)
			return
		}

		s.id2ObserverFunc[eventID] = observerFunc
		okFlag = true
	}()

	if okFlag {
		s.eventHub.Subscribe(eventID, s)
	}
}

func (s *simpleObserver) Unsubscribe(eventID string) {
	okFlag := false
	func() {
		s.idLock.Lock()
		defer s.idLock.Unlock()

		_, ok := s.id2ObserverFunc[eventID]
		if !ok {
			log.Warnf("not exist eventID:%v", eventID)
			return
		}

		delete(s.id2ObserverFunc, eventID)
		okFlag = true
	}()

	if okFlag {
		s.eventHub.Unsubscribe(eventID, s)
	}
}
