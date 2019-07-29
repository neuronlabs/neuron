package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFilterOperators tests the filter operators registration process.
func TestFilterOperators(t *testing.T) {
	assert.True(t, OpEqual.isBasic())
	assert.False(t, OpEqual.isRangable())
	assert.False(t, OpEqual.isStringOnly())
	assert.True(t, OpContains.isStringOnly())
	assert.Equal(t, AnnotationOperatorGreaterThan, OpGreaterThan.Raw)
}
