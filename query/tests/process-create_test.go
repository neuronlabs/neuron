package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/filters"
	"github.com/neuronlabs/neuron/query/mocks"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/controller"
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

func (c *beforeCreateTestModel) BeforeCreate(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func (c *afterCreateTestModel) AfterCreate(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errors.New("Not called")
	}

	return nil
}

// TestCreate tests the create Queries with default processor
func TestCreate(t *testing.T) {
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

}

// TestCreateTransactions tests the create process with transactions
func TestCreateTransactions(t *testing.T) {

	type foreignKeyModel struct {
		ID int `neuron:"type=primary"`
		FK int `neuron:"type=foreign"`
	}

	type hasOneModel struct {
		ID     int              `neuron:"type=primary"`
		HasOne *foreignKeyModel `neuron:"type=relation;foreign=FK"`
	}

	type hasManyModel struct {
		ID      int                `neuron:"type=primary"`
		HasMany []*foreignKeyModel `neuron:"type=relation;foreign=FK"`
	}

	t.Run("Valid", func(t *testing.T) {
		// Valid transaction should start with begin then the create and commit at the end
		t.Run("NoForeignRelationships", func(t *testing.T) {
			c := newController(t)
			err := c.RegisterModels(&foreignKeyModel{})
			require.NoError(t, err)

			model := &foreignKeyModel{
				FK: 24,
			}

			s, err := query.NewC((*ctrl.Controller)(c), model)
			require.NoError(t, err)

			fkModel, err := repository.GetRepository(c, &foreignKeyModel{})
			require.NoError(t, err)

			foreignKeyRepo, ok := fkModel.(*mocks.Repository)
			require.True(t, ok)

			// begin
			foreignKeyRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			// create
			foreignKeyRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				s := a[1].(*query.Scope)

				v := s.Value.(*foreignKeyModel)
				v.ID = 1
			}).Return(nil)
			// commit
			foreignKeyRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

			err = s.Begin()
			require.NoError(t, err)

			foreignKeyRepo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			err = s.Create()
			require.NoError(t, err)

			foreignKeyRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)

			err = s.Commit()
			require.NoError(t, err)
			foreignKeyRepo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)

		})
		t.Run("WithForeignRelationships", func(t *testing.T) {

			t.Run("HasOne", func(t *testing.T) {
				// valid transactions with the has one model should act like a normal valid create process
				// and as a subsequent process the begin, patch and commit on the related field should occur
				c := newController(t)
				err := c.RegisterModels(&hasOneModel{}, &foreignKeyModel{})
				require.NoError(t, err)

				tm := &hasOneModel{
					HasOne: &foreignKeyModel{
						ID: 1,
					},
				}

				s, err := query.NewC((*ctrl.Controller)(c), tm)
				require.NoError(t, err)
				r, _ := repository.GetRepository(s.Controller(), s.Struct())

				// get the repository for the has one model
				repo := r.(*mocks.Repository)

				repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
					log.Debug("Begin on hasOneModel")
				}).Return(nil)
				repo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
					log.Debug("Create on hasOneModel")
					s := a[1].(*query.Scope)
					_, ok := s.StoreGet(internal.TxStateStoreKey)
					assert.True(t, ok)

					sv := s.Value.(*hasOneModel)
					sv.ID = 1
				}).Return(nil)
				// Do the Commit
				repo.On("Commit", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					log.Debug("Commit on hasOneModel")
				}).Return(nil)

				// do the create

				repo2, err := repository.GetRepository(s.Controller(), &foreignKeyModel{})
				require.NoError(t, err)

				fkModelRepo, ok := repo2.(*mocks.Repository)
				require.True(t, ok)

				fkModelRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				fkModelRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					log.Debug("Patch on foreignKeyModel")
				}).Return(nil)
				fkModelRepo.On("Commit", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					log.Debug("Commit on foreignKeyModel")
				}).Return(nil)

				// Begin the transaction
				err = s.Begin()
				require.NoError(t, err)

				repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

				// Create the scope
				require.NoError(t, s.Create())

				repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
				// foreignkey model should begin also
				fkModelRepo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
				fkModelRepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

				// Commit the results
				require.NoError(t, s.Commit())

				// Assert Called Commit
				fkModelRepo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
				repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)

			})

			t.Run("HasMany", func(t *testing.T) {
				// the valid hasmany transaction should begin as the valid create transactioned process
				// and as the subsequent step it should patch all the related models that were in the relationship of the root model
				c := newController(t)
				err := c.RegisterModels(&hasManyModel{}, &foreignKeyModel{})
				require.NoError(t, err)

				m1 := hasManyModel{
					HasMany: []*foreignKeyModel{
						{
							ID: 2,
						},
						{
							ID: 4,
						},
					},
				}

				// get the scope for the has many model
				s, err := query.NewC((*ctrl.Controller)(c), &m1)
				require.NoError(t, err)

				// get the repositories for the both models in order to mock it
				hmModel, err := repository.GetRepository(c, &m1)
				require.NoError(t, err)

				hasManyRepo, ok := hmModel.(*mocks.Repository)
				require.True(t, ok)

				fkModel, err := repository.GetRepository(c, &foreignKeyModel{})
				require.NoError(t, err)

				foreignKeyRepo, ok := fkModel.(*mocks.Repository)
				require.True(t, ok)

				// must begin
				hasManyRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				// needs to create
				hasManyRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					// create and set the ID
					s := a[1].(*query.Scope)
					model := s.Value.(*hasManyModel)
					model.ID = 1
				}).Return(nil)
				// allow to commit
				hasManyRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				// foreign key repo should at first clear the old fk's, get the begin, patch, commit processes
				// while it would get the filters on the FK then it should list the patched id first
				foreignKeyRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				// then it should reduce the filters into primaries
				foreignKeyRepo.On("List", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					s := a[1].(*query.Scope)
					log.Debugf("Primary Filters: %v", s.PrimaryFilters())
					log.Debugf("Foreign Filters: %v", s.ForeignFilters())
					models := (s.Value.(*[]*foreignKeyModel))
					(*models) = append((*models), &foreignKeyModel{ID: 3}, &foreignKeyModel{ID: 5}, &foreignKeyModel{ID: 6})
				}).Return(nil)
				// first patch would clear the foreign keys
				foreignKeyRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					s := a[1].(*query.Scope)
					model := s.Value.(*foreignKeyModel)
					assert.Equal(t, 0, model.FK)
					primaryFilters := s.PrimaryFilters()

					if assert.Len(t, primaryFilters, 1) {
						if assert.Len(t, primaryFilters[0].Values(), 1) {
							v := primaryFilters[0].Values()[0]
							assert.Equal(t, filters.OpIn, v.Operator())
							assert.Contains(t, v.Values, 3)
							assert.Contains(t, v.Values, 5)
							assert.Contains(t, v.Values, 6)

						}
					}
				}).Return(nil)
				foreignKeyRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					s := a[1].(*query.Scope)
					model := s.Value.(*foreignKeyModel)
					assert.Equal(t, m1.ID, model.FK)
					primaryFilters := s.PrimaryFilters()
					if assert.Len(t, primaryFilters, 1) {
						if assert.Len(t, primaryFilters[0].Values(), 1) {
							v := primaryFilters[0].Values()[0]
							assert.Equal(t, filters.OpIn, v.Operator())

							assert.Contains(t, v.Values, 2)
							assert.Contains(t, v.Values, 4)
						}
					}
				}).Return(nil)
				// double commits - first clears the second adds the foreign keys
				foreignKeyRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				foreignKeyRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				err = s.Create()
				require.NoError(t, err)

			})

			t.Run("Many2Many", func(t *testing.T) {

			})

		})
	})

	t.Run("Rollbacked", func(t *testing.T) {

		t.Run("ByRoot", func(t *testing.T) {

		})

		t.Run("ByForeignRelationships", func(t *testing.T) {

			t.Run("HasOne", func(t *testing.T) {
				c := newController(t)
				err := c.RegisterModels(&hasOneModel{}, &foreignKeyModel{})
				require.NoError(t, err)

				tm := &hasOneModel{
					ID: 2,
					HasOne: &foreignKeyModel{
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
					log.Debug("Begin on hasOneModel")
				}).Return(nil)

				require.NoError(t, s.Begin())

				repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

				repo.On("Create", mock.Anything, mock.Anything).Return(nil)

				model, err := c.GetModelStruct(&foreignKeyModel{})
				require.NoError(t, err)

				m2Repo, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
				require.NoError(t, err)

				repo2, ok := m2Repo.(*mocks.Repository)
				require.True(t, ok)

				// Begin the transaction on subscope
				repo2.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
					log.Debug("Begin on foreignKeyModel")
				}).Return(nil)
				repo2.On("Patch", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
					log.Debug("Patch on foreignKeyModel")
				}).Return(errors.New("Some error"))

				// Rollback the result
				repo2.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
					log.Debug("Rollback on foreignKeyModel")
				}).Return(nil)
				repo.On("Rollback", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
					log.Debug("Rollback on hasOneModel")
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

			t.Run("HasMany", func(t *testing.T) {

			})

			t.Run("Many2Many", func(t *testing.T) {

			})
		})
	})
}

func newController(t testing.TB) *controller.Controller {
	t.Helper()

	c := controller.DefaultTesting(t, nil)
	if testing.Verbose() {
		log.SetLevel(log.LDEBUG2)
	}

	return c
}
