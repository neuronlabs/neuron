package controller

import (
	"context"
	"time"

	"github.com/neuronlabs/uni-logger"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal/controller"
)

// new -> config ( set repositories -> set default if any exists )  -> register repositories -> register models (if no default repository found error)

// DefaultController is the Default controller used if no 'controller' is provided for operations
var DefaultController *Controller

// Default returns current default controller.
func Default() *Controller {
	if DefaultController == nil {
		DefaultController = (*Controller)(controller.Default())
	}
	return DefaultController
}

// NewDefault creates and returns new default Controller.
func NewDefault() *Controller {
	return (*Controller)(controller.NewDefault())
}

// SetDefault sets given Controller 'c' as the default.
func SetDefault(c *Controller) {
	controller.SetDefault(c.internal())
}

// Controller is the structure that controls whole jsonapi behavior.
// It contains repositories, model definitions, query builders and it's own config.
type Controller controller.Controller

// MustGetNew creates new controller for given config 'cfg' and 'logger' (optionally).
// Panics on error.
func MustGetNew(cfg *config.Controller, logger ...unilogger.LeveledLogger) *Controller {
	c, err := new(cfg, logger...)
	if err != nil {
		panic(err)
	}

	return (*Controller)(c)

}

// New creates new controller for given config 'cfg' and 'logger' (optionally).
func New(cfg *config.Controller, logger ...unilogger.LeveledLogger) (*Controller, error) {
	c, err := new(cfg, logger...)
	if err != nil {
		return nil, err
	}

	return (*Controller)(c), nil
}

// GetRepository gets the repository for the provided model.
func (c *Controller) GetRepository(model interface{}) (repository.Repository, error) {
	return c.internal().GetRepository(model)
}

// ModelStruct gets the model struct on the base of the provided model
func (c *Controller) ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	m, err := c.internal().GetModelStruct(model)
	if err != nil {
		return nil, err
	}
	return (*mapping.ModelStruct)(m), nil
}

// RegisterModels registers provided models within the context of the provided Controller
func (c *Controller) RegisterModels(models ...interface{}) error {
	return c.internal().RegisterModels(models...)
}

// RegisterRepository registers provided repository for given 'name' and with with given 'cfg' config.
func (c *Controller) RegisterRepository(name string, cfg *config.Repository) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	return c.internal().RegisterRepository(name, cfg)
}

// Close closes all repository instances.
func (c *Controller) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	repository.CloseAll(ctx)

	return nil
}

func new(cfg *config.Controller, logger ...unilogger.LeveledLogger) (*controller.Controller, error) {
	var l unilogger.LeveledLogger

	if len(logger) == 1 {
		l = logger[0]
	}

	return controller.New(cfg, l)
}

func (c *Controller) internal() *controller.Controller {
	return (*controller.Controller)(c)
}
