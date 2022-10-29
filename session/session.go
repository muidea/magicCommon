package session

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Status int
type Token string

const (
	StatusUpdate = iota
	StatusTerminate
)

const (
	// sessionID 会话ID
	sessionID = "sessionID"
	// RemoteAddress 远端地址
	RemoteAddress = "$$sessionRemoteAddress"
	// ExpiryValue 会话有效期
	ExpiryValue = "$$sessionExpiryValue"
	// refreshTime 会话刷新时间
	refreshTime = "$$sessionRefreshTime"
)

const (
	jwtToken      = "Bearer"
	endpointToken = "Sig"

	DefaultSessionTimeOutValue = 10 * time.Minute  // 10 minute
	tempSessionTimeOutValue    = 1 * time.Minute   // 1 minute
	ForeverSessionTimeOutValue = time.Duration(-1) // forever time out value
)

// Observer session Observer
type Observer interface {
	ID() string
	OnStatusChange(session Session, status Status)
}

// Session 会话
type Session interface {
	ID() string
	Signature() (Token, error)
	Reset()
	BindObserver(observer Observer)
	UnbindObserver(observer Observer)

	RefreshTime() int64
	ExpireTime() int64
	GetString(key string) (string, bool)
	GetInt(key string) (int64, bool)
	GetUint(key string) (uint64, bool)
	GetFloat(key string) (float64, bool)
	GetBool(key string) (bool, bool)
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

func (s *sessionImpl) innerKey(key string) bool {
	switch key {
	case RemoteAddress, ExpiryValue, refreshTime:
		return true
	}

	return false
}

func (s *sessionImpl) Signature() (Token, error) {
	mc := jwt.MapClaims{}
	if s.id != "" {
		mc[sessionID] = s.id
	}

	for k, v := range s.context {
		if s.innerKey(k) {
			continue
		}

		mc[k] = v
	}

	return SignatureJWT(mc)
}

func (s *sessionImpl) Reset() {
	func() {
		s.registry.sessionLock.RLock()
		defer s.registry.sessionLock.RUnlock()

		s.context = map[string]interface{}{refreshTime: time.Now(), ExpiryValue: tempSessionTimeOutValue}
		s.observer = map[string]Observer{}
	}()

	s.save()
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

func (s *sessionImpl) RefreshTime() int64 {
	timeVal, timeOK := s.GetOption(refreshTime)
	if timeOK {
		return timeVal.(time.Time).UTC().Unix()
	}

	return time.Now().UTC().Unix()
}

func (s *sessionImpl) ExpireTime() int64 {
	timeVal, timeOK := s.GetOption(ExpiryValue)
	if !timeOK {
		return 0
	}

	if timeVal.(time.Duration) == ForeverSessionTimeOutValue {
		return -1
	}

	refreshTime, refreshOK := s.GetOption(refreshTime)
	if refreshOK {
		return refreshTime.(time.Time).Add(timeVal.(time.Duration)).UTC().Unix()
	}

	return 0
}

func (s *sessionImpl) GetString(key string) (string, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return "", ok
	}

	strVal, strOK := val.(string)
	return strVal, strOK
}

func (s *sessionImpl) GetInt(key string) (int64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, ok
	}

	switch val.(type) {
	case int8:
		return int64(val.(int8)), true
	case int16:
		return int64(val.(int16)), true
	case int32:
		return int64(val.(int32)), true
	case int64:
		return val.(int64), true
	case int:
		return int64(val.(int)), true
	case float64:
		return int64(val.(float64)), true
	case float32:
		return int64(val.(float32)), true
	case string:
		val, err := strconv.ParseInt(val.(string), 10, 64)
		return val, err == nil
	}

	return 0, false
}

func (s *sessionImpl) GetUint(key string) (uint64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, ok
	}

	switch val.(type) {
	case uint8:
		return uint64(val.(uint8)), true
	case uint16:
		return uint64(val.(uint16)), true
	case uint32:
		return uint64(val.(uint32)), true
	case uint64:
		return val.(uint64), true
	case uint:
		return uint64(val.(uint)), true
	case float64:
		return uint64(val.(float64)), true
	case float32:
		return uint64(val.(float32)), true
	case string:
		val, err := strconv.ParseUint(val.(string), 10, 64)
		return val, err == nil
	}

	return 0, false
}

func (s *sessionImpl) GetFloat(key string) (float64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0.00, ok
	}

	switch val.(type) {
	case float64:
		return val.(float64), true
	case float32:
		return float64(val.(float32)), true
	case string:
		val, err := strconv.ParseFloat(val.(string), 64)
		return val, err == nil
	}

	return 0.00, false
}

func (s *sessionImpl) GetBool(key string) (bool, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return false, ok
	}

	switch val.(type) {
	case bool:
		return val.(bool), true
	case string:
		val, err := strconv.ParseBool(val.(string))
		return val, err == nil
	}

	return false, false
}

func (s *sessionImpl) GetOption(key string) (interface{}, bool) {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	value, found := s.context[key]

	return value, found
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
	s.registry.sessionLock.Lock()
	defer s.registry.sessionLock.Unlock()
	s.context[refreshTime] = time.Now()
}

func (s *sessionImpl) timeout() bool {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	expiryDate, _ := s.context[ExpiryValue]
	if expiryDate.(time.Duration) == ForeverSessionTimeOutValue {
		return false
	}

	preTime, _ := s.context[refreshTime]

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

// Context context info
type Context interface {
	Decode(req *http.Request)
	Encode(vals url.Values) url.Values
}
