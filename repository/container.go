package repository

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"

	"github.com/neuronlabs/neuron/internal/models"
)

var ctr = newContainer()

// RegisterFactory registers provided Factory within the container.
func RegisterFactory(f Factory) error {
	return ctr.registerFactory(f)
}

// GetFactory gets the repository factory with given 'name'.
func GetFactory(name string) Factory {
	f, ok := ctr.factories[name]
	if ok {
		return f
	}
	return nil
}

// GetRepository gets the repository instance for the provided model.
func GetRepository(structer ModelStructer, model interface{}) (Repository, error) {
	mstruct, ok := model.(*mapping.ModelStruct)
	if !ok {
		var err error
		mstruct, err = structer.ModelStruct(model)
		if err != nil {
			return nil, err
		}
	}

	repo, ok := ctr.models[mstruct]
	if !ok {
		if err := ctr.mapModel(structer, mstruct); err != nil {
			return nil, err
		}
		repo = ctr.models[mstruct]
	}

	return repo, nil

}

// container is the container for the model repositories.
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
		return errors.Newf(class.RepositoryFactoryAlreadyRegistered, "factory: '%s' already registered", repoName)
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
		return errors.New(class.ModelSchemaNotFound, "no default repository factory found")
	}

	factory = c.factories[repoName]
	if factory == nil {
		err := errors.Newf(class.RepositoryFactoryNotFound, "repository factory: '%s' not found.", repoName)
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

// CloseAll closes all repositories
func CloseAll(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan interface{}, len(ctr.factories))
	for _, factory := range ctr.factories {
		go factory.Close(ctx, done)
	}
	var ct int
fl:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v := <-done:
			if err, ok := v.(error); ok {
				return err
			}
			ct++
			if ct == len(ctr.factories) {
				break fl
			}
		}
	}
	return nil
}
