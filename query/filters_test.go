package query

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
)

// TestFilterFormatQuery checks the FormatQuery function for the filters.
//noinspection GoNilness
func TestFilterFormatQuery(t *testing.T) {
	mp := mapping.NewModelMap(mapping.SnakeCase)

	require.NoError(t, mp.RegisterModels(&TestingModel{}, &FilterRelationModel{}))

	mStruct, ok := mp.GetModelStruct(&TestingModel{})
	require.True(t, ok)

	t.Run("MultipleValue", func(t *testing.T) {
		tm := time.Now()
		f := NewFilterField(mStruct.Primary(), OpIn, 1, 2.01, 30, "something", []string{"i", "am"}, true, tm, &tm)
		q := f.FormatQuery()
		require.NotNil(t, q)

		assert.Len(t, q, 1)
		var k string
		var v []string

		for k, v = range q {
		}

		assert.Equal(t, fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), mStruct.Primary().NeuronName(), OpIn.URLAlias), k)

		if assert.Len(t, v, 1) {
			v = strings.Split(v[0], mapping.AnnotationSeparator)
			assert.Equal(t, "1", v[0])
			assert.Contains(t, v[1], "2.01")
			assert.Equal(t, "30", v[2])
			assert.Equal(t, "something", v[3])
			assert.Equal(t, "i", v[4])
			assert.Equal(t, "am", v[5])
			assert.Equal(t, "true", v[6])
			assert.Equal(t, fmt.Sprintf("%d", tm.Unix()), v[7])
			assert.Equal(t, fmt.Sprintf("%d", tm.Unix()), v[8])
		}
	})

	t.Run("WithNested", func(t *testing.T) {
		rel, ok := mStruct.RelationByName("relation")
		require.True(t, ok)

		relFilter := newRelationshipFilter(rel, NewFilterField(rel.ModelStruct().Primary(), OpIn, uint(1), uint64(2)))
		q := relFilter.FormatQuery()

		require.Len(t, q, 1)
		var k string
		var v []string

		for k, v = range q {
		}

		assert.Equal(t, fmt.Sprintf("filter[%s][%s][%s][%s]", mStruct.Collection(), relFilter.StructField.NeuronName(), relFilter.StructField.Relationship().Struct().Primary().NeuronName(), OpIn.URLAlias), k)
		if assert.Len(t, v, 1) {
			assert.NotNil(t, v)
			v = strings.Split(v[0], mapping.AnnotationSeparator)

			assert.Equal(t, "1", v[0])
			assert.Equal(t, "2", v[1])
		}
	})
}
