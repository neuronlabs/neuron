package controller

import (
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/uni-logger"
	"log"
	"os"
)

// DefaultConfig is the controller default config used with the Default function
var DefaultConfig *config.ControllerConfig = &config.ControllerConfig{
	NamingConvention:   "snake",
	DefaultSchema:      "api",
	QueryErrorLimits:   5,
	IncludeNestedLimit: 2,
	FilterValueLimit:   30,
}

func DefaultTesting() *Controller {
	c := Default()

	if internal.Verbose != nil && *internal.Verbose {
		c.Config.Debug = true

		c.logger = unilogger.NewBasicLogger(os.Stderr, "[PKG][CONTROLLER] ", log.Lshortfile)

		c.logger.(*unilogger.BasicLogger).SetLevel(unilogger.DEBUG)
	}

	return c
}
