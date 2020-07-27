package core

import (
	"github.com/neuronlabs/neuron/controller"
)

// Initializer is an interface used to initialize services, repositories etc.
type Initializer interface {
	Initialize(c *controller.Controller) error
}
