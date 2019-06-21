package repository

import (
	"context"
)

// Repository is the interface that defines the base neuron Repository.
type Repository interface {
	Namer
	Close(ctx context.Context) error
}

// Namer is the interface that gets the repository name for the 'Namer'.
type Namer interface {
	RepositoryName() string
}
