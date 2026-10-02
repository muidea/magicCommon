package event

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/execute"
	"github.com/muidea/magicCommon/foundation/util"
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
	LaneKey() string
	Header() Values
	Context() context.Context
	BindContext(ctx context.Context)
	BindLaneKey(laneKey string)
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
	// ID returns the observer's unique subscription identity. The hub uses it
	// only for subscription de-duplication and independent unsubscription.
	ID() string
	Notify(event Event, result Result)
}

// destinationMatcher is an optional internal routing contract. It keeps an
// observer's destination pattern independent from its subscription identity.
type destinationMatcher interface {
	MatchID() string
}

type SimpleObserver interface {
	Observer
	Subscribe(eventID string, observerFunc ObserverFunc) *cd.Error
	Unsubscribe(eventID string) *cd.Error
}

type Hub interface {
	// Subscription errors mean no change was applied. Once admitted, the call
	// waits for actual completion; acknowledgement is never discarded.
	Subscribe(eventID string, observer Observer) *cd.Error
	Unsubscribe(eventID string, observer Observer) *cd.Error
	Post(event Event)
	Send(event Event) Result
	Terminate(ctx context.Context)
}

// DrainingHub exposes checked runtime completion. Owners must stop external
// producers before Drain; TerminateChecked additionally rejects new roots but
// lets already executing handlers finish their synchronous dependencies.
type DrainingHub interface {
	Hub
	Drain(context.Context) *cd.Error
	TerminateChecked(context.Context) *cd.Error
}

// HubOption Hub 配置项，用于控制内部缓冲和并发策略
type HubOption func(*hubOptions)

type hubOptions struct {
	perLaneChanSize   int
	hubActionChanSize int
	workerPoolSize    int
	laneIdleTimeout   time.Duration
}

const defaultMaxPerLaneChanSize = 64
const defaultMaxHubActionChanSize = 1024
const defaultMaxWorkerPoolSize = 256
const defaultLaneIdleTimeout = 30 * time.Second

func defaultHubOptions(capacitySize int) *hubOptions {
	if capacitySize <= 0 {
		capacitySize = 1
	}

	perLaneChanSize := capacitySize
	if perLaneChanSize > defaultMaxPerLaneChanSize {
		perLaneChanSize = defaultMaxPerLaneChanSize
	}

	return &hubOptions{
		// hubActionChannel 仅承担订阅控制面，不应跟随 500000 之类的大容量配置常驻扩张。
		// per-lane channel 若跟随 500000 之类的容量，会在每个新 lane 上常驻数 MB 内存。
		// workerPoolSize 控制同源事件异步投递的 goroutine 并发上限，需要与 hub 总队列容量解耦。
		perLaneChanSize:   perLaneChanSize,
		hubActionChanSize: minInt(capacitySize, defaultMaxHubActionChanSize),
		workerPoolSize:    minInt(capacitySize, defaultMaxWorkerPoolSize),
		laneIdleTimeout:   defaultLaneIdleTimeout,
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}

	return right
}

// WithPerDestinationChanSize 兼容旧命名，配置每个 lane 的内部 actionChannel 缓冲大小。
func WithPerDestinationChanSize(size int) HubOption {
	return WithPerLaneChanSize(size)
}

// WithPerLaneChanSize 配置每个顺序 lane 的内部 actionChannel 缓冲大小。
func WithPerLaneChanSize(size int) HubOption {
	return func(o *hubOptions) {
		if size > 0 {
			o.perLaneChanSize = size
		}
	}
}

// WithHubActionChanSize 配置 Hub 级别的 actionChannel 缓冲大小
func WithHubActionChanSize(size int) HubOption {
	return func(o *hubOptions) {
		if size > 0 {
			o.hubActionChanSize = size
		}
	}
}

// WithWorkerPoolSize 配置内部 Execute 的 worker 池大小
func WithWorkerPoolSize(size int) HubOption {
	return func(o *hubOptions) {
		if size > 0 {
			o.workerPoolSize = size
		}
	}
}

// WithLaneIdleTimeout 配置每个 lane 空闲多久后自动回收。
func WithLaneIdleTimeout(timeout time.Duration) HubOption {
	return func(o *hubOptions) {
		if timeout > 0 {
			o.laneIdleTimeout = timeout
		}
	}
}

type ObserverList []Observer
type ID2ObserverMap map[string]ObserverList
type ObserverFunc func(Event, Result)
type ID2ObserverFuncMap map[string]ObserverFunc
type actionChannel chan action
type LaneKey2ActionChannelMap map[string]*laneActionChannel

