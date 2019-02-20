package controller

import (
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/db-manager"
	"github.com/kucjac/jsonapi/pkg/internal/controller"
	"github.com/kucjac/jsonapi/pkg/internal/repositories"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/uni-logger"
)

var DefaultController *Controller

// Default returns the default controller
func Default() *Controller {
	if DefaultController == nil {
		DefaultController = (*Controller)(controller.Default())
	}
	return DefaultController
}

// SetDefault sets the default Controller to the provided
func SetDefault(c *Controller) {
	d := Default()

	*d = *c
}

type Controller controller.Controller

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

func new(cfg *config.ControllerConfig, logger ...unilogger.LeveledLogger) (*controller.Controller, error) {
	var l unilogger.LeveledLogger

	if len(logger) == 1 {
		l = logger[0]
	}

	return controller.New(cfg, l)
}
