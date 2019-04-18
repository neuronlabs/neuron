package scope_test

import (
	"context"
	"github.com/kucjac/uni-logger"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query/scope"
	"github.com/neuronlabs/neuron/query/scope/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type beforeLister struct {
	ID int `neuron:"type=primary"`
}

func (b *beforeLister) HBeforeList(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type afterLister struct {
	ID int `neuron:"type=primary"`
}

func (a *afterLister) HAfterList(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type lister struct {
	ID int `neuron:"type=primary"`
}

func TestList(t *testing.T) {
	if testing.Verbose() {
		err := log.SetLevel(unilogger.DEBUG)
		require.NoError(t, err)
	}

	repo := &mocks.Repository{}

	c := newController(t, repo)

	err := c.RegisterModels(&beforeLister{}, &afterLister{}, &lister{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		v := []*lister{}
		s, err := scope.NewWithC((*ctrl.Controller)(c), &v)
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("List", mock.Anything).Return(nil)

		err = s.List()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		v := []*beforeLister{}
		s, err := scope.NewWithC((*ctrl.Controller)(c), &v)
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("List", mock.Anything).Return(nil)

		err = s.List()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		v := []*afterLister{}
		s, err := scope.NewWithC((*ctrl.Controller)(c), &v)
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("List", mock.Anything).Return(nil)

		err = s.List()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything)
		}
	})
}
