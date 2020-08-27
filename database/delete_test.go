package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"

	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestDelete(t *testing.T) {
	c := core.NewDefault()
	err := c.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	err = c.SetDefaultRepository(repo)
	require.NoError(t, err)

	db := New(c)

	m := &testmodels.Blog{ID: 3}
	mStruct, err := c.ModelStruct(m)
	require.NoError(t, err)

	repo.OnDelete(func(ctx context.Context, s *query.Scope) (int64, error) {
		if assert.Len(t, s.Filters, 1) {
			f, ok := s.Filters[0].(filter.Simple)
			require.True(t, ok)

			assert.Equal(t, mStruct.Primary(), f.StructField)
			assert.Equal(t, filter.OpEqual, f.Operator)
			if assert.Len(t, f.Values, 1) {
				assert.Equal(t, m.ID, f.Values[0])
			}
		}
		return 1, nil
	})

	res, err := db.Delete(context.Background(), mStruct, m)
	require.NoError(t, err)

	assert.Equal(t, int64(1), res)
}

func TestSoftDelete(t *testing.T) {
	c := core.NewDefault()
	err := c.RegisterModels(Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	err = c.SetDefaultRepository(repo)
	require.NoError(t, err)

	db := New(c)

	mStruct, err := c.ModelStruct(&TestModel{})
	require.NoError(t, err)

	t.Run("Models", func(t *testing.T) {
		t.Run("WithHooks", func(t *testing.T) {
			m := &TestModel{ID: 3}
			// Soft delete is an update with the deleted at field selected.
			repo.OnUpdateModels(func(ctx context.Context, s *query.Scope) (int64, error) {
				if assert.Len(t, s.Models, 1) {
					model, ok := s.Models[0].(*TestModel)
					require.True(t, ok)

					assert.NotNil(t, model.DeletedAt)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 3) {
						assert.Equal(t, fieldSet[0], mStruct.Primary())
						deletedAt, _ := mStruct.DeletedAt()
						assert.Equal(t, fieldSet[1], deletedAt)
						assert.Equal(t, fieldSet[2], mStruct.MustFieldByName("FieldSetBefore"))
					}
				}
				return 1, nil
			})

			res, err := db.Delete(context.Background(), mStruct, m)
			require.NoError(t, err)

			assert.Equal(t, int64(1), res)
			assert.Equal(t, "before", m.FieldSetBefore)
			assert.Equal(t, "after", m.FieldSetAfter)
		})

		t.Run("NoHooks", func(t *testing.T) {
			m := &SoftDeletableNoHooks{ID: 3}
			mStruct, err := c.ModelStruct(m)
			require.NoError(t, err)

			// Soft delete is an update with the deleted at field selected.
			repo.OnUpdateModels(func(ctx context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, mStruct, s.ModelStruct)
				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpEqual, sf.Operator)
					assert.ElementsMatch(t, []interface{}{3}, sf.Values)
				}

				if assert.Len(t, s.Models, 1) {
					model, ok := s.Models[0].(*SoftDeletableNoHooks)
					require.True(t, ok)

					assert.NotNil(t, model.DeletedAt)
				}

				deletedAt, _ := mStruct.DeletedAt()
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fieldSet[0], deletedAt)
					}
				}
				return 1, nil
			})

			res, err := db.Delete(context.Background(), mStruct, m)
			require.NoError(t, err)

			assert.Equal(t, int64(1), res)
		})
	})

	t.Run("Filtered", func(t *testing.T) {
		t.Run("WithHooks", func(t *testing.T) {
			q := query.NewScope(mStruct)
			q.Filter(filter.New(mStruct.Primary(), filter.OpIn, 12, 15))

			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				require.Equal(t, mStruct, s.ModelStruct)

				if assert.Len(t, s.Filters, 2) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{12, 15}, sf.Values)

					sf, ok = s.Filters[1].(filter.Simple)
					require.True(t, ok)

					deletedAt, ok := mStruct.DeletedAt()
					require.True(t, ok)

					assert.Equal(t, deletedAt, sf.StructField)
					assert.Equal(t, filter.OpIsNull, sf.Operator)
				}
				s.Models = []mapping.Model{
					&TestModel{
						ID:        12,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					&TestModel{
						ID:        15,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				return nil
			})

			repo.OnUpdateModels(func(ctx context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, mStruct, s.ModelStruct)

				if assert.Len(t, s.Models, 2) {
					s1, ok := s.Models[0].(*TestModel)
					require.True(t, ok)
					assert.Equal(t, 12, s1.ID)
					assert.NotNil(t, s1.DeletedAt)

					s2, ok := s.Models[1].(*TestModel)
					require.True(t, ok)
					assert.Equal(t, 15, s2.ID)
					assert.NotNil(t, s2.DeletedAt)
				}
				return 2, nil
			})

			res, err := db.DeleteQuery(context.Background(), q)
			require.NoError(t, err)

			assert.Equal(t, int64(2), res)
		})
		t.Run("NoHooks", func(t *testing.T) {
			mStruct, err := c.ModelStruct(&SoftDeletableNoHooks{})
			require.NoError(t, err)

			q := query.NewScope(mStruct)
			q.Filter(filter.New(mStruct.Primary(), filter.OpIn, 12, 15))

			repo.OnFind(func(ctx context.Context, s *query.Scope) error {
				require.Equal(t, mStruct, s.ModelStruct)

				if assert.Len(t, s.Filters, 2) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{12, 15}, sf.Values)

					sf, ok = s.Filters[1].(filter.Simple)
					require.True(t, ok)

					deletedAt, ok := mStruct.DeletedAt()
					require.True(t, ok)

					assert.Equal(t, deletedAt, sf.StructField)
					assert.Equal(t, filter.OpIsNull, sf.Operator)
				}
				s.Models = []mapping.Model{
					&SoftDeletableNoHooks{
						ID: 12,
					},
					&SoftDeletableNoHooks{
						ID: 15,
					},
				}
				return nil
			})

			repo.OnUpdate(func(ctx context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, mStruct, s.ModelStruct)

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{12, 15}, sf.Values)
				}

				if assert.Len(t, s.Models, 1) {
					sd, ok := s.Models[0].(*SoftDeletableNoHooks)
					require.True(t, ok)
					assert.Zero(t, sd.ID)
					assert.NotNil(t, sd.DeletedAt)
				}
				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						deletedAt, _ := mStruct.DeletedAt()
						assert.Equal(t, deletedAt, fieldSet[0])
					}
				}
				return 2, nil
			})

			res, err := db.DeleteQuery(context.Background(), q)
			require.NoError(t, err)

			assert.Equal(t, int64(2), res)
		})
	})
}
