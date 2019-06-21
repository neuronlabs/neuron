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

// DefaultController is the Default controller used if no 'controller' is provided for operations
var DefaultController *Controller

// Default returns the default controller
func Default() *Controller {
	if DefaultController == nil {
		DefaultController = (*Controller)(controller.Default())
	}
	return DefaultController
}

// NewDefault returns new default controller
func NewDefault() *Controller {
	return (*Controller)(controller.NewDefault())
}

// SetDefault sets the default Controller to the provided
func SetDefault(c *Controller) {
	controller.SetDefault(c.internal())
}

// Controller is the structure that controls whole jsonapi behavior.
// It contains repositories, model definitions, query builders and it's own config
type Controller controller.Controller

// MustGetNew gets the
func MustGetNew(cfg *config.Controller, logger ...unilogger.LeveledLogger) *Controller {
	c, err := new(cfg, logger...)
	if err != nil {
		panic(err)
	}

	return (*Controller)(c)

}

// New creates new controller for given config
func New(cfg *config.Controller, logger ...unilogger.LeveledLogger) (*Controller, error) {
	c, err := new(cfg, logger...)
	if err != nil {
		return nil, err
	}

	return (*Controller)(c), nil
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

// Schema gets the schema by it's name
func (c *Controller) Schema(schemaName string) (*mapping.Schema, bool) {
	s, ok := c.internal().ModelSchemas().Schema(schemaName)
	if ok {
		return (*mapping.Schema)(s), ok
	}
	return nil, ok
}

// Schemas gets the controller defined schemas
func (c *Controller) Schemas() (schemas []*mapping.Schema) {
	for _, s := range c.internal().ModelSchemas().Schemas() {
		schemas = append(schemas, (*mapping.Schema)(s))
	}
	return
}

// Close closes all repositories
func (c *Controller) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
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
