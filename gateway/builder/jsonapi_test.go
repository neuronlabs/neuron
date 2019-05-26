package builder

import (
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	ictrl "github.com/neuronlabs/neuron/internal/controller"
	_ "github.com/neuronlabs/neuron/query"
	_ "github.com/neuronlabs/neuron/query/mocks"
	"github.com/stretchr/testify/require"

	"testing"
)

// DefaultJSONAPI returns builder with default config and no i18n support
func DefaultJSONAPI(t *testing.T, c *ictrl.Controller) *JSONAPI {
	gw := config.ReadDefaultGatewayConfig()
	return NewJSONAPI((*controller.Controller)(c), gw.QueryBuilder, nil)
}

// DefaultJSONAPIWithModels returns builder with provided models
func DefaultJSONAPIWithModels(t *testing.T, models ...interface{}) *JSONAPI {
	t.Helper()

	c := ictrl.DefaultTesting(t, nil)

	err := c.RegisterModels(models...)
	require.NoError(t, err)

	b := DefaultJSONAPI(t, c)

	return b
}
