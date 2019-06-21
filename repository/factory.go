package repository

import (
	"context"

	"github.com/neuronlabs/neuron/mapping"
)

// Factory is the interface used for creating the repositories.
// It implements FactoryCloser and Namer interface.
type Factory interface {
	FactoryCloser

	// Namer gets the repository name for given factory.
	Namer

	// New creates new instance of the Repository for given 'model'.
	New(structer ModelStructer, model *mapping.ModelStruct) (Repository, error)
}

// FactoryCloser is the interface used to close the connections.
type FactoryCloser interface {
	Close(ctx context.Context, done chan<- interface{})
}
