package controller

import (
	"context"
	"time"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/mapping"
	"github.com/neuronlabs/neuron-core/repository"

	"github.com/neuronlabs/neuron-core/internal/controller"
)

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

// MustGetNew creates new controller for given provided 'cfg' config.
// Panics on error.
func MustGetNew(cfg *config.Controller) *Controller {
	c, err := new(cfg)
	if err != nil {
		panic(err)
	}

	return (*Controller)(c)

}

// New creates new controller for given config 'cfg'.
func New(cfg *config.Controller) (*Controller, error) {
	c, err := new(cfg)
	if err != nil {
		return nil, err
	}

	return (*Controller)(c), nil
}

// Close closes all repository instances.
func (c *Controller) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	return repository.CloseAll(ctx)
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

func new(cfg *config.Controller) (*controller.Controller, error) {
	return controller.New(cfg)
}

func (c *Controller) internal() *controller.Controller {
	return (*controller.Controller)(c)
}
