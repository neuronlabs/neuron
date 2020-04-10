package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testDeleter struct {
	ID int `neuron:"type=primary"`
}

type testBeforeDeleter struct {
	ID int `neuron:"type=primary"`
}

func (b *testBeforeDeleter) BeforeDelete(ctx context.Context, s *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type testAfterDeleter struct {
	ID int `neuron:"type=primary"`
}

func (a *testAfterDeleter) AfterDelete(ctx context.Context, s *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

// TestDelete test the Delete Processes.
func TestDelete(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&testDeleter{}, &testAfterDeleter{}, &testBeforeDeleter{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := NewC(c, &testDeleter{ID: 1})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Delete()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := NewC(c, &testBeforeDeleter{ID: 1})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.DeleteContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := NewC(c, &testAfterDeleter{ID: 1})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.DeleteContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
		}
	})

	t.Run("Transactions", func(t *testing.T) {
		// define the helper models
		type deleteTMRelated struct {
			ID int `neuron:"type=primary"`
			FK int `neuron:"type=foreign"`
		}

		type deleteTMRelations struct {
			ID  int              `neuron:"type=primary"`
			Rel *deleteTMRelated `neuron:"type=relation;foreign=FK"`
		}

		c := newController(t)

		err := c.RegisterModels(&deleteTMRelations{}, &deleteTMRelated{})
		require.NoError(t, err)

		t.Run("Valid", func(t *testing.T) {
			r, _ := c.GetRepository(&deleteTMRelations{})
			repo := r.(*Repository)

			defer clearRepository(repo)

			// Begin define
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
			}).Once().Return(nil)

			tx := Begin()

			repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)

			// get the related model
			model, err := c.ModelStruct(&deleteTMRelated{})
			require.NoError(t, err)

			mr, err := c.GetRepository(model)
			require.NoError(t, err)

			repo2, ok := mr.(*Repository)
			require.True(t, ok)

			defer clearRepository(repo2)

			repo2.On("Patch", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			repo2.On("Begin", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
			}).Return(nil)

			repo2.On("List", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				s := a[1].(*Scope)
				values := []*deleteTMRelated{{ID: 2}}
				s.Value = &values
			}).Once().Return(nil)

			tm := &deleteTMRelations{ID: 2}
			err = tx.QueryC(c, tm).Delete()
			require.NoError(t, err)

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			// Commit
			repo.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
			}).Return(nil)
			repo2.On("Commit", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
			}).Return(nil)
			err = tx.Commit()
			require.NoError(t, err)

			repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
		})

		t.Run("Rollback", func(t *testing.T) {

			r, err := c.GetRepository(deleteTMRelations{})
			require.NoError(t, err)

			// prepare the transaction
			repo := r.(*Repository)
			defer clearRepository(repo)

			// Begin the transaction
			repo.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			tx := Begin()

			repo.On("Delete", mock.Anything, mock.Anything).Once().Return(nil)

			model, err := c.ModelStruct(&deleteTMRelated{})
			require.NoError(t, err)

			m2Repo, err := c.GetRepository(model)
			require.NoError(t, err)

			repo2, ok := m2Repo.(*Repository)
			require.True(t, ok)

			defer clearRepository(repo2)

			// Begin the transaction on sub query.
			repo2.On("Begin", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			repo2.On("List", mock.Anything, mock.Anything).Once().Run(func(a mock.Arguments) {
				s := a[1].(*Scope)
				values := []*deleteTMRelated{{ID: 1}}
				s.Value = &values
			}).Return(nil)

			repo2.On("Patch", mock.Anything, mock.Anything).Once().Return(errors.New("Some error"))

			// Rollback the result
			repo2.On("Rollback", mock.Anything, mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
			}).Return(nil)

			repo.On("Rollback", mock.Anything, mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
			}).Return(nil)

			tm := &deleteTMRelations{
				ID: 2,
				Rel: &deleteTMRelated{
					ID: 1,
				},
			}

			err = tx.QueryC(c, tm).Delete()
			require.Error(t, err)

			err = tx.Rollback()
			assert.NoError(t, err)
		})
	})

	t.Run("ForeignRelationships", func(t *testing.T) {
		c := newController(t)

		err = c.RegisterModels(HasOneModel{}, HasManyModel{}, Many2ManyModel{}, JoinModel{}, ForeignModel{}, RelatedModel{})
		require.NoError(t, err)

		t.Run("HasOne", func(t *testing.T) {
			t.Run("Valid", func(t *testing.T) {
				// patch the model
				model := &HasOneModel{
					ID: 3,
				}

				s, err := NewC(c, model)
				require.NoError(t, err)

				hasOneRepo, err := c.GetRepository(model)
				require.NoError(t, err)

				repo, ok := hasOneRepo.(*Repository)
				require.True(t, ok)

				defer clearRepository(repo)

				foreignRepo, err := c.GetRepository(model.HasOne)
				require.NoError(t, err)

				frepo, ok := foreignRepo.(*Repository)
				require.True(t, ok)

				defer clearRepository(frepo)

				repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				repo.On("Delete", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					primaries := s.PrimaryFilters

					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values, 1) {
							fv := single.Values[0]

							assert.Equal(t, OpIn, fv.Operator)
							if assert.Len(t, fv.Values, 1) {
								assert.Contains(t, fv.Values, model.ID)
							}
						}
					}
				}).Return(nil)

				repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				frepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

				frepo.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					foreigns := s.ForeignFilters
					if assert.Len(t, foreigns, 1) {
						if assert.Len(t, foreigns[0].Values, 1) {
							v := foreigns[0].Values[0]
							assert.Equal(t, OpIn, v.Operator)
							assert.Equal(t, model.ID, v.Values[0])
						}
					}

					v, ok := s.Value.(*[]*ForeignModel)
					require.True(t, ok)

					(*v) = append((*v), &ForeignModel{ID: 5})
				}).Return(nil)

				frepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(arg mock.Arguments) {
					s, ok := arg[1].(*Scope)
					require.True(t, ok)

					pm := s.PrimaryFilters
					if assert.NotEmpty(t, pm) {
						if assert.Len(t, pm, 1) {
							if assert.Len(t, pm[0].Values, 1) {
								v := pm[0].Values[0]
								assert.Equal(t, OpIn, v.Operator)
								assert.Equal(t, 5, v.Values[0])
							}
						}
					}

					_, isSelected := s.InFieldset("ForeignKey")
					assert.True(t, isSelected)
				}).Return(nil)

				frepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				require.NoError(t, s.Delete())

				repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
				repo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)

				frepo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
				frepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

				frepo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
				repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			})
		})

		t.Run("HasMany", func(t *testing.T) {
			model := &HasManyModel{ID: 3}

			s, err := NewC(c, model)
			require.NoError(t, err)

			mr, err := c.GetRepository(model)
			require.NoError(t, err)

			hasMany, ok := mr.(*Repository)
			require.True(t, ok)

			defer clearRepository(hasMany)

			hasMany.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

			// delete the 'hasMany' for model's primaries
			hasMany.On("Delete", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*Scope)
				primaries := s.PrimaryFilters

				if assert.Len(t, primaries, 1) {
					single := primaries[0]

					if assert.Len(t, single.Values, 1) {
						fv := single.Values[0]

						assert.Equal(t, OpIn, fv.Operator)
						if assert.Len(t, fv.Values, 1) {
							assert.Contains(t, fv.Values, model.ID)
						}
					}
				}
			}).Return(nil)

			fr, err := c.GetRepository(&ForeignModel{})
			require.NoError(t, err)

			foreignModel := fr.(*Repository)
			foreignModel.Calls = []mock.Call{}

			foreignModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

			// get the foreign relationships with the foreign key equal to the primaries of the root
			foreignModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*Scope)

				foreignKeys := s.ForeignFilters
				if assert.Len(t, foreignKeys, 1) {
					single := foreignKeys[0]

					if assert.Len(t, single.Values, 1) {
						fv := single.Values[0]

						assert.Equal(t, OpIn, fv.Operator)
						if assert.Len(t, fv.Values, 1) {
							assert.Contains(t, fv.Values, model.ID)
						}
					}
				}

				if assert.Len(t, s.Fieldset, 1) {
					_, ok := s.Fieldset["id"]
					assert.True(t, ok)
				}

				sv := s.Value.(*[]*ForeignModel)
				(*sv) = append((*sv), &ForeignModel{ID: 4}, &ForeignModel{ID: 7})
			}).Return(nil)

			// clear the relationship foreign keys
			foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*Scope)

				primaries := s.PrimaryFilters
				if assert.Len(t, primaries, 1) {
					single := primaries[0]

					if assert.Len(t, single.Values, 1) {
						fv := single.Values[0]

						assert.Equal(t, OpIn, fv.Operator)
						if assert.Len(t, fv.Values, 2) {
							assert.Contains(t, fv.Values, 4)
							assert.Contains(t, fv.Values, 7)
						}
					}
				}

				_, isFKSelected := s.InFieldset("ForeignKey")
				assert.True(t, isFKSelected)

				sv := s.Value.(*ForeignModel)
				assert.Equal(t, 0, sv.ForeignKey)
			}).Return(nil)

			foreignModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
			foreignModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

			hasMany.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

			err = s.Delete()
			require.NoError(t, err)

			hasMany.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			// one call to hasMany - check if exists
			hasMany.AssertNumberOfCalls(t, "Delete", 1)

			foreignModel.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			// one call to foreign List - get the primary id's - a part of clear scope
			foreignModel.AssertNumberOfCalls(t, "List", 1)

			// 1. clear scope
			foreignModel.AssertNumberOfCalls(t, "Patch", 1)
			foreignModel.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			hasMany.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
		})

		t.Run("Many2Many", func(t *testing.T) {
			model := &Many2ManyModel{ID: 4}

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
			many2many.On("Delete", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				primaries := s.PrimaryFilters
				if assert.Len(t, primaries, 1) {
					if assert.Len(t, primaries[0].Values, 1) {
						pv := primaries[0].Values[0]

						assert.Equal(t, OpIn, pv.Operator)
						if assert.Len(t, pv.Values, 1) {
							assert.Equal(t, 4, pv.Values[0])
						}
					}
				}
			}).Return(nil)

			joinModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

			// List is the reduce primaries lister
			joinModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				foreigns := s.ForeignFilters
				if assert.Len(t, foreigns, 1) {
					foreignValues := foreigns[0].Values
					fk, ok := s.Struct().ForeignKey("ForeignKey")
					if assert.True(t, ok) {
						assert.Equal(t, fk, foreigns[0].StructField)
					}

					if assert.Len(t, foreignValues, 1) {
						foreignFirst := foreignValues[0]

						assert.Equal(t, OpIn, foreignFirst.Operator)
						if assert.Len(t, foreignFirst.Values, 1) {
							assert.Equal(t, 4, foreignFirst.Values[0])
						}
					}
				}

				v, ok := s.Value.(*[]*JoinModel)
				require.True(t, ok)

				(*v) = append((*v),
					&JoinModel{ID: 6, ForeignKey: 4, MtMForeignKey: 17},
					&JoinModel{ID: 7, ForeignKey: 4, MtMForeignKey: 33},
				)
			}).Return(nil)

			joinModel.On("Delete", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				primaries := s.PrimaryFilters
				if assert.Len(t, primaries, 1) {
					pf := primaries[0]
					pfValues := pf.Values

					if assert.Len(t, pfValues, 1) {
						pfOpValue := pfValues[0]

						assert.Equal(t, OpIn, pfOpValue.Operator)

						if assert.Len(t, pfOpValue.Values, 2) {
							assert.Contains(t, pfOpValue.Values, 6)
							assert.Contains(t, pfOpValue.Values, 7)
						}
					}
				}
			}).Return(nil)

			joinModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
			many2many.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

			s, err := NewC(c, model)
			require.NoError(t, err)

			err = s.Delete()
			require.NoError(t, err)

			many2many.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			many2many.AssertCalled(t, "Delete", mock.Anything, mock.Anything)

			joinModel.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			joinModel.AssertCalled(t, "List", mock.Anything, mock.Anything)
			joinModel.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
			joinModel.AssertCalled(t, "Commit", mock.Anything, mock.Anything)

			many2many.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
		})
	})

	t.Run("TimeRelated", func(t *testing.T) {
		// test models 'DeletedAt' field.
		type timer struct {
			ID        int
			DeletedAt *time.Time
		}

		c := newController(t)
		err := c.RegisterModels(timer{})
		require.NoError(t, err)

		s, err := NewC(c, &timer{ID: 3})
		require.NoError(t, err)

		repo, err := c.GetRepository(timer{})
		require.NoError(t, err)

		timerRepo, ok := repo.(*Repository)
		require.True(t, ok)

		timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
		timerRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			tm, ok := s.Value.(*timer)
			require.True(t, ok)

			if assert.NotNil(t, tm.DeletedAt) {
				assert.True(t, time.Since(*tm.DeletedAt) < time.Second)
			}
		}).Return(nil)

		err = s.Delete()
		require.NoError(t, err)

		timerRepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
	})
}
