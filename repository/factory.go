package repository

import (
	"github.com/neuronlabs/neuron/mapping"
)

// Factory is the interface used for creating the repositories
type Factory interface {
	New(structer ModelStructer, model *mapping.ModelStruct) (Repository, error)
	Namer
	FactoryCloser
}
