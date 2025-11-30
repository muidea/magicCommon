package session

import (
	"net/http"
	"net/url"
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
	// sessionID 会话ID
	sessionID = "_sessionID"
	// expiryTime 会话有效期
	expiryTime = "_expiryTime"
	// Authorization info, from request header
	Authorization = "Authorization"
)

const (
	jwtToken = "Bearer"
	sigToken = "Sig"

	DefaultSessionTimeOutValue = 10 * time.Minute // 10 minute
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
	GetOption(key string) (interface{}, bool)
	SetOption(key string, value interface{})
	RemoveOption(key string)
	SubmitOptions()
}

type sessionImpl struct {
	id            string // session id
	context       map[string]interface{}
	observer      map[string]Observer
	registry      *sessionRegistryImpl
	optionsChange bool
}

func (s *sessionImpl) ID() string {
	return s.id
}

func (s *sessionImpl) innerKey(key string) bool {
	switch key {
	case Authorization:
		return true
	}

	return false
}

func (s *sessionImpl) Signature() (Token, error) {
	mc := jwt.MapClaims{}
	if s.id != "" {
		mc[sessionID] = s.id
	}

	func() {
		s.registry.sessionLock.RLock()
		defer s.registry.sessionLock.RUnlock()
		for k, v := range s.context {
			if s.innerKey(k) {
				continue
			}

			mc[k] = v
		}
	}()

	return SignatureJWT(mc)
}

func (s *sessionImpl) Reset() {
	expiryValue := time.Now().Add(DefaultSessionTimeOutValue).UTC().Unix()
	func() {
		s.registry.sessionLock.RLock()
		defer s.registry.sessionLock.RUnlock()

		s.context = map[string]interface{}{expiryTime: expiryValue}
		s.observer = map[string]Observer{}
		s.optionsChange = true
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
		s.optionsChange = true
	}()

	s.save()
}

func (s *sessionImpl) RemoveOption(key string) {
	func() {
		s.registry.sessionLock.Lock()
		defer s.registry.sessionLock.Unlock()

		delete(s.context, key)
		s.optionsChange = true

	}()

	s.save()
}

func (s *sessionImpl) SubmitOptions() {
	if !s.optionsChange {
		return
	}

	s.optionsChange = false
	for _, val := range s.observer {
		go val.OnStatusChange(s, StatusUpdate)
	}
}

func (s *sessionImpl) refresh() {
	expiryValue := time.Now().Add(DefaultSessionTimeOutValue).UTC().Unix()

	s.registry.sessionLock.Lock()
	defer s.registry.sessionLock.Unlock()
	s.context[expiryTime] = expiryValue
}

func (s *sessionImpl) timeout() bool {
	nowTime := time.Now().UTC().Unix()

	s.registry.sessionLock.RLock()
	defer s.registry.sessionLock.RUnlock()

	expiryTimeVal, ok := s.context[expiryTime]
	if !ok {
		return true // 如果没有设置过期时间，默认认为已超时
	}

	expiryTimeInt64, ok := expiryTimeVal.(int64)
	if !ok {
		log.Errorf("invalid type for expiryTime: %T", expiryTimeVal)
		return true // 类型不正确，默认认为已超时
	}

	return expiryTimeInt64 < nowTime
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
	Get(key string) (string, bool)
	Set(key, value string)
	Remove(key string)
	Clear()
}

// defaultSessionContext 默认会话上下文实现
type defaultSessionContext struct {
	values url.Values
}

// NewDefaultSessionContext 创建新的默认会话上下文
func NewDefaultSessionContext() Context {
	return &defaultSessionContext{
		values: url.Values{},
	}
}

// Decode 从HTTP请求解码会话上下文
// 从Header中抽取所有X-开头的参数
func (c *defaultSessionContext) Decode(req *http.Request) {
	c.values = url.Values{}

	// 遍历所有Header，抽取X-开头的参数
	for key, values := range req.Header {
		if strings.HasPrefix(key, "X-") && len(values) > 0 {
			c.values[key] = values
		}
	}
}

// Encode 将会话上下文编码为URL值
func (c *defaultSessionContext) Encode(vals url.Values) url.Values {
	if vals == nil {
		vals = make(url.Values)
	}

	for key, value := range c.values {
		vals[key] = value
	}

	return vals
}

// Get 获取指定键的值
func (c *defaultSessionContext) Get(key string) (string, bool) {
	value, ok := c.values[key]
	return value[0], ok
}

// Set 设置指定键的值
func (c *defaultSessionContext) Set(key, value string) {
	c.values[key] = []string{value}
}

// Remove 移除指定键
func (c *defaultSessionContext) Remove(key string) {
	delete(c.values, key)
}

// Clear 清空所有值
func (c *defaultSessionContext) Clear() {
	c.values = url.Values{}
}

// GetAll 获取所有值
func (c *defaultSessionContext) GetAll() url.Values {
	result := url.Values{}
	for k, v := range c.values {
		result[k] = v
	}
	return result
}
