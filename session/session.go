package session

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/muidea/magicCommon/foundation/log"
)

type Status int
type Token string

const (
	StatusUpdate = iota
	StatusTerminate
)

const (
	// innerSessionID 会话ID
	innerSessionID = "_sessionID"
	// InnerRemoteAccessAddr 会话来源地址
	InnerRemoteAccessAddr = "_remoteAccessAddr"
	// InnerUseAgent 会话来源UA
	InnerUseAgent = "_userAgent"
	// account/endpoint 认证方式
	InnerAuthType = "_authType"
	// innserSessionStartTime 会话开始时间
	InnerStartTime = "innerSessionStartTime"
	// innerExpireTime 会话有效期，该有效性必须要定期刷新，否则就会在超过该有效期时失效
	innerExpireTime = "innerExpireTime"
	// AuthExpireTime 会话强制有效期，该有效期通过session Option进行强制设置，与innerExpireTime在使用时，取两者之间最大值为实际会话有效期
	AuthExpireTime = "authExpireTime"
	// Authorization info, from request header
	Authorization = "Authorization"
)

const (
	AuthJWTSession      = "jwt"
	AuthEndpointSession = "endpoint"
)

const (
	jwtToken = "Bearer"
	sigToken = "Sig"

	DefaultSessionTimeOutValue = 10 * time.Minute // 10 minute
)

const (
	sessionActive    = 0
	sessionUpdate    = 1
	sessionTerminate = 2
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

	GetString(key string) (string, bool)
	GetInt(key string) (int64, bool)
	GetUint(key string) (uint64, bool)
	GetFloat(key string) (float64, bool)
	GetBool(key string) (bool, bool)
	GetOption(key string) (any, bool)
	SetOption(key string, value any)
	RemoveOption(key string)
	SubmitOptions()
}

type sessionImpl struct {
	id       string // session id
	context  map[string]any
	observer map[string]Observer
	registry *sessionRegistryImpl
	status   int
}

func (s *sessionImpl) ID() string {
	return s.id
}

func (s *sessionImpl) excludeKey(key string) bool {
	switch key {
	case Authorization:
		return true
	}

	// 以下划线开头的key也要进行排除
	return strings.HasPrefix(key, "_")
}

func (s *sessionImpl) Signature() (Token, error) {
	mc := jwt.MapClaims{}
	if s.id != "" {
		mc[innerSessionID] = s.id
	}

	func() {
		s.registry.sessionLock.RLock()
		defer s.registry.sessionLock.RUnlock()
		for k, v := range s.context {
			if s.excludeKey(k) {
				continue
			}

			mc[k] = v
		}
	}()

	return SignatureJWT(mc)
}

func (s *sessionImpl) Reset() {
	expireValue := time.Now().Add(GetSessionTimeOutValue()).UTC().UnixMilli()
	func() {
		startTime := s.context[InnerStartTime]
		remoteAccessAddr := s.context[InnerRemoteAccessAddr]
		useAgent := s.context[InnerUseAgent]
		s.registry.sessionLock.RLock()
		defer s.registry.sessionLock.RUnlock()

		s.context = map[string]any{
			InnerStartTime:        startTime,
			InnerRemoteAccessAddr: remoteAccessAddr,
			InnerUseAgent:         useAgent,
			innerExpireTime:       expireValue,
		}
		s.observer = map[string]Observer{}
		s.status = sessionUpdate
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
		return 0, false
	}

	switch v := val.(type) {
	case int8, int16, int32, int64, int:
		return reflect.ValueOf(v).Int(), true
	case float32, float64:
		return int64(reflect.ValueOf(v).Float()), true
	case string:
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return val, true
	default:
		log.Errorf("unsupported type for key %s: %T", key, val)
		return 0, false
	}
}