type laneEnqueueResult int

const (
	laneEnqueueOK laneEnqueueResult = iota
	laneEnqueueClosed
	laneEnqueueTimeout
	laneEnqueueCanceled
)

type laneActionChannel struct {
	key        string
	ch         actionChannel
	mu         sync.RWMutex
	closed     bool
	lastActive atomic.Int64
}

type laneExecutionContextKey struct{}

// A synchronous dispatch may revisit an active ancestor lane (A -> B -> A).
// Frames expire when their handler returns; retained contexts cannot bypass
// serialization later, nor can a context from another Hub authorize reentry.
type laneExecutionFrame struct {
	hub    *hubImpl
	key    string
	parent *laneExecutionFrame
	active atomic.Bool
}

type laneContextEvent struct {
	Event
	ctx context.Context
}

func (s *laneContextEvent) Context() context.Context {
	if s.ctx == nil {
		return context.Background()
	}

	return s.ctx
}

func (s *laneContextEvent) BindContext(ctx context.Context) {
	s.ctx = ctx
}

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
	return NewHubWithOptions(capacitySize)
}

// NewHubWithOptions 创建带可选配置的 Hub，实现更灵活的缓冲和并发控制
func NewHubWithOptions(capacitySize int, opts ...HubOption) Hub {
	hubOpts := defaultHubOptions(capacitySize)
	for _, opt := range opts {
		if opt != nil {
			opt(hubOpts)
		}
	}

	hub := &hubImpl{
		Execute:               execute.NewExecute(hubOpts.workerPoolSize),
		event2Observer:        ID2ObserverMap{},
		hubActionChannel:      make(chan action, hubOpts.hubActionChanSize),
		laneKey2ActionChannel: LaneKey2ActionChannelMap{},
		perLaneChanSize:       hubOpts.perLaneChanSize,
		laneIdleTimeout:       hubOpts.laneIdleTimeout,
		eventMatchCache:       map[string]ObserverList{},
		idle:                  make(chan struct{}),
	}
	close(hub.idle)
	hub.workers.Add(1)
	go hub.run()
	return hub
}

func NewSimpleObserver(id string, hub Hub) SimpleObserver {
	return &simpleObserver{id: id, matchID: id, eventHub: hub, eventID2ObserverFunc: ID2ObserverFuncMap{}}
}

// NewSimpleObserverWithMatchID 允许将观察者的唯一订阅标识与 destination 匹配模式分离。
func NewSimpleObserverWithMatchID(id, matchID string, hub Hub) SimpleObserver {
	if matchID == "" {
		matchID = id
	}

	return &simpleObserver{id: id, matchID: matchID, eventHub: hub, eventID2ObserverFunc: ID2ObserverFuncMap{}}
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
)

type action interface {
	Code() int
}

type subscribeData struct {
	eventID  string
	observer Observer
	result   chan *cd.Error
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
	state  atomic.Int32 // 0: queued, 1: executing, 2: canceled before execution
}

func (s *sendData) Code() int {
	return send
}

func (s actionChannel) run(hubPtr *hubImpl) {
	for {
		actionData, actionOK := <-s
		if !actionOK {
			return
		}

		if hubPtr.handleAction(actionData) {
			return
		}
	}
}

func (s *laneActionChannel) run(hubPtr *hubImpl) {
	defer hubPtr.workers.Done()
	if hubPtr.laneIdleTimeout <= 0 {
		for {
			actionData, actionOK := <-s.ch
			if !actionOK {
				return
			}
			if hubPtr.handleAction(actionData) {
				return
			}
		}
	}

	timer := time.NewTimer(hubPtr.laneIdleTimeout)
	defer timer.Stop()
	for {
		select {
		case actionData, actionOK := <-s.ch:
			if !actionOK {
				return
			}
			if hubPtr.handleAction(actionData) {
				return
			}
			s.touch()
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(hubPtr.laneIdleTimeout)
		case <-timer.C:
			if s.retireIfIdle(hubPtr) {
				return
			}
			timer.Reset(hubPtr.laneIdleTimeout)
		}
	}
}

