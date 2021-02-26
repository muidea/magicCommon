package event

import "strings"

type Event interface {
	ID() string
	Source() string
	Destination() string
	Data() interface{}
	Match(pattern string) bool
}

type Observer interface {
	ID() string
	Notify(event Event)
}

type Hub interface {
	Subscribe(eventID string, observer Observer)
	Unsubscribe(eventID string, observer Observer)
	Post(event Event)
	Send(event Event)
	Terminate()
}

type ObserverList []Observer
type ID2ObserverMap map[string]ObserverList

func NewHub() Hub {
	hub := &hImpl{actionChannel: make(chan action)}
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
	result chan bool
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
	actionChannel chan action
}

func (s *hImpl) run() {
	event2Observer := ID2ObserverMap{}
	for action := range s.actionChannel {
		switch action.Code() {
		case subscribe:
			data := action.(*subscribeData)
			event2Observer = s.subscribeInternal(data.eventID, data.observer, event2Observer)
		case unsubscribe:
			data := action.(*unsubscribeData)
			event2Observer = s.unsubscribeInternal(data.eventID, data.observer, event2Observer)
		case post:
			data := action.(*postData)
			event2Observer = s.postInternal(data.event, event2Observer)
		case send:
			data := action.(*sendData)
			event2Observer = s.sendInternal(data.event, event2Observer)
			data.result <- true
		case terminate:
			data := action.(*terminateData)
			data.result <- true
			break
		default:
		}
	}

	close(s.actionChannel)
}

func (s *hImpl) Subscribe(eventID string, observer Observer) {
	action := &subscribeData{eventID: eventID, observer: observer}

	s.actionChannel <- action
	return
}

func (s *hImpl) Unsubscribe(eventID string, observer Observer) {
	action := &unsubscribeData{eventID: eventID, observer: observer}

	s.actionChannel <- action
	return
}

func (s *hImpl) Post(event Event) {
	action := &postData{event: event}

	s.actionChannel <- action
	return
}

func (s *hImpl) Send(event Event) {
	replay := make(chan bool)
	action := &sendData{event: event, result: replay}

	s.actionChannel <- action
	<-replay

	return
}

func (s *hImpl) Terminate() {
	replay := make(chan bool)
	action := &terminateData{result: replay}

	s.actionChannel <- action
	<-replay

	return
}

func (s *hImpl) subscribeInternal(eventID string, observer Observer, event2Observer ID2ObserverMap) ID2ObserverMap {
	observerList, observerOK := event2Observer[eventID]
	if !observerOK {
		observerList = ObserverList{}
	}
	observerList = append(observerList, observer)
	event2Observer[eventID] = observerList

	return event2Observer
}

func (s *hImpl) unsubscribeInternal(eventID string, observer Observer, event2Observer ID2ObserverMap) ID2ObserverMap {
	observerList, observerOK := event2Observer[eventID]
	if !observerOK {
		return event2Observer
	}

	newList := ObserverList{}
	for _, val := range observerList {
		if val.ID() == observer.ID() {
			continue
		}

		newList = append(newList, val)
	}
	if len(newList) > 0 {
		event2Observer[eventID] = newList
		return event2Observer
	}

	delete(event2Observer, eventID)
	return event2Observer
}

func (s *hImpl) postInternal(event Event, event2Observer ID2ObserverMap) ID2ObserverMap {
	for key, value := range event2Observer {
		if matchID(key, event.ID()) {
			for _, sv := range value {
				if matchID(event.Destination(), sv.ID()) {
					sv.Notify(event)
				}
			}
		}
	}
	return event2Observer
}

func (s *hImpl) sendInternal(event Event, event2Observer ID2ObserverMap) ID2ObserverMap {
	for key, value := range event2Observer {
		if matchID(key, event.ID()) {
			for _, sv := range value {
				if matchID(event.Destination(), sv.ID()) {
					sv.Notify(event)
				}
			}
		}
	}
	return event2Observer
}
