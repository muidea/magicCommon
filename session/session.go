package session

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	commonConst "github.com/muidea/magicCommon/common"
)

// Session 会话
type Session interface {
	ID() string
	Flush(res http.ResponseWriter, req *http.Request)

	GetSessionInfo() *commonConst.SessionInfo
	SetSessionInfo(info *commonConst.SessionInfo)
	GetOption(key string) (interface{}, bool)
	SetOption(key string, value interface{})
	RemoveOption(key string)
	OptionKey() []string
}

const (
	maxTimeOut = 10
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

func (s *sessionImpl) GetSessionInfo() (ret *commonConst.SessionInfo) {
	val, ok := s.GetOption(commonConst.AuthSessionInfo)
	if !ok {
		ret = nil
		return
	}

	ret, ok = val.(*commonConst.SessionInfo)
	if !ok {
		s.RemoveOption(commonConst.AuthSessionInfo)
		ret = nil
		return
	}

	return
}

func (s *sessionImpl) SetSessionInfo(info *commonConst.SessionInfo) {
	if info == nil {
		return
	}

	s.SetOption(commonConst.AuthSessionInfo, info)
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

	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	for key := range s.context {
		keys = append(keys, key)
	}

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

	expiryDate, found := s.context[commonConst.ExpiryDate]
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
	s.registry.updateSession(s)
}
