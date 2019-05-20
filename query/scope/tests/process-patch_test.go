package tests

import (
	"context"
	"errors"
	"github.com/kucjac/uni-logger"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query/scope"
	"github.com/neuronlabs/neuron/query/scope/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type testPatcher struct {
	ID int `neuron:"type=primary"`
}

type testBeforePatcher struct {
	ID int `neuron:"type=primary"`
}

func (b *testBeforePatcher) HBeforePatch(ctx context.Context, s *scope.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type testAfterPatcher struct {
	ID int `neuron:"type=primary"`
}

func (a *testAfterPatcher) HAfterPatch(ctx context.Context, s *scope.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func TestPatch(t *testing.T) {
	if testing.Verbose() {
		err := log.SetLevel(unilogger.DEBUG)
		require.NoError(t, err)
	}

	repo := &mocks.Repository{}

	c := newController(t, repo)

	err := c.RegisterModels(&testPatcher{}, &testAfterPatcher{}, &testBeforePatcher{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := scope.NewC((*ctrl.Controller)(c), &testPatcher{})
		require.NoError(t, err)

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Return(nil)

		err = s.Patch()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := scope.NewC((*ctrl.Controller)(c), &testBeforePatcher{})
		require.NoError(t, err)

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Return(nil)

		err = s.PatchContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := scope.NewC((*ctrl.Controller)(c), &testAfterPatcher{})
		require.NoError(t, err)

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Return(nil)

		err = s.PatchContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
		}
	})

	t.Run("Transactions", func(t *testing.T) {

		// define the helper models
		type patchTMRelated struct {
			ID int `neuron:"type=primary"`
			FK int `neuron:"type=foreign"`
		}

		type patchTMRelations struct {
			ID  int             `neuron:"type=primary"`
			Rel *patchTMRelated `neuron:"type=relation;foreign=FK"`
		}

		repo := &mocks.Repository{}

		c := newController(t, repo)

		err := c.RegisterModels(&patchTMRelations{}, &patchTMRelated{})
		require.NoError(t, err)

		t.Run("Valid", func(t *testing.T) {

			tm := &patchTMRelations{
				ID: 2,
				Rel: &patchTMRelated{
					ID: 1,
				},
			}

			s, err := scope.NewC((*ctrl.Controller)(c), tm)
			require.NoError(t, err)

			r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))
			repo = r.(*mocks.Repository)

			// Begin define
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on patchTMRelations")
			}).Once().Return(nil)

			require.NoError(t, s.Begin())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.On("Patch", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Patch on patchTMRelations")
				s := a[1].(*scope.Scope)
				tx := s.Store[internal.TxStateCtxKey]
				assert.NotNil(t, tx)
			}).Once().Return(nil)

			// get the related model
			model, err := c.GetModelStruct(&patchTMRelated{})
			require.NoError(t, err)

			mr, ok := c.RepositoryByModel(model)
			require.True(t, ok)

			repo2, ok := mr.(*mocks.Repository)
			require.True(t, ok)

			repo2.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				log.Debug("Patch on patchTMRelated")
			}).Return(nil)

			repo2.On("Begin", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				log.Debug("Begin on patchTMRelations")
			}).Return(nil)

			require.NoError(t, s.Patch())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			// Commit
			repo.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Commit on patchTMRelations")
			}).Return(nil)
			repo2.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Commit on patchTMRelated")
			}).Return(nil)
			require.NoError(t, s.Commit())

			repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Commit", mock.Anything, mock.Anything)

		})

		t.Run("Rollback", func(t *testing.T) {
			tm := &patchTMRelations{
				ID: 2,
				Rel: &patchTMRelated{
					ID: 1,
				},
			}

			s, err := scope.NewC((*ctrl.Controller)(c), tm)
			require.NoError(t, err)
			r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

			// prepare the transaction
			repo = r.(*mocks.Repository)

			// Begin the transaction
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on patchTMRelations")
			}).Return(nil)

			require.NoError(t, s.Begin())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.On("Patch", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Patch on patchTMRelated")
			}).Once().Return(nil)

			model, err := c.GetModelStruct(&patchTMRelated{})
			require.NoError(t, err)

			m2Repo, ok := c.RepositoryByModel(model)
			require.True(t, ok)

			repo2, ok := m2Repo.(*mocks.Repository)
			require.True(t, ok)

			// Begin the transaction on subscope
			repo2.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on patchTMRelated")
			}).Return(nil)
			repo2.On("Patch", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Patch on patchTMRelated")
			}).Return(errors.New("Some error"))

			// Rollback the result
			repo2.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on patchTMRelated")
			}).Return(nil)
			repo.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on patchTMRelations")
			}).Return(nil)

			err = s.Patch()
			require.Error(t, err)

			// Assert calls

			repo2.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
		})
	})
}
