package routers

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/i18n"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
	"net/http"
)

var routers = newContainer()

// Errors used by the router setter
var (
	ErrRouterNotRegistered error = errors.New("Router not registered.")
	ErrNoRouterNameDefined error = errors.New("No router name provided.")
)

// SetterFunc is the function that sets all the api routes for the provided handler
type SetterFunc func(
	c *controller.Controller,
	cfg *config.Gateway,
	sup *i18n.Support,
) (http.Handler, error)

// RegisterRouter registers provided router with the provided name
// The router is obtainable after calling the RouterSetterFunc
func RegisterRouter(name string, f SetterFunc) {
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
	cfg *config.Gateway,
	i *i18n.Support,
) (http.Handler, error) {
	if cfg.Router.Name == "" {
		return nil, ErrNoRouterNameDefined
	}
	f, ok := routers.RegisteredRouters[cfg.Router.Name]
	if !ok {
		return nil, ErrRouterNotRegistered
	}

	return f(c, cfg, i)
}

// RouterSetterContainer is the container for the registered router setter functions
type routerSetterContainer struct {
	RegisteredRouters map[string]SetterFunc
}

func newContainer() *routerSetterContainer {
	r := &routerSetterContainer{
		RegisteredRouters: make(map[string]SetterFunc),
	}
	return r
}