type hubImpl struct {
	execute.Execute
	event2ObserverlLock sync.RWMutex
	event2Observer      ID2ObserverMap
	eventMatchCacheLock sync.RWMutex

	hubActionChannel      actionChannel
	laneKey2ChannelLock   sync.RWMutex
	laneKey2ActionChannel LaneKey2ActionChannelMap

	perLaneChanSize int
	laneIdleTimeout time.Duration

	// eventMatchCache 以 eventID 为 key 缓存匹配到的 ObserverList
	// 仅作为加速读路径使用，订阅关系变更时整体失效
	eventMatchCache map[string]ObserverList

	terminateFlag     atomic.Bool
	operationsMu      sync.Mutex
	operations        int
	idle              chan struct{}
	terminationActive bool
	channelsClosed    bool
	workers           sync.WaitGroup
	waitsMu           sync.Mutex
	waits             map[string]map[string]int
}

func newLaneActionChannel(key string, size int) *laneActionChannel {
	if size <= 0 {
		size = 1
	}

	ret := &laneActionChannel{
		key: key,
		ch:  make(actionChannel, size),
	}
	ret.touch()
	return ret
}

func (s *laneActionChannel) touch() {
	s.lastActive.Store(time.Now().UnixNano())
}

func (s *laneActionChannel) enqueue(actionData action, timeout time.Duration) laneEnqueueResult {
	return s.enqueueContext(context.Background(), actionData, timeout)
}

func (s *laneActionChannel) enqueueContext(ctx context.Context, actionData action, timeout time.Duration) laneEnqueueResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return laneEnqueueClosed
	}

	select {
	case <-ctx.Done():
		return laneEnqueueCanceled
	case s.ch <- actionData:
		s.touch()
		return laneEnqueueOK
	case <-time.After(timeout):
		return laneEnqueueTimeout
	}
}

func (s *laneActionChannel) retireIfIdle(hubPtr *hubImpl) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return true
	}
	if len(s.ch) > 0 {
		return false
	}
	if time.Since(time.Unix(0, s.lastActive.Load())) < hubPtr.laneIdleTimeout {
		return false
	}

	s.closed = true
	hubPtr.laneKey2ChannelLock.Lock()
	if currentPtr, currentOK := hubPtr.laneKey2ActionChannel[s.key]; currentOK && currentPtr == s {
		delete(hubPtr.laneKey2ActionChannel, s.key)
	}
	hubPtr.laneKey2ChannelLock.Unlock()

	close(s.ch)
	return true
}

func (s *laneActionChannel) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	close(s.ch)
}

func eventLaneKey(ev Event) string {
	if ev == nil {
		return ""
	}

	laneKey := ev.LaneKey()
	if laneKey != "" {
		return laneKey
	}

	return ev.Destination()
}

func (s *hubImpl) eventWithLaneContext(ev Event, inherit bool) (Event, func()) {
	frame := &laneExecutionFrame{hub: s, key: eventLaneKey(ev)}
	if inherit {
		frame.parent, _ = ev.Context().Value(laneExecutionContextKey{}).(*laneExecutionFrame)
	}
	frame.active.Store(true)
	return &laneContextEvent{
		Event: ev,
		ctx:   context.WithValue(ev.Context(), laneExecutionContextKey{}, frame),
	}, func() { frame.active.Store(false) }
}

func (s *hubImpl) isReentrantLaneExecution(ev Event, laneKey string) bool {
	if ev == nil || laneKey == "" {
		return false
	}

	frame, _ := ev.Context().Value(laneExecutionContextKey{}).(*laneExecutionFrame)
	for ; frame != nil; frame = frame.parent {
		if frame.hub != s || !frame.active.Load() {
			return false
		}
		if frame.key == laneKey {
			return true
		}
	}
	return false
}

func (s *hubImpl) getOrCreateLaneActionChannel(laneKey string) *laneActionChannel {
	s.laneKey2ChannelLock.Lock()
	defer s.laneKey2ChannelLock.Unlock()

	channelVal, channelOK := s.laneKey2ActionChannel[laneKey]
	if channelOK {
		return channelVal
	}

	channelVal = newLaneActionChannel(laneKey, s.perLaneChanSize)
	s.workers.Add(1)
	go channelVal.run(s)
	s.laneKey2ActionChannel[laneKey] = channelVal
	return channelVal
}

