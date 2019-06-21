package jsonapi

import (
	"github.com/neuronlabs/neuron/internal/controller"
)

var c = controller.Default()

const (
	// MediaType is the identifier for the JSON API media type
	// see http://jsonapi.org/format/#document-structure
	MediaType = "application/vnd.api+json"
)
