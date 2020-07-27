package db

import (
	"github.com/neuronlabs/neuron/controller"
)

// Collection is the interface used for the orm model collections.
type Collection interface {
	InitializeCollection(c *controller.Controller) error
}
