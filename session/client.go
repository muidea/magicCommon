package session

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strings"
)

type AuthSecret struct {
	Endpoint  string `json:"endpoint"`
	AuthToken string `json:"authToken"`
}

type Client interface {
	GetServerURL() string
	GetHTTPClient() *http.Client

	// Context 通过Header进行传递至服务器
	AttachContext(ctx Context)
	DetachContext()

	AttachAuthorization(authorization string)
	DetachAuthorization()

	// for endpoint
	BindAuthSecret(authSecret *AuthSecret)
	UnBindAuthSecret()

	Release()
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

// defaultHeaderContext 默认会话上下文实现
type defaultHeaderContext struct {
	values url.Values
}

// NewDefaultHeaderContext 创建新的默认会话上下文
func NewDefaultHeaderContext() Context {
	return &defaultHeaderContext{
		values: url.Values{},
	}
}

// Decode 从HTTP请求解码会话上下文
// 从Header中抽取所有X-开头的参数
func (c *defaultHeaderContext) Decode(req *http.Request) {
	c.values = url.Values{}

	// 遍历所有Header，抽取X-开头的参数
	for key, values := range req.Header {
		if strings.HasPrefix(key, "X-Mp-") && len(values) > 0 {
			c.values[key] = values
		}
	}
}

// Encode 将会话上下文编码为URL值
func (c *defaultHeaderContext) Encode(vals url.Values) url.Values {
	if vals == nil {
		vals = make(url.Values)
	}

	maps.Copy(vals, c.values)

	return vals
}

// Get 获取指定键的值
func (c *defaultHeaderContext) Get(key string) (string, bool) {
	value, ok := c.values[key]
	return value[0], ok
}

// Set 设置指定键的值
func (c *defaultHeaderContext) Set(key, value string) {
	c.values[key] = []string{value}
}

// Remove 移除指定键
func (c *defaultHeaderContext) Remove(key string) {
	delete(c.values, key)
}

// Clear 清空所有值
func (c *defaultHeaderContext) Clear() {
	c.values = url.Values{}
}

// GetAll 获取所有值
func (c *defaultHeaderContext) GetAll() url.Values {
	result := url.Values{}
	maps.Copy(result, c.values)
	return result
}

func NewBaseClient(serverUrl string) BaseClient {
	return BaseClient{serverURL: serverUrl, httpClient: &http.Client{}}
}

type BaseClient struct {
	serverURL  string
	httpClient *http.Client

	sessionAuthorization string
	sessionAuthSecret    *AuthSecret
	headerContext        Context
}

func cloneContext(ctx Context) Context {
	if ctx == nil {
		return nil
	}

	cloned := NewDefaultHeaderContext()
	values := ctx.Encode(url.Values{})
	for key, items := range values {
		for _, item := range items {
			cloned.Set(key, item)
		}
	}

	return cloned
}

func (s *BaseClient) GetServerURL() string {
	return s.serverURL
}

func (s *BaseClient) GetHTTPClient() *http.Client {
	return s.httpClient
}

func (s *BaseClient) Clone() BaseClient {
	clone := *s
	clone.headerContext = cloneContext(s.headerContext)
	if s.sessionAuthSecret != nil {
		authSecret := *s.sessionAuthSecret
		clone.sessionAuthSecret = &authSecret
	}

	return clone
}

func (s *BaseClient) WithContext(ctx Context) BaseClient {
	clone := s.Clone()
	clone.headerContext = cloneContext(ctx)
	return clone
}

func (s *BaseClient) WithAuthorization(authorization string) BaseClient {
	clone := s.Clone()
	clone.sessionAuthorization = authorization
	return clone
}

func (s *BaseClient) WithAuthSecret(authSecret *AuthSecret) BaseClient {
	clone := s.Clone()
	if authSecret == nil {
		clone.sessionAuthSecret = nil
		return clone
	}

	secretCopy := *authSecret
	clone.sessionAuthSecret = &secretCopy
	return clone
}

func (s *BaseClient) AttachContext(ctx Context) {
	s.headerContext = ctx
}

func (s *BaseClient) DetachContext() {
	s.headerContext = nil
}

func (s *BaseClient) GetContextValues() url.Values {
	ret := url.Values{}
	if s.headerContext != nil {
		ret = s.headerContext.Encode(ret)
	}

	if s.sessionAuthSecret != nil {
		ret.Set(Authorization, fmt.Sprintf("%s %s", sigToken, s.sessionAuthSecret.AuthToken))
	}
	if s.sessionAuthorization != "" {
		ret.Set(Authorization, s.sessionAuthorization)
	}

	return ret
}

func (s *BaseClient) AttachAuthorization(authorization string) {
	s.sessionAuthorization = authorization
}

func (s *BaseClient) DetachAuthorization() {
	s.sessionAuthorization = ""
}

func (s *BaseClient) BindAuthSecret(authSecret *AuthSecret) {
	s.sessionAuthSecret = authSecret
}

func (s *BaseClient) UnBindAuthSecret() {
	s.sessionAuthSecret = nil
}

func (s *BaseClient) Release() {
	if s.httpClient != nil {
		s.httpClient.CloseIdleConnections()
		s.httpClient = nil
	}
}
