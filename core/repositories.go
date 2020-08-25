package core

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// SetDefaultRepository sets the default repository for given controller.
func (c *Controller) SetDefaultRepository(r repository.Repository) error {
	if c.DefaultRepository != nil {
		return errors.WrapDetf(ErrRepositoryAlreadyRegistered, "default repository already set to: %T", c.DefaultRepository)
	}
	if _, ok := c.Repositories[r.ID()]; !ok {
		c.Repositories[r.ID()] = r
	}
	c.DefaultRepository = r
	return nil
}

// GetRepository gets the repository for the provided model.
func (c *Controller) GetRepository(model mapping.Model) (repository.Repository, error) {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		return nil, err
	}
	return c.GetRepositoryByModelStruct(mStruct)
}

// GetRepositoryByModelStruct gets the service by model struct.
func (c *Controller) GetRepositoryByModelStruct(mStruct *mapping.ModelStruct) (repository.Repository, error) {
	repo, ok := c.ModelRepositories[mStruct]
	if !ok {
		if c.DefaultRepository == nil {
			return nil, errors.Wrapf(ErrRepositoryNotFound, "service not found for the model: %s", mStruct)
		}
		repo = c.DefaultRepository
	}
	return repo, nil
}

// RegisterRepositories registers provided repositories.
func (c *Controller) RegisterRepositories(repositories ...repository.Repository) {
	for _, repo := range repositories {
		_, ok := c.Repositories[repo.ID()]
		if ok {
			continue
		}
		c.Repositories[repo.ID()] = repo
	}
}

// SetUnmappedModelRepositories checks if all models have their repositories mapped.
func (c *Controller) SetUnmappedModelRepositories() error {
	var nonRepositoryModels []*mapping.ModelStruct
	for _, model := range c.ModelMap.Models() {
		_, ok := c.ModelRepositories[model]
		if !ok {
			nonRepositoryModels = append(nonRepositoryModels, model)
			if c.DefaultRepository != nil {
				c.ModelRepositories[model] = c.DefaultRepository
			}
		}
	}
	if c.DefaultRepository == nil && len(nonRepositoryModels) > 0 {
		return errors.WrapDetf(ErrRepositoryNotFound, "no repositories found for the models: %v", nonRepositoryModels)
	}
	return nil
}

// RegisterRepositoryModels registers models in the repositories that implements repository.ModelRegistrar.
func (c *Controller) RegisterRepositoryModels() error {
	repositories := map[repository.Repository][]*mapping.ModelStruct{}
	for _, model := range c.ModelMap.Models() {
		r, err := c.GetRepositoryByModelStruct(model)
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
