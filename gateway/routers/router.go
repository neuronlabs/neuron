package routers

import (
	"github.com/kucjac/jsonapi/config"
	"github.com/kucjac/jsonapi/controller"
	"github.com/kucjac/jsonapi/log"
	"github.com/pkg/errors"
	"net/http"
)

var routers *routerSetterContainer = newContainer()

var (
	ErrRouterNotRegistered error = errors.New("Router not registered.")
	ErrNoRouterNameDefined error = errors.New("No router name provided.")
)

// RouteSetterFunc is the function that sets all the api routes for the provided handler
type RouterSetterFunc func(
	c *controller.Controller,
	cfg *config.RouterConfig,
) (http.Handler, error)

// RegisterRouter registers provided router with the provided name
// The router is obtainable after calling the RouterSetterFunc
func RegisterRouter(name string, f RouterSetterFunc) {
	_, ok := routers.RegisteredRouters[name]
	if ok {
		panic(errors.Errorf("Router: '%s' is already registered.", name))
	}

	routers.RegisteredRouters[name] = f
	log.Debugf("Router: '%s' registered with success.", name)
}

// GetSettedRouter gets the registered router for the provided controller, middlewares
// and configuration.
func GetSettedRouter(
	c *controller.Controller,
	cfg *config.RouterConfig,
) (http.Handler, error) {
	if cfg.Name == "" {
		return nil, ErrNoRouterNameDefined
	}
	f, ok := routers.RegisteredRouters[cfg.Name]
	if !ok {
		return nil, ErrRouterNotRegistered
	}

	return f(c, cfg)
}

// RouterSetterContainer is the container for the registered router setter functions
type routerSetterContainer struct {
	RegisteredRouters map[string]RouterSetterFunc
}

func newContainer() *routerSetterContainer {
	r := &routerSetterContainer{
		RegisteredRouters: make(map[string]RouterSetterFunc),
	}
	return r
}
