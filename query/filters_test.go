package query

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/annotation"
	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/controller"
)

type testingModel struct {
	ID         int                  `neuron:"type=primary"`
	Attr       string               `neuron:"type=attr"`
	Relation   *filterRelationModel `neuron:"type=relation;foreign=ForeignKey"`
	ForeignKey int                  `neuron:"type=foreign"`
	Nested     *filterNestedModel   `neuron:"type=attr"`
}

type filterRelationModel struct {
	ID int `neuron:"type=primary"`
}

type filterNestedModel struct {
	Field string
}

// TestNewStringFilter tests the NewUrlStringFilter function.
func TestNewUrlStringFilter(t *testing.T) {
	c := controller.NewDefault()

	err := c.RegisterRepository(repoName, &config.Repository{DriverName: repoName})
	require.NoError(t, err)

	require.NoError(t, c.RegisterModels(&testingModel{}, &filterRelationModel{}))

	mStruct, err := c.ModelStruct(&testingModel{})
	require.NoError(t, err)

	t.Run("Primary", func(t *testing.T) {
		t.Run("WithoutOperator", func(t *testing.T) {
			filter, err := NewUrlStringFilter(c, "filter[testing_models][id]", 521)
			require.NoError(t, err)

			assert.Equal(t, mStruct.Primary(), filter.StructField)
			require.Len(t, filter.Values, 1)

			fv := filter.Values[0]
			assert.Equal(t, OpEqual, fv.Operator)
			require.Len(t, fv.Values, 1)
			assert.Equal(t, 521, fv.Values[0])
		})

		t.Run("WithoutFilterWord", func(t *testing.T) {
			filter, err := NewUrlStringFilter(c, "[testing_models][id][$ne]", "some string value")
			require.NoError(t, err)

			assert.Equal(t, mStruct.Primary(), filter.StructField)
			require.Len(t, filter.Values, 1)

			fv := filter.Values[0]
			assert.Equal(t, OpNotEqual, fv.Operator)
			require.Len(t, fv.Values, 1)
			assert.Equal(t, "some string value", fv.Values[0])
		})
	})

	t.Run("Invalid", func(t *testing.T) {
		t.Run("Collection", func(t *testing.T) {
			_, err := NewUrlStringFilter(c, "filter[invalid-collection][field_name][$eq]", 1)
			require.Error(t, err)
		})

		t.Run("Operator", func(t *testing.T) {
			_, err := NewUrlStringFilter(c, "filter[testing_models][id][$unknown]", 1)
			require.Error(t, err)
		})

		t.Run("FieldName", func(t *testing.T) {
			_, err := NewUrlStringFilter(c, "filter[testing_models][field-unknown][$eq]", "", 1)
			require.Error(t, err)
		})
	})

	t.Run("Attribute", func(t *testing.T) {
		filter, err := NewUrlStringFilter(c, "[testing_models][attr][$ne]", "some string value")
		require.NoError(t, err)

		attrField, ok := mStruct.FieldByName("Attr")
		require.True(t, ok)

		assert.Equal(t, attrField, filter.StructField)
		require.Len(t, filter.Values, 1)

		fv := filter.Values[0]
		assert.Equal(t, OpNotEqual, fv.Operator)
		require.Len(t, fv.Values, 1)
		assert.Equal(t, "some string value", fv.Values[0])
	})

	t.Run("ForeignKey", func(t *testing.T) {
		_, err := NewUrlStringFilter(c, "[testing_models][foreign_key][$ne]", "some string value")
		require.Error(t, err)

		filter, err := NewStringFilterWithForeignKey(c, "[testing_models][foreign_key][$ne]", "some string value")
		require.NoError(t, err)

		attrField, ok := mStruct.FieldByName("ForeignKey")
		require.True(t, ok)

		assert.Equal(t, attrField, filter.StructField)
		require.Len(t, filter.Values, 1)

		fv := filter.Values[0]
		assert.Equal(t, OpNotEqual, fv.Operator)
		require.Len(t, fv.Values, 1)
		assert.Equal(t, "some string value", fv.Values[0])
	})

	t.Run("Relationship", func(t *testing.T) {
		filter, err := NewUrlStringFilter(c, "[testing_models][relation][id][$ne]", "some string value")
		require.NoError(t, err)
		attrField, ok := mStruct.FieldByName("Relation")
		require.True(t, ok)

		assert.Equal(t, attrField, filter.StructField)
		require.Len(t, filter.Nested, 1)

		nested := filter.Nested[0]
		require.Len(t, nested.Values, 1)

		fv := nested.Values[0]
		assert.Equal(t, OpNotEqual, fv.Operator)
		require.Len(t, fv.Values, 1)
		assert.Equal(t, "some string value", fv.Values[0])
	})
}

// TestFilterFormatQuery checks the FormatQuery function for the filters.
//noinspection GoNilness
func TestFilterFormatQuery(t *testing.T) {
	c := controller.NewDefault()

	err := c.RegisterRepository(repoName, &config.Repository{DriverName: repoName})
	require.NoError(t, err)

	require.NoError(t, c.RegisterModels(&testingModel{}, &filterRelationModel{}))

	mStruct, err := c.ModelStruct(&testingModel{})
	require.NoError(t, err)

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
			v = strings.Split(v[0], annotation.Separator)
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
		rel, ok := mStruct.RelationField("relation")
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
			v = strings.Split(v[0], annotation.Separator)

			assert.Equal(t, "1", v[0])
			assert.Equal(t, "2", v[1])
		}
	})
}
