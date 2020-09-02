package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestTransactions(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{}

	db, err := New(WithDefaultRepository(repo), WithModelMap(mm))
	require.NoError(t, err)

	mStruct, err := mm.ModelStruct(&testmodels.HasOneModel{})
	require.NoError(t, err)

	t.Run("Commit", func(t *testing.T) {
		tx := db.Begin(context.Background(), nil)
		var begin, commit bool
		repo.OnBegin(func(_ context.Context, transaction *query.Transaction) error {
			assert.NotEqual(t, transaction.ID, uuid.Nil)
			begin = true
			return nil
		})

		repo.OnFind(func(_ context.Context, s *query.Scope) error {
			assert.Equal(t, mStruct, s.ModelStruct)
			s.Models = []mapping.Model{
				&testmodels.HasOneModel{ID: 2},
			}
			return nil
		})
		repo.OnCommit(func(_ context.Context, transaction *query.Transaction) error {
			assert.NotEqual(t, transaction.ID, uuid.Nil)
			commit = true
			return nil
		})

		_, err = tx.Query(mStruct).Find()
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		assert.True(t, begin)
		assert.True(t, commit)

		err = tx.Commit()
		require.Error(t, err)
	})

	t.Run("Rollback", func(t *testing.T) {
		tx := db.Begin(context.Background(), nil)
		var begin, rollback bool
		repo.OnBegin(func(_ context.Context, transaction *query.Transaction) error {
			assert.NotEqual(t, transaction.ID, uuid.Nil)
			begin = true
			return nil
		})

		repo.OnFind(func(_ context.Context, s *query.Scope) error {
			assert.Equal(t, mStruct, s.ModelStruct)
			s.Models = []mapping.Model{
				&testmodels.HasOneModel{ID: 2},
			}
			return nil
		})
		repo.OnRollback(func(_ context.Context, transaction *query.Transaction) error {
			assert.NotEqual(t, transaction.ID, uuid.Nil)
			rollback = true
			return nil
		})

		_, err = tx.Query(mStruct).Find()
		require.NoError(t, err)

		err = tx.Rollback()
		require.NoError(t, err)

		assert.True(t, begin)
		assert.True(t, rollback)

		err = tx.Rollback()
		require.Error(t, err)
	})
}

func TestTransactionMultiRepo(t *testing.T) {
	mm := mapping.New()
	err := mm.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{IDValue: "first"}
	repo2 := &mockrepo.Repository{IDValue: "second"}

	db, err := New(
		WithModelMap(mm),
		WithDefaultRepository(repo),
		WithRepositoryModels(repo2, &testmodels.Blog{}),
	)
	require.NoError(t, err)

	postMStruct, err := mm.ModelStruct(&testmodels.Post{})
	require.NoError(t, err)

	blogMStruct, err := mm.ModelStruct(&testmodels.Blog{})
	require.NoError(t, err)

	tx := db.Begin(context.Background(), nil)
	var begin, begin2, commit, commit2 bool
	repo.OnBegin(func(_ context.Context, transaction *query.Transaction) error {
		assert.NotEqual(t, transaction.ID, uuid.Nil)
		begin = true
		return nil
	})

	repo.OnFind(func(_ context.Context, s *query.Scope) error {
		assert.Equal(t, postMStruct, s.ModelStruct)
		s.Models = []mapping.Model{
			&testmodels.Post{ID: 2},
		}
		return nil
	})
	repo.OnCommit(func(_ context.Context, transaction *query.Transaction) error {
		assert.NotEqual(t, transaction.ID, uuid.Nil)
		commit = true
		return nil
	})

	repo2.OnBegin(func(_ context.Context, transaction *query.Transaction) error {
		assert.NotEqual(t, transaction.ID, uuid.Nil)
		begin2 = true
		return nil
	})

	repo2.OnUpdateModels(func(_ context.Context, s *query.Scope) (int64, error) {
		return 1, nil
	})
	repo2.OnCommit(func(_ context.Context, transaction *query.Transaction) error {
		assert.NotEqual(t, transaction.ID, uuid.Nil)
		commit2 = true
		return nil
	})

	_, err = tx.Query(postMStruct).Find()
	require.NoError(t, err)

	_, err = tx.Query(blogMStruct, &testmodels.Blog{ID: 10, CurrentPostID: 2}).Update()
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	assert.True(t, begin)
	assert.True(t, commit)
	assert.True(t, begin2)
	assert.True(t, commit2)

	err = tx.Commit()
	require.Error(t, err)
}
