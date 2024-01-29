package event

import (
	"context"
	"fmt"
	cd "github.com/muidea/magicCommon/def"
)

const innerDataKey = "_innerDataKey_"
const innerValKey = "_innerValKey_"

const (
	Action = "_action_"
	Add    = "add"
	Del    = "del"
	Mod    = "mod"
	Notify = "notify"
)

type baseEvent struct {
	eventID          string
	eventSource      string
	eventDestination string
	eventHeader      Values
	eventContext     context.Context
	eventData        map[string]interface{}
	eventResult      interface{}
}

type baseResult struct {
	resultData map[string]interface{}
	resultErr  *cd.Result
}

func NewValues() Values {
	return map[string]interface{}{}
}

func NewEvent(id, source, destination string, header Values, data interface{}) Event {
	return &baseEvent{
		eventID:          id,
		eventSource:      source,
		eventDestination: destination,
		eventHeader:      header,
		eventData:        map[string]interface{}{innerDataKey: data},
	}
}

func NewEventWitchContext(id, source, destination string, header Values, context context.Context, data interface{}) Event {
	return &baseEvent{
		eventID:          id,
		eventSource:      source,
		eventDestination: destination,
		eventHeader:      header,
		eventContext:     context,
		eventData:        map[string]interface{}{innerDataKey: data},
	}
}

func NewResult(id, source, destination string) Result {
	msg := fmt.Sprintf("illegal event, no result returned, id:%s, source:%s, destination:%s", id, source, destination)
	return &baseResult{resultErr: cd.NewError(cd.Failed, msg), resultData: map[string]interface{}{}}
}

func (s *baseEvent) ID() string {
	return s.eventID
}

func (s *baseEvent) Source() string {
	return s.eventSource
}

func (s *baseEvent) Destination() string {
	return s.eventDestination
}

func (s *baseEvent) Header() Values {
	return s.eventHeader
}

func (s *baseEvent) Context() context.Context {
	return s.eventContext
}

func (s *baseEvent) BindContext(ctx context.Context) {
	s.eventContext = ctx
}

func (s *baseEvent) Data() interface{} {
	val, ok := s.eventData[innerDataKey]
	if ok {
		return val
	}
	return nil
}

func (s *baseEvent) SetData(key string, val interface{}) {
	s.eventData[key] = val
}

func (s *baseEvent) GetData(key string) interface{} {
	val, ok := s.eventData[key]
	if ok {
		return val
	}

	return nil
}

func (s *baseEvent) Result(result interface{}) {
	s.eventResult = result
}

func (s *baseEvent) Match(pattern string) bool {
	return MatchValue(pattern, s.eventID)
}

func (s *baseResult) Set(data interface{}, err *cd.Result) {
	s.resultData[innerValKey] = data
	s.resultErr = err
}

func (s *baseResult) Error() *cd.Result {
	return s.resultErr
}

func (s *baseResult) Get() (ret interface{}, err *cd.Result) {
	ret = s.resultData[innerValKey]
	err = s.resultErr
	return
}

func (s *baseResult) SetVal(key string, val interface{}) {
	s.resultData[key] = val
}

func (s *baseResult) GetVal(key string) interface{} {
	val, ok := s.resultData[key]
	if ok {
		return val
	}

	return nil
}