func (s *sessionImpl) GetUint(key string) (uint64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, false
	}

	switch v := val.(type) {
	case uint8, uint16, uint32, uint64, uint:
		return reflect.ValueOf(v).Uint(), true
	case float32, float64:
		return uint64(reflect.ValueOf(v).Float()), true
	case string:
		val, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return val, true
	default:
		log.Errorf("unsupported type for key %s: %T", key, val)
		return 0, false
	}
}

func (s *sessionImpl) GetFloat(key string) (float64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, false
	}

	switch v := val.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float(), true
	case string:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return val, true
	default:
		log.Errorf("unsupported type for key %s: %T", key, val)
		return 0, false
	}
}

func (s *sessionImpl) GetBool(key string) (bool, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return false, false
	}

	switch v := val.(type) {
	case bool:
		return reflect.ValueOf(v).Bool(), true
	case string:
		val, err := strconv.ParseBool(v)
		if err != nil {
			return false, false
		}
		return val, true
	default:
		log.Errorf("unsupported type for key %s: %T", key, val)
		return false, false
	}
}

func (s *sessionImpl) GetOption(key string) (any, bool) {
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	value, found := s.context[key]

	return value, found
}

func (s *sessionImpl) SetOption(key string, value any) {
	func() {
		s.registry.sessionLock.Lock()
		defer s.registry.sessionLock.Unlock()

		s.context[key] = value
		s.status = sessionUpdate
	}()

	s.save()
}

func (s *sessionImpl) RemoveOption(key string) {
	func() {
		s.registry.sessionLock.Lock()
		defer s.registry.sessionLock.Unlock()

		delete(s.context, key)
		s.status = sessionUpdate

	}()

	s.save()
}

func (s *sessionImpl) SubmitOptions() {
	if s.status != sessionUpdate {
		return
	}

	s.status = sessionActive
	for _, val := range s.observer {
		go val.OnStatusChange(s, StatusUpdate)
	}
}

func (s *sessionImpl) refresh() {
	if s.status == sessionTerminate {
		return
	}

	expireValue := time.Now().Add(GetSessionTimeOutValue()).UTC().UnixMilli()
	// 刷新有效期，每次刷新，在当前时间基础上延长有效期
	s.registry.sessionLock.Lock()
	defer s.registry.sessionLock.Unlock()
	s.context[innerExpireTime] = expireValue
}

func (s *sessionImpl) timeout() (ret bool) {
	var innerExpireTimeInt64 int64
	nowTime := time.Now().UTC().UnixMilli()
	defer func() {
		if ret {
			s.status = sessionTerminate
		}
	}()

	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	innerExpireTimeInt64 = s.getExpireTime()

	// 过期时间小于当前时间就说明已经过期
	ret = innerExpireTimeInt64 < nowTime
	return
}

// 该函数调用前必须确保sessionLock已加锁
func (s *sessionImpl) getExpireTime() int64 {
	var innerExpireTimeInt64 int64
	innerExpireTimeVal, ok := s.getInt(innerExpireTime)
	if ok {
		innerExpireTimeInt64 = innerExpireTimeVal
	}

	// 如果主动设置了过期时间，就检查这两个值谁大，没有超过最大值就认为未超时
	authExpireTimeVal, authExpireTimeOK := s.getInt(AuthExpireTime)
	if authExpireTimeOK {
		if authExpireTimeVal > innerExpireTimeInt64 {
			innerExpireTimeInt64 = authExpireTimeVal
		}
	}
	return innerExpireTimeInt64
}

func (s *sessionImpl) getInt(key string) (int64, bool) {
	optVal, optOK := s.context[key]
	if !optOK {
		return 0, false
	}
	switch val := optVal.(type) {
	case int64:
		return val, true
	case float64:
		return int64(val), true
	default:
	}

	return 0, false
}

func (s *sessionImpl) terminate() {
	s.status = sessionTerminate
	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	for _, val := range s.observer {
		go val.OnStatusChange(s, StatusTerminate)
	}
}

func (s *sessionImpl) save() {
	s.registry.updateSession(s)
}

func (s *sessionImpl) isFinal() bool {
	return s.status == sessionTerminate
}
