package controller

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// ListModels returns a list of registered models for given controller.
func (c *Controller) ListModels() []*mapping.ModelStruct {
	return c.ModelMap.Models()
}

// ModelStruct gets the model struct on the base of the provided model
func (c *Controller) ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	return c.getModelStruct(model)
}

// MustModelStruct gets the model struct from the cached model Map.
// Panics if the model does not exists in the map.
func (c *Controller) MustModelStruct(model interface{}) *mapping.ModelStruct {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		panic(err)
	}
	return mStruct
}

// MigrateModels updates or creates provided models representation in their related repositories.
// A representation of model might be a database table, collection etc.
// Model's repository must implement repository.Migrator.
func (c *Controller) MigrateModels(ctx context.Context, models ...interface{}) error {
	migratorModels := map[repository.Migrator][]*mapping.ModelStruct{}
	// map models to their repositories.
	for _, model := range models {
		modelStruct, err := c.ModelStruct(model)
		if err != nil {
			return err
		}
		repo, err := c.GetRepository(modelStruct)
		if err != nil {
			return err
		}
		migrator, ok := repo.(repository.Migrator)
		if !ok {
			return errors.Newf(repository.ClassNotImplements,
				"models: '%s' repository doesn't not allow to Migrate", modelStruct.Type().Name())
		}
		migratorModels[migrator] = append(migratorModels[migrator], modelStruct)
	}

	// migrate models structures in their migrator repositories.
	for migrator, modelsStructures := range migratorModels {
		if err := migrator.MigrateModels(ctx, modelsStructures...); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

// RegisterModels registers provided models within the context of the provided Controller.
// All repositories must be registered up to this moment.
func (c *Controller) RegisterModels(models ...interface{}) (err error) {
	log.Debug2f("Registering '%d' models", len(models))
	start := time.Now()
	// register models
	if err = c.ModelMap.RegisterModels(models...); err != nil {
		return err
	}
	// map models to their repositories.
	if err = c.mapAndRegisterInRepositories(); err != nil {
		return err
	}
	log.Debug2f("Models registered in: %s", time.Since(start))
	return nil
}

func (c *Controller) mapAndRegisterInRepositories(models ...interface{}) error {
	// map models to their repositories in the config
	if err := c.Config.MapModelsRepositories(); err != nil {
		return err
	}

	// match models to their repository instances.
	modelsRepositories := make(map[repository.Repository][]*mapping.ModelStruct)
	for _, model := range models {
		mStruct, err := c.getModelStruct(model)
		if err != nil {
			return err
		}
		repo, ok := c.Repositories[mStruct.Config().RepositoryName]
		if !ok {
			return errors.NewDetf(ClassRepositoryNotFound, "repository not found for the model: '%s'", mStruct.String())
		}
		modelsRepositories[repo] = append(modelsRepositories[repo], mStruct)
	}
	for repo, modelsStructures := range modelsRepositories {
		if err := repo.RegisterModels(modelsStructures...); err != nil {
			log.Errorf("Registering models in repository: %v failed: %v", repo.FactoryName(), modelsStructures)
			return err
		}
	}
	return nil
}

func (c *Controller) getModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	if model == nil {
		return nil, errors.NewDet(ClassInvalidModel, "provided nil model value")
	}

	switch tp := model.(type) {
	case *mapping.ModelStruct:
		return tp, nil
	case string:
		m := c.ModelMap.GetByCollection(tp)
		if m == nil {
			return nil, errors.NewDetf(mapping.ClassModelNotFound, "model: '%s' is not found", tp)
		}
		return m, nil
	default:
		mStruct, err := c.ModelMap.GetModelStruct(model)
		if err != nil {
			return nil, err
		}
		return mStruct, nil
	}
}
