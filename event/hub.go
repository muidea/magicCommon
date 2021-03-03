package event

import (
	"strings"
	"sync"
)

type Event interface {
	ID() string
	Source() string
	Destination() string
	Data() interface{}
	Match(pattern string) bool
}

type Result interface {
	Set(interface{})
	Get() interface{}
}

type Observer interface {
	ID() string
	Notify(event Event, result Result)
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

func NewHub() Hub {
	hub := &hImpl{event2Observer: ID2ObserverMap{}, actionChannel: make(chan action)}
	go hub.run()
	return hub
}

func matchID(pattern, id string) bool {
	pItems := strings.Split(pattern, "/")
	iItems := strings.Split(id, "/")
	for ik, iv := range iItems {
		if ik >= len(pItems) {
			return false
		}
		if pItems[ik] == "+" {
			continue
		}
		if pItems[ik] == "#" {
			return true
		}
		if pItems[ik] == iv {
			continue
		}
		return false
	}

	return len(pItems) == len(iItems)
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

	action := &subscribeData{eventID: eventID, observer: observer}

	s.actionChannel <- action
	return
}

func (s *hImpl) Unsubscribe(eventID string, observer Observer) {
	if s.terminateFlag {
		return
	}

	action := &unsubscribeData{eventID: eventID, observer: observer}

	s.actionChannel <- action
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

	result := NewResult()
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
		case unsubscribe:
			data := action.(*unsubscribeData)
			s.unsubscribeInternal(data.eventID, data.observer)
		case post:
			data := action.(*postData)
			s.postInternal(data.event)
		case send:
			data := action.(*sendData)
			result := NewResult()
			s.sendInternal(data.event, result)
			data.result <- result
		case terminate:
			data := action.(*terminateData)
			data.result <- true
			break
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
	observerList = append(observerList, observer)
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
