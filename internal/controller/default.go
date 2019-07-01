package controller

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/config"
)

// DefaultConfig is the controller default config used with the Default function
var (
	DefaultTestingConfig *config.Controller
)

func init() {
	DefaultTestingConfig = config.ReadDefaultControllerConfig()
	DefaultTestingConfig.Repositories = map[string]*config.Repository{
		"mock": &config.Repository{DriverName: "mockery"},
	}

	DefaultTestingConfig.DefaultRepositoryName = "mock"
}

// DefaultTesting is the default controller used for testing
func DefaultTesting(t testing.TB, cfg *config.Controller) *Controller {
	if cfg == nil {
		cfg = DefaultTestingConfig
	}
	if testing.Verbose() {
		cfg.LogLevel = "debug3"
	}

	c, err := newController(cfg)
	require.NoError(t, err)

	return c
}

// NewDefault creates new default controller based on the default config
func NewDefault() *Controller {
	c, err := newController(config.ReadDefaultControllerConfig())
	if err != nil {
		panic(err)
	}
	return c
}
