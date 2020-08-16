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
	migratorModels := map[repository.Migrator][]*mapping.ModelStruct{}
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
		migrator, ok := repo.(repository.Migrator)
		if !ok {
			return errors.WrapDetf(repository.ErrNotImplements,
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
	log.Debug3f("Registering '%d' models", len(models))
	start := time.Now()
	// Register models.
	if err = c.ModelMap.RegisterModels(models...); err != nil {
		return err
	}

	log.Debug3f("Models registered in: %s", time.Since(start))
	return nil
}

// MapRepositoryModels maps models to their repositories. If the model doesn't have repository name mapped
// then the controller would match default repository.
func (c *Controller) MapRepositoryModels(r repository.Repository, models ...mapping.Model) (err error) {
	if _, ok := c.Repositories[r.ID()]; !ok {
		c.Repositories[r.ID()] = r
	}
	if len(models) == 0 {
		return nil
	}

	modelStructs := make([]*mapping.ModelStruct, len(models))
	for i, model := range models {
		mStruct, err := c.getModelStruct(model)
		if err != nil {
			return err
		}
		c.ModelRepositories[mStruct] = r
		modelStructs[i] = mStruct
	}

	return c.registerRepositoryModels(r, modelStructs...)
}

func (c *Controller) registerRepositoryModels(r repository.Repository, modelStructs ...*mapping.ModelStruct) error {
	if registrar, isRegistrar := r.(repository.ModelRegistrar); isRegistrar {
		if err := registrar.RegisterModels(modelStructs...); err != nil {
			return err
		}
	}
	return nil
}

// defaultRepositoryModels gest default repository models.
func (c *Controller) defaultRepositoryModels() []*mapping.ModelStruct {
	if c.DefaultRepository == nil {
		return nil
	}
	var defaultModels []*mapping.ModelStruct
	for _, model := range c.ModelMap.Models() {
		_, ok := c.ModelRepositories[model]
		if !ok {
			defaultModels = append(defaultModels, model)
		}
	}
	return defaultModels
}

func (c *Controller) getModelStruct(model mapping.Model) (*mapping.ModelStruct, error) {
	if model == nil {
		return nil, errors.WrapDet(mapping.ErrModelDefinition, "provided nil model value")
	}
	mStruct, ok := c.ModelMap.GetModelStruct(model)
	if !ok {
		return nil, errors.Wrapf(mapping.ErrModelNotFound, "provided model: '%T' is not found within given controller", model)
	}
	return mStruct, nil
}
