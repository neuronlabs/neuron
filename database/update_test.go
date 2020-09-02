package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filter"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestUpdateModels(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(WithDefaultRepository(repo), WithModelMap(mm))
	require.NoError(t, err)

	t.Run("WithHooks", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&TestModel{})
		require.NoError(t, err)

		m1 := &TestModel{
			ID:             1,
			FieldSetBefore: "",
			FieldSetAfter:  "",
		}
		m2 := &TestModel{
			ID:             2,
			FieldSetBefore: "",
			FieldSetAfter:  "",
		}

		repo.OnUpdateModels(func(_ context.Context, s *query.Scope) (int64, error) {
			if assert.Len(t, s.Models, 2) {
				model, ok := s.Models[0].(*TestModel)
				require.True(t, ok)

				assert.Equal(t, model, m1)
				assert.False(t, model.UpdatedAt.IsZero())
				assert.Equal(t, "before", model.FieldSetBefore)

				model, ok = s.Models[1].(*TestModel)
				require.True(t, ok)

				assert.Equal(t, model, m2)
				assert.False(t, model.UpdatedAt.IsZero())
				assert.Equal(t, "before", model.FieldSetBefore)
			}
			if assert.Len(t, s.FieldSets, 2) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 3) {
					assert.Equal(t, fieldSet[0], mStruct.Primary())
					updatedAt, _ := mStruct.UpdatedAt()
					assert.Equal(t, fieldSet[1], updatedAt)
					assert.Equal(t, fieldSet[2], mStruct.MustFieldByName("FieldSetBefore"))
				}
				assert.Equal(t, s.FieldSets[0].Hash(), s.FieldSets[1].Hash())
			}
			return 2, nil
		})
		res, err := db.Update(context.Background(), mStruct, m1, m2)
		require.NoError(t, err)

		assert.Equal(t, int64(2), res)
	})

	t.Run("NoHooksNoUpdatedAt", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&SoftDeletableNoHooks{})
		require.NoError(t, err)

		m1 := &SoftDeletableNoHooks{ID: 1}
		m2 := &SoftDeletableNoHooks{ID: 2, Integer: 20}

		repo.OnUpdateModels(func(_ context.Context, s *query.Scope) (int64, error) {
			if assert.Len(t, s.Models, 2) {
				model, ok := s.Models[0].(*SoftDeletableNoHooks)
				require.True(t, ok)

				assert.Equal(t, model, m1)

				model, ok = s.Models[1].(*SoftDeletableNoHooks)
				require.True(t, ok)

				assert.Equal(t, model, m2)
			}
			if assert.Len(t, s.FieldSets, 2) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 1) {
					assert.Equal(t, fieldSet[0], mStruct.Primary())
				}
				fieldSet = s.FieldSets[1]
				if assert.Len(t, fieldSet, 2) {
					assert.Equal(t, fieldSet[0], mStruct.Primary())
					assert.Equal(t, fieldSet[1], mStruct.MustFieldByName("Integer"))
				}
			}
			return 2, nil
		})
		res, err := db.Update(context.Background(), mStruct, m1, m2)
		require.NoError(t, err)

		assert.Equal(t, int64(2), res)
	})
}

func TestUpdate(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(Neuron_Models...)
	require.NoError(t, err)
	err = mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(WithDefaultRepository(repo), WithModelMap(mm))
	require.NoError(t, err)

	t.Run("WithHooks", func(t *testing.T) {
		mStruct, err := mm.ModelStruct(&TestModel{})
		require.NoError(t, err)

		input := &TestModel{Integer: 50}
		q := query.NewScope(mStruct, input)
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
			now := time.Now().Add(-time.Hour)
			s.Models = []mapping.Model{
				&TestModel{
					ID:        12,
					CreatedAt: now,
					UpdatedAt: now,
				},
				&TestModel{
					ID:        15,
					CreatedAt: now,
					UpdatedAt: now,
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
				assert.False(t, s1.UpdatedAt.Equal(s1.CreatedAt))
				assert.True(t, s1.UpdatedAt.After(s1.CreatedAt))
				assert.Nil(t, s1.DeletedAt)
				assert.Equal(t, input.Integer, s1.Integer)

				s2, ok := s.Models[1].(*TestModel)
				require.True(t, ok)
				assert.Equal(t, 15, s2.ID)
				assert.False(t, s2.UpdatedAt.Equal(s2.CreatedAt))
				assert.True(t, s2.UpdatedAt.After(s2.CreatedAt))
				assert.Nil(t, s2.DeletedAt)
				assert.Equal(t, input.Integer, s2.Integer)
			}
			return 2, nil
		})

		res, err := db.UpdateQuery(context.Background(), q)
		require.NoError(t, err)

		assert.Equal(t, int64(2), res)
	})

	t.Run("NoHooks", func(t *testing.T) {
		t.Run("NoUpdatedAt", func(t *testing.T) {
			mStruct, err := mm.ModelStruct(&SoftDeletableNoHooks{})
			require.NoError(t, err)

			input := &SoftDeletableNoHooks{Integer: 50}
			q := query.NewScope(mStruct, input)
			q.Filter(filter.New(mStruct.Primary(), filter.OpIn, 12, 15))

			repo.OnUpdate(func(ctx context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, mStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 1) {
					assert.Equal(t, s.Models[0], input)
				}

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{12, 15}, sf.Values)
				}

				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fieldSet[0], mStruct.MustFieldByName("Integer"))
					}
				}
				return 2, nil
			})

			res, err := db.UpdateQuery(context.Background(), q)
			require.NoError(t, err)

			assert.Equal(t, int64(2), res)
		})

		t.Run("WithUpdatedAt", func(t *testing.T) {
			mStruct, err := mm.ModelStruct(&Updateable{})
			require.NoError(t, err)

			input := &Updateable{Integer: 50}
			q := query.NewScope(mStruct, input)
			q.Filter(filter.New(mStruct.Primary(), filter.OpIn, 12, 15))

			repo.OnUpdate(func(ctx context.Context, s *query.Scope) (int64, error) {
				require.Equal(t, mStruct, s.ModelStruct)
				if assert.Len(t, s.Models, 1) {
					if assert.Equal(t, s.Models[0], input) {
						assert.False(t, input.UpdatedAt.IsZero())
					}
				}

				if assert.Len(t, s.Filters, 1) {
					sf, ok := s.Filters[0].(filter.Simple)
					require.True(t, ok)

					assert.Equal(t, mStruct.Primary(), sf.StructField)
					assert.Equal(t, filter.OpIn, sf.Operator)
					assert.ElementsMatch(t, []interface{}{12, 15}, sf.Values)
				}

				if assert.Len(t, s.FieldSets, 1) {
					fieldSet := s.FieldSets[0]
					if assert.Len(t, fieldSet, 2) {
						updatedAt, _ := mStruct.UpdatedAt()
						assert.Equal(t, fieldSet[0], mStruct.MustFieldByName("Integer"))
						assert.Equal(t, fieldSet[1], updatedAt)
					}
				}
				return 2, nil
			})

			res, err := db.UpdateQuery(context.Background(), q)
			require.NoError(t, err)

			assert.Equal(t, int64(2), res)
		})
	})
}
