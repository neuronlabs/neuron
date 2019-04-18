package controller

import (
	"fmt"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/db-manager"
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/repositories"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/kucjac/uni-logger"
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

// DBManager returns the database error manager
func (c *Controller) DBManager() *dbmanager.ErrorManager {
	return (*controller.Controller)(c).DBManager()
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

// RegisterModels registers provided models within the context of the provided Controller
func (c *Controller) RegisterModels(models ...interface{}) error {
	return (*controller.Controller)(c).RegisterModels(models...)
}

// RegisterRepositories registers provided repositories
func (c *Controller) RegisterRepositories(repos ...interface{}) error {
	for _, repo := range repos {
		r, ok := repo.(repositories.Repository)
		if !ok {
			log.Errorf("Cannot register repository: %T. It doesn't implement repository interface.", repo)
			return repositories.ErrNewNotRepository
		}
		if err := (*controller.Controller)(c).RegisterRepository(r); err != nil {
			log.Debugf("Registering Repository: '%s' failed. %v", r.RepositoryName(), err)
			return err
		}
		log.Debugf("Repository: '%s' registered succesfully.", r.RepositoryName())
	}
	return nil
}

// SetDefaultRepository sets the default repository.
// By default the first registered repository is set to default.
// This method allows to change the behaviour
func (c *Controller) SetDefaultRepository(repo interface{}) error {
	r, ok := repo.(repositories.Repository)
	if !ok {
		return fmt.Errorf("Provided 'repo' struct is not a Repository: %T", repo)
	}

	(*controller.Controller)(c).SetDefaultRepository(r)
	return nil
}

// ModelStruct gets the model struct on the base of the provided model
func (c *Controller) ModelStruct(model interface{}) (*mapping.ModelStruct, error) {
	m, err := (*controller.Controller)(c).GetModelStruct(model)
	if err != nil {
		return nil, err
	}
	return (*mapping.ModelStruct)(m), nil
}

func new(cfg *config.ControllerConfig, logger ...unilogger.LeveledLogger) (*controller.Controller, error) {
	var l unilogger.LeveledLogger

	if len(logger) == 1 {
		l = logger[0]
	}

	return controller.New(cfg, l)
}
