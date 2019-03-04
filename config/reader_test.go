package config

import (
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestReadDefaultConfig(t *testing.T) {
	if testing.Verbose() {
		log.SetLevel(unilogger.DEBUG)
	}
	var c *Config
	require.NotPanics(t, func() { c = ReadDefaultConfig() })
	require.NotNil(t, c)

	t.Run("Controller", func(t *testing.T) {
		assert.Equal(t, "snake", c.Controller.NamingConvention)
		assert.Equal(t, "create", c.Controller.CreateValidatorAlias)
		assert.Equal(t, "patch", c.Controller.PatchValidatorAlias)
		assert.Equal(t, "api", c.Controller.DefaultSchema)

		t.Run("Builder", func(t *testing.T) {
			assert.Equal(t, 5, c.Controller.Builder.ErrorLimits)
			assert.Equal(t, 3, c.Controller.Builder.IncludeNestedLimit)
			assert.Equal(t, 50, c.Controller.Builder.FilterValueLimit)
		})
		t.Run("Flags", func(t *testing.T) {

			f, err := c.Controller.Flags.Container()
			require.NoError(t, err)

			v, ok := f.Get(FlUseLinks)
			assert.True(t, ok)
			assert.True(t, v)

			v, ok = f.Get(FlUseFilterValueLimit)
			assert.True(t, ok)
			assert.True(t, v)

			v, ok = f.Get(FlReturnPatchContent)
			assert.True(t, ok)
			assert.True(t, v)
		})
	})

	t.Run("Gateway", func(t *testing.T) {
		assert.Equal(t, 8080, c.Gateway.Port)
		assert.Equal(t, time.Second*10, c.Gateway.ReadTimeout)
		assert.Equal(t, time.Second*5, c.Gateway.ReadHeaderTimeout)
		assert.Equal(t, time.Second*10, c.Gateway.WriteTimeout)
		assert.Equal(t, time.Second*120, c.Gateway.IdleTimeout)
		assert.Equal(t, time.Second*10, c.Gateway.ShutdownTimeout)

		t.Run("Router", func(t *testing.T) {
			assert.Equal(t, "v1", c.Gateway.Router.Prefix)
		})
	})

}
