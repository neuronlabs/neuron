package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/query"
	"github.com/neuronlabs/neuron-core/query/filters"
)

import (
	// mocks import and register mock repository
	_ "github.com/neuronlabs/neuron-core/query/mocks"
)

type formatter struct {
	ID     int                `neuron:"type=primary"`
	Attr   string             `neuron:"type=attr"`
	Rel    *formatterRelation `neuron:"type=relation;foreign=FK"`
	FK     int                `neuron:"type=foreign"`
	Filter string             `neuron:"type=filterkey"`
	Lang   string             `neuron:"type=attr;flags=lang"`
}

type formatterRelation struct {
	ID int `neuron:"type=primary"`
}

// TestFormatQuery tests the format query methods
func TestFormatQuery(t *testing.T) {

	c := (*controller.Controller)(newController(t))
	err := c.RegisterModels(&formatter{}, &formatterRelation{})
	require.NoError(t, err)

	mStruct, err := c.ModelStruct(&formatter{})
	require.NoError(t, err)

	t.Run("Filters", func(t *testing.T) {
		t.Run("Primary", func(t *testing.T) {
			s, err := query.NewC(c, &formatter{})
			require.NoError(t, err)

			err = s.FilterField(filters.NewFilter(mStruct.Primary(), filters.OpEqual, 1))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), mStruct.Primary().NeuronName(), filters.OpEqual.Raw)))
		})

		t.Run("Foreign", func(t *testing.T) {
			s, err := query.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.ForeignKey("fk")
			require.True(t, ok)

			err = s.FilterField(filters.NewFilter(field, filters.OpEqual, 1))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.NeuronName(), filters.OpEqual.Raw)))
		})

		t.Run("Attribute", func(t *testing.T) {
			s, err := query.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.Attr("attr")
			require.True(t, ok)

			err = s.FilterField(filters.NewFilter(field, filters.OpEqual, "some-value"))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "some-value", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.NeuronName(), filters.OpEqual.Raw)))
		})

		t.Run("Relationship", func(t *testing.T) {
			s, err := query.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.RelationField("rel")
			require.True(t, ok)

			relPrim := field.Relationship().ModelStruct().Primary()

			err = s.FilterField(filters.NewRelationshipFilter(field, filters.NewFilter(relPrim, filters.OpEqual, 12)))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "12", q.Get(fmt.Sprintf("filter[%s][%s][%s][%s]", mStruct.Collection(), field.NeuronName(), relPrim.NeuronName(), filters.OpEqual.Raw)))
		})

		t.Run("FilterKey", func(t *testing.T) {
			s, err := query.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.FilterKey("filter")
			require.True(t, ok)

			err = s.FilterField(filters.NewFilter(field, filters.OpEqual, "some-key"))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "some-key", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.NeuronName(), filters.OpEqual.Raw)))
		})

		t.Run("Language", func(t *testing.T) {
			s, err := query.NewC(c, &formatter{})
			require.NoError(t, err)

			field := mStruct.LanguageField()
			require.NotNil(t, field)

			err = s.FilterField(filters.NewFilter(field, filters.OpNotEqual, "pl"))
			require.NoError(t, err)

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "pl", q.Get(filters.QueryParamLanguage), fmt.Sprintf("%v", q))
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		s, err := query.NewC(c, &formatter{})
		require.NoError(t, err)

		err = s.SetPagination(query.LimitOffsetPagination(12, 0))
		require.NoError(t, err)
		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "12", q.Get(query.ParamPageLimit))
	})

	t.Run("Sorts", func(t *testing.T) {
		s, err := query.NewC(c, &formatter{})
		require.NoError(t, err)

		err = s.Sort("-id")
		require.NoError(t, err)

		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "-id", q.Get(query.ParamSort))
	})
}
