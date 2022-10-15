package session

import (
	"net/http"
	"net/url"
)

const (
	// sessionID 会话ID
	sessionID = "sessionID"
	// RemoteAddress 远端地址
	RemoteAddress = "$$sessionRemoteAddress"
	// ExpiryValue 会话有效期
	ExpiryValue = "$$sessionExpiryValue"
	// refreshTime 会话刷新时间
	refreshTime = "$$sessionRefreshTime"
)

// ContextInfo context info
type ContextInfo interface {
	Decode(req *http.Request)
	Encode(vals url.Values) url.Values
}
