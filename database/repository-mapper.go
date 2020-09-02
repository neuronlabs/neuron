package database

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// RepositoryMapper is the database repository mapping structure.
type RepositoryMapper struct {
	// Repositories are the controller registered repositories.
	Repositories map[string]repository.Repository
	// DefaultRepository is the default repository for given controller.
	DefaultRepository repository.Repository
	// ModelRepositories is the mapping between the model and related repositories.
	ModelRepositories map[*mapping.ModelStruct]repository.Repository
	// ModelMap contains the information about the models.
	ModelMap *mapping.ModelMap
}

// GetRepository gets the repository for the provided model.
func (r *RepositoryMapper) GetRepository(model mapping.Model) (repository.Repository, error) {
	mStruct, err := r.ModelMap.ModelStruct(model)
	if err != nil {
		return nil, err
	}
	return r.GetRepositoryByModelStruct(mStruct)
}

// GetRepositoryByModelStruct gets the service by model struct.
func (r *RepositoryMapper) GetRepositoryByModelStruct(mStruct *mapping.ModelStruct) (repository.Repository, error) {
	repo, ok := r.ModelRepositories[mStruct]
	if !ok {
		if r.DefaultRepository == nil {
			return nil, errors.Wrapf(ErrRepositoryNotFound, "service not found for the model: %s", mStruct)
		}
		repo = r.DefaultRepository
	}
	return repo, nil
}

// RegisterRepositories registers provided repositories.
func (r *RepositoryMapper) RegisterRepositories(repositories ...repository.Repository) {
	for _, repo := range repositories {
		_, ok := r.Repositories[repo.ID()]
		if ok {
			continue
		}
		r.Repositories[repo.ID()] = repo
	}
}

// SetUnmappedModelRepositories checks if all models have their repositories mapped.
func (r *RepositoryMapper) SetUnmappedModelRepositories() error {
	var nonRepositoryModels []*mapping.ModelStruct
	for _, model := range r.ModelMap.Models() {
		_, ok := r.ModelRepositories[model]
		if !ok {
			nonRepositoryModels = append(nonRepositoryModels, model)
			if r.DefaultRepository != nil {
				r.ModelRepositories[model] = r.DefaultRepository
			}
		}
	}
	if r.DefaultRepository == nil && len(nonRepositoryModels) > 0 {
		return errors.WrapDetf(ErrRepositoryNotFound, "no repositories found for the models: %v", nonRepositoryModels)
	}
	return nil
}

// RegisterRepositoryModels registers models in the repositories that implements repository.ModelRegistrar.
func (r *RepositoryMapper) RegisterRepositoryModels() error {
	repositories := map[repository.Repository][]*mapping.ModelStruct{}
	for _, model := range r.ModelMap.Models() {
		r, err := r.GetRepositoryByModelStruct(model)
		if err != nil {
			return err
		}
		repositories[r] = append(repositories[r], model)
	}
	for repo, models := range repositories {
		if registrar, isRegistrar := repo.(repository.ModelRegistrar); isRegistrar {
			if err := registrar.RegisterModels(models...); err != nil {
				return err
			}
		}
	}
	return nil
}
