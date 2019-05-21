package repository

import (
	"errors"
	"fmt"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
)

var (
	ctr = newContainer()
)

// RegisterFactory registers provided Factory within the container
func RegisterFactory(f Factory) error {
	return ctr.registerFactory(f)
}

// GetFactory gets the repository factory
func GetFactory(name string) Factory {
	f, ok := ctr.factories[name]
	if ok {
		return f
	}
	return nil
}

// GetRepository gets the repository instance for the provided model
func GetRepository(structer ModelStructer, model *mapping.ModelStruct) (Repository, error) {
	repo, ok := ctr.models[model]
	if !ok {
		if err := ctr.mapModel(structer, model); err != nil {
			return nil, err
		}
		repo = ctr.models[model]
	}

	return repo, nil

}

// container is the container for the model repositories
// It contains mapping between repository name as well as the repository mapped to
// the given ModelStruct
type container struct {
	factories map[string]Factory

	// models are the mapping for the model's defined
	models map[*mapping.ModelStruct]Repository

	defaultFactory Factory
}

// newContainer creates the repository container
func newContainer() *container {
	r := &container{
		factories: map[string]Factory{},
		models:    map[*mapping.ModelStruct]Repository{},
	}

	return r
}

// registerFactory registers the given factory within the container
func (c *container) registerFactory(f Factory) error {
	repoName := f.RepositoryName()

	_, ok := c.factories[repoName]
	if ok {
		log.Debugf("Repository already registered: %s", repoName)
		log.Debugf("Factories: %v", c.factories)
		return fmt.Errorf("Factory already registered: %s", repoName)
	}

	c.factories[repoName] = f

	log.Debugf("Repository Factory: '%s' registered succesfully.", repoName)
	return nil
}

// MapModel maps the model with the provided repositories
func (c *container) mapModel(structer ModelStructer, model *mapping.ModelStruct) error {
	repoName := (*models.ModelStruct)(model).Config().Repository.DriverName

	var factory Factory

	if repoName == "" {
		return errors.New("No default repository factory found")
	}

	factory = c.factories[repoName]
	if factory == nil {
		err := fmt.Errorf("Repository Factory: '%s' is not found.", repoName)
		log.Debug(err)
		return err
	}

	repo, err := factory.New(structer, model)
	if err != nil {
		return err
	}

	c.models[model] = repo

	log.Debugf("Model %s mapped, to repository: %s", model.Collection(), repoName)
	return nil
}
