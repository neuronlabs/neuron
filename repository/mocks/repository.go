package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/neuronlabs/neuron/repository"
)

var _ repository.Repository = &Repository{}

// Repository is the structure that implements repositories.Repository
type Repository struct {
	mock.Mock
}

// RepositoryName returns the repository name
// Implements repositories.RepositoryNamer interface
func (r *Repository) RepositoryName() string {
	return "mocks"
}

// Close closes the repository connection
func (r *Repository) Close(ctx context.Context) error {
	return nil
}
