package repository

import (
	"context"

	"github.com/neuronlabs/neuron-core/mapping"
)

// Repository is the interface that defines the base neuron Repository.
type Repository interface {
	// Dial establish all possible repository connections.
	Dial(ctx context.Context) error
	// FactoryName returns the factory name for given repository.
	FactoryName() string
	// RegisterModels registers provided 'models' into Repository specific mappings.
	RegisterModels(models ...*mapping.ModelStruct) error
	// HealthCheck defines the health status of the repository.
	HealthCheck(ctx context.Context) (*HealthResponse, error)
	// Close closes the connection for given repository.
	Close(ctx context.Context) error
}

// Migrator migrates the models into the repository.
type Migrator interface {
	MigrateModels(models ...*mapping.ModelStruct) error
}
