package controller

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/service"
)

// ListModels returns a list of registered models for given controller.
func (c *Controller) ListModels() []*mapping.ModelStruct {
	return c.ModelMap.Models()
}

// ModelStruct gets the model struct on the base of the provided model
func (c *Controller) ModelStruct(model mapping.Model) (*mapping.ModelStruct, error) {
	return c.getModelStruct(model)
}

// MustModelStruct gets the model struct from the cached model Map.
// Panics if the model does not exists in the map.
func (c *Controller) MustModelStruct(model mapping.Model) *mapping.ModelStruct {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		panic(err)
	}
	return mStruct
}

// MigrateModels updates or creates provided models representation in their related repositories.
// A representation of model might be a database table, collection etc.
// Model's repository must implement repository.Migrator.
func (c *Controller) MigrateModels(ctx context.Context, models ...mapping.Model) error {
	migratorModels := map[service.Migrator][]*mapping.ModelStruct{}
	// map models to their repositories.
	for _, model := range models {
		modelStruct, err := c.ModelStruct(model)
		if err != nil {
			return err
		}
		repo, err := c.GetRepository(model)
		if err != nil {
			return err
		}
		migrator, ok := repo.(service.Migrator)
		if !ok {
			return errors.Newf(service.ClassNotImplements,
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
func (c *Controller) RegisterModels(models ...mapping.Model) (err error) {
	log.Debug2f("Registering '%d' models", len(models))
	start := time.Now()
	// Register models.
	if err = c.ModelMap.RegisterModels(models...); err != nil {
		return err
	}

	log.Debug2f("Models registered in: %s", time.Since(start))
	return nil
}

// MapModelsRepositories maps models to their repositories. If the model doesn't have repository name mapped
// then the controller would match default repository.
func (c *Controller) MapModelsRepositories(models ...mapping.Model) (err error) {
	// match models to their repository instances.
	modelsRepositories := make(map[string][]*mapping.ModelStruct)
	for _, model := range models {
		mStruct, err := c.getModelStruct(model)
		if err != nil {
			return err
		}
		if mStruct.RepositoryName == "" {
			if c.Config.DisallowDefaultRepository {
				return errors.Newf(ClassRepositoryNotMatched, "no repository set for model: '%s'", mStruct.String())
			}
			mStruct.RepositoryName = c.defaultRepository
			modelsRepositories[c.defaultRepository] = append(modelsRepositories[c.defaultRepository], mStruct)
			continue
		}
		_, ok := c.Services[mStruct.RepositoryName]
		if !ok {
			return errors.NewDetf(ClassRepositoryNotFound, "repository not found for the model: '%s'", mStruct.String())
		}
		modelsRepositories[mStruct.RepositoryName] = append(modelsRepositories[mStruct.RepositoryName], mStruct)
	}
	for repoName, modelsStructures := range modelsRepositories {
		repo := c.Repositories[repoName]
		registrar, ok := repo.(service.ModelRegistrar)
		if !ok {
			continue
		}
		if err := registrar.RegisterModels(modelsStructures...); err != nil {
			log.Errorf("Registering models in repository failed: %v", modelsStructures)
			return err
		}
	}
	return nil
}

func (c *Controller) getModelStruct(model mapping.Model) (*mapping.ModelStruct, error) {
	if model == nil {
		return nil, errors.NewDet(ClassInvalidModel, "provided nil model value")
	}
	mStruct, ok := c.ModelMap.GetModelStruct(model)
	if !ok {
		return nil, errors.Newf(mapping.ClassModelNotFound, "provided model: '%T' is not found within given controller", model)
	}
	return mStruct, nil
}
