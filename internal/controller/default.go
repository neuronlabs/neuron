package controller

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/stretchr/testify/require"
	"testing"
)

// DefaultConfig is the controller default config used with the Default function
var (
	DefaultConfig        *config.ControllerConfig = config.ReadDefaultControllerConfig()
	DefaultTestingConfig *config.ControllerConfig
)

func init() {
	DefaultTestingConfig = config.ReadDefaultControllerConfig()
	DefaultTestingConfig.Repositories = map[string]*config.Repository{
		"mock": &config.Repository{DriverName: "mockery"},
	}

	DefaultTestingConfig.DefaultRepositoryName = "mock"
}

// DefaultTesting is the default controller used for testing
func DefaultTesting(t testing.TB, cfg *config.ControllerConfig) *Controller {
	if cfg == nil {
		cfg = DefaultTestingConfig
	}
	if testing.Verbose() {
		cfg.Debug = true
	}

	c, err := newController(cfg)
	require.NoError(t, err)

	return c
}

// NewDefault creates new default controller based on the default config
func NewDefault() *Controller {
	c, err := newController(DefaultConfig)
	if err != nil {
		panic(err)
	}
	return c
}
