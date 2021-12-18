package session

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

// Session 会话
type Session interface {
	ID() string
	Flush(res http.ResponseWriter, req *http.Request)

	GetSessionInfo() *SessionInfo
	SetSessionInfo(info *SessionInfo)
	GetOption(key string) (interface{}, bool)
	SetOption(key string, value interface{})
	RemoveOption(key string)
	OptionKey() []string
}

const (
	defaultSessionTimeOutValue = 10 * time.Minute // 10 minute
)

type sessionImpl struct {
	id       string // session id
	context  map[string]interface{}
	registry *sessionRegistryImpl

	callBack CallBack
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

func (s *sessionImpl) SetOption(key string, value interface{}) {
	func() {
		s.registry.sessionLock.Lock()
		defer s.registry.sessionLock.Unlock()

		s.context[key] = value
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
	s.context["$$refreshTime"] = time.Now()
}

func (s *sessionImpl) timeOut() bool {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	expiryDate, found := s.context[ExpiryValue]
	if found && expiryDate.(time.Duration) == -1 {
		return false
	}
	if !found {
		expiryDate = defaultSessionTimeOutValue
	}

	preTime, found := s.context["$$refreshTime"]
	if !found {
		return true
	}

	nowTime := time.Now()
	elapse := nowTime.Sub(preTime.(time.Time))

	return elapse > expiryDate.(time.Duration)
}

func (s *sessionImpl) save() {
	s.registry.updateSession(s)
}
