package repository

import (
	"github.com/neuronlabs/neuron-core/mapping"
)

// Controller gets the model struct from the given model.
type Controller interface {
	ModelStruct(model interface{}) (*mapping.ModelStruct, error)
}
