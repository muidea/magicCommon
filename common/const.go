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

// Encode encode session info
func (s *SessionInfo) Encode() (ret string) {
	ret = ""

	values := url.Values{}

	if s.ID != "" {
		values.Set(sessionID, s.ID)
	}

	if s.Token != "" {
		values.Set(sessionToken, s.Token)
	}

	if s.Scope != "" {
		values.Set(sessionScope, s.Scope)
	}

	ret = values.Encode()
	return
}

// Merge merge values
func (s *SessionInfo) Merge(values url.Values) (ret url.Values) {
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
