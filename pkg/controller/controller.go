package controller

import (
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/db-manager"
	"github.com/kucjac/jsonapi/pkg/internal/controller"
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

func new(cfg *config.ControllerConfig, logger ...unilogger.LeveledLogger) (*controller.Controller, error) {
	var l unilogger.LeveledLogger

	if len(logger) == 1 {
		l = logger[0]
	}

	return controller.New(cfg, l)
}
