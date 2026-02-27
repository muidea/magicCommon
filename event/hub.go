package event

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/execute"
	"github.com/muidea/magicCommon/foundation/util"
	"log/slog"
)

type Values map[string]any

func (s Values) Set(key string, value any) {
	s[key] = value
}

func (s Values) Get(key string) any {
	val, ok := s[key]
	if ok {
		return val
	}

	return nil
}

func (s Values) GetString(key string) string {
	return GetTypedValue[string](s, key, "", "string")
}

func (s Values) GetInt(key string) int {
	return GetTypedValue[int](s, key, 0, "int")
}

func (s Values) GetBool(key string) bool {
	return GetTypedValue[bool](s, key, false, "bool")
}

// GetTypedValue 泛型方法获取指定类型的值
func GetTypedValue[T any](s Values, key string, defaultValue T, typeName string) T {
	val := s.Get(key)
	if val == nil {
		return defaultValue
	}

	if v, ok := val.(T); ok {
		return v
	}

	slog.Warn("illegal value, not expected type", "type", typeName, "value", val)
	return defaultValue
}

type Event interface {
	ID() string
	Source() string
	Destination() string
	Header() Values
	Context() context.Context
	BindContext(ctx context.Context)
	Data() any
	SetData(key string, val any)
	GetData(key string) any
	Match(pattern string) bool
}

type Result interface {
	Error() *cd.Error
	Set(data any, err *cd.Error)
	Get() (any, *cd.Error)
	SetVal(key string, val any)
	GetVal(key string) any
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
	Terminate()
}

type ObserverList []Observer
type ID2ObserverMap map[string]ObserverList
type ObserverFunc func(Event, Result)
type ID2ObserverFuncMap map[string]ObserverFunc
type actionChannel chan action
type ID2ActionChanelMap map[string]actionChannel

func notificationEvent(sv Observer, ev Event, re Result) {
	defer func() {
		if err := recover(); err != nil {
			stackInfo := util.GetStack(3)
			slog.Warn("notify event exception", "event_id", ev.ID(), "source", ev.Source(), "destination", ev.Destination(), "panic", err, "stack", stackInfo)

			if re != nil {
				re.Set(nil, cd.NewError(cd.Unexpected, fmt.Sprintf("%v", err)))
			}
		}
	}()

	sv.Notify(ev, re)
}

func NewHub(capacitySize int) Hub {
	hub := &hubImpl{
		Execute:                  execute.NewExecute(capacitySize),
		event2Observer:           ID2ObserverMap{},
		hubActionChannel:         make(chan action),
		observerID2ActionChannel: ID2ActionChanelMap{},
		terminateFlag:            false,
	}
	go hub.run()
	return hub
}

func NewSimpleObserver(id string, hub Hub) SimpleObserver {
	return &simpleObserver{id: id, eventHub: hub, eventID2ObserverFunc: ID2ObserverFuncMap{}}
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
			// channel 被正常关闭，退出循环
			break
		}

		// 所有操作都顺序执行，以保证同一个观察者的事件顺序
		// 包括 post 和 send 操作都需要顺序执行
		switch actionData.Code() {
		case subscribe:
			data := actionData.(*subscribeData)
			hubPtr.subscribeInternal(data.eventID, data.observer)
			select {
			case data.result <- true:
				// 成功发送
			case <-time.After(100 * time.Millisecond):
				slog.Warn("timeout sending subscribe result")
			}
		case unsubscribe:
			data := actionData.(*unsubscribeData)
			hubPtr.unsubscribeInternal(data.eventID, data.observer)
			select {
			case data.result <- true:
				// 成功发送
			case <-time.After(100 * time.Millisecond):
				slog.Warn("timeout sending unsubscribe result")
			}
		case post:
			data := actionData.(*postData)
			hubPtr.postInternal(data.event)
		case send:
			data := actionData.(*sendData)
			result := NewResult(data.event.ID(), data.event.Source(), data.event.Destination())
			hubPtr.sendInternal(data.event, result)
			select {
			case data.result <- result:
				// 成功发送
			case <-time.After(100 * time.Millisecond):
				slog.Warn("timeout sending result")
			}
		case terminate:
			data := actionData.(*terminateData)
			data.waitGroup.Done()
			select {
			case data.result <- true:
				// 成功发送
			case <-time.After(100 * time.Millisecond):
				slog.Warn("timeout sending terminate result")
			}
			terminateFlag = true
		default:
			slog.Error("unknown action code", "code", actionData.Code())
		}

		if terminateFlag {
			break
		}
	}
}

