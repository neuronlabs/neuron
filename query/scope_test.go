package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/internal"
)

// TestFormatQuery tests the format query methods
func TestFormatQuery(t *testing.T) {
	mp := mapping.NewModelMap(mapping.SnakeCase)

	err := mp.RegisterModels(&internal.Formatter{}, &internal.FormatterRelation{})
	require.NoError(t, err)

	mStruct, ok := mp.GetModelStruct(&internal.Formatter{})
	require.True(t, ok)

	t.Run("Pagination", func(t *testing.T) {
		s := newScope(mStruct)
		require.NoError(t, err)

		s.Limit(12)
		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "12", q.Get(ParamPageLimit))
	})

	t.Run("Fieldset", func(t *testing.T) {
		t.Run("DefaultController", func(t *testing.T) {
			s := newScope(mStruct)
			require.NoError(t, err)

			s.FieldSets = append(s.FieldSets, mStruct.Fields())

			q := s.FormatQuery()
			require.Len(t, q, 1)

			fieldsString := q.Get(fmt.Sprintf("fields[%s]", s.ModelStruct.Collection()))
			assert.Equal(t, "[[id,attr,fk,lang]]", fieldsString)
		})
	})
}
