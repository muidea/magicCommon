package toolkit

import (
	"net/http"

	engine "github.com/muidea/magicEngine"
)

// RouteRegistry private route registry
type RouteRegistry interface {
	AddHandler(pattern, method string, handler func(http.ResponseWriter, *http.Request))

	AddRoute(route engine.Route, filters ...engine.MiddleWareHandler)
}

// NewRouteRegistry create routeRegistry
func NewRouteRegistry(router engine.Router) RouteRegistry {
	return &routeRegistryImpl{router: router}
}

// routeRegistryImpl route registry
type routeRegistryImpl struct {
	router engine.Router
}

// AddHandler add route handler
func (s *routeRegistryImpl) AddHandler(
	pattern, method string,
	handler func(http.ResponseWriter, *http.Request)) {

	s.router.AddRoute(engine.CreateRoute(pattern, method, handler))
}

func (s *routeRegistryImpl) AddRoute(route engine.Route, filters ...engine.MiddleWareHandler) {
	s.router.AddRoute(route, filters...)
}