func (s *hubImpl) handleAction(actionData action) bool {
	// 同一个 lane 上的所有操作都顺序执行。
	// 包括 post 和 send 操作都需要顺序执行，以保证 lane 内事件顺序。
	switch actionData.Code() {
	case subscribe:
		data := actionData.(*subscribeData)
		data.result <- s.applySubscription(data.eventID, data.observer, true)
	case unsubscribe:
		data := actionData.(*unsubscribeData)
		data.result <- s.applySubscription(data.eventID, data.observer, false)
	case post:
		defer s.endOperation()
		data := actionData.(*postData)
		ev, finish := s.eventWithLaneContext(data.event, false)
		defer finish()
		s.postInternal(ev)
	case send:
		defer s.endOperation()
		data := actionData.(*sendData)
		if !data.state.CompareAndSwap(0, 1) {
			return false
		}
		eventWithContext, finish := s.eventWithLaneContext(data.event, true)
		defer finish()
		result := NewResult(data.event.ID(), data.event.Source(), data.event.Destination())
		if eventWithContext.Context().Err() != nil {
			result.Set(nil, cd.NewError(cd.Timeout, "event context expired before dispatch"))
		} else {
			s.sendInternal(eventWithContext, result)
		}
		data.result <- result
	default:
		slog.Error("unknown action code", "code", actionData.Code())
	}

	return false
}

func (s *hubImpl) Subscribe(eventID string, observer Observer) *cd.Error {
	return s.changeSubscription(eventID, observer, true)
}

func (s *hubImpl) Unsubscribe(eventID string, observer Observer) *cd.Error {
	return s.changeSubscription(eventID, observer, false)
}

func (s *hubImpl) changeSubscription(eventID string, observer Observer, add bool) *cd.Error {
	if eventID == "" || observer == nil {
		return cd.NewError(cd.IllegalParam, "event ID and observer are required")
	}
	if !s.beginOperation(context.TODO()) {
		return cd.NewError(cd.InvalidOperation, "event hub is stopping")
	}
	defer s.endOperation()

	result := make(chan *cd.Error, 1)
	data := &subscribeData{eventID: eventID, observer: observer, result: result}
	var request action = data
	if !add {
		request = (*unsubscribeData)(data)
	}
	// Admission must not first block on Execute's worker capacity. A rejected
	// request is never queued and therefore cannot mutate subscriptions later.
	timer := time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()
	if !s.admitSubscription(request, timer.C) {
		return cd.NewError(cd.ResourceExhausted, "event hub subscription admission timed out: control queue unavailable")
	}
	return <-result
}

func (s *hubImpl) admitSubscription(request action, deadline <-chan time.Time) bool {
	select {
	case s.hubActionChannel <- request:
		return true
	case <-deadline:
		// 到期与可入队可能同时就绪；最后只做即时探测，不延长等待或安排迟到发送。
		return s.tryAdmitSubscription(request)
	}
}

func (s *hubImpl) tryAdmitSubscription(request action) bool {
	select {
	case s.hubActionChannel <- request:
		return true
	default:
		return false
	}
}

func (s *hubImpl) applySubscription(eventID string, observer Observer, add bool) (err *cd.Error) {
	// An invalid custom observer must not kill the control worker or strand
	// callers. Registry mutation happens only after all ID comparisons succeed.
	defer func() {
		if value := recover(); value != nil {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("event subscription failed: %v", value))
		}
	}()
	if observer.ID() == "" {
		return cd.NewError(cd.IllegalParam, "observer ID is required")
	}
	if add {
		s.subscribeInternal(eventID, observer)
	} else {
		s.unsubscribeInternal(eventID, observer)
	}
	return nil
}

func (s *hubImpl) Post(ev Event) {
	if ev == nil || !s.beginOperation(ev.Context()) {
		return
	}
	queued := false
	defer func() {
		if !queued {
			s.endOperation()
		}
	}()

	laneKey := eventLaneKey(ev)
	actionData := &postData{event: ev}
	for {
		laneChannel := s.getOrCreateLaneActionChannel(laneKey)

		switch laneChannel.enqueue(actionData, 100*time.Millisecond) {
		case laneEnqueueOK:
			queued = true
			return
		case laneEnqueueClosed:
			continue
		case laneEnqueueTimeout:
			slog.Warn("timeout sending post data to channel")
			return
		}
	}
}

