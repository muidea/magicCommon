package session

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strings"

	fnet "github.com/muidea/magicCommon/foundation/net"
)

type Client interface {
	GetServerURL() string
	GetHTTPClient() *http.Client

	// Context 通过Header进行传递至服务器
	AttachContext(ctx Context)
	DetachContext()

	AttachAuthorization(authorization string)
	DetachAuthorization()

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
	return BaseClient{serverURL: serverUrl, httpClient: fnet.NewDNSCacheHttpClient()}
}

type BaseClient struct {
	serverURL  string
	httpClient *http.Client

	sessionAuthorization string
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

// AttachBearer sets Authorization to an explicit Bearer access token.
// Prefer this over constructing the header string at call sites.
func (s *BaseClient) AttachBearer(accessToken string) {
	s.sessionAuthorization = FormatBearerAuthorization(accessToken)
}

// WithBearer returns a clone that injects Authorization: Bearer <accessToken>.
func (s *BaseClient) WithBearer(accessToken string) BaseClient {
	return s.WithAuthorization(FormatBearerAuthorization(accessToken))
}

// FormatBearerAuthorization builds a Bearer Authorization header value.
// Empty accessToken yields an empty string (no header).
func FormatBearerAuthorization(accessToken string) string {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(accessToken), "bearer ") {
		return accessToken
	}
	return fmt.Sprintf("%s %s", jwtToken, accessToken)
}

// ParseAuthorizationScheme splits "Scheme credentials" Authorization values.
// Unknown or empty input returns empty scheme/credentials.
func ParseAuthorizationScheme(authorization string) (scheme, credentials string) {
	authorization = strings.TrimSpace(authorization)
	if authorization == "" {
		return "", ""
	}
	scheme, credentials, ok := strings.Cut(authorization, " ")
	if !ok {
		return "", ""
	}
	return scheme, strings.TrimSpace(credentials)
}

// IsBearerAuthorization reports whether the header uses the Bearer scheme.
func IsBearerAuthorization(authorization string) bool {
	scheme, _ := ParseAuthorizationScheme(authorization)
	return strings.EqualFold(scheme, jwtToken)
}

func (s *BaseClient) Release() {
	if s.httpClient != nil {
		s.httpClient.CloseIdleConnections()
		s.httpClient = nil
	}
}
