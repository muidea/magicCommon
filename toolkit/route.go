package toolkit

import (
	"context"
	"net/http"

	engine "github.com/muidea/magicEngine"
)

// RouteRegistry route registry
type RouteRegistry interface {
	SetApiVersion(version string)

	AddHandler(pattern, method string, handler func(context.Context, http.ResponseWriter, *http.Request))

	AddRoute(route engine.Route, filters ...engine.MiddleWareHandler)
}

// NewRouteRegistry create routeRegistry
func NewRouteRegistry(router engine.Router) (ret RouteRegistry) {
	ret = &routeRegistryImpl{router: router}
	return
}

// routeRegistryImpl route registry
type routeRegistryImpl struct {
	router engine.Router
}

func (s *routeRegistryImpl) SetApiVersion(version string) {
	s.router.SetApiVersion(version)
}

// AddHandler add route handler
func (s *routeRegistryImpl) AddHandler(
	pattern, method string,
	handler func(context.Context, http.ResponseWriter, *http.Request)) {

	s.router.AddRoute(engine.CreateRoute(pattern, method, handler))
}

func (s *routeRegistryImpl) AddRoute(route engine.Route, filters ...engine.MiddleWareHandler) {
	s.router.AddRoute(route, filters...)
}
