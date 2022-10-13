package session

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Status int

const (
	StatusUpdate = iota
	StatusTerminate
)

// Observer session Observer
type Observer interface {
	ID() string
	OnStatusChange(session Session, status Status)
}

// Session 会话
type Session interface {
	ID() string
	SignedString() (string, error)
	BindObserver(observer Observer)
	UnbindObserver(observer Observer)

	GetOption(key string) (interface{}, bool)
	SetOption(key string, value interface{})
	RemoveOption(key string)
}

type sessionImpl struct {
	id       string // session id
	context  map[string]interface{}
	observer map[string]Observer
	registry *sessionRegistryImpl
}

func (s *sessionImpl) ID() string {
	return s.id
}

func (s *sessionImpl) SignedString() (string, error) {
	mc := jwt.MapClaims{}
	if s.id != "" {
		mc[sessionID] = s.id
	}

	for k, v := range s.context {
		if k == AuthRemoteAddress || k == AuthExpiryValue {
			continue
		}

		mc[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mc)
	return token.SignedString([]byte(hmacSampleSecret))
}

func (s *sessionImpl) BindObserver(observer Observer) {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	_, ok := s.observer[observer.ID()]
	if ok {
		return
	}
	s.observer[observer.ID()] = observer
}

func (s *sessionImpl) UnbindObserver(observer Observer) {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	_, ok := s.observer[observer.ID()]
	if !ok {
		return
	}

	delete(s.observer, observer.ID())
}

func (s *sessionImpl) SetOption(key string, value interface{}) {
	func() {
		s.registry.sessionLock.Lock()
		defer s.registry.sessionLock.Unlock()

		s.context[key] = value

		for _, val := range s.observer {
			go val.OnStatusChange(s, StatusUpdate)
		}
	}()

	s.save()
}

func (s *sessionImpl) GetOption(key string) (interface{}, bool) {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	value, found := s.context[key]

	return value, found
}

func (s *sessionImpl) RemoveOption(key string) {
	func() {
		s.registry.sessionLock.Lock()
		defer s.registry.sessionLock.Unlock()

		delete(s.context, key)

		for _, val := range s.observer {
			go val.OnStatusChange(s, StatusUpdate)
		}
	}()

	s.save()
}

func (s *sessionImpl) refresh() {
	s.SetOption(refreshTime, time.Now())
}

func (s *sessionImpl) timeout() bool {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	expiryDate, found := s.context[AuthExpiryValue]
	if found {
		return true
	}

	preTime, found := s.context[refreshTime]
	if !found {
		return true
	}

	nowTime := time.Now()
	elapse := nowTime.Sub(preTime.(time.Time))

	return elapse > expiryDate.(time.Duration)
}

func (s *sessionImpl) terminate() {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	for _, val := range s.observer {
		go val.OnStatusChange(s, StatusTerminate)
	}
}

func (s *sessionImpl) save() {
	s.registry.updateSession(s)
}
