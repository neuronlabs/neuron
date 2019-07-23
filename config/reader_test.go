package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadDefaultConfig test the read default config function.
func TestReadDefaultConfig(t *testing.T) {
	var c *Controller
	require.NotPanics(t, func() {
		c = ReadDefaultConfig()
	})
	require.NotNil(t, c)

	t.Run("Controller", func(t *testing.T) {
		assert.Equal(t, "snake", c.NamingConvention)
		assert.Equal(t, "create", c.CreateValidatorAlias)
		assert.Equal(t, "patch", c.PatchValidatorAlias)
	})
}
