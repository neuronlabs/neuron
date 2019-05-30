package repository

import (
	"context"
)

// Repository is the interface that defines the basic neuron Repository
// it may be extended by the interfaces from the scope package
type Repository interface {
	Namer
	Close(ctx context.Context) error
}

// Namer is the interface that defines the repository name
type Namer interface {
	RepositoryName() string
}

// FactoryCloser is the interface used to close the connections
type FactoryCloser interface {
	Close(ctx context.Context, done chan<- interface{})
}
