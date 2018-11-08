package session

import (
	"time"

	common_const "muidea.com/magicCommon/common"
)

const (
	// AccountKey account key
	AccountKey = "$account_key$"
	// TokenKey session token key
	TokenKey = "$session_token$"
)

// Session 会话
type Session interface {
	ID() string

	GetOption(key string) (interface{}, bool)
	SetOption(key string, value interface{})
	RemoveOption(key string)

	OptionKey() []string

	Flush()
}

const (
	maxTimeOut = 10
)

type sessionImpl struct {
	id       string // session id
	context  map[string]interface{}
	registry Registry
}

func (s *sessionImpl) ID() string {
	return s.id
}

func (s *sessionImpl) SetOption(key string, value interface{}) {
	s.context[key] = value

	s.save()
}

func (s *sessionImpl) GetOption(key string) (interface{}, bool) {
	value, found := s.context[key]

	return value, found
}

func (s *sessionImpl) RemoveOption(key string) {
	delete(s.context, key)

	s.save()
}

func (s *sessionImpl) OptionKey() []string {
	keys := []string{}
	for key := range s.context {
		keys = append(keys, key)
	}

	return keys
}

func (s *sessionImpl) Flush() {
	s.registry.FlushSession(s)
}

func (s *sessionImpl) refresh() {
	s.context["$$refreshTime"] = time.Now()
}

func (s *sessionImpl) timeOut() bool {
	expiryDate, found := s.context[common_const.ExpiryDate]
	if found && expiryDate.(int) == -1 {
		return false
	}

	preTime, found := s.context["$$refreshTime"]
	if !found {
		return true
	}

	nowTime := time.Now()
	elapse := nowTime.Sub(preTime.(time.Time)).Minutes()

	return elapse > maxTimeOut
}

func (s *sessionImpl) save() {
	s.registry.UpdateSession(s)
}