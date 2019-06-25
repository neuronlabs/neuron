package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/mocks"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal"
)

type testDeleter struct {
	ID int `neuron:"type=primary"`
}

type testBeforeDeleter struct {
	ID int `neuron:"type=primary"`
}

func (b *testBeforeDeleter) BeforeDelete(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type testAfterDeleter struct {
	ID int `neuron:"type=primary"`
}

func (a *testAfterDeleter) AfterDelete(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

// TestDelete test the Delete Processes.
func TestDelete(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&testDeleter{}, &testAfterDeleter{}, &testBeforeDeleter{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testDeleter{ID: 1})
		require.NoError(t, err)

		r, err := repository.GetRepository(s.Controller(), s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*mocks.Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Delete()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testBeforeDeleter{ID: 1})
		require.NoError(t, err)

		r, err := repository.GetRepository(s.Controller(), s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*mocks.Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.DeleteContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testAfterDeleter{ID: 1})
		require.NoError(t, err)

		r, err := repository.GetRepository(s.Controller(), s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*mocks.Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.DeleteContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("Transactions", func(t *testing.T) {
		// define the helper models
		type deleteTMRelated struct {
			ID int `neuron:"type=primary"`
			FK int `neuron:"type=foreign"`
		}

		type deleteTMRelations struct {
			ID  int              `neuron:"type=primary"`
			Rel *deleteTMRelated `neuron:"type=relation;foreign=FK"`
		}

		c := newController(t)

		err := c.RegisterModels(&deleteTMRelations{}, &deleteTMRelated{})
		require.NoError(t, err)

		t.Run("Valid", func(t *testing.T) {

			tm := &deleteTMRelations{ID: 2}

			s, err := query.NewC((*controller.Controller)(c), tm)
			require.NoError(t, err)

			r, _ := repository.GetRepository(s.Controller(), s.Struct())
			repo := r.(*mocks.Repository)

			defer clearRepository(repo)

			// Begin define
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on deleteTMRelations")
			}).Once().Return(nil)

			_, err = s.Begin()
			require.NoError(t, err)

			repo.On("Delete", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Delete on deleteTMRelations")
				s := a[1].(*query.Scope)

				_, ok := s.StoreGet(internal.TxStateStoreKey)
				assert.True(t, ok)
			}).Once().Return(nil)

			// get the related model
			model, err := c.GetModelStruct(&deleteTMRelated{})
			require.NoError(t, err)

			mr, err := repository.GetRepository(s.Controller(), ((*mapping.ModelStruct)(model)))
			require.NoError(t, err)

			repo2, ok := mr.(*mocks.Repository)
			require.True(t, ok)

			defer clearRepository(repo2)

			repo2.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				log.Debug("Patch on deleteTMRelated")
			}).Return(nil)

			repo2.On("Begin", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				log.Debug("Begin on deleteTMRelated")
			}).Return(nil)

			repo2.On("List", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				s := a[1].(*query.Scope)
				values := []*deleteTMRelated{{ID: 2}}
				s.Value = &values
			}).Once().Return(nil)

			require.NoError(t, s.Delete())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			// Commit
			repo.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Commit on deleteTMRelations")
			}).Return(nil)
			repo2.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Commit on deleteTMRelated")
			}).Return(nil)
			require.NoError(t, s.Commit())

			repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Commit", mock.Anything, mock.Anything)

		})

		t.Run("Rollback", func(t *testing.T) {
			tm := &deleteTMRelations{
				ID: 2,
				Rel: &deleteTMRelated{
					ID: 1,
				},
			}

			s, err := query.NewC((*controller.Controller)(c), tm)
			require.NoError(t, err)
			r, _ := repository.GetRepository(s.Controller(), s.Struct())

			// prepare the transaction
			repo := r.(*mocks.Repository)
			defer clearRepository(repo)

			// Begin the transaction
			repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

			_, err = s.Begin()
			require.NoError(t, err)

			repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)

			model, err := c.GetModelStruct(&deleteTMRelated{})
			require.NoError(t, err)

			m2Repo, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
			require.NoError(t, err)

			repo2, ok := m2Repo.(*mocks.Repository)
			require.True(t, ok)

			clearRepository(repo2)

			// Begin the transaction on subquery
			repo2.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

			repo2.On("List", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				s := a[1].(*query.Scope)
				values := []*deleteTMRelated{{ID: 1}}
				s.Value = &values
			}).Return(nil)

			repo2.On("Patch", mock.Anything, mock.Anything).Once().Return(errors.New("Some error"))

			// Rollback the result
			repo2.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on deleteTMRelated")
			}).Return(nil)

			repo.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on deleteTMRelations")
			}).Return(nil)

			err = s.Delete()
			require.Error(t, err)

			err = s.Rollback()
			assert.NoError(t, err)

			// Assert calls
			repo2.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
		})
	})
}
