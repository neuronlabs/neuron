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
	"github.com/neuronlabs/neuron/query/filters"
	"github.com/neuronlabs/neuron/query/mocks"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal"
)

type testPatcher struct {
	ID    int    `neuron:"type=primary"`
	Field string `neuron:"type=attr"`
}

type testBeforePatcher struct {
	ID int `neuron:"type=primary"`
}

func (b *testBeforePatcher) BeforePatch(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type testAfterPatcher struct {
	ID int `neuron:"type=primary"`
}

func (a *testAfterPatcher) AfterPatch(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func TestPatch(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&testPatcher{}, &testAfterPatcher{}, &testBeforePatcher{})
	require.NoError(t, err)

	t.Run("NoSelectedValues", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testPatcher{ID: 5})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Patch()
		if assert.NoError(t, err) {
			repo.AssertNotCalled(t, "Patch", mock.Anything, mock.Anything)
		}
	})

	t.Run("NoHooks", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testPatcher{ID: 5, Field: "Something"})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Patch()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testBeforePatcher{ID: 1})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)

		require.NoError(t, s.PatchContext(context.WithValue(context.Background(), testCtxKey, t)))

	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testAfterPatcher{ID: 2})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Return(nil)

		require.NoError(t, s.PatchContext(context.WithValue(context.Background(), testCtxKey, t)))

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

		c := newController(t)

		err := c.RegisterModels(&patchTMRelations{}, &patchTMRelated{})
		require.NoError(t, err)

		t.Run("Valid", func(t *testing.T) {

			tm := &patchTMRelations{
				ID: 2,
				Rel: &patchTMRelated{
					ID: 1,
				},
			}

			s, err := query.NewC((*controller.Controller)(c), tm)
			require.NoError(t, err)

			r, _ := repository.GetRepository(s.Controller(), s.Struct())
			repo := r.(*mocks.Repository)

			// Begin define
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on patchTMRelations")
			}).Once().Return(nil)

			require.NoError(t, s.Begin())

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.On("Patch", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Patch on patchTMRelations")
				s := a[1].(*query.Scope)
				_, ok := s.StoreGet(internal.TxStateStoreKey)
				assert.True(t, ok)
			}).Once().Return(nil)

			// get the related model
			model, err := c.GetModelStruct(&patchTMRelated{})
			require.NoError(t, err)

			mr, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
			require.NoError(t, err)

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

			s, err := query.NewC((*controller.Controller)(c), tm)
			require.NoError(t, err)
			r, _ := repository.GetRepository(s.Controller(), s.Struct())

			// prepare the transaction
			repo := r.(*mocks.Repository)

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

			m2Repo, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
			require.NoError(t, err)

			repo2, ok := m2Repo.(*mocks.Repository)
			require.True(t, ok)

			// Begin the transaction on subquery
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

			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			repo.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
		})
	})

	t.Run("Relationships", func(t *testing.T) {
		c := newController(t)

		err := c.RegisterModels(HasOneModel{}, HasManyModel{}, ForeignModel{}, Many2ManyModel{}, JoinModel{}, RelatedModel{})
		require.NoError(t, err)

		t.Run("HasOne", func(t *testing.T) {
			t.Run("Valid", func(t *testing.T) {
				// patch the model
				model := &HasOneModel{
					ID:     3,
					HasOne: &ForeignModel{ID: 5},
				}

				s, err := query.NewC((*controller.Controller)(c), model)
				require.NoError(t, err)

				hasOneRepo, err := repository.GetRepository((*controller.Controller)(c), model)
				require.NoError(t, err)

				repo, ok := hasOneRepo.(*mocks.Repository)
				require.True(t, ok)

				foreignRepo, err := repository.GetRepository((*controller.Controller)(c), model.HasOne)
				require.NoError(t, err)

				frepo, ok := foreignRepo.(*mocks.Repository)
				require.True(t, ok)

				repo.On("Patch", mock.Anything, mock.Anything).Once().Return(errors.New("Should not be called"))

				frepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(arg mock.Arguments) {
					s, ok := arg[1].(*query.Scope)
					require.True(t, ok)

					pm := s.PrimaryFilters()
					if assert.NotEmpty(t, pm) {
						if assert.Len(t, pm, 1) {
							if assert.Len(t, pm[0].Values(), 1) {
								v := pm[0].Values()[0]
								assert.Equal(t, filters.OpIn, v.Operator())
								assert.Equal(t, model.HasOne.ID, v.Values[0])
							}
						}
					}

				}).Return(nil)

				require.NoError(t, s.Patch())

				repo.AssertNotCalled(t, "Patch", mock.Anything, mock.Anything)
				frepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
			})
		})
		t.Run("HasMany", func(t *testing.T) {
			// TODO: patch hasMany tests
		})
		t.Run("Many2Many", func(t *testing.T) {
			// TODO: patch many2many tests
		})
	})
}
