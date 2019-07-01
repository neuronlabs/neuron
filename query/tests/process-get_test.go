package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/mocks"
)

type beforeGetter struct {
	ID int `neuron:"type=primary"`
}

func (b *beforeGetter) BeforeGet(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type afterGetter struct {
	ID int `neuron:"type=primary"`
}

func (a *afterGetter) AfterGet(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type getter struct {
	ID int `neuron:"type=primary"`
}

func TestGet(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&beforeGetter{}, &afterGetter{}, &getter{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &getter{})
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Get", mock.Anything, mock.Anything).Return(nil)

		err = s.Get()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything, mock.Anything)
		}
	})

	t.Run("BeforeGet", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &beforeGetter{})
		require.NoError(t, err)

		require.NotNil(t, s.Value)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Get", mock.Anything, mock.Anything).Return(nil)

		err = s.GetContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything, mock.Anything)
		}
	})

	t.Run("AfterGet", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &afterGetter{})
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Get", mock.Anything, mock.Anything).Return(nil)

		err = s.GetContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything, mock.Anything)
		}
	})
}
