package jsonapi

import (
	"github.com/neuronlabs/neuron-core/internal/controller"
)

var c = controller.Default()

const (
	// MediaType is the identifier for the JSON API media type
	// see http://jsonapi.org/format/#document-structure
	MediaType = "application/vnd.api+json"

	// Iso8601TimeFormat is the time formatting for the ISO 8601.
	Iso8601TimeFormat = "2006-01-02T15:04:05Z"
)
