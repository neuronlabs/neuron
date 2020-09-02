package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestInsert(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}
	db, err := New(
		WithDefaultRepository(repo),
		WithModelMap(mm),
	)
	require.NoError(t, err)

	mStruct, err := mm.ModelStruct(&TestModel{})
	require.NoError(t, err)

	createdAt, _ := mStruct.CreatedAt()
	updatedAt, _ := mStruct.UpdatedAt()

	t.Run("WithCommonFieldSet", func(t *testing.T) {
		m1 := &TestModel{
			ID:      1,
			Integer: 1,
		}
		m2 := &TestModel{
			ID:      2,
			Integer: 2,
		}
		repo.OnInsert(func(_ context.Context, s *query.Scope) error {
			if assert.Len(t, s.FieldSets, 1) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 4) {
					assert.Contains(t, fieldSet, mStruct.Primary())
					assert.Contains(t, fieldSet, mStruct.MustFieldByName("Integer"))
					assert.Contains(t, fieldSet, createdAt)
					assert.Contains(t, fieldSet, updatedAt)
				}
			}
			if assert.Len(t, s.Models, 2) {
				ms1, ok := s.Models[0].(*TestModel)
				require.True(t, ok)
				assert.Equal(t, m1, ms1)

				assert.False(t, ms1.CreatedAt.IsZero())
				assert.False(t, ms1.UpdatedAt.IsZero())

				ms2, ok := s.Models[1].(*TestModel)
				require.True(t, ok)
				assert.Equal(t, m2, ms2)

				assert.False(t, ms2.CreatedAt.IsZero())
				assert.False(t, ms2.UpdatedAt.IsZero())
			}
			return nil
		})

		// Insert fields with only selected primary and 'Integer' field. This would result with common fieldset for many models.
		err = db.Query(mStruct, m1, m2).
			Select(mStruct.Primary(), mStruct.MustFieldByName("Integer")).
			Insert()
		require.NoError(t, err)
	})

	t.Run("WithFieldSets", func(t *testing.T) {
		// No Primary key defined.
		m1 := &TestModel{
			FieldSetBefore: "before",
		}
		// Defined primary key and the 'Integer'.
		m2 := &TestModel{
			ID:      2,
			Integer: 2,
		}
		repo.OnInsert(func(_ context.Context, s *query.Scope) error {
			if assert.Len(t, s.FieldSets, 2) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 3) {
					assert.Contains(t, fieldSet, mStruct.MustFieldByName("FieldSetBefore"))
					assert.Contains(t, fieldSet, createdAt)
					assert.Contains(t, fieldSet, updatedAt)
				}

				fieldSet = s.FieldSets[1]
				if assert.Len(t, fieldSet, 4) {
					assert.Contains(t, fieldSet, mStruct.Primary())
					assert.Contains(t, fieldSet, mStruct.MustFieldByName("Integer"))
					assert.Contains(t, fieldSet, createdAt)
					assert.Contains(t, fieldSet, updatedAt)
				}
			}
			if assert.Len(t, s.Models, 2) {
				ms1, ok := s.Models[0].(*TestModel)
				require.True(t, ok)
				assert.Equal(t, m1, ms1)
				// Set up the ID field for the model.
				ms1.ID = 3

				assert.False(t, ms1.CreatedAt.IsZero())
				assert.False(t, ms1.UpdatedAt.IsZero())

				ms2, ok := s.Models[1].(*TestModel)
				require.True(t, ok)
				assert.Equal(t, m2, ms2)

				assert.False(t, ms2.CreatedAt.IsZero())
				assert.False(t, ms2.UpdatedAt.IsZero())
			}
			return nil
		})

		// Insert fields with only selected primary and 'Integer' field. This would result with common fieldset for many models.
		q := query.NewScope(mStruct, m1, m2)
		q.FieldSets = []mapping.FieldSet{
			{mStruct.MustFieldByName("FieldSetBefore")},
			{mStruct.Primary(), mStruct.MustFieldByName("Integer")},
		}
		err = db.InsertQuery(context.Background(), q)
		require.NoError(t, err)

		assert.Equal(t, 3, m1.ID)
	})

	t.Run("NoFieldSets", func(t *testing.T) {
		// No Primary key defined.
		m1 := &TestModel{
			FieldSetBefore: "not overwritten",
		}
		// Defined primary key and the 'Integer'.
		m2 := &TestModel{
			ID:      2,
			Integer: 2,
		}
		repo.OnInsert(func(_ context.Context, s *query.Scope) error {
			if assert.Len(t, s.Models, 2) {
				ms1, ok := s.Models[0].(*TestModel)
				require.True(t, ok)
				assert.Equal(t, m1, ms1)
				assert.False(t, ms1.CreatedAt.IsZero())
				assert.False(t, ms1.UpdatedAt.IsZero())
				assert.Equal(t, "not overwritten", ms1.FieldSetBefore)
				// Set up the ID field for the model.
				ms1.ID = 3

				ms2, ok := s.Models[1].(*TestModel)
				require.True(t, ok)
				assert.Equal(t, m2, ms2)
				assert.False(t, ms2.CreatedAt.IsZero())
				assert.False(t, ms2.UpdatedAt.IsZero())
				assert.Equal(t, "before", ms2.FieldSetBefore)
			}

			if assert.Len(t, s.FieldSets, 2) {
				fieldSet := s.FieldSets[0]
				if assert.Len(t, fieldSet, 3) {
					assert.Contains(t, fieldSet, mStruct.MustFieldByName("FieldSetBefore"))
					assert.Contains(t, fieldSet, createdAt)
					assert.Contains(t, fieldSet, updatedAt)
				}

				fieldSet = s.FieldSets[1]
				if assert.Len(t, fieldSet, 5) {
					assert.Contains(t, fieldSet, mStruct.Primary())
					assert.Contains(t, fieldSet, mStruct.MustFieldByName("Integer"))
					// BeforeInsert hook changes this field's value from zero - this should be added to fieldset.
					assert.Contains(t, fieldSet, mStruct.MustFieldByName("FieldSetBefore"))
					assert.Contains(t, fieldSet, createdAt)
					assert.Contains(t, fieldSet, updatedAt)
				}
			}

			return nil
		})

		err = db.Insert(context.Background(), mStruct, m1, m2)
		require.NoError(t, err)

		assert.Equal(t, 3, m1.ID)
		assert.Equal(t, "after", m1.FieldSetAfter)
		assert.Equal(t, "after", m2.FieldSetAfter)
	})
}
