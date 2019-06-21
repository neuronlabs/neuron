package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neuronlabs/neuron/common"
)

// TestSplitBracketParameter tests the split bracket parameter function.
func TestSplitBracketParameter(t *testing.T) {
	type stringBool struct {
		Str string
		Val bool
	}

	values := []stringBool{
		{"[some][thing]", true},
		{"[no][closing", false},
		{"no][opening]", false},
		{"]justclosing", false},
		{"[doubleopen[]", false},
		{"[doubleclose]]", false},
	}

	var splitted []string
	var err error
	for _, v := range values {
		splitted, err = common.SplitBracketParameter(v.Str)
		if !v.Val {
			assert.Error(t, err)
			// t.Log(err)
			if err == nil {
				t.Log(v.Str)
			}
		} else {
			assert.Nil(t, err)
			assert.NotEmpty(t, splitted)
		}
	}
}

// TestFilterOperators tests the filter operators registration process.
func TestFilterOperators(t *testing.T) {
	container := NewOpContainer()
	container.registerManyOperators(defaultOperators...)

	assert.True(t, OpEqual.isBasic())
	assert.False(t, OpEqual.isRangable())
	assert.False(t, OpEqual.isStringOnly())
	assert.True(t, OpContains.isStringOnly())
	assert.Equal(t, AnnotationOperatorGreaterThan, OpGreaterThan.Raw)
}
