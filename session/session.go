package session

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"log/slog"
)

type Status int

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
	SessionToken        = "session_token"
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
	Signature() (string, error)
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
	mu       sync.RWMutex
	context  map[string]any
	observer map[string]Observer
	registry *sessionRegistryImpl
	status   int
}

func (s *sessionImpl) ID() string {
	return s.id
}

func excludeSessionSignatureKey(key string) bool {
	// 以下划线开头的key也要进行排除
	return strings.HasPrefix(key, "_")
}

func (s *sessionImpl) excludeKey(key string) bool {
	return excludeSessionSignatureKey(key)
}

func (s *sessionImpl) Signature() (string, error) {
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

func (s *sessionImpl) Reset() {
	expireValue := time.Now().Add(GetSessionTimeOutValue()).UTC().UnixMilli()
	s.mu.Lock()
	startTime := s.context[InnerStartTime]
	remoteAccessAddr := s.context[InnerRemoteAccessAddr]
	useAgent := s.context[InnerUseAgent]
	s.context = map[string]any{
		InnerStartTime:        startTime,
		InnerRemoteAccessAddr: remoteAccessAddr,
		InnerUseAgent:         useAgent,
		innerExpireTime:       expireValue,
	}
	s.observer = map[string]Observer{}
	s.status = sessionUpdate
	s.mu.Unlock()

	s.save()
}

func (s *sessionImpl) BindObserver(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.observer[observer.ID()]
	if ok {
		return
	}
	s.observer[observer.ID()] = observer
}

func (s *sessionImpl) UnbindObserver(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	return convertToInt64(val, key)
}

func (s *sessionImpl) GetUint(key string) (uint64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, false
	}
	return convertToUint64(val, key)
}

func (s *sessionImpl) GetFloat(key string) (float64, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return 0, false
	}
	return convertToFloat64(val, key)
}

func (s *sessionImpl) GetBool(key string) (bool, bool) {
	val, ok := s.GetOption(key)
	if !ok {
		return false, false
	}
	return convertToBool(val, key)
}

func (s *sessionImpl) GetOption(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, found := s.context[key]

	return value, found
}

func (s *sessionImpl) SetOption(key string, value any) {
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.context[key] = value
		s.status = sessionUpdate
	}()

	s.save()
}

func (s *sessionImpl) RemoveOption(key string) {
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.context, key)
		s.status = sessionUpdate

	}()

	s.save()
}

func (s *sessionImpl) SubmitOptions() {
	s.mu.Lock()
	if s.status != sessionUpdate {
		s.mu.Unlock()
		return
	}

	s.status = sessionActive
	observers := make([]Observer, 0, len(s.observer))
	for _, val := range s.observer {
		observers = append(observers, val)
	}
	s.mu.Unlock()

	for _, val := range observers {
		go val.OnStatusChange(s, StatusUpdate)
	}
}

func (s *sessionImpl) refresh() {
	if s.status == sessionTerminate {
		return
	}

	expireValue := time.Now().Add(GetSessionTimeOutValue()).UTC().UnixMilli()
	// 刷新有效期，每次刷新，在当前时间基础上延长有效期
	s.mu.Lock()
	defer s.mu.Unlock()
	s.context[innerExpireTime] = expireValue
}

func (s *sessionImpl) timeout() (ret bool) {
	var innerExpireTimeInt64 int64
	nowTime := time.Now().UTC().UnixMilli()

	s.mu.RLock()
	innerExpireTimeInt64 = s.getExpireTime()
	s.mu.RUnlock()

	// 过期时间小于当前时间就说明已经过期
	ret = innerExpireTimeInt64 < nowTime
	if ret {
		s.mu.Lock()
		s.status = sessionTerminate
		s.mu.Unlock()
	}
	return
}

// 该函数调用前必须确保 session 已加锁
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
	s.mu.Lock()
	s.status = sessionTerminate
	observers := make([]Observer, 0, len(s.observer))
	for _, val := range s.observer {
		observers = append(observers, val)
	}
	s.mu.Unlock()

	for _, val := range observers {
		go val.OnStatusChange(s, StatusTerminate)
	}
}

func (s *sessionImpl) save() {
	s.registry.updateSession(s)
}

func (s *sessionImpl) isFinal() bool {
	return s.status == sessionTerminate
}

// 类型转换辅助函数

// convertToInt64 将任意值转换为int64
func convertToInt64(val any, key string) (int64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case int8, int16, int32, int64, int:
		return reflect.ValueOf(v).Int(), true
	case uint8, uint16, uint32, uint64, uint:
		return int64(reflect.ValueOf(v).Uint()), true
	case float32, float64:
		return int64(reflect.ValueOf(v).Float()), true
	case string:
		result, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			slog.Error("failed to parse string as int64", "key", key, "value", v, "error", err)
			return 0, false
		}
		return result, true
	default:
		slog.Error("unsupported type for int64 conversion", "key", key, "type", fmt.Sprintf("%T", val), "value", val)
		return 0, false
	}
}

// convertToUint64 将任意值转换为uint64
func convertToUint64(val any, key string) (uint64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case uint8, uint16, uint32, uint64, uint:
		return reflect.ValueOf(v).Uint(), true
	case int8, int16, int32, int64, int:
		// 检查是否为负数
		intVal := reflect.ValueOf(v).Int()
		if intVal < 0 {
			slog.Error("negative value cannot be converted to uint64", "key", key, "value", intVal)
			return 0, false
		}
		return uint64(intVal), true
	case float32, float64:
		floatVal := reflect.ValueOf(v).Float()
		if floatVal < 0 {
			slog.Error("negative value cannot be converted to uint64", "key", key, "value", floatVal)
			return 0, false
		}
		return uint64(floatVal), true
	case string:
		result, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			slog.Error("failed to parse string as uint64", "key", key, "value", v, "error", err)
			return 0, false
		}
		return result, true
	default:
		slog.Error("unsupported type for uint64 conversion", "key", key, "type", fmt.Sprintf("%T", val), "value", val)
		return 0, false
	}
}

// convertToFloat64 将任意值转换为float64
func convertToFloat64(val any, key string) (float64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float(), true
	case int8, int16, int32, int64, int:
		return float64(reflect.ValueOf(v).Int()), true
	case uint8, uint16, uint32, uint64, uint:
		return float64(reflect.ValueOf(v).Uint()), true
	case string:
		result, err := strconv.ParseFloat(v, 64)
		if err != nil {
			slog.Error("failed to parse string as float64", "key", key, "value", v, "error", err)
			return 0, false
		}
		return result, true
	default:
		slog.Error("unsupported type for float64 conversion", "key", key, "type", fmt.Sprintf("%T", val), "value", val)
		return 0, false
	}
}

// convertToBool 将任意值转换为bool
func convertToBool(val any, key string) (bool, bool) {
	if val == nil {
		return false, false
	}

	switch v := val.(type) {
	case bool:
		return v, true
	case string:
		result, err := strconv.ParseBool(v)
		if err != nil {
			slog.Error("failed to parse string as bool", "key", key, "value", v, "error", err)
			return false, false
		}
		return result, true
	case int8, int16, int32, int64, int:
		intVal := reflect.ValueOf(v).Int()
		return intVal != 0, true
	case uint8, uint16, uint32, uint64, uint:
		uintVal := reflect.ValueOf(v).Uint()
		return uintVal != 0, true
	case float32, float64:
		floatVal := reflect.ValueOf(v).Float()
		return floatVal != 0, true
	default:
		slog.Error("unsupported type for bool conversion", "key", key, "type", fmt.Sprintf("%T", val), "value", val)
		return false, false
	}
}
