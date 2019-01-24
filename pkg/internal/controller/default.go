package controller

import (
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/uni-logger"
)

// DefaultConfig is the controller default config used with the Default function
var DefaultConfig *config.ControllerConfig = &config.ControllerConfig{
	NamingConvention: "snake",
	DefaultSchema:    "api",
	Builder: &config.BuilderConfig{
		FilterValueLimit:   30,
		ErrorLimits:        5,
		IncludeNestedLimit: 1,
	},
}

func DefaultTesting() *Controller {
	c := Default()

	log.Default()

	if internal.Verbose != nil && *internal.Verbose {
		c.Config.Debug = true

		log.SetLevel(unilogger.DEBUG)
	}

	return c
}
