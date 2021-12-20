package toolkit

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/session"
	engine "github.com/muidea/magicEngine"
)

// CasVerifier cas verifier
type CasVerifier interface {
	Verify(ctx context.Context, res http.ResponseWriter, req *http.Request) (*def.Entity, error)
}

// CasRegistry private route registry
type CasRegistry interface {
	SetApiVersion(version string)

	AddHandler(pattern, method string, handler func(context.Context, http.ResponseWriter, *http.Request))

	AddRoute(route engine.Route, filters ...engine.MiddleWareHandler)
}

// NewCasRegistry create cas Registry
func NewCasRegistry(verifier CasVerifier, router engine.Router) (ret CasRegistry) {
	ret = &casRegistryImpl{casVerifier: verifier, router: router}
	return
}

// casRegistryImpl cas route registry
type casRegistryImpl struct {
	casVerifier CasVerifier
	router      engine.Router
}

func (s *casRegistryImpl) SetApiVersion(version string) {
	s.router.SetApiVersion(version)
}

// AddHandler add route handler
func (s *casRegistryImpl) AddHandler(
	pattern, method string,
	handler func(context.Context, http.ResponseWriter, *http.Request)) {

	s.router.AddRoute(engine.CreateRoute(pattern, method, handler), s)
}

func (s *casRegistryImpl) AddRoute(route engine.Route, filters ...engine.MiddleWareHandler) {
	filters = append(filters, s)
	s.router.AddRoute(route, filters...)
}

// Handle middleware handler
func (s *casRegistryImpl) Handle(ctx engine.RequestContext, res http.ResponseWriter, req *http.Request) {
	result := &def.Result{ErrorCode: def.Success}

	for {
		entityView, entityErr := s.casVerifier.Verify(ctx.Context(), res, req)
		if entityErr != nil {
			result.ErrorCode = def.InvalidAuthority
			result.Reason = entityErr.Error()
			break
		}

		casCtx := context.WithValue(ctx.Context(), session.AuthAccount, entityView)
		ctx.Update(casCtx)
		break
	}

	if result.Fail() {
		block, err := json.Marshal(result)
		if err == nil {
			res.Write(block)
			return
		}

		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx.Next()
}
