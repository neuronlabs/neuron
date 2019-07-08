package repository

import (
	"context"
)

// Repository is the interface that defines the base neuron Repository.
type Repository interface {
	RepositoryName() string
	Close(ctx context.Context) error
}
