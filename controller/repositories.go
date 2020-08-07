package controller

import (
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// SetDefaultRepository sets the default repository for given controller.
func (c *Controller) SetDefaultRepository(r repository.Repository) error {
	if c.DefaultRepository != nil {
		return errors.NewDetf(ClassRepositoryAlreadyRegistered, "default repository already set to: %T", c.DefaultRepository)
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
			return nil, errors.Newf(ClassRepositoryNotFound, "service not found for the model: %s", mStruct)
		}
		repo = c.DefaultRepository
	}
	return repo, nil
}

// RegisterService registers provided service for given 'name' and with with given 'cfg' config.
func (c *Controller) RegisterRepositories(repositories ...repository.Repository) {
	for _, repo := range repositories {
		_, ok := c.Repositories[repo.ID()]
		if ok {
			continue
		}
		c.Repositories[repo.ID()] = repo
	}
}

func (c *Controller) VerifyModelRepositories() error {
	if c.DefaultRepository != nil {
		return nil
	}
	var nonRepositoryModels string
	for _, model := range c.ModelMap.Models() {
		_, ok := c.ModelRepositories[model]
		if !ok {
			if nonRepositoryModels != "" {
				nonRepositoryModels += " "
			}
			nonRepositoryModels += model.String()
		}
	}
	if nonRepositoryModels != "" {
		return errors.NewDetf(ClassRepositoryNotFound, "no repositories found for the models: %s", nonRepositoryModels)
	}
	return nil
}
