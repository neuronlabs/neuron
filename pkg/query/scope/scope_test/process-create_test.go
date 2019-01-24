package scope_test

import (
	"context"
	"errors"
	ctrl "github.com/kucjac/jsonapi/pkg/controller"
	"github.com/kucjac/jsonapi/pkg/internal/controller"
	"github.com/kucjac/jsonapi/pkg/internal/models"

	"github.com/kucjac/jsonapi/pkg/internal/repositories"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/jsonapi/pkg/query/scope"
	"github.com/kucjac/jsonapi/pkg/query/scope/mocks"
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	testCtxKey   string = "testCtxKey"
	errNotCalled error  = errors.New("Not called")
)

type createTestModel struct {
	ID int `jsonapi:"type=primary"`
}

type beforeCreateTestModel struct {
	ID int `jsonapi:"type=primary"`
}

type afterCreateTestModel struct {
	ID int `jsonapi:"type=primary"`
}

func (c *beforeCreateTestModel) WBeforeCreate(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func (c *afterCreateTestModel) WAfterCreate(s *scope.Scope) error {
	v := s.Context().Value(testCtxKey)
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
	repo := &mocks.Repository{}

	c := newController(t, repo)

	err := c.RegisterModels(&createTestModel{}, &beforeCreateTestModel{}, &afterCreateTestModel{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &createTestModel{})
		require.NoError(t, err)

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Create", mock.Anything).Return(nil)

		err = s.Create()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything)
		}

	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &beforeCreateTestModel{})
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))

		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)
		repo.On("Create", mock.Anything).Return(nil)

		err = s.Create()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything)
		}

	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := scope.NewWithC((*ctrl.Controller)(c), &afterCreateTestModel{})
		require.NoError(t, err)

		s.WithContext(context.WithValue(s.Context(), testCtxKey, t))
		r, _ := c.RepositoryByModel((*models.ModelStruct)(s.Struct()))

		repo = r.(*mocks.Repository)

		repo.On("Create", mock.Anything).Return(nil)
		err = s.Create()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything)
		}

	})

}

func newController(t testing.TB, repo repositories.Repository) *controller.Controller {
	t.Helper()
	cfg := &(*controller.DefaultConfig)
	cfg.DefaultRepository = repo.RepositoryName()

	c, err := controller.New(cfg, nil)
	require.NoError(t, err)

	err = c.RegisterRepository(repo)
	require.NoError(t, err)

	return c
}
