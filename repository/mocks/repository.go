package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/neuronlabs/neuron-core/repository"
)

var _ repository.Repository = &Repository{}

// Repository is the structure that implements repositories.Repository
type Repository struct {
	mock.Mock
}

// FactoryName implements repository.Repository interface
func (r *Repository) FactoryName() string {
	return "mocks"
}

// Close closes the repository connection
func (r *Repository) Close(ctx context.Context) error {
	return nil
}
