package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/log"
)

func TestReadDefaultConfig(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG)
	}
	var c *Controller
	require.NotPanics(t, func() {
		c = ReadDefaultConfig()
	})
	require.NotNil(t, c)

	t.Run("Controller", func(t *testing.T) {
		assert.Equal(t, "snake", c.NamingConvention)
		assert.Equal(t, "create", c.CreateValidatorAlias)
		assert.Equal(t, "patch", c.PatchValidatorAlias)
		assert.Equal(t, "api", c.DefaultSchema)
	})
}
