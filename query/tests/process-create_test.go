package tests

import (
	"context"
	"errors"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/repository"

	"github.com/kucjac/uni-logger"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	testCtxKey   = testKeyStruct{}
	errNotCalled = errors.New("Not called")
)

type testKeyStruct struct{}

type createTestModel struct {
	ID int `neuron:"type=primary"`
}

type beforeCreateTestModel struct {
	ID int `neuron:"type=primary"`
}

var _ query.BeforeCreator = &beforeCreateTestModel{}

type afterCreateTestModel struct {
	ID int `neuron:"type=primary"`
}

var _ query.AfterCreator = &afterCreateTestModel{}

func (c *beforeCreateTestModel) HBeforeCreate(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func (c *afterCreateTestModel) HAfterCreate(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errors.New("Not called")
	}

	return nil
}

func TestCreate(t *testing.T) {
	if testing.Verbose() {
		err := log.SetLevel(unilogger.DEBUG)
		require.NoError(t, err)
	}

	c := newController(t)

	err := c.RegisterModels(&createTestModel{}, &beforeCreateTestModel{}, &afterCreateTestModel{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := query.NewC((*ctrl.Controller)(c), &createTestModel{})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Create()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		}

	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := query.NewC((*ctrl.Controller)(c), &beforeCreateTestModel{})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)
		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.CreateContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		}

	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := query.NewC((*ctrl.Controller)(c), &afterCreateTestModel{})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)
		err = s.CreateContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		}

	})

	t.Run("Transactions", func(t *testing.T) {

		type createTMRelated struct {
			ID int `neuron:"type=primary"`
			FK int `neuron:"type=foreign"`
		}

		type createTMRelations struct {
			ID  int              `neuron:"type=primary"`
			Rel *createTMRelated `neuron:"type=relation;foreign=FK"`
		}

		c := newController(t)
		err := c.RegisterModels(&createTMRelations{}, &createTMRelated{})
		require.NoError(t, err)

		t.Run("Valid", func(t *testing.T) {
			tm := &createTMRelations{
				ID: 2,
				Rel: &createTMRelated{
					ID: 1,
				},
			}

			s, err := query.NewC((*ctrl.Controller)(c), tm)
			require.NoError(t, err)
			r, _ := repository.GetRepository(s.Controller(), s.Struct())

			repo := r.(*mocks.Repository)

			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on createTMRelations")
			}).Return(nil)
			require.NoError(t, s.Begin())

			repo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Create on createTMRelations")
				s := a[1].(*query.Scope)
				tx := s.Store[internal.TxStateCtxKey]
				assert.NotNil(t, tx)
			}).Return(nil)

			// r, _ := repository.GetRepository(s.Controller(),s.Struct())

			model, err := c.GetModelStruct(&createTMRelated{})
			require.NoError(t, err)

			repo2, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
			require.NoError(t, err)

			mr, ok := repo2.(*mocks.Repository)
			require.True(t, ok)

			mr.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				log.Debug("Patch on createTMRelated")
			}).Return(nil)
			mr.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

			require.NoError(t, s.Create())

			// Assert the calls for the repository
			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			mr.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
			mr.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			// Do the Commit
			repo.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Commit on createTMRelations")
			}).Return(nil)
			mr.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Commit on createTMRelated")
			}).Return(nil)

			require.NoError(t, s.Commit())

			// Assert Called Commit
			mr.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)

		})

		t.Run("Rollback", func(t *testing.T) {
			tm := &createTMRelations{
				ID: 2,
				Rel: &createTMRelated{
					ID: 1,
				},
			}

			s, err := query.NewC((*ctrl.Controller)(c), tm)
			require.NoError(t, err)
			r, _ := repository.GetRepository(s.Controller(), s.Struct())

			// prepare the transaction
			repo := r.(*mocks.Repository)

			// Begin the transaction
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on createTMRelations")
			}).Return(nil)

			require.NoError(t, s.Begin())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.On("Create", mock.Anything, mock.Anything).Return(nil)

			model, err := c.GetModelStruct(&createTMRelated{})
			require.NoError(t, err)

			m2Repo, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
			require.NoError(t, err)

			repo2, ok := m2Repo.(*mocks.Repository)
			require.True(t, ok)

			// Begin the transaction on subscope
			repo2.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on createTMRelated")
			}).Return(nil)
			repo2.On("Patch", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Patch on createTMRelated")
			}).Return(errors.New("Some error"))

			// Rollback the result
			repo2.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on createTMRelated")
			}).Return(nil)
			repo.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Rollback on createTMRelations")
			}).Return(nil)

			err = s.Create()
			require.Error(t, err)

			// Assert calls

			repo2.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)

		})

	})

}

func newController(t testing.TB) *controller.Controller {
	t.Helper()

	c := controller.DefaultTesting(t, nil)

	return c
}
