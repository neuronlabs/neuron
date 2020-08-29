package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/core"
	"github.com/neuronlabs/neuron/internal/testmodels"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository/mockrepo"
)

func TestSavepoints(t *testing.T) {
	c := core.NewDefault()
	err := c.RegisterModels(testmodels.Neuron_Models...)
	require.NoError(t, err)

	repo := &mockrepo.Repository{IDValue: "first"}
	err = c.SetDefaultRepository(repo)
	require.NoError(t, err)

	repo2 := &mockrepo.Repository{IDValue: "second"}

	err = c.MapRepositoryModels(repo2, &testmodels.Blog{})
	require.NoError(t, err)

	err = c.RegisterRepositoryModels()
	require.NoError(t, err)

	db := New(c)

	postMStruct, err := c.ModelStruct(&testmodels.Post{})
	require.NoError(t, err)

	blogMStruct, err := c.ModelStruct(&testmodels.Blog{})
	require.NoError(t, err)

	ctx := context.Background()
	tx, err := Begin(ctx, db, nil)
	require.NoError(t, err)

	repo.OnInsert(func(ctx context.Context, s *query.Scope) error {
		return nil
	})

	var beginPost bool
	repo.OnBegin(func(_ context.Context, _ *query.Transaction) error {
		beginPost = true
		return nil
	})

	err = tx.Insert(ctx, postMStruct, &testmodels.Post{Title: "Name"})
	require.NoError(t, err)

	require.True(t, beginPost)

	var isSpFirst bool
	repo.OnSavepoint(func(_ context.Context, _ *query.Transaction, s string) error {
		assert.Equal(t, "first", s)
		isSpFirst = true
		return nil
	})

	err = tx.Savepoint(ctx, "first")
	require.NoError(t, err)

	require.True(t, isSpFirst)

	require.Len(t, tx.savePoints, 1)
	sp := tx.savePoints[0]
	assert.Equal(t, "first", sp.Name)
	assert.Len(t, sp.Transactions, 1)

	repo.OnInsert(func(_ context.Context, _ *query.Scope) error {
		return nil
	})
	err = tx.Insert(ctx, postMStruct, &testmodels.Post{Title: "Name2"})
	require.NoError(t, err)

	var isSecondSP bool
	repo.OnSavepoint(func(_ context.Context, _ *query.Transaction, s string) error {
		assert.Equal(t, "second", s)
		isSecondSP = true
		return nil
	})

	err = tx.Savepoint(ctx, "second")
	require.NoError(t, err)

	require.True(t, isSecondSP)

	require.Len(t, tx.savePoints, 2)
	sp2 := tx.savePoints[1]
	assert.Equal(t, "second", sp2.Name)
	assert.Len(t, sp2.Transactions, 1)

	var isBlogBegin, isBlogSP2 bool
	repo2.OnBegin(func(_ context.Context, _ *query.Transaction) error {
		isBlogBegin = true
		return nil
	})

	repo2.OnSavepoint(func(_ context.Context, _ *query.Transaction, s string) error {
		assert.Equal(t, "second", s)
		isBlogSP2 = true
		return nil
	})

	repo2.OnInsert(func(_ context.Context, _ *query.Scope) error {
		return nil
	})

	err = tx.Insert(ctx, blogMStruct, &testmodels.Blog{Title: "Blog"})
	require.NoError(t, err)

	require.True(t, isBlogBegin)
	require.True(t, isBlogSP2)

	assert.Len(t, sp2.Transactions, 2)

	var rollbackBlog, rollbackPostSP bool
	repo2.OnRollback(func(_ context.Context, _ *query.Transaction) error {
		rollbackBlog = true
		return nil
	})

	repo.OnRollbackSavepoint(func(_ context.Context, _ *query.Transaction, s string) error {
		assert.Equal(t, "first", s)
		rollbackPostSP = true
		return nil
	})

	err = tx.RollbackSavepoint(ctx, "first")
	require.NoError(t, err)

	require.True(t, rollbackBlog)
	require.True(t, rollbackPostSP)

	require.Len(t, tx.savePoints, 1)
	sp = tx.savePoints[0]
	assert.Equal(t, "first", sp.Name)
	assert.Len(t, sp.Transactions, 1)
}