func (s *hubImpl) Send(ev Event) (ret Result) {
	if ev == nil {
		result := NewResult("", "", "")
		result.Set(nil, cd.NewError(cd.IllegalParam, "event is nil"))
		return result
	}
	if ev.Context().Err() != nil {
		result := NewResult(ev.ID(), ev.Source(), ev.Destination())
		result.Set(nil, cd.NewError(cd.Timeout, "event context is canceled"))
		return result
	}
	if !s.beginOperation(ev.Context()) {
		return nil
	}
	queued := false
	defer func() {
		if !queued {
			s.endOperation()
		}
	}()

	replay := make(chan Result, 1)

	laneKey := eventLaneKey(ev)
	if s.isReentrantLaneExecution(ev, laneKey) {
		eventWithContext, finish := s.eventWithLaneContext(ev, true)
		defer finish()
		result := NewResult(ev.ID(), ev.Source(), ev.Destination())
		s.sendInternal(eventWithContext, result)
		ret = result
		return
	}
	finishWait, err := s.beginLaneWait(ev, laneKey)
	if err != nil {
		result := NewResult(ev.ID(), ev.Source(), ev.Destination())
		result.Set(nil, err)
		return result
	}
	defer finishWait()

	actionData := &sendData{event: ev, result: replay}
	for {
		laneChannel := s.getOrCreateLaneActionChannel(laneKey)

		switch laneChannel.enqueueContext(ev.Context(), actionData, 100*time.Millisecond) {
		case laneEnqueueOK:
			queued = true
			select {
			case ret = <-replay:
				return
			case <-ev.Context().Done():
				if actionData.state.CompareAndSwap(0, 2) {
					result := NewResult(ev.ID(), ev.Source(), ev.Destination())
					result.Set(nil, cd.NewError(cd.Timeout, "event canceled before dispatch"))
					return result
				}
				// Execution won the race. Only its real completion can release
				// the caller's resources, even when the handler ignores cancel.
				return <-replay
			}
		case laneEnqueueClosed:
			continue
		case laneEnqueueTimeout, laneEnqueueCanceled:
			timeoutResult := NewResult(ev.ID(), ev.Source(), ev.Destination())
			timeoutResult.Set(nil, cd.NewError(cd.Timeout, "event admission timed out or was canceled"))
			return timeoutResult
		}
	}
}

func (s *hubImpl) Terminate(ctx context.Context) {
	if err := s.TerminateChecked(ctx); err != nil {
		slog.Warn("hub shutdown incomplete; handlers retained", "error", err)
	}
}

func (s *hubImpl) TerminateChecked(ctx context.Context) *cd.Error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.operationsMu.Lock()
	if s.terminationActive {
		s.operationsMu.Unlock()
		return cd.NewError(cd.InvalidOperation, "hub shutdown is in progress")
	}
	s.terminationActive = true
	s.terminateFlag.Store(true)
	s.operationsMu.Unlock()
	defer func() { s.operationsMu.Lock(); s.terminationActive = false; s.operationsMu.Unlock() }()

	if err := s.Drain(ctx); err != nil {
		return err
	}
	if !s.WaitContext(ctx) {
		return cd.NewError(cd.Timeout, "hub control tasks are still running")
	}

	if !s.channelsClosed {
		s.laneKey2ChannelLock.Lock()
		laneChannels := make([]*laneActionChannel, 0, len(s.laneKey2ActionChannel))
		for _, val := range s.laneKey2ActionChannel {
			laneChannels = append(laneChannels, val)
		}
		s.laneKey2ActionChannel = LaneKey2ActionChannelMap{}
		s.laneKey2ChannelLock.Unlock()
		for _, val := range laneChannels {
			val.close()
		}
		close(s.hubActionChannel)
		s.channelsClosed = true
	}
	if !waitGroupContext(&s.workers, ctx) {
		return cd.NewError(cd.Timeout, "hub workers are still stopping")
	}
	s.event2ObserverlLock.Lock()
	s.event2Observer = ID2ObserverMap{}
	s.event2ObserverlLock.Unlock()
	s.eventMatchCacheLock.Lock()
	s.eventMatchCache = map[string]ObserverList{}
	s.eventMatchCacheLock.Unlock()
	return nil
}

