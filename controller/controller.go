package controller

import (
	"github.com/kucjac/uni-logger"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/mapping"
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
	d := Default()

	*d = *c
}

// Controller is the structure that controls whole jsonapi behavior.
// It contains repositories, model definitions, query builders and it's own config
type Controller controller.Controller

// DBErrorMapper returns the database error manager
func (c *Controller) DBErrorMapper() *errors.ErrorMapper {
	return (*controller.Controller)(c).DBErrorMapper()
}

// MustGetNew gets the
func MustGetNew(cfg *config.ControllerConfig, logger ...unilogger.LeveledLogger) *Controller {
	c, err := new(cfg, logger...)
	if err != nil {
		panic(err)
	}

	return (*Controller)(c)

}

// New creates new controller for given config
func New(cfg *config.ControllerConfig, logger ...unilogger.LeveledLogger) (*Controller, error) {
	c, err := new(cfg, logger...)
	if err != nil {
		return nil, err
	}

	return (*Controller)(c), nil
}

// ModelStruct gets the model struct on the base of the provided model
func (c *Controller) ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	m, err := (*controller.Controller)(c).GetModelStruct(model)
	if err != nil {
		return nil, err
	}
	return (*mapping.ModelStruct)(m), nil
}

// RegisterModels registers provided models within the context of the provided Controller
func (c *Controller) RegisterModels(models ...interface{}) error {
	return (*controller.Controller)(c).RegisterModels(models...)
}

// Schema gets the schema by it's name
func (c *Controller) Schema(schemaName string) (*mapping.Schema, bool) {
	s, ok := (*controller.Controller)(c).ModelSchemas().Schema(schemaName)
	if ok {
		return (*mapping.Schema)(s), ok
	}
	return nil, ok
}

// Schemas gets the controller defined schemas
func (c *Controller) Schemas() (schemas []*mapping.Schema) {
	for _, s := range (*controller.Controller)(c).ModelSchemas().Schemas() {
		schemas = append(schemas, (*mapping.Schema)(s))
	}
	return
}

func new(cfg *config.ControllerConfig, logger ...unilogger.LeveledLogger) (*controller.Controller, error) {
	var l unilogger.LeveledLogger

	if len(logger) == 1 {
		l = logger[0]
	}

	return controller.New(cfg, l)
}
