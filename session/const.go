package session

import (
	"net/http"
	"net/url"
)

const (
	// sessionID 会话ID
	sessionID = "sessionID"
	// sessionToken 会话Token
	sessionToken = "sessionToken"
	// sessionScope 会话域，标识是否是共享会话
	sessionScope = "sessionScope"
)

const (
	// ShareSession share session flag value
	ShareSession = "shareSession"
)

// AuthSession auth Session
const AuthSession = "$$contextSession"

// AuthAccount 鉴权Account
const AuthAccount = "$$sessionAccount"

// AuthRole Account 权限组
const AuthRole = "$$sessionRole"

// AuthNamespace Account namespace
const AuthNamespace = "$$sessionNamespace"

// AuthRemoteAddress 远端地址
const AuthRemoteAddress = "$$sessionRemoteAddress"

// expiryValue 会话有效期
const expiryValue = "$$sessionExpiryValue"

// refreshTime 会话刷新时间
const refreshTime = "$$sessionRefreshTime"

// authSessionInfo session info
const authSessionInfo = "$$sessionInfo"

const defaultSessionCookieID = "$$session_info"

// SessionInfo session info
type SessionInfo struct {
	ID    string `json:"sessionID"`
	Token string `json:"sessionToken"`
	Scope string `json:"sessionScope"`
}

// Encode encode values
func (s *SessionInfo) Encode(values url.Values) (ret url.Values) {
	if s.ID != "" {
		values.Set(sessionID, s.ID)
	}

	if s.Token != "" {
		values.Set(sessionToken, s.Token)
	}

	if s.Scope != "" {
		values.Set(sessionScope, s.Scope)
	}

	ret = values
	return
}

// Decode decode session info
func (s *SessionInfo) Decode(req *http.Request) {
	s.ID = req.Header.Get(sessionID)
	s.Token = req.Header.Get(sessionToken)
	s.Scope = req.Header.Get(sessionScope)
}

// Same compare SessionInfo
func (s *SessionInfo) Same(right *SessionInfo) bool {
	return s.ID == right.ID && s.Token == right.Token && s.Scope == right.Scope
}

// ContextInfo context info
type ContextInfo interface {
	Decode(req *http.Request)
	Encode(vals url.Values) url.Values
}
