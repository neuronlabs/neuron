package controller

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"
)

// defaultController is the DefaultController controller used if no 'controller' is provided for operations
var defaultController *Controller

// Default returns current default controller.
func Default() *Controller {
	if defaultController == nil {
		defaultController = newDefault()
	}
	return defaultController
}

// New creates new controller for given config 'cfg'.
func New(cfg *config.Controller) (err error) {
	defaultController, err = newController(cfg)
	return err
}

// MustNew creates new controller for given config 'cfg'. On error panics.
func MustNew(cfg *config.Controller) {
	var err error
	defaultController, err = newController(cfg)
	if err != nil {
		panic(err)
	}
}

// NewDefault creates new default controller based on the default config.
func NewDefault() *Controller {
	c, err := newController(config.DefaultController())
	if err != nil {
		panic(err)
	}
	return c
}

//
// DefaultController Controller methods
//

// CloseAll gently closes repository connections.
func CloseAll(ctx context.Context) error {
	return Default().CloseAll(ctx)
}

// DialAll establish connections for all repositories.
func DialAll(ctx context.Context) error {
	return Default().DialAll(ctx)
}

// GetRepository gets 'model' repository.
func GetRepository(model interface{}) (repository.Repository, error) {
	return Default().GetRepository(model)
}

// HealthCheck checks all repositories health.
func HealthCheck(ctx context.Context) (*repository.HealthResponse, error) {
	return Default().HealthCheck(ctx)
}

// MigrateModels updates or creates provided models representation in their related repositories.
// A representation of model might be a database table, collection etc.
// Model's repository must implement repository.Migrator.
func MigrateModels(ctx context.Context, models ...interface{}) error {
	return Default().MigrateModels(ctx, models...)
}

// ModelStruct gets the model struct on the base of the provided model
func ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	return Default().getModelStruct(model)
}

// MustModelStruct gets the model struct from the cached model Map.
// Panics if the model does not exists in the map.
func MustModelStruct(model interface{}) *mapping.ModelStruct {
	return Default().MustModelStruct(model)
}

// Now gets and returns current timestamp. If the configs specify the function might return UTC timestamp.
func Now() time.Time {
	ts := time.Now()
	if Default().Config.UTCTimestamps {
		ts = ts.UTC()
	}
	return ts
}

// RegisterModels registers provided models within the context of the provided Controller.
// All repositories must be registered up to this moment.
func RegisterModels(models ...interface{}) (err error) {
	return Default().RegisterModels(models...)
}

// RegisterRepository registers provided repository for given 'name' and with with given 'cfg' config.
func RegisterRepository(name string, cfg *config.Repository) (err error) {
	return Default().RegisterRepository(name, cfg)
}

func newDefault() *Controller {
	c, err := newController(config.DefaultController())
	if err != nil {
		panic(err)
	}
	return c
}