type hubImpl struct {
	execute.Execute
	event2ObserverlLock sync.RWMutex
	event2Observer      ID2ObserverMap

	hubActionChannel         actionChannel
	observerID2ChanelLock    sync.RWMutex
	observerID2ActionChannel ID2ActionChanelMap

	terminateFlag bool
}

func (s *hubImpl) Subscribe(eventID string, observer Observer) {
	if s.terminateFlag {
		return
	}

	replay := make(chan bool)
	defer close(replay)
	s.Run(func() {
		actionData := &subscribeData{eventID: eventID, observer: observer, result: replay}

		s.hubActionChannel <- actionData
	})
	<-replay
}

func (s *hubImpl) Unsubscribe(eventID string, observer Observer) {
	if s.terminateFlag {
		return
	}

	replay := make(chan bool)
	defer close(replay)

	s.Run(func() {
		actionData := &unsubscribeData{eventID: eventID, observer: observer, result: replay}

		s.hubActionChannel <- actionData
	})
	<-replay
}

func (s *hubImpl) Post(ev Event) {
	if s.terminateFlag {
		return
	}

	actionData := &postData{event: ev}
	var eventChannel actionChannel
	func() {
		s.observerID2ChanelLock.Lock()
		defer s.observerID2ChanelLock.Unlock()
		channelVal, channelOK := s.observerID2ActionChannel[ev.Destination()]
		if !channelOK {
			channelVal = make(actionChannel)
			go channelVal.run(s)

			s.observerID2ActionChannel[ev.Destination()] = channelVal
		}

		eventChannel = channelVal
	}()

	// 再次检查 terminateFlag，防止竞态条件
	if s.terminateFlag {
		return
	}

	if ev.Source() == ev.Destination() {
		s.Run(func() {
			select {
			case eventChannel <- actionData:
				// 成功发送
			case <-time.After(100 * time.Millisecond):
				slog.Warn("timeout sending post data to channel")
			}
		})
	} else {
		select {
		case eventChannel <- actionData:
			// 成功发送
		case <-time.After(100 * time.Millisecond):
			slog.Warn("timeout sending post data to channel")
		}
	}
}

func (s *hubImpl) Send(ev Event) (ret Result) {
	if s.terminateFlag {
		return
	}

	replay := make(chan Result)
	defer close(replay)

	actionData := &sendData{event: ev, result: replay}

	var eventChannel actionChannel
	func() {
		s.observerID2ChanelLock.Lock()
		defer s.observerID2ChanelLock.Unlock()
		channelVal, channelOK := s.observerID2ActionChannel[ev.Destination()]
		if !channelOK {
			channelVal = make(actionChannel)
			go channelVal.run(s)

			s.observerID2ActionChannel[ev.Destination()] = channelVal
		}

		eventChannel = channelVal
	}()

	// 再次检查 terminateFlag，防止竞态条件
	if s.terminateFlag {
		return
	}

	if ev.Source() == ev.Destination() {
		s.Run(func() {
			select {
			case eventChannel <- actionData:
				// 成功发送
			case <-time.After(100 * time.Millisecond):
				slog.Warn("timeout sending data to channel")
			}
		})
	} else {
		select {
		case eventChannel <- actionData:
			// 成功发送
		case <-time.After(100 * time.Millisecond):
			slog.Warn("timeout sending data to channel")
		}
	}

	ret = <-replay
	return
}

func (s *hubImpl) Terminate() {
	if s.terminateFlag {
		return
	}

	s.terminateFlag = true
	var waitGroup sync.WaitGroup
	replay := make(chan bool)
	actionData := &terminateData{result: replay, waitGroup: &waitGroup}

	// 先发送终止信号到所有 channel
	go func() {
		s.observerID2ChanelLock.Lock()
		defer s.observerID2ChanelLock.Unlock()

		for _, val := range s.observerID2ActionChannel {
			waitGroup.Add(1)
			select {
			case val <- actionData:
				// 成功发送
			default:
				// channel 可能已满，等待一下再尝试
				go func(ch actionChannel) {
					ch <- actionData
				}(val)
			}
		}
	}()

	waitGroup.Add(1)
	select {
	case s.hubActionChannel <- actionData:
		// 成功发送
	default:
		// hub action channel 可能已满
		go func() {
			s.hubActionChannel <- actionData
		}()
	}

	<-replay

	waitGroup.Wait()

	// 等待所有 goroutine 完成处理
	s.observerID2ChanelLock.Lock()
	defer s.observerID2ChanelLock.Unlock()
	for _, val := range s.observerID2ActionChannel {
		close(val)
	}
	s.event2Observer = ID2ObserverMap{}
	s.observerID2ActionChannel = ID2ActionChanelMap{}

	close(replay)
}

