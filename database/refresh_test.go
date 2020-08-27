package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestRefreshFind(t *testing.T) {
	c := core.NewDefault()
	err := c.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	err = c.SetDefaultRepository(repo)
	require.NoError(t, err)

	db := New(c)

	mStruct, err := c.ModelStruct(&testmodels.Blog{})
	require.NoError(t, err)

	t.Run("Many", func(t *testing.T) {
		repo.OnFind(func(ctx context.Context, s *query.Scope) error {
			// There shouldn't be any model in the query - even though it is about to refresh them.
			assert.Equal(t, mStruct, s.ModelStruct)
			// Filter.
			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)
				assert.Equal(t, mStruct.Primary(), sf.StructField)
				assert.Equal(t, filter.OpIn, sf.Operator)
				if assert.Len(t, sf.Values, 2) {
					assert.ElementsMatch(t, sf.Values, []interface{}{12, 11})
				}
			}
			// Fieldset
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.Equal(t, fieldSet[0], mStruct.Primary())
					assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Title"))
				}
			}
			// Even though the query scope would contain different models - the input models would contain different
			s.Models = []mapping.Model{
				&testmodels.Blog{ID: 12, Title: "Simple title"},
				&testmodels.Blog{ID: 11, Title: "Second title"},
			}
			return nil
		})

		// Put input models and refresh their values.
		b1 := &testmodels.Blog{ID: 12}
		b2 := &testmodels.Blog{ID: 11}

		err = db.Query(mStruct, b1, b2).
			Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
			Refresh()
		require.NoError(t, err)

		assert.Equal(t, "Simple title", b1.Title)
		assert.Equal(t, "Second title", b2.Title)

		t.Run("NotAllFound", func(t *testing.T) {
			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				// There shouldn't be any model in the query - even though it is about to refresh them.
				assert.Equal(t, mStruct, s.ModelStruct)
				// Filter.
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)
					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					if assert.Len(t, sf.Values, 2) {
						assert.ElementsMatch(t, sf.Values, []interface{}{12, 11})
					}
				}
				// Fieldset
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 2) {
						assert.Equal(t, fieldSet[0], mStruct.Primary())
						assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Title"))
					}
				}
				// Even though the query scope would contain different models - the input models would contain different
				s.Models = []mapping.Model{
					&testmodels.Blog{ID: 11, Title: "Second title"},
				}
				return nil
			})
			err = db.Query(mStruct, b1, b2).
				Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
				Refresh()
			assert.Error(t, err)
		})
	})

	t.Run("Single", func(t *testing.T) {
		repo.OnFind(func(ctx context.Context, s *query.Scope) error {
			// There shouldn't be any model in the query - even though it is about to refresh them.
			assert.Equal(t, mStruct, s.ModelStruct)
			// Filter.
			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)
				assert.Equal(t, mStruct.Primary(), sf.StructField)
				assert.Equal(t, filter.OpEqual, sf.Operator)
				if assert.Len(t, sf.Values, 1) {
					assert.Equal(t, 20, sf.Values[0])
				}
			}
			// Fieldset
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.Equal(t, fieldSet[0], mStruct.Primary())
					assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Title"))
				}
			}
			// Even though the query scope would contain different models - the input models would contain different
			s.Models = []mapping.Model{
				&testmodels.Blog{ID: 20, Title: "20 title"},
			}
			return nil
		})
		b3 := &testmodels.Blog{ID: 20}
		err = db.Query(mStruct, b3).
			Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
			Refresh()
		require.NoError(t, err)

		assert.Equal(t, "20 title", b3.Title)

		t.Run("NotFound", func(t *testing.T) {
			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				// There shouldn't be any model in the query - even though it is about to refresh them.
				assert.Equal(t, mStruct, s.ModelStruct)
				// Filter.
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)
					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpEqual, sf.Operator)
					if assert.Len(t, sf.Values, 1) {
						assert.Equal(t, 20, sf.Values[0])
					}
				}
				// Fieldset
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 2) {
						assert.Equal(t, fieldSet[0], mStruct.Primary())
						assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Title"))
					}
				}
				// Even though the query scope would contain different models - the input models would contain different
				s.Models = []mapping.Model{}
				return nil
			})
			b3 := &testmodels.Blog{ID: 20}
			err = db.Query(mStruct, b3).
				Select(mStruct.Primary(), mStruct.MustFieldByName("Title")).
				Refresh()
			assert.Error(t, err)
			assert.True(t, errors.Is(err, query.ErrNoResult))
		})
	})

	t.Run("Models", func(t *testing.T) {
		repo.OnFind(func(ctx context.Context, s *query.Scope) error {
			// There shouldn't be any model in the query - even though it is about to refresh them.
			assert.Equal(t, mStruct, s.ModelStruct)
			// Filter.
			if assert.Len(t, s.Filters, 1) {
				sf, ok := s.Filters[0].(filter.Simple)
				require.True(t, ok)
				assert.Equal(t, mStruct.Primary(), sf.StructField)
				assert.Equal(t, filter.OpIn, sf.Operator)
				if assert.Len(t, sf.Values, 2) {
					assert.ElementsMatch(t, sf.Values, []interface{}{12, 11})
				}
			}
			// Fieldset
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 2) {
					assert.Equal(t, fieldSet[0], mStruct.Primary())
					assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Title"))
				}
			}
			// Even though the query scope would contain different models - the input models would contain different
			s.Models = []mapping.Model{
				&testmodels.Blog{ID: 12, Title: "Simple title"},
				&testmodels.Blog{ID: 11, Title: "Second title"},
			}
			return nil
		})

		// Put input models and refresh their values.
		b1 := &testmodels.Blog{ID: 12}
		b2 := &testmodels.Blog{ID: 11}

		err = RefreshModels(context.Background(), db, mStruct, []mapping.Model{b1, b2}, mStruct.Primary(), mStruct.MustFieldByName("Title"))
		require.NoError(t, err)

		assert.Equal(t, "Simple title", b1.Title)
		assert.Equal(t, "Second title", b2.Title)
	})
}
