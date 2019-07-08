package repository

import (
	"context"

	"github.com/neuronlabs/neuron-core/mapping"
)

// Factory is the interface used for creating the repositories.
type Factory interface {
	// Namer gets the repository name for given factory.
	DriverName() string
	// New creates new instance of the Repository for given 'model'.
	New(c Controller, model *mapping.ModelStruct) (Repository, error)
	// Close should close all the repository instances for this factory.
	Close(ctx context.Context, done chan<- interface{})
}
