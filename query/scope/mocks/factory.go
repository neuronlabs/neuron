package mocks

import (
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
	"github.com/stretchr/testify/mock"
)

const (
	repoName = "mockery"
)

func init() {
	repository.RegisterFactory(&Factory{})
}

// Factory is the repository.Factory mock implementation
type Factory struct {
	mock.Mock
}

// New creates new repository
// Implements repository.Factory method
func (f *Factory) New(s repository.ModelStructer, model *mapping.ModelStruct) (repository.Repository, error) {
	return &Repository{}, nil
}

// RepositoryName returns the factory repository name
// Implements repository.Repository
func (f *Factory) RepositoryName() string {
	return repoName
}
