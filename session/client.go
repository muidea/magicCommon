package session

import (
	"fmt"
	"net/http"
	"net/url"
)

type Client interface {
	AttachContext(ctx Context)
	DetachContext()

	BindToken(token Token)
	UnBindToken()

	BindEndpoint(endpoint *Endpoint)
	UnBindEndpoint()

	Release()
}

type BaseClient struct {
	serverURL       string
	sessionToken    Token
	sessionEndpoint *Endpoint
	contextInfo     Context
	httpClient      *http.Client
}

func (s *BaseClient) AttachContext(ctx Context) {
	s.contextInfo = ctx
}

func (s *BaseClient) DetachContext() {
	s.contextInfo = nil
}

func (s *BaseClient) GetContextValues() url.Values {
	ret := url.Values{}
	if s.contextInfo != nil {
		ret = s.contextInfo.Encode(ret)
	}

	if s.sessionToken != "" {
		ret.Set("Authorization", fmt.Sprintf("%s %s", jwtToken, s.sessionToken))
	}
	if s.sessionEndpoint != nil {
		ret.Set("Authorization", fmt.Sprintf("%s %s", sigToken, signature(s.sessionEndpoint, ret)))
	}

	return ret
}

func (s *BaseClient) BindToken(sessionToken Token) {
	s.sessionToken = sessionToken
}

func (s *BaseClient) UnBindToken() {
	s.sessionToken = ""
}

func (s *BaseClient) Release() {
	if s.httpClient != nil {
		s.httpClient.CloseIdleConnections()
		s.httpClient = nil
	}
}
