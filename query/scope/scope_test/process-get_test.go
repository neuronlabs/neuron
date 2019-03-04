package scope_test

import (
	"context"
	ctrl "github.com/kucjac/jsonapi/controller"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/jsonapi/query/scope/mocks"
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type beforeGetter struct {
	ID int `jsonapi:"type=primary"`
}

func (b *beforeGetter) HBeforeGet(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type afterGetter struct {
	ID int `jsonapi:"type=primary"`
}

func (a *afterGetter) HAfterGet(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type getter struct {
	ID int `jsonapi:"type=primary"`
}

func TestGet(t *testing.T) {
	if testing.Verbose() {
		err := log.SetLevel(unilogger.DEBUG)
		require.NoError(t, err)
	}

	repo := &mocks.Repository{}

	c := newController(t, repo)

	err := c.RegisterModels(&beforeGetter{}, &afterGetter{}, &getter{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &getter{})
		require.NoError(t, err)

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Get", mock.Anything).Return(nil)

		err = s.Get()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything)
		}
	})

	t.Run("BeforeGet", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &beforeGetter{})
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		require.NotNil(t, s.Value)

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Get", mock.Anything).Return(nil)

		err = s.Get()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything)
		}
	})

	t.Run("AfterGet", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &afterGetter{})
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Get", mock.Anything).Return(nil)

		err = s.Get()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything)
		}
	})
}
