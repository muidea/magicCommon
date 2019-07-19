package route

import (
	"encoding/json"
	"net/http"

	"github.com/muidea/magicCommon/common"
	"github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/session"
	engine "github.com/muidea/magicEngine"
)

type rtItem struct {
	pattern         string
	method          string
	sessionRegistry session.Registry
	handler         func(http.ResponseWriter, *http.Request)
}

func (s *rtItem) Pattern() string {
	return s.pattern
}

func (s *rtItem) Method() string {
	return s.method
}

func (s *rtItem) Handler() func(http.ResponseWriter, *http.Request) {
	return s.casHandler
}

func (s *rtItem) casHandler(res http.ResponseWriter, req *http.Request) {
	sessionInfo := &common.SessionInfo{}
	sessionInfo.Decode(req)

	result := &def.Result{ErrorCode: def.Success}
	session := s.sessionRegistry.GetSession(res, req)
	sessionInfoVal, ok := session.GetOption(common.SessionIdentity)
	if !ok {
		result.ErrorCode = def.InvalidAuthority
		result.Reason = "当前会话无权限"
	} else {
		if sessionInfoVal.(*common.SessionInfo).Token != sessionInfo.Token {
			result.ErrorCode = def.InvalidAuthority
			result.Reason = "当前会话无权限"
		}
	}

	if result.Success() {
		s.handler(res, req)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	block, err := json.Marshal(result)
	if err == nil {
		res.Write(block)
		return
	}

	res.WriteHeader(http.StatusInternalServerError)
}

// CreateCasRoute create cas Route
func CreateCasRoute(pattern, method string, sessionRegistry session.Registry, handler func(http.ResponseWriter, *http.Request)) engine.Route {
	return &rtItem{pattern: pattern, method: method, sessionRegistry: sessionRegistry, handler: handler}
}