func waitGroupContext(waitGroup *sync.WaitGroup, ctx context.Context) bool {
	done := make(chan struct{})
	go func() {
		defer close(done)
		waitGroup.Wait()
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

func (s *hubImpl) run() {
	defer s.workers.Done()
	s.hubActionChannel.run(s)
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

	// 订阅关系变更，清空匹配缓存
	s.eventMatchCacheLock.Lock()
	s.eventMatchCache = map[string]ObserverList{}
	s.eventMatchCacheLock.Unlock()
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
	} else {
		delete(s.event2Observer, eventID)
	}

	// 订阅关系变更，清空匹配缓存
	s.eventMatchCacheLock.Lock()
	s.eventMatchCache = map[string]ObserverList{}
	s.eventMatchCacheLock.Unlock()
}

func (s *hubImpl) postInternal(ev Event) {
	matchList := s.matchingObservers(ev)

	for _, sv := range matchList {
		notificationEvent(sv, ev, nil)
	}
}

func (s *hubImpl) sendInternal(ev Event, re Result) {
	matchList := s.matchingObservers(ev)

	for _, sv := range matchList {
		notificationEvent(sv, ev, re)
	}

	if len(matchList) == 0 && re != nil {
		re.Set(nil, cd.NewError(cd.Unexpected, fmt.Sprintf("missing observer, event:[id-%v, source-%s, destination-%s]", ev.ID(), ev.Source(), ev.Destination())))
	}
}

func (s *hubImpl) matchingObservers(ev Event) ObserverList {
	// Keep registry lookup and cache publication in the same read transaction.
	// Otherwise an old lookup can repopulate the cache after invalidation.
	s.event2ObserverlLock.RLock()
	defer s.event2ObserverlLock.RUnlock()
	key := matchCacheKey(ev.ID(), ev.Destination())
	if cached, ok := s.getCachedObservers(key); ok {
		return cached
	}
	matched := s.findMatchingObservers(ev)
	s.setCachedObservers(key, matched)
	return matched
}

func matchCacheKey(eventID, destination string) string {
	return eventID + "\x00" + destination
}

func (s *hubImpl) getCachedObservers(cacheKey string) (ObserverList, bool) {
	s.eventMatchCacheLock.RLock()
	defer s.eventMatchCacheLock.RUnlock()

	cached, ok := s.eventMatchCache[cacheKey]
	if !ok {
		return nil, false
	}

	result := make(ObserverList, len(cached))
	copy(result, cached)
	return result, true
}

func (s *hubImpl) setCachedObservers(cacheKey string, observers ObserverList) {
	s.eventMatchCacheLock.Lock()
	defer s.eventMatchCacheLock.Unlock()

	if s.eventMatchCache == nil {
		s.eventMatchCache = map[string]ObserverList{}
	}
	result := make(ObserverList, len(observers))
	copy(result, observers)
	s.eventMatchCache[cacheKey] = result
}

func (s *hubImpl) findMatchingObservers(ev Event) ObserverList {
	matchList := make(ObserverList, 0, 4)

	for key, value := range s.event2Observer {
		if MatchValue(key, ev.ID()) {
			for _, sv := range value {
				if matchDestination(ev.Destination(), observerMatchID(sv)) {
					matchList = append(matchList, sv)
				}
			}
		}
	}

	return matchList
}

func matchDestination(destination, matchID string) bool {
	return MatchValue(destination, matchID) || MatchValue(matchID, destination)
}

func observerMatchID(observer Observer) string {
	if matcher, ok := observer.(destinationMatcher); ok {
		if matchID := matcher.MatchID(); matchID != "" {
			return matchID
		}
	}

	return observer.ID()
}

type simpleObserver struct {
	id                   string
	matchID              string
	eventHub             Hub
	eventID2ObserverFunc ID2ObserverFuncMap
	eventIDLock          sync.RWMutex
}

var _ destinationMatcher = (*simpleObserver)(nil)

func (s *simpleObserver) ID() string {
	return s.id
}

func (s *simpleObserver) MatchID() string {
	return s.matchID
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

func (s *simpleObserver) Subscribe(eventID string, observerFunc ObserverFunc) *cd.Error {
	if eventID == "" || observerFunc == nil || s.eventHub == nil {
		return cd.NewError(cd.IllegalParam, "event ID, handler and hub are required")
	}
	// Serialize local changes with Hub completion. Notify cannot observe a
	// partially registered handler, and failed operations leave state retryable.
	s.eventIDLock.Lock()
	defer s.eventIDLock.Unlock()
	if _, exists := s.eventID2ObserverFunc[eventID]; exists {
		return cd.NewError(cd.Duplicated, "event handler is already subscribed")
	}
	if err := s.eventHub.Subscribe(eventID, s); err != nil {
		return err
	}
	s.eventID2ObserverFunc[eventID] = observerFunc
	return nil
}

func (s *simpleObserver) Unsubscribe(eventID string) *cd.Error {
	s.eventIDLock.Lock()
	defer s.eventIDLock.Unlock()
	if _, exists := s.eventID2ObserverFunc[eventID]; !exists {
		return nil
	}
	if err := s.eventHub.Unsubscribe(eventID, s); err != nil {
		return err
	}
	delete(s.eventID2ObserverFunc, eventID)
	return nil
}
