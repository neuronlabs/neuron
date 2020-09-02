package repository

import (
	"context"

	"github.com/neuronlabs/neuron/mapping"
)

// Migrator migrates the models into the repository.
type Migrator interface {
	MigrateModels(ctx context.Context, models ...*mapping.ModelStruct) error
}

// Runner is the interface for the services that are runnable.
type Runner interface {
	Run(ctx context.Context) error
}