func (s *hubImpl) run() {
	s.hubActionChannel.run(s)

	close(s.hubActionChannel)
}

func (s *hubImpl) subscribeInternal(eventID string, observer Observer) {
	s.event2ObserverlLock.Lock()
	defer s.event2ObserverlLock.Unlock()

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
	s.event2ObserverlLock.Lock()
	defer s.event2ObserverlLock.Unlock()

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

func (s *hubImpl) postInternal(ev Event) {
	matchList := ObserverList{}

	func() {
		s.event2ObserverlLock.RLock()
		defer s.event2ObserverlLock.RUnlock()
		for key, value := range s.event2Observer {
			if MatchValue(key, ev.ID()) {
				for _, sv := range value {
					if MatchValue(ev.Destination(), sv.ID()) {
						matchList = append(matchList, sv)
					}
				}
			}
		}
	}()

	for _, sv := range matchList {
		notificationEvent(sv, ev, nil)
	}
}

func (s *hubImpl) sendInternal(ev Event, re Result) {
	matchList := ObserverList{}
	finalFlag := false

	func() {
		s.event2ObserverlLock.RLock()
		defer s.event2ObserverlLock.RUnlock()
		for key, value := range s.event2Observer {
			if MatchValue(key, ev.ID()) {
				for _, sv := range value {
					if MatchValue(ev.Destination(), sv.ID()) {
						matchList = append(matchList, sv)
						finalFlag = true
					}
				}
			}
		}
	}()

	for _, sv := range matchList {
		notificationEvent(sv, ev, re)
	}

	if !finalFlag && re != nil {
		re.Set(nil, cd.NewError(cd.Unexpected, fmt.Sprintf("missing observer, event:[id-%v, source-%s, destination-%s]", ev.ID(), ev.Source(), ev.Destination())))
	}
}

type simpleObserver struct {
	id                   string
	eventHub             Hub
	eventID2ObserverFunc ID2ObserverFuncMap
	eventIDLock          sync.RWMutex
}

func (s *simpleObserver) ID() string {
	return s.id
}

func (s *simpleObserver) Notify(ev Event, re Result) {
	var funcVal ObserverFunc
	func() {
		s.eventIDLock.RLock()
		defer s.eventIDLock.RUnlock()

		for k, v := range s.eventID2ObserverFunc {
			if ev.Match(k) {
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
					slog.Warn("notify event exception", "event_id", ev.ID(), "source", ev.Source(), "destination", ev.Destination(), "panic", err, "stack", stackInfo)

					if re != nil {
						re.Set(nil, cd.NewError(cd.Unexpected, fmt.Sprintf("%v", err)))
					}
				}
			}()

			funcVal(ev, re)
		}()
	}
}

func (s *simpleObserver) Subscribe(eventID string, observerFunc ObserverFunc) {
	okFlag := false
	func() {
		s.eventIDLock.Lock()
		defer s.eventIDLock.Unlock()

		_, ok := s.eventID2ObserverFunc[eventID]
		if ok {
			slog.Warn("duplicate eventID", "value", eventID)
			return
		}

		s.eventID2ObserverFunc[eventID] = observerFunc
		okFlag = true
	}()

	if okFlag {
		s.eventHub.Subscribe(eventID, s)
	}
}

func (s *simpleObserver) Unsubscribe(eventID string) {
	okFlag := false
	func() {
		s.eventIDLock.Lock()
		defer s.eventIDLock.Unlock()

		_, ok := s.eventID2ObserverFunc[eventID]
		if !ok {
			slog.Warn("not exist eventID", "value", eventID)
			return
		}

		delete(s.eventID2ObserverFunc, eventID)
		okFlag = true
	}()

	if okFlag {
		s.eventHub.Unsubscribe(eventID, s)
	}
}
