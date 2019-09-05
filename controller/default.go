package controller

import (
	"github.com/neuronlabs/neuron-core/config"
)

// DefaultController is the Default controller used if no 'controller' is provided for operations
var DefaultController *Controller

// Default returns current default controller.
func Default() *Controller {
	if DefaultController == nil {
		c, err := newController(config.ReadDefaultControllerConfig())
		if err != nil {
			panic(err)
		}
		DefaultController = c
	}
	return DefaultController
}

// NewDefault creates new default controller based on the default config.
func NewDefault() *Controller {
	c, err := newController(config.ReadDefaultControllerConfig())
	if err != nil {
		panic(err)
	}
	return c
}
