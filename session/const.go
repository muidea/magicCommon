package session

import (
	"net/http"
	"net/url"
)

const (
	// sessionID 会话ID
	sessionID     = "sessionID"
	AuthNamespace = "authNamespace"
)

// AuthRemoteAddress 远端地址
const AuthRemoteAddress = "$$sessionRemoteAddress"

// AuthExpiryValue 会话有效期
const AuthExpiryValue = "$$sessionExpiryValue"

// refreshTime 会话刷新时间
const refreshTime = "$$sessionRefreshTime"

// ContextInfo context info
type ContextInfo interface {
	Decode(req *http.Request)
	Encode(vals url.Values) url.Values
}
