package middlewares

import (
	"github.com/neuronlabs/neuron/log"
)

var c = newcontainer()

// container is the middleware container for the middleware functions
type container struct {
	// middlewares contains the middleware functions with their names
	middlewares map[string]MiddlewareFunc
}

func newcontainer() *container {
	c := &container{
		middlewares: make(map[string]MiddlewareFunc),
	}

	return c
}

// RegisterMiddleware registers provided middleware within the container
func RegisterMiddleware(name string, mid MiddlewareFunc) error {
	return c.registerMiddleware(name, mid)
}

// Get returns the middleware function by it's name
// If the middleware with provided name doesn't exists returns an error
func Get(name string) (MiddlewareFunc, error) {
	return c.get(name)
}

// Get returns the middleware function by it's name
func (c *container) get(name string) (MiddlewareFunc, error) {
	f, ok := c.middlewares[name]
	if !ok {
		log.Errorf("Middleware with the name: '%s' is not registered.", name)
		return nil, ErrMiddlewareNotRegistered
	}
	return f, nil
}

// registermiddlewares registers provided middlewares for the given gateway
func (c *container) registerMiddleware(name string, mid MiddlewareFunc) error {
	_, ok := c.middlewares[name]
	if ok {
		log.Errorf("MiddlewareFunc with name: '%s' already registered.", name)
		return ErrMiddlewareAlreadyRegistered
	}
	c.middlewares[name] = mid
	log.Debugf("Registered: '%s' middleware successfully.", name)
	return nil
}
