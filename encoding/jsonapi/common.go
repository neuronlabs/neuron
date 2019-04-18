package jsonapi

import (
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/controller"
)

var c *controller.Controller = controller.Default()

const (
	// MediaType defines the HTTP ContentType value for the jsonapi specification
	MediaType string = internal.MediaType
)
