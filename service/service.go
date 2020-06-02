package service

import (
	"context"

	"github.com/neuronlabs/neuron/mapping"
)

// Repository is the interface that defines the base neuron Repository.
type Service interface {
	// ID gets the repository unique identification.
	ID() string
	// FactoryName returns the factory name for given repository.
	FactoryName() string
}

// Closer is the interface used to close the repository connections.
type Closer interface {
	// Close closes the connection for given repository.
	Close(ctx context.Context) error
}

// Dialer is the interface used to establish a connection for the repository.
type Dialer interface {
	Dial(ctx context.Context) error
}

// ModelRegistrar is the interface used to register the models in the repository.
type ModelRegistrar interface {
	// RegisterModels registers provided 'models' into Repository specific mappings.
	RegisterModels(models ...*mapping.ModelStruct) error
}

// Migrator migrates the models into the repository.
type Migrator interface {
	MigrateModels(ctx context.Context, models ...*mapping.ModelStruct) error
}

// Runner is the interface for the services that are runnable.
type Runner interface {
	Run(ctx context.Context) error
}
