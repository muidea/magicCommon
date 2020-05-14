package common

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

// SessionIdentity session identity
const SessionIdentity = "sessionIdentity"

// AuthAccount 鉴权Account
const AuthAccount = "authAccount"

// AuthPrivateGroup Account 权限组
const AuthPrivateGroup = "authPrivateGroup"

// AuthToken 鉴权Token
const AuthToken = "authToken"

// IdentifyID 标识ID
const IdentifyID = "identifyID"

// ExpiryDate 会话有效期
const ExpiryDate = "expiryDate"

const (
	// ShareSession share session flag value
	ShareSession = "shareSession"
)

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
	s.ID = req.URL.Query().Get(sessionID)
	s.Token = req.URL.Query().Get(sessionToken)
	s.Scope = req.URL.Query().Get(sessionScope)
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

// SessionContext session context
type SessionContext struct {
	sessionInfo *SessionInfo
}

// NewSessionContext new session context
func NewSessionContext(sessionInfo *SessionInfo) *SessionContext {
	return &SessionContext{sessionInfo: sessionInfo}
}

// Decode session context
func (s *SessionContext) Decode(req *http.Request) {

	sessionInfo := &SessionInfo{}
	sessionInfo.ID = req.Header.Get(sessionID)
	sessionInfo.Token = req.Header.Get(sessionToken)
	sessionInfo.Scope = req.Header.Get(sessionScope)
	if sessionInfo.ID != "" {
		s.sessionInfo = sessionInfo
	}
}

// Encode session context
func (s *SessionContext) Encode(values url.Values) (ret url.Values) {
	if s.sessionInfo != nil {
		if s.sessionInfo.ID != "" {
			values.Set(sessionID, s.sessionInfo.ID)
		}

		if s.sessionInfo.Token != "" {
			values.Set(sessionToken, s.sessionInfo.Token)
		}

		if s.sessionInfo.Scope != "" {
			values.Set(sessionScope, s.sessionInfo.Scope)
		}
	}

	ret = values
	return
}

// GetSessionInfo get session info
func (s *SessionContext) GetSessionInfo() *SessionInfo {
	return s.sessionInfo
}
