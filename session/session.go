package session

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

type Status int

const (
	StatusRunning = iota
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
	Flush(res http.ResponseWriter, req *http.Request)
	BindObserver(observer Observer)
	UnbindObserver(observer Observer)

	Namespace() string
	GetSessionInfo() *SessionInfo
	SetSessionInfo(info *SessionInfo)
	RefreshTime() int64
	ExpireTime() int64
	GetOption(key string) (interface{}, bool)
	SetOption(key string, value interface{})
	RemoveOption(key string)
	OptionKey() []string
}

const (
	defaultSessionTimeOutValue = 10 * time.Minute  // 10 minute
	tempSessionTimeOutValue    = 1 * time.Minute   // 1 minute
	ForeverSessionTimeOutValue = time.Duration(-1) // forever time out value
)

type sessionImpl struct {
	id       string // session id
	context  map[string]interface{}
	observer map[string]Observer
	registry *sessionRegistryImpl
}

func (s *sessionImpl) ID() string {
	return s.id
}

func (s *sessionImpl) Flush(res http.ResponseWriter, req *http.Request) {
	info := s.GetSessionInfo()
	if info == nil {
		return
	}

	// 存入cookie,使用cookie存储
	dataValue, dataErr := json.Marshal(info)
	if dataErr == nil {
		sessionCookie := http.Cookie{
			Name:   sessionCookieID,
			Value:  base64.StdEncoding.EncodeToString(dataValue),
			Path:   "/",
			MaxAge: 600,
		}

		http.SetCookie(res, &sessionCookie)

		req.AddCookie(&sessionCookie)
	}
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

func (s *sessionImpl) Namespace() string {
	val, ok := s.GetOption(AuthNamespace)
	if !ok {
		return ""
	}

	return val.(string)
}

func (s *sessionImpl) GetSessionInfo() (ret *SessionInfo) {
	val, ok := s.GetOption(authSessionInfo)
	if !ok {
		ret = nil
		return
	}

	ret, ok = val.(*SessionInfo)
	if !ok {
		s.RemoveOption(authSessionInfo)
		ret = nil
		return
	}

	return
}

func (s *sessionImpl) SetSessionInfo(info *SessionInfo) {
	if info == nil {
		return
	}

	s.SetOption(authSessionInfo, info)
}

func (s *sessionImpl) RefreshTime() int64 {
	timeVal, timeOK := s.GetOption(refreshTime)
	if timeOK {
		return timeVal.(time.Time).UTC().Unix()
	}

	return time.Now().UTC().Unix()
}

func (s *sessionImpl) ExpireTime() int64 {
	timeVal, timeOK := s.GetOption(AuthExpiryValue)
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

func (s *sessionImpl) SetOption(key string, value interface{}) {
	func() {
		s.registry.sessionLock.Lock()
		defer s.registry.sessionLock.Unlock()

		s.context[key] = value

		if key == AuthEntity {
			s.context[AuthExpiryValue] = defaultSessionTimeOutValue
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
		if key == AuthEntity {
			s.context[AuthExpiryValue] = tempSessionTimeOutValue
		}
	}()

	s.save()
}

func (s *sessionImpl) OptionKey() []string {
	keys := []string{}

	func() {
		s.registry.sessionLock.RLock()
		defer s.registry.sessionLock.RUnlock()

		for key := range s.context {
			keys = append(keys, key)
		}
	}()

	return keys
}

func (s *sessionImpl) refresh() {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	// 这里是在sessionRegistry里更新的，所以这里不用save
	s.context[refreshTime] = time.Now()
}

func (s *sessionImpl) timeOut() bool {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	expiryDate, found := s.context[AuthExpiryValue]
	if found && expiryDate.(time.Duration) == ForeverSessionTimeOutValue {
		return false
	}
	if !found {
		expiryDate = defaultSessionTimeOutValue
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
