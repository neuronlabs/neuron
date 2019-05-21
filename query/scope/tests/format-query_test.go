package tests

import (
	"fmt"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/query/filters"
	"github.com/neuronlabs/neuron/query/pagination"
	"github.com/neuronlabs/neuron/query/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

import (
	// mocks import and register mock repository
	_ "github.com/neuronlabs/neuron/query/scope/mocks"
)

type formatter struct {
	ID     int                `neuron:"type=primary"`
	Attr   string             `neuron:"type=attr"`
	Rel    *formatterRelation `neuron:"type=relation;foreign=FK"`
	FK     int                `neuron:"type=foreign"`
	Filter string             `neuron:"type=filterkey"`
	Lang   string             `neuron:"type=attr;flags=langtag"`
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
			s, err := scope.NewC(c, &formatter{})
			require.NoError(t, err)

			s.AddFilter(filters.NewFilter(mStruct.Primary(), filters.OpEqual, 1))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), mStruct.Primary().ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Foreign", func(t *testing.T) {
			s, err := scope.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.ForeignKey("fk")
			require.True(t, ok)

			s.AddFilter(filters.NewFilter(field, filters.OpEqual, 1))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Attribute", func(t *testing.T) {
			s, err := scope.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.Attr("attr")
			require.True(t, ok)

			s.AddFilter(filters.NewFilter(field, filters.OpEqual, "some-value"))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "some-value", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Relationship", func(t *testing.T) {
			s, err := scope.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.RelationField("rel")
			require.True(t, ok)

			relPrim := field.Relationship().ModelStruct().Primary()

			s.AddFilter(filters.NewRelationshipFilter(field, filters.NewFilter(relPrim, filters.OpEqual, 12)))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "12", q.Get(fmt.Sprintf("filter[%s][%s][%s][%s]", mStruct.Collection(), field.ApiName(), relPrim.ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("FilterKey", func(t *testing.T) {
			s, err := scope.NewC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.FilterKey("filter")
			require.True(t, ok)

			s.AddFilter(filters.NewFilter(field, filters.OpEqual, "some-key"))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "some-key", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Language", func(t *testing.T) {
			s, err := scope.NewC(c, &formatter{})
			require.NoError(t, err)

			field := mStruct.LanguageField()
			require.NotNil(t, field)

			s.AddFilter(filters.NewFilter(field, filters.OpNotEqual, "pl"))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "pl", q.Get(internal.QueryParamLanguage), fmt.Sprintf("%v", q))
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		s, err := scope.NewC(c, &formatter{})
		require.NoError(t, err)

		s.SetPagination(pagination.NewLimitOffset(12, 0))
		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "12", q.Get(internal.QueryParamPageLimit))
	})

	t.Run("Sorts", func(t *testing.T) {
		s, err := scope.NewC(c, &formatter{})
		require.NoError(t, err)

		s.SortBy("-id")

		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "-id", q.Get(internal.QueryParamSort))
	})
}
