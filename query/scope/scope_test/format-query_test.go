package scope_test

import (
	"fmt"
	"github.com/kucjac/jsonapi/controller"
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/query/filters"
	"github.com/kucjac/jsonapi/query/pagination"
	"github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/jsonapi/query/scope/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type formatter struct {
	ID     int                `jsonapi:"type=primary"`
	Attr   string             `jsonapi:"type=attr"`
	Rel    *formatterRelation `jsonapi:"type=relation;foreign=FK"`
	FK     int                `jsonapi:"type=foreign"`
	Filter string             `jsonapi:"type=filterkey"`
	Lang   string             `jsonapi:"type=attr;flags=langtag"`
}

type formatterRelation struct {
	ID int `jsonapi:"type=primary"`
}

// TestFormatQuery tests the format query methods
func TestFormatQuery(t *testing.T) {
	repo := &mocks.Repository{}

	c := (*controller.Controller)(newController(t, repo))
	err := c.RegisterModels(&formatter{}, &formatterRelation{})
	require.NoError(t, err)

	mStruct, err := c.ModelStruct(&formatter{})
	require.NoError(t, err)

	t.Run("Filters", func(t *testing.T) {
		t.Run("Primary", func(t *testing.T) {
			s, err := scope.NewWithC(c, &formatter{})
			require.NoError(t, err)

			s.AddFilter(filters.NewFilter(mStruct.Primary(), filters.OpEqual, 1))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), mStruct.Primary().ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Foreign", func(t *testing.T) {
			s, err := scope.NewWithC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.ForeignKey("fk")
			require.True(t, ok)

			s.AddFilter(filters.NewFilter(field, filters.OpEqual, 1))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "1", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Attribute", func(t *testing.T) {
			s, err := scope.NewWithC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.Attr("attr")
			require.True(t, ok)

			s.AddFilter(filters.NewFilter(field, filters.OpEqual, "some-value"))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "some-value", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Relationship", func(t *testing.T) {
			s, err := scope.NewWithC(c, &formatter{})
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
			s, err := scope.NewWithC(c, &formatter{})
			require.NoError(t, err)

			field, ok := mStruct.FilterKey("filter")
			require.True(t, ok)

			s.AddFilter(filters.NewFilter(field, filters.OpEqual, "some-key"))

			q := s.FormatQuery()
			require.Len(t, q, 1)

			assert.Equal(t, "some-key", q.Get(fmt.Sprintf("filter[%s][%s][%s]", mStruct.Collection(), field.ApiName(), filters.OpEqual.Raw)))
		})

		t.Run("Language", func(t *testing.T) {
			s, err := scope.NewWithC(c, &formatter{})
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
		s, err := scope.NewWithC(c, &formatter{})
		require.NoError(t, err)

		s.SetPagination(pagination.NewLimitOffset(12, 0))
		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "12", q.Get(internal.QueryParamPageLimit))
	})

	t.Run("Sorts", func(t *testing.T) {
		s, err := scope.NewWithC(c, &formatter{})
		require.NoError(t, err)

		s.SortBy("-id")

		q := s.FormatQuery()
		require.Len(t, q, 1)

		assert.Equal(t, "-id", q.Get(internal.QueryParamSort))
	})
}
