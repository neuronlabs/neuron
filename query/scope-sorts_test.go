package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScopeSorts tests the scope sort functions.
func TestScopeSorts(t *testing.T) {
	type SortForeignModel struct {
		ID     int
		Length int
	}

	type SortModel struct {
		ID         int
		Name       string
		Age        int
		Relation   *SortForeignModel
		RelationID int
	}

	c := newController(t)

	err := c.RegisterModels(SortModel{}, SortForeignModel{})
	require.NoError(t, err)

	t.Run("EmptyFields", func(t *testing.T) {
		s, err := NewC(c, &SortModel{})
		require.NoError(t, err)

		err = s.SortField("-id")
		require.NoError(t, err)

		err = s.Sort("-age")
		assert.NoError(t, err)
	})

	t.Run("Duplicated", func(t *testing.T) {
		s, err := NewC(c, &SortModel{})
		require.NoError(t, err)

		err = s.Sort("age")
		require.NoError(t, err)

		err = s.Sort("-id", "id", "-id")
		assert.Error(t, err)
	})
}
