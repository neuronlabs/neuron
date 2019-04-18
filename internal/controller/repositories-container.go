package controller

import (
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/namer"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
)

var (
	ErrRepoAlreadyRegistered error = errors.New("Repository already registered.")
	ErrNoRepoForModel        error = errors.New("Model have no repository name")
)

// RepositoryContainer is the container for the model repositories
// It contains mapping between repository name as well as the repository mapped to
// the given ModelStruct
type RepositoryContainer struct {
	// defaultRepository
	defaultRepositoryName string

	// names are the mapping for the repository interfaces for given string name
	names map[string]interface{}

	// models are the mapping for the model's defined
	models map[*models.ModelStruct]string
}

// NewRepoContainer creates the repository container
func NewRepoContainer() *RepositoryContainer {
	r := &RepositoryContainer{
		names:  map[string]interface{}{},
		models: map[*models.ModelStruct]string{},
	}

	return r
}

// MapModel maps the model with the provided repositories
func (r *RepositoryContainer) MapModel(model *models.ModelStruct) error {
	repoName := model.RepositoryName()

	if repoName == "" {
		return ErrNoRepoForModel
	}

	r.models[model] = repoName

	log.Debugf("Model %s mapped, to repository: %s", model.Collection(), repoName)
	return nil
}

// RegisterRepository registers the repostiory by it's name
func (r *RepositoryContainer) RegisterRepository(repo namer.RepositoryNamer) error {
	// if names are zero
	// get the default repository name

	repoName := repo.RepositoryName()

	_, ok := r.names[repoName]
	if ok {
		return ErrRepoAlreadyRegistered
	}

	r.names[repoName] = repo

	log.Debugf("Repository: '%s' registered succesfully.", repoName)
	return nil
}

// RepositoryByName returns repository by it's name
func (r *RepositoryContainer) RepositoryByName(name string) (interface{}, bool) {
	repo, ok := r.names[name]
	return repo, ok
}

// RepositoryByModel returns repository by the provided ModelsStruct
func (r *RepositoryContainer) RepositoryByModel(model *models.ModelStruct) (interface{}, bool) {
	repoName, ok := r.models[model]
	if !ok {
		return nil, false
	}
	repo, ok := r.names[repoName]
	return repo, ok
}
