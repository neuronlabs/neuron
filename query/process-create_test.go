package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal"
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

var _ BeforeCreator = &beforeCreateTestModel{}

type afterCreateTestModel struct {
	ID int `neuron:"type=primary"`
}

var _ AfterCreator = &afterCreateTestModel{}

func (c *beforeCreateTestModel) BeforeCreate(ctx context.Context, s *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func (c *afterCreateTestModel) AfterCreate(ctx context.Context, s *Scope) error {
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
		s, err := NewC(c, &createTestModel{})
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Create()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		}

	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := NewC(c, &beforeCreateTestModel{})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)
		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.CreateContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		}

	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := NewC(c, &afterCreateTestModel{})
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*Repository)
		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

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

			s, err := NewC(c, model)
			require.NoError(t, err)

			fkModel, err := c.GetRepository(&foreignKeyModel{})
			require.NoError(t, err)

			foreignKeyRepo, ok := fkModel.(*Repository)
			require.True(t, ok)

			defer clearRepository(foreignKeyRepo)

			// begin
			foreignKeyRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			// create
			foreignKeyRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				s := a[1].(*Scope)

				v := s.Value.(*foreignKeyModel)
				v.ID = 1
			}).Return(nil)
			// commit
			foreignKeyRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

			_, err = s.Begin()
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

				s, err := NewC(c, tm)
				require.NoError(t, err)
				r, err := s.Controller().GetRepository(s.Struct())
				require.NoError(t, err)

				// get the repository for the has one model
				repo, ok := r.(*Repository)
				require.True(t, ok)

				defer clearRepository(repo)

				repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				repo.On("Create", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					s := a[1].(*Scope)
					_, ok := s.StoreGet(internal.TxStateStoreKey)
					assert.True(t, ok)

					sv := s.Value.(*hasOneModel)
					sv.ID = 1
				}).Return(nil)
				repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				// do the create

				repo2, err := s.Controller().GetRepository(&foreignKeyModel{})
				require.NoError(t, err)

				fkModelRepo, ok := repo2.(*Repository)
				require.True(t, ok)

				defer clearRepository(fkModelRepo)

				fkModelRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				fkModelRepo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)
				fkModelRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				// Create the scope
				require.NoError(t, s.Create())

				repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
				repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)

				fkModelRepo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
				fkModelRepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
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
						{ID: 2},
						{ID: 4},
					},
				}

				// get the scope for the has many model
				s, err := NewC(c, &m1)
				require.NoError(t, err)

				// get the repositories for the both models in order to mock it
				hmModel, err := c.GetRepository(&m1)
				require.NoError(t, err)

				hasManyRepo, ok := hmModel.(*Repository)
				require.True(t, ok)

				defer clearRepository(hasManyRepo)

				fkModel, err := c.GetRepository(&foreignKeyModel{})
				require.NoError(t, err)

				foreignKeyRepo, ok := fkModel.(*Repository)
				require.True(t, ok)

				defer clearRepository(foreignKeyRepo)

				// must begin
				hasManyRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				// needs to create
				hasManyRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					// create and set the ID
					s := a[1].(*Scope)
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
					s := a[1].(*Scope)
					models := (s.Value.(*[]*foreignKeyModel))
					(*models) = append((*models), &foreignKeyModel{ID: 3}, &foreignKeyModel{ID: 5}, &foreignKeyModel{ID: 6})
				}).Return(nil)

				foreignKeyRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				// first patch would clear the foreign keys
				foreignKeyRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					s := a[1].(*Scope)
					model := s.Value.(*foreignKeyModel)
					assert.Equal(t, 0, model.FK)
					primaryFilters := s.PrimaryFilters

					if assert.Len(t, primaryFilters, 1) {
						if assert.Len(t, primaryFilters[0].Values, 1) {
							v := primaryFilters[0].Values[0]
							assert.Equal(t, OpIn, v.Operator)
							assert.Contains(t, v.Values, 3)
							assert.Contains(t, v.Values, 5)
							assert.Contains(t, v.Values, 6)

						}
					}
				}).Return(nil)
				foreignKeyRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					s := a[1].(*Scope)
					model := s.Value.(*foreignKeyModel)
					assert.Equal(t, m1.ID, model.FK)
					primaryFilters := s.PrimaryFilters
					if assert.Len(t, primaryFilters, 1) {
						if assert.Len(t, primaryFilters[0].Values, 1) {
							v := primaryFilters[0].Values[0]
							assert.Equal(t, OpIn, v.Operator)

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
				c := newController(t)
				err := c.RegisterModels(Many2ManyModel{}, RelatedModel{}, JoinModel{})
				require.NoError(t, err)

				model := &Many2ManyModel{Many2Many: []*RelatedModel{{ID: 1}}}

				r, err := c.GetRepository(model)
				require.NoError(t, err)

				many2many, ok := r.(*Repository)
				require.True(t, ok)

				defer clearRepository(many2many)

				r, err = c.GetRepository(RelatedModel{})
				require.NoError(t, err)

				relatedModel, ok := r.(*Repository)
				require.True(t, ok)

				defer clearRepository(relatedModel)

				r, err = c.GetRepository(JoinModel{})
				require.NoError(t, err)

				joinModel, ok := r.(*Repository)
				require.True(t, ok)

				defer clearRepository(joinModel)

				many2many.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				many2many.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					v, ok := s.Value.(*Many2ManyModel)
					require.True(t, ok)

					v.ID = 4
				}).Return(nil)

				relatedModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				relatedModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					primaries := s.PrimaryFilters
					if assert.Len(t, primaries, 1) {
						pf := primaries[0]
						pfValues := pf.Values

						if assert.Len(t, pfValues, 1) {
							pfOpValue := pfValues[0]

							assert.Equal(t, OpIn, pfOpValue.Operator)

							if assert.Len(t, pfOpValue.Values, 1) {
								assert.Equal(t, 1, pfOpValue.Values[0])
							}
						}
					}
					fieldset := s.Fieldset
					assert.Len(t, fieldset, 1)
					_, ok = s.Fieldset["id"]
					assert.True(t, ok)

					v, ok := s.Value.(*[]*RelatedModel)
					require.True(t, ok)

					(*v) = append((*v), &RelatedModel{ID: 1})
				}).Return(nil)

				joinModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				joinModel.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					v, ok := s.Value.(*JoinModel)
					require.True(t, ok)

					assert.Equal(t, 4, v.ForeignKey)
					assert.Equal(t, 1, v.MtMForeignKey)

					v.ID = 33
				}).Return(nil)

				joinModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				relatedModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				many2many.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				s, err := NewC(c, model)
				require.NoError(t, err)

				err = s.Create()
				require.NoError(t, err)

				assert.Equal(t, 4, model.ID)

				many2many.AssertNumberOfCalls(t, "Begin", 1)
				many2many.AssertNumberOfCalls(t, "Create", 1)

				relatedModel.AssertNumberOfCalls(t, "Begin", 1)
				relatedModel.AssertNumberOfCalls(t, "List", 1)

				joinModel.AssertNumberOfCalls(t, "Begin", 1)
				joinModel.AssertNumberOfCalls(t, "Create", 1)
				joinModel.AssertNumberOfCalls(t, "Commit", 1)

				relatedModel.AssertNumberOfCalls(t, "Commit", 1)
				many2many.AssertNumberOfCalls(t, "Commit", 1)
			})
		})
	})

	t.Run("Rollback", func(t *testing.T) {
		t.Run("ByForeignRelationships", func(t *testing.T) {
			c := newController(t)
			err := c.RegisterModels(&hasOneModel{}, &foreignKeyModel{})
			require.NoError(t, err)

			tm := &hasOneModel{
				ID: 2,
				HasOne: &foreignKeyModel{
					ID: 1,
				},
			}

			s, err := NewC(c, tm)
			require.NoError(t, err)
			r, _ := s.Controller().GetRepository(s.Struct())

			// prepare the transaction
			hasOneModel := r.(*Repository)

			// Begin the transaction
			hasOneModel.On("Begin", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			defer func() { hasOneModel.Calls = []mock.Call{} }()

			_, err = s.Begin()
			require.NoError(t, err)
			hasOneModel.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			hasOneModel.On("Create", mock.Anything, mock.Anything).Once().Return(nil)

			model, err := c.ModelStruct(&foreignKeyModel{})
			require.NoError(t, err)

			m2Repo, err := s.Controller().GetRepository((*mapping.ModelStruct)(model))
			require.NoError(t, err)

			foreignKey, ok := m2Repo.(*Repository)
			require.True(t, ok)

			defer func() {
				foreignKey.Calls = []mock.Call{}
			}()

			// Begin the transaction on subscope
			foreignKey.On("Begin", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			foreignKey.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(errors.New("Some error"))

			// Rollback the result
			foreignKey.On("Rollback", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			hasOneModel.On("Rollback", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			err = s.Create()
			require.Error(t, err)

			err = s.Rollback()
			require.NoError(t, err)

			// Assert calls

			foreignKey.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			hasOneModel.AssertCalled(t, "Create", mock.Anything, mock.Anything)
			foreignKey.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			hasOneModel.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
			foreignKey.AssertCalled(t, "Rollback", mock.Anything, mock.Anything)
		})
	})

	t.Run("CreatedAt", func(t *testing.T) {
		type timer struct {
			ID        int
			CreatedAt time.Time
		}

		type ptrTimer struct {
			ID        int
			CreatedAt *time.Time
		}
		c := newController(t)

		err := c.RegisterModels(&timer{}, &ptrTimer{})
		require.NoError(t, err)

		repo, err := c.GetRepository(timer{})
		require.NoError(t, err)

		timerRepo, ok := repo.(*Repository)
		require.True(t, ok)

		t.Run("AutoSelected", func(t *testing.T) {
			t.Run("Zero", func(t *testing.T) {
				s, err := NewC(c, &ptrTimer{ID: 1})
				require.NoError(t, err)

				repo, err := c.GetRepository(ptrTimer{})
				require.NoError(t, err)

				timerRepo, ok := repo.(*Repository)
				require.True(t, ok)

				// timerRepo
				timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				timerRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					createdAt, ok := s.Struct().CreatedAt()
					require.True(t, ok)

					ok, err := s.IsSelected(createdAt)
					require.NoError(t, err)

					// if the value is auto selected  with zero value it should be created
					assert.True(t, ok, "%v", s.SelectedFields)

					tm := s.Value.(*ptrTimer)
					tm.ID = 3
					if assert.NotNil(t, tm.CreatedAt) {
						assert.True(t, time.Since(*tm.CreatedAt) < time.Second)
					}
				}).Return(nil)
				err = s.Create()
				require.NoError(t, err)
			})

			t.Run("NonZero", func(t *testing.T) {
				s, err := NewC(c, &timer{ID: 1, CreatedAt: time.Now().Add(-time.Hour)})
				require.NoError(t, err)

				// timerRepo
				timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				timerRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					createdAt, ok := s.Struct().CreatedAt()
					require.True(t, ok)

					ok, err := s.IsSelected(createdAt)
					require.NoError(t, err)

					assert.True(t, ok)

					tm := s.Value.(*timer)
					tm.ID = 3

					assert.True(t, time.Since(tm.CreatedAt) > time.Hour)
				}).Return(nil)
				err = s.Create()
				require.NoError(t, err)
			})
		})

		t.Run("NotSelected", func(t *testing.T) {
			s, err := NewC(c, &timer{ID: 3})
			require.NoError(t, err)

			err = s.SelectField("ID")
			require.NoError(t, err)

			// timerRepo
			timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

			timerRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				createdAt, ok := s.Struct().CreatedAt()
				require.True(t, ok)

				ok, err := s.IsSelected(createdAt)
				require.NoError(t, err)

				// Field should be selected by the create process.
				assert.False(t, s.autoSelectedFields)
				assert.True(t, ok)

				tm := s.Value.(*timer)
				tm.ID = 3

				assert.True(t, time.Since(tm.CreatedAt) < time.Second)
			}).Return(nil)

			err = s.Create()
			require.NoError(t, err)
		})

		t.Run("Selected", func(t *testing.T) {
			s, err := NewC(c, &timer{ID: 3, CreatedAt: time.Now().Add(-time.Hour)})
			require.NoError(t, err)

			err = s.SelectFields("ID", "CreatedAt")
			require.NoError(t, err)

			// timerRepo
			timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

			timerRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				createdAt, ok := s.Struct().CreatedAt()
				require.True(t, ok)

				ok, err := s.IsSelected(createdAt)
				require.NoError(t, err)

				// Field should be selected by the create process.
				assert.False(t, s.autoSelectedFields)
				assert.True(t, ok)

				tm := s.Value.(*timer)
				tm.ID = 3

				assert.True(t, time.Since(tm.CreatedAt) > time.Hour)
			}).Return(nil)

			err = s.Create()
			require.NoError(t, err)
		})
	})
}

func newController(t testing.TB) *controller.Controller {
	t.Helper()
	c := controller.NewDefault()
	c.Config.DefaultRepositoryName = repoName
	c.Config.Repositories = map[string]*config.Repository{
		repoName: {DriverName: repoName},
	}
	return c
}
