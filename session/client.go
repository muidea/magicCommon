package session

import (
	"fmt"
	"net/http"
	"net/url"
)

type Client interface {
	GetServerURL() string
	GetHTTPClient() *http.Client

	AttachContext(ctx Context)
	DetachContext()

	AttachAuthorization(authorization string)
	DetachAuthorization()

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

func NewBaseClient(serverUrl string) BaseClient {
	return BaseClient{serverURL: serverUrl, httpClient: &http.Client{}}
}

type BaseClient struct {
	serverURL  string
	httpClient *http.Client

	sessionAuthorization string
	sessionToken         Token
	sessionEndpoint      *endpointInfo
	contextInfo          Context
}

func (s *BaseClient) GetServerURL() string {
	return s.serverURL
}

func (s *BaseClient) GetHTTPClient() *http.Client {
	return s.httpClient
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
		ret.Set(Authorization, fmt.Sprintf("%s %s", jwtToken, s.sessionToken))
	}
	if s.sessionEndpoint != nil {
		tokenVal, _ := SignatureEndpoint(s.sessionEndpoint.endpoint, s.sessionEndpoint.authToken)
		ret.Set(Authorization, fmt.Sprintf("%s %s", endpointToken, tokenVal))
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
