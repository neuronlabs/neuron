package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrderedFielset test the ordered fieldset function.
func TestOrderedFielset(t *testing.T) {
	type Ordered struct {
		ID     int
		First  string
		Second string
		Third  string
	}

	c := newController(t)
	err := c.RegisterModels(Ordered{})
	require.NoError(t, err)

	t.Run("All", func(t *testing.T) {
		s, err := NewC(c, &Ordered{})
		require.NoError(t, err)

		s.setAllFields()

		ordered := s.OrderedFieldset()
		for i, field := range ordered {
			switch i {
			case 0:
				assert.Equal(t, "ID", field.Name())
			case 1:
				assert.Equal(t, "First", field.Name())
			case 2:
				assert.Equal(t, "Second", field.Name())
			case 3:
				assert.Equal(t, "Third", field.Name())
			}
		}
	})

	t.Run("Custom", func(t *testing.T) {
		s, err := NewC(c, &Ordered{})
		require.NoError(t, err)

		s.SetFields("first", "id", "third")

		ordered := s.OrderedFieldset()
		for i, field := range ordered {
			switch i {
			case 0:
				assert.Equal(t, "ID", field.Name())
			case 1:
				assert.Equal(t, "First", field.Name())
			case 2:
				assert.Equal(t, "Third", field.Name())
			}
		}
	})
}
