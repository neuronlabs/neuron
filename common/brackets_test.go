package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSplitBracketParameter tests the SplitBracketParameter function.
func TestSplitBracketParameter(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		testCase := "[collection][field][$operator]"
		splitted, err := SplitBracketParameter(testCase)
		require.NoError(t, err)

		if assert.Len(t, splitted, 3) {
			assert.Equal(t, "collection", splitted[0])
			assert.Equal(t, "field", splitted[1])
			assert.Equal(t, "$operator", splitted[2])
		}
	})

	t.Run("DoubleOpen", func(t *testing.T) {
		testCase := "[[collection][field][$operator]"
		_, err := SplitBracketParameter(testCase)
		require.Error(t, err)
	})

	t.Run("DoubleClose", func(t *testing.T) {
		testCase := "[collection]][field][$operator]"
		_, err := SplitBracketParameter(testCase)
		require.Error(t, err)
	})

	t.Run("NoClose", func(t *testing.T) {
		testCase := "[collection"
		_, err := SplitBracketParameter(testCase)
		require.Error(t, err)
	})
}
