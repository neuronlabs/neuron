package repository

import (
	"context"

	"github.com/neuronlabs/neuron/mapping"
)

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
