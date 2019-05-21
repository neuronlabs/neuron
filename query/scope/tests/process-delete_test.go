package tests

import (
	"context"
	"errors"
	"github.com/kucjac/uni-logger"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/scope"
	"github.com/neuronlabs/neuron/query/scope/mocks"
	"github.com/neuronlabs/neuron/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type testDeleter struct {
	ID int `neuron:"type=primary"`
}

type testBeforeDeleter struct {
	ID int `neuron:"type=primary"`
}

func (b *testBeforeDeleter) HBeforeDelete(ctx context.Context, s *scope.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type testAfterDeleter struct {
	ID int `neuron:"type=primary"`
}

func (a *testAfterDeleter) HAfterDelete(ctx context.Context, s *scope.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func TestDelete(t *testing.T) {
	if testing.Verbose() {
		err := log.SetLevel(unilogger.DEBUG)
		require.NoError(t, err)
	}

	c := newController(t)

	err := c.RegisterModels(&testDeleter{}, &testAfterDeleter{}, &testBeforeDeleter{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := scope.NewC((*ctrl.Controller)(c), &testDeleter{})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Delete", mock.Anything, mock.Anything).Return(nil)

		err = s.Delete()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := scope.NewC((*ctrl.Controller)(c), &testBeforeDeleter{})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Delete", mock.Anything, mock.Anything).Return(nil)

		err = s.DeleteContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := scope.NewC((*ctrl.Controller)(c), &testAfterDeleter{})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Delete", mock.Anything, mock.Anything).Return(nil)

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

			tm := &deleteTMRelations{
				ID: 2,
				Rel: &deleteTMRelated{
					ID: 1,
				},
			}

			s, err := scope.NewC((*ctrl.Controller)(c), tm)
			require.NoError(t, err)

			r, _ := repository.GetRepository(s.Controller(), s.Struct())
			repo := r.(*mocks.Repository)

			// Begin define
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on deleteTMRelations")
			}).Once().Return(nil)

			require.NoError(t, s.Begin())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.On("Delete", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Delete on deleteTMRelations")
				s := a[1].(*scope.Scope)

				tx := s.Store[internal.TxStateCtxKey]
				assert.NotNil(t, tx)
			}).Once().Return(nil)

			// get the related model
			model, err := c.GetModelStruct(&deleteTMRelated{})
			require.NoError(t, err)

			mr, err := repository.GetRepository(s.Controller(), ((*mapping.ModelStruct)(model)))
			require.NoError(t, err)

			repo2, ok := mr.(*mocks.Repository)
			require.True(t, ok)

			repo2.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				log.Debug("Patch on deleteTMRelated")
			}).Return(nil)

			repo2.On("Begin", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				log.Debug("Begin on deleteTMRelated")
			}).Return(nil)

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

			s, err := scope.NewC((*ctrl.Controller)(c), tm)
			require.NoError(t, err)
			r, _ := repository.GetRepository(s.Controller(), s.Struct())

			// prepare the transaction
			repo := r.(*mocks.Repository)

			// Begin the transaction
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on deleteTMRelations")
			}).Return(nil)

			require.NoError(t, s.Begin())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.On("Delete", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Delete on deleteTMRelations")
			}).Once().Return(nil)

			model, err := c.GetModelStruct(&deleteTMRelated{})
			require.NoError(t, err)

			m2Repo, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
			require.NoError(t, err)

			repo2, ok := m2Repo.(*mocks.Repository)
			require.True(t, ok)

			// Begin the transaction on subscope
			repo2.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on deleteTMRelated")
			}).Return(nil)
			repo2.On("Patch", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Patch on deleteTMRelated")
			}).Return(errors.New("Some error"))

			// Rollback the result
			repo2.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on deleteTMRelated")
			}).Return(nil)
			repo.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on deleteTMRelations")
			}).Return(nil)

			err = s.Delete()
			require.Error(t, err)

			// Assert calls

			repo2.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
		})

	})
}
