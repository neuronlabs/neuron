package controller

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repositories/mocks"
	"testing"
)

// DefaultConfig is the controller default config used with the Default function
var DefaultConfig *config.ControllerConfig = config.ReadDefaultControllerConfig()

// DefaultTesting is the default controller used for testing
func DefaultTesting(t *testing.T) *Controller {
	c := NewDefault()

	log.Default()

	if internal.Verbose != nil && *internal.Verbose {
		c.Config.Debug = true

		log.SetLevel(log.LDEBUG)
	}
	c.RegisterRepository(&mocks.Repository{})
	return c
}
