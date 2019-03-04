package scope_test

import (
	"context"
	ctrl "github.com/kucjac/jsonapi/pkg/controller"
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/jsonapi/pkg/query/scope"
	"github.com/kucjac/jsonapi/pkg/query/scope/mocks"
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type testDeleter struct {
	ID int `jsonapi:"type=primary"`
}

type testBeforeDeleter struct {
	ID int `jsonapi:"type=primary"`
}

func (b *testBeforeDeleter) HBeforeDelete(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type testAfterDeleter struct {
	ID int `jsonapi:"type=primary"`
}

func (a *testAfterDeleter) HAfterDelete(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
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

	repo := &mocks.Repository{}

	c := newController(t, repo)

	err := c.RegisterModels(&testDeleter{}, &testAfterDeleter{}, &testBeforeDeleter{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &testDeleter{})
		require.NoError(t, err)

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Delete", mock.Anything).Return(nil)

		err = s.Delete()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &testBeforeDeleter{})
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Delete", mock.Anything).Return(nil)

		err = s.Delete()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &testAfterDeleter{})
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Delete", mock.Anything).Return(nil)

		err = s.Delete()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything)
		}
	})
}
