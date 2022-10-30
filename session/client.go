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

	BindEndpoint(endpoint, authToken string)
	UnBindEndpoint()

	Release()
}

type endpointInfo struct {
	endpoint  string
	authToken string
}

type BaseClient struct {
	serverURL       string
	sessionToken    Token
	sessionEndpoint *endpointInfo
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
		tokenVal, _ := SignatureEndpoint(s.sessionEndpoint.endpoint, s.sessionEndpoint.authToken)
		ret.Set("Authorization", fmt.Sprintf("%s %s", endpointToken, tokenVal))
	}

	return ret
}

func (s *BaseClient) BindToken(sessionToken Token) {
	s.sessionToken = sessionToken
}

func (s *BaseClient) UnBindToken() {
	s.sessionToken = ""
}

func (s *BaseClient) BindEndpoint(endpoint, authToken string) {
	s.sessionEndpoint = &endpointInfo{endpoint: endpoint, authToken: authToken}
}

func (s *BaseClient) UnBindEndpoint() {
	s.sessionEndpoint = nil
}

func (s *BaseClient) Release() {
	if s.httpClient != nil {
		s.httpClient.CloseIdleConnections()
		s.httpClient = nil
	}
}
