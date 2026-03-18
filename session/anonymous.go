package session

import (
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type anonymousSession struct {
	id       string
	context  map[string]any
	observer map[string]Observer
	mu       sync.RWMutex
}

func NewAnonymousSession(remoteAddress, userAgent string) Session {
	now := time.Now().UTC()
	return &anonymousSession{
		id: createUUID(),
		context: map[string]any{
			InnerStartTime:        now.UnixMilli(),
			InnerRemoteAccessAddr: remoteAddress,
			InnerUseAgent:         userAgent,
			innerExpireTime:       now.Add(GetSessionTimeOutValue()).UnixMilli(),
		},
		observer: map[string]Observer{},
	}
}

func (s *anonymousSession) ID() string {
	return s.id
}

func (s *anonymousSession) excludeKey(key string) bool {
	return strings.HasPrefix(key, "_")
}

func (s *anonymousSession) Signature() (string, error) {
	mc := jwt.MapClaims{}
	if s.id != "" {
		mc[innerSessionID] = s.id
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.context {
		if s.excludeKey(k) {
			continue
		}
		mc[k] = v
	}

	return SignatureJWT(mc)
}

func (s *anonymousSession) Reset() {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	startTime := s.context[InnerStartTime]
	remoteAccessAddr := s.context[InnerRemoteAccessAddr]
	useAgent := s.context[InnerUseAgent]
	s.context = map[string]any{
		InnerStartTime:        startTime,
		InnerRemoteAccessAddr: remoteAccessAddr,
		InnerUseAgent:         useAgent,
		innerExpireTime:       now.Add(GetSessionTimeOutValue()).UnixMilli(),
	}
	s.observer = map[string]Observer{}
}

func (s *anonymousSession) BindObserver(observer Observer) {
	if observer == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.observer[observer.ID()] = observer
}

func (s *anonymousSession) UnbindObserver(observer Observer) {
	if observer == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.observer, observer.ID())
}

func (s *anonymousSession) GetString(key string) (string, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return "", false
	}

	strVal, strOK := val.(string)
	return strVal, strOK
}

func (s *anonymousSession) GetInt(key string) (int64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, false
	}
	return convertToInt64(val, key)
}

func (s *anonymousSession) GetUint(key string) (uint64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, false
	}
	return convertToUint64(val, key)
}

func (s *anonymousSession) GetFloat(key string) (float64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, false
	}
	return convertToFloat64(val, key)
}

func (s *anonymousSession) GetBool(key string) (bool, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return false, false
	}
	return convertToBool(val, key)
}

func (s *anonymousSession) GetOption(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.context[key]
	return value, found
}

func (s *anonymousSession) SetOption(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.context[key] = value
}

func (s *anonymousSession) RemoveOption(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.context, key)
}

func (s *anonymousSession) SubmitOptions() {
	s.mu.RLock()
	observers := make([]Observer, 0, len(s.observer))
	for _, val := range s.observer {
		observers = append(observers, val)
	}
	s.mu.RUnlock()

	for _, val := range observers {
		go val.OnStatusChange(s, StatusUpdate)
	}
}
