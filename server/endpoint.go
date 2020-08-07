package server

import (
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
)

// Endpoint is a structure that represents api endpoint.
type Endpoint struct {
	Path        string
	HTTPMethod  string
	QueryMethod query.Method
	ModelStruct *mapping.ModelStruct
	Relation    *mapping.StructField
}
