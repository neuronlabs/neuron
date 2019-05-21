package mocks

import (
	"github.com/neuronlabs/neuron/repository"
	"github.com/stretchr/testify/mock"
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
