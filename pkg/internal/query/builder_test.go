package query

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// DefaultBuilderWithModels returns builder with provided models
func DefaultBuilderWithModels(t *testing.T, models ...interface{}) *Builder {
	t.Helper()
	b := DefaultBuilder()

	err := b.schemas.RegisterModels(models...)
	require.NoError(t, err)

	return b
}
