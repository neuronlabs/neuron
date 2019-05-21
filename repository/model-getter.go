package repository

import (
	"github.com/neuronlabs/neuron/mapping"
)

// ModelStructer gets the model struct
type ModelStructer interface {
	ModelStruct(model interface{}) (*mapping.ModelStruct, error)
}
