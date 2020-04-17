package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/config"
	"github.com/neuronlabs/neuron-core/controller"
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

func (c *beforeCreateTestModel) BeforeCreate(ctx context.Context, orm *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

func (c *afterCreateTestModel) AfterCreate(ctx context.Context, orm *Scope) error {
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

		repo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

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

		repo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

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

		repo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

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

	type hasBoth struct {
		ID      int
		HasOne  *foreignKeyModel   `neuron:"foreign=FK"`
		HasMany []*foreignKeyModel `neuron:"foreign=FK"`
	}

	t.Run("Valid", func(t *testing.T) {
		// Valid transaction should start with begin then the create and commit at the end
		t.Run("NoForeignRelationships", func(t *testing.T) {
			c := newController(t)
			err := c.RegisterModels(&foreignKeyModel{})
			require.NoError(t, err)

			fkModelRepo, err := c.GetRepository(&foreignKeyModel{})
			require.NoError(t, err)

			foreignKeyRepo, ok := fkModelRepo.(*Repository)
			require.True(t, ok)

			defer clearRepository(foreignKeyRepo)

			// begin
			foreignKeyRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			// create
			foreignKeyRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				s := a[1].(*Scope)

				v := s.Value.(*foreignKeyModel)
				v.ID = 1
			}).Return(nil)
			// commit
			foreignKeyRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			tx := Begin(context.Background(), c, nil)

			model := &foreignKeyModel{FK: 24}

			err = tx.Query(model).Create()
			require.NoError(t, err)

			foreignKeyRepo.AssertNumberOfCalls(t, "Begin", 1)
			foreignKeyRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)

			err = tx.Commit()
			require.NoError(t, err)
			foreignKeyRepo.AssertCalled(t, "Commit", mock.Anything, mock.Anything, mock.Anything)
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

				repo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				repo.On("Create", mock.Anything, s).Once().Run(func(a mock.Arguments) {
					s := a[1].(*Scope)

					sv := s.Value.(*hasOneModel)
					sv.ID = 1
				}).Return(nil)
				repo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				// do the create

				repo2, err := s.Controller().GetRepository(&foreignKeyModel{})
				require.NoError(t, err)

				fkModelRepo, ok := repo2.(*Repository)
				require.True(t, ok)

				defer clearRepository(fkModelRepo)

				fkModelRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				fkModelRepo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)
				fkModelRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				// Create the scope
				require.NoError(t, s.Create())

				repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything, mock.Anything)
				repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)

				fkModelRepo.AssertCalled(t, "Begin", mock.Anything, mock.Anything, mock.Anything)
				fkModelRepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
				fkModelRepo.AssertCalled(t, "Commit", mock.Anything, mock.Anything, mock.Anything)

				repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything, mock.Anything)
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
				hasManyRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				// needs to create
				hasManyRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					// create and set the ID
					s := a[1].(*Scope)
					model := s.Value.(*hasManyModel)
					model.ID = 1
				}).Return(nil)
				// allow to commit
				hasManyRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				// foreign key repo should at first clear the old fk's, get the begin, patch, commit processes
				// while it would get the filters on the FK then it should list the patched id first
				foreignKeyRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				// first patch would clear the foreign keys
				foreignKeyRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
					s := a[1].(*Scope)
					model := s.Value.(*foreignKeyModel)
					assert.Equal(t, 1, model.FK)

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
				foreignKeyRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

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

				many2many.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				many2many.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					v, ok := s.Value.(*Many2ManyModel)
					require.True(t, ok)

					v.ID = 4
				}).Return(nil)

				relatedModel.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
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

				joinModel.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				joinModel.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					v, ok := s.Value.(*JoinModel)
					require.True(t, ok)

					assert.Equal(t, 4, v.ForeignKey)
					assert.Equal(t, 1, v.MtMForeignKey)

					v.ID = 33
				}).Return(nil)

				joinModel.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				relatedModel.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				many2many.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				s, err := NewC(c, model)
				require.NoError(t, err)

				err = s.Create()
				require.NoError(t, err)

				assert.Equal(t, 4, model.ID)
			})

			t.Run("Both", func(t *testing.T) {
				c := newController(t)
				err := c.RegisterModels(foreignKeyModel{}, hasBoth{})
				require.NoError(t, err)

				s, err := NewC(c, &hasBoth{HasOne: &foreignKeyModel{ID: 1}, HasMany: []*foreignKeyModel{{ID: 3}, {ID: 4}}})
				require.NoError(t, err)

				repo, err := c.GetRepository(hasBoth{})
				require.NoError(t, err)

				hbRepo, ok := repo.(*Repository)
				require.True(t, ok)

				hbRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				hbRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					v, ok := s.Value.(*hasBoth)
					require.True(t, ok)

					v.ID = 4
				}).Return(nil)
				hbRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				repo, err = c.GetRepository(foreignKeyModel{})
				require.NoError(t, err)

				fkRepo, ok := repo.(*Repository)
				require.True(t, ok)

				var tx1ID, tx2ID uuid.UUID
				fkRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				fkRepo.On("List", mock.Anything, mock.Anything).Twice().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					var primaries []interface{}
					if assert.Len(t, s.ForeignFilters, 1, "%s", s) {
						fv := s.ForeignFilters[0]
						for _, v := range fv.Values {
							primaries = append(primaries, v.Values...)
						}
					}

					v, ok := s.Value.(*[]*foreignKeyModel)
					require.True(t, ok)

					for _, primary := range primaries {
						(*v) = append((*v), &foreignKeyModel{ID: primary.(int)})
					}
				}).Return(nil)
				fkRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					tx := s.Tx()
					tx1ID = tx.id
				}).Return(nil)
				fkRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					tx := s.Tx()
					tx2ID = tx.id
				}).Return(nil)
				fkRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				err = s.Create()
				require.NoError(t, err)

				assert.Equal(t, tx1ID, tx2ID)
			})
		})
	})

	t.Run("Rollback", func(t *testing.T) {
		t.Run("ByForeignRelationships", func(t *testing.T) {
			c := newController(t)
			err := c.RegisterModels(&hasOneModel{}, &foreignKeyModel{})
			require.NoError(t, err)

			r, _ := c.GetRepository(hasOneModel{})

			// prepare the transaction
			hasOneModelRepo := r.(*Repository)

			// Begin the transaction
			hasOneModelRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			defer func() { hasOneModelRepo.Calls = []mock.Call{} }()

			hasOneModelRepo.On("Create", mock.Anything, mock.Anything).Once().Return(nil)

			model, err := c.ModelStruct(&foreignKeyModel{})
			require.NoError(t, err)

			m2Repo, err := c.GetRepository(model)
			require.NoError(t, err)

			foreignKey, ok := m2Repo.(*Repository)
			require.True(t, ok)

			defer func() {
				foreignKey.Calls = []mock.Call{}
			}()

			// Begin the transaction on subscope
			foreignKey.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			foreignKey.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(errors.New("some error"))

			// Rollback the result
			foreignKey.On("Rollback", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			hasOneModelRepo.On("Rollback", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			tm := &hasOneModel{
				ID: 2,
				HasOne: &foreignKeyModel{
					ID: 1,
				},
			}
			tx := Begin(context.Background(), c, nil)
			err = tx.Query(tm).Create()
			require.Error(t, err)

			err = tx.Rollback()
			require.NoError(t, err)
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
			s, err := NewC(c, &timer{ID: 1, CreatedAt: time.Now().Add(-time.Hour)})
			require.NoError(t, err)

			// timerRepo
			timerRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			timerRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				createdAt, ok := s.Struct().CreatedAt()
				require.True(t, ok)

				_, ok = s.InFieldset(createdAt)
				assert.True(t, ok)

				tm := s.Value.(*timer)
				tm.ID = 3

				assert.True(t, time.Since(tm.CreatedAt) > time.Hour)
			}).Return(nil)
			err = s.Create()
			require.NoError(t, err)
		})

		t.Run("NotSelected", func(t *testing.T) {
			s, err := NewC(c, &timer{ID: 3})
			require.NoError(t, err)

			err = s.SetFields("ID")
			require.NoError(t, err)

			// timerRepo
			timerRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			timerRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				createdAt, ok := s.Struct().CreatedAt()
				require.True(t, ok)

				_, ok = s.InFieldset(createdAt)
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

			err = s.SetFields("ID", "CreatedAt")
			require.NoError(t, err)

			// timerRepo
			timerRepo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Commit", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			timerRepo.On("Create", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				createdAt, ok := s.Struct().CreatedAt()
				require.True(t, ok)

				_, ok = s.InFieldset(createdAt)
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
	cfg := config.Default()
	cfg.DefaultRepositoryName = repoName
	cfg.Repositories = map[string]*config.Repository{
		repoName: {DriverName: repoName},
	}
	c, err := controller.New(cfg)
	require.NoError(t, err)
	require.NoError(t, c.DialAll(context.Background()))
	return c
}
