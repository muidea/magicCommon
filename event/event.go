package event

import (
	"errors"
	"fmt"
)

const innerDataKey = "_innerDataKey_"

type baseEvent struct {
	eventID          string
	eventSource      string
	eventDestination string
	eventHeader      Values
	eventData        map[string]interface{}
	eventResult      interface{}
}

type baseResult struct {
	resultData interface{}
	resultErr  error
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

func NewResult(id, source, destination string) Result {
	msg := fmt.Sprintf("illegal event, id:%s, source:%s, destination:%s", id, source, destination)
	return &baseResult{resultErr: errors.New(msg)}
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
	return matchID(pattern, s.eventID)
}

func (s *baseResult) Set(data interface{}, err error) {
	s.resultData = data
	s.resultErr = err
}

func (s *baseResult) Error() error {
	return s.resultErr
}

func (s *baseResult) Get() (ret interface{}, err error) {
	ret = s.resultData
	err = s.resultErr
	return
}
