package repository

import (
	"github.com/neuronlabs/neuron-core/mapping"
)

// ModelStructer gets the model struct from the given model.
type ModelStructer interface {
	ModelStruct(model interface{}) (*mapping.ModelStruct, error)
}
