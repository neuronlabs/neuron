package query

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
)

type testPatcher struct {
	ID    int    `neuron:"type=primary"`
	Field string `neuron:"type=attr"`
}

type testBeforePatcher struct {
	ID   int    `neuron:"type=primary"`
	Attr string `neuron:"type=attr"`
}

func (b *testBeforePatcher) BeforePatch(ctx context.Context, s *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type testAfterPatcher struct {
	ID   int    `neuron:"type=primary"`
	Attr string `neuron:"type=attr"`
}

func (a *testAfterPatcher) AfterPatch(ctx context.Context, s *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

// TestPatch tests the patch process for the default processor.
func TestPatch(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&testPatcher{}, &testAfterPatcher{}, &testBeforePatcher{})
	require.NoError(t, err)

	t.Run("NoSelectedValues", func(t *testing.T) {
		s, err := NewC(c, &testPatcher{ID: 5})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Patch()
		if assert.Error(t, err) {
			repo.AssertNotCalled(t, "Patch", mock.Anything, mock.Anything)
		}
	})

	t.Run("NoHooks", func(t *testing.T) {
		s, err := NewC(c, &testPatcher{ID: 5, Field: "Something"})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Patch()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		s, err := NewC(c, &testBeforePatcher{ID: 1, Attr: "MustBeSomething"})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		require.NoError(t, s.PatchContext(context.WithValue(context.Background(), testCtxKey, t)))
	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := NewC(c, &testAfterPatcher{ID: 2, Attr: "MustBeSomething"})
		require.NoError(t, err)

		r, err := s.Controller().GetRepository(s.Struct())
		require.NoError(t, err)

		repo, ok := r.(*Repository)
		require.True(t, ok)

		defer clearRepository(repo)

		repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)
		repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

		require.NoError(t, s.PatchContext(context.WithValue(context.Background(), testCtxKey, t)))
	})

	t.Run("DedicatedTransactions", func(t *testing.T) {
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
			r, _ := c.GetRepository(patchTMRelations{})
			repo := r.(*Repository)

			defer clearRepository(repo)

			// Begin define
			repo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

			tx := Begin(context.Background(), c, nil)

			repo.On("List", mock.Anything, mock.Anything).Once().Return(nil)

			// get the related model
			model, err := c.ModelStruct(&patchTMRelated{})
			require.NoError(t, err)

			mr, err := c.GetRepository(model)
			require.NoError(t, err)

			repo2, ok := mr.(*Repository)
			require.True(t, ok)

			defer clearRepository(repo2)

			repo2.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			repo2.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)

			tm := &patchTMRelations{
				ID: 2,
				Rel: &patchTMRelated{
					ID: 1,
				},
			}
			err = tx.Query(tm).Patch()
			require.NoError(t, err)

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

			// Commit
			repo.On("Commit", mock.Anything, mock.Anything).Return(nil)
			repo2.On("Commit", mock.Anything, mock.Anything).Return(nil)

			require.NoError(t, tx.Commit())

			repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			repo2.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
		})

		t.Run("Rollback", func(t *testing.T) {
			r, _ := c.GetRepository(patchTMRelations{})

			// prepare the transaction
			repo := r.(*Repository)
			defer clearRepository(repo)

			// Begin the transaction
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
			}).Return(nil)

			repo.On("List", mock.Anything, mock.Anything).Once().Return(nil)

			model, err := c.ModelStruct(&patchTMRelated{})
			require.NoError(t, err)

			m2Repo, err := c.GetRepository(model)
			require.NoError(t, err)

			repo2, ok := m2Repo.(*Repository)
			require.True(t, ok)

			defer clearRepository(repo2)

			// Begin the transaction on subquery
			repo2.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			repo2.On("Patch", mock.Anything, mock.Anything).Once().Return(stdErrors.New("Some error"))

			tx := Begin(context.Background(), c, nil)
			tm := &patchTMRelations{
				ID: 2,
				Rel: &patchTMRelated{
					ID: 1,
				},
			}
			err = tx.Query(tm).Patch()
			require.Error(t, err)

			// Rollback the result
			repo2.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)
			repo.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)

			err = tx.Rollback()
			require.NoError(t, err)

			// Assert calls
			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
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
				repo.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
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

					if fieldSet := s.Fieldset; assert.Len(t, fieldSet, 1) {
						_, ok := s.Fieldset["id"]
						assert.True(t, ok)
					}

					sv, ok := s.Value.(*[]*HasOneModel)
					require.True(t, ok)

					*sv = append((*sv), &HasOneModel{ID: model.ID})

				}).Return(nil)
				repo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				frepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				frepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(arg mock.Arguments) {
					s, ok := arg[1].(*Scope)
					require.True(t, ok)

					pm := s.PrimaryFilters
					if assert.NotEmpty(t, pm) {
						if assert.Len(t, pm, 1) {
							if assert.Len(t, pm[0].Values, 1) {
								v := pm[0].Values[0]
								assert.Equal(t, OpIn, v.Operator)
								assert.Equal(t, model.HasOne.ID, v.Values[0])
							}
						}
					}

					_, isSelected := s.InFieldset("ForeignKey")
					assert.True(t, isSelected)
				}).Return(nil)

				frepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				require.NoError(t, s.Patch())

				repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything, mock.Anything)
				repo.AssertCalled(t, "List", mock.Anything, mock.Anything)

				frepo.AssertCalled(t, "Begin", mock.Anything, mock.Anything, mock.Anything)
				frepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)

				frepo.AssertCalled(t, "Commit", mock.Anything, mock.Anything, mock.Anything)
				repo.AssertCalled(t, "Commit", mock.Anything, mock.Anything, mock.Anything)
			})
		})

		t.Run("HasMany", func(t *testing.T) {
			t.Run("NonEmpty", func(t *testing.T) {
				model := &HasManyModel{
					ID: 3,
					HasMany: []*ForeignModel{
						{ID: 1},
						{ID: 5},
					},
				}

				s, err := NewC(c, model)
				require.NoError(t, err)

				mr, err := c.GetRepository(model)
				require.NoError(t, err)

				hasMany, ok := mr.(*Repository)
				require.True(t, ok)

				defer clearRepository(hasMany)

				hasMany.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				// patch the 'hasMany' model.
				hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
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

					if fieldSet := s.Fieldset; assert.Len(t, fieldSet, 1) {
						_, ok := s.Fieldset["id"]
						assert.True(t, ok)
					}

					sv, ok := s.Value.(*[]*HasManyModel)
					require.True(t, ok)

					(*sv) = append((*sv), &HasManyModel{ID: model.ID})
				}).Return(nil)

				fr, err := c.GetRepository(&ForeignModel{})
				require.NoError(t, err)

				foreignModel, ok := fr.(*Repository)
				require.True(t, ok)

				defer clearRepository(foreignModel)

				foreignModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				// patch model's with provided id's and set their's foreign keys into 'root' primary
				foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*Scope)

					primaries := s.PrimaryFilters
					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values, 1) {
							fv := single.Values[0]

							assert.Equal(t, OpIn, fv.Operator)
							if assert.Len(t, fv.Values, 2) {
								assert.Contains(t, fv.Values, 1)
								assert.Contains(t, fv.Values, 5)
							}
						}
					}
					_, isFKSelected := s.InFieldset("ForeignKey")
					assert.True(t, isFKSelected)

					sv, ok := s.Value.(*ForeignModel)
					require.True(t, ok)

					assert.Equal(t, model.ID, sv.ForeignKey)
				}).Return(nil)

				foreignModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				hasMany.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				err = s.Patch()
				require.NoError(t, err)
			})

			t.Run("Clear", func(t *testing.T) {
				model := &HasManyModel{
					ID:      3,
					HasMany: []*ForeignModel{},
				}

				s, err := NewC(c, model)
				require.NoError(t, err)

				mr, err := c.GetRepository(model)
				require.NoError(t, err)

				hasMany, ok := mr.(*Repository)
				require.True(t, ok)

				defer clearRepository(hasMany)

				hasMany.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

				// list the 'hasMany' model primaries - check if exists.
				hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
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

					if fieldSet := s.Fieldset; assert.Len(t, fieldSet, 1) {
						_, ok := s.Fieldset["id"]
						assert.True(t, ok)
					}

					sv, ok := s.Value.(*[]*HasManyModel)
					require.True(t, ok)

					(*sv) = append((*sv), &HasManyModel{ID: model.ID})
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

					if fieldSet := s.Fieldset; assert.Len(t, fieldSet, 1) {
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

				err = s.Patch()
				require.NoError(t, err)
			})

			t.Run("Error", func(t *testing.T) {
				// no primary found - list when no attributes were selected
				t.Run("NoRootFound", func(t *testing.T) {
					model := &HasManyModel{
						ID:      3,
						HasMany: []*ForeignModel{},
					}

					s, err := NewC(c, model)
					require.NoError(t, err)

					mr, err := c.GetRepository(model)
					require.NoError(t, err)

					hasMany, ok := mr.(*Repository)
					require.True(t, ok)

					defer clearRepository(hasMany)

					hasMany.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
					// list the 'hasMany' model primaries - check if exists.
					hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
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

						if fieldSet := s.Fieldset; assert.Len(t, fieldSet, 1) {
							_, ok := s.Fieldset["id"]
							assert.True(t, ok)
						}

						_, ok := s.Value.(*[]*HasManyModel)
						require.True(t, ok)
					}).Return(errors.NewDet(class.QueryValueNoResult, "no results"))

					hasMany.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)

					err = s.Patch()
					assert.Error(t, err)
				})

				// no related found
				t.Run("NoRelatedFound", func(t *testing.T) {
					model := &HasManyModel{
						ID: 3,
						HasMany: []*ForeignModel{
							{ID: 6},
							{ID: 7},
						},
					}

					s, err := NewC(c, model)
					require.NoError(t, err)

					mr, err := c.GetRepository(model)
					require.NoError(t, err)

					hasMany, ok := mr.(*Repository)
					require.True(t, ok)

					defer clearRepository(hasMany)

					hasMany.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
					// list the 'hasMany' model primaries - check if exists.
					hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
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

						if fieldSet := s.Fieldset; assert.Len(t, fieldSet, 1) {
							_, ok := s.Fieldset["id"]
							assert.True(t, ok)
						}

						sv, ok := s.Value.(*[]*HasManyModel)
						require.True(t, ok)

						(*sv) = append((*sv), &HasManyModel{ID: model.ID})
					}).Return(nil)

					fr, err := c.GetRepository(&ForeignModel{})
					require.NoError(t, err)

					foreignModel, ok := fr.(*Repository)
					require.True(t, ok)

					defer clearRepository(foreignModel)

					foreignModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
					// patch model's with provided id's and set their's foreign keys into 'root' primary
					foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
						s := args[1].(*Scope)

						primaries := s.PrimaryFilters
						if assert.Len(t, primaries, 1) {
							single := primaries[0]

							if assert.Len(t, single.Values, 1) {
								fv := single.Values[0]

								assert.Equal(t, OpIn, fv.Operator)
								if assert.Len(t, fv.Values, 2) {
									assert.Contains(t, fv.Values, 6)
									assert.Contains(t, fv.Values, 7)
								}
							}
						}

						_, isFKSelected := s.InFieldset("ForeignKey")
						assert.True(t, isFKSelected)

						sv, ok := s.Value.(*ForeignModel)
						require.True(t, ok)

						assert.Equal(t, model.ID, sv.ForeignKey)
					}).Return(errors.NewDet(class.QueryValueNoResult, "no results"))
					foreignModel.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)

					hasMany.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)
					err = s.Patch()
					assert.Error(t, err)
				})
			})
		})

		t.Run("Many2Many", func(t *testing.T) {
			t.Run("NonEmpty", func(t *testing.T) {
				c := newController(t)
				err := c.RegisterModels(Many2ManyModel{}, RelatedModel{}, JoinModel{})
				require.NoError(t, err)

				model := &Many2ManyModel{
					ID:        4,
					Many2Many: []*RelatedModel{{ID: 1}},
				}

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

				// check the values
				many2many.On("List", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					switch s.Struct() {
					case c.MustGetModelStruct(Many2ManyModel{}):
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

						v, ok := s.Value.(*[]*Many2ManyModel)
						require.True(t, ok)

						(*v) = append((*v), &Many2ManyModel{ID: 4})
					case c.MustGetModelStruct(RelatedModel{}):
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
						assert.Contains(t, s.Fieldset, "id")

						v, ok := s.Value.(*[]*RelatedModel)
						require.True(t, ok)

						(*v) = append((*v), &RelatedModel{ID: 1})
					case c.MustGetModelStruct(JoinModel{}):
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

						*v = append(*v,
							&JoinModel{ID: 6, ForeignKey: 4, MtMForeignKey: 17},
							&JoinModel{ID: 7, ForeignKey: 4, MtMForeignKey: 33},
						)
					}

				}).Return(nil)

				relatedModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				joinModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

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
				joinModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				relatedModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				many2many.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				s, err := NewC(c, model)
				require.NoError(t, err)

				err = s.Patch()
				require.NoError(t, err)
			})

			t.Run("Clear", func(t *testing.T) {
				c := newController(t)
				err := c.RegisterModels(Many2ManyModel{}, RelatedModel{}, JoinModel{})
				require.NoError(t, err)

				model := &Many2ManyModel{
					ID:        4,
					Many2Many: []*RelatedModel{},
				}

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
				many2many.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
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

					v, ok := s.Value.(*[]*Many2ManyModel)
					require.True(t, ok)

					(*v) = append((*v), &Many2ManyModel{ID: 4})
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

				err = s.Patch()
				require.NoError(t, err)

				many2many.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
				many2many.AssertCalled(t, "List", mock.Anything, mock.Anything)

				joinModel.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
				joinModel.AssertCalled(t, "List", mock.Anything, mock.Anything)
				joinModel.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
				joinModel.AssertCalled(t, "Commit", mock.Anything, mock.Anything)

				many2many.AssertCalled(t, "Commit", mock.Anything, mock.Anything)
			})
		})
	})

	t.Run("UpdatedAt", func(t *testing.T) {
		type timer struct {
			ID        int
			UpdatedAt time.Time
		}

		c := newController(t)
		err = c.RegisterModels(timer{})
		require.NoError(t, err)

		repo, err := c.GetRepository(timer{})
		require.NoError(t, err)

		timerRepo, ok := repo.(*Repository)
		require.True(t, ok)

		t.Run("AutoSelected", func(t *testing.T) {
			t.Run("Zero", func(t *testing.T) {
				s, err := NewC(c, &timer{ID: 3})
				require.NoError(t, err)

				updatedAt, ok := s.Struct().UpdatedAt()
				require.True(t, ok)

				timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				timerRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					v, ok := s.Value.(*timer)
					require.True(t, ok)

					_, ok = s.InFieldset(updatedAt)
					assert.True(t, ok)

					assert.NotZero(t, v.UpdatedAt)
				}).Return(nil)
				defer clearRepository(timerRepo)

				err = s.Patch()
				require.NoError(t, err)

				timerRepo.AssertNumberOfCalls(t, "Patch", 1)

			})

			t.Run("NonZero", func(t *testing.T) {
				s, err := NewC(c, &timer{ID: 3, UpdatedAt: time.Now().Add(-time.Hour)})
				require.NoError(t, err)

				updatedAt, ok := s.Struct().UpdatedAt()
				require.True(t, ok)

				timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
				timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
				timerRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*Scope)
					require.True(t, ok)

					v, ok := s.Value.(*timer)
					require.True(t, ok)

					_, ok = s.InFieldset(updatedAt)
					assert.True(t, ok)

					assert.True(t, time.Since(v.UpdatedAt) > time.Hour)
				}).Return(nil)

				defer clearRepository(timerRepo)

				err = s.Patch()
				require.NoError(t, err)

				timerRepo.AssertNumberOfCalls(t, "Patch", 1)
			})
		})

		t.Run("Selected", func(t *testing.T) {
			s, err := NewC(c, &timer{ID: 3, UpdatedAt: time.Now().Add(-time.Hour)})
			require.NoError(t, err)

			updatedAt, ok := s.Struct().UpdatedAt()
			require.True(t, ok)

			err = s.SetFields("UpdatedAt")
			require.NoError(t, err)

			timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				v, ok := s.Value.(*timer)
				require.True(t, ok)

				_, ok = s.InFieldset(updatedAt)
				assert.True(t, ok)

				assert.True(t, time.Since(v.UpdatedAt) > time.Hour)
			}).Return(nil)

			defer clearRepository(timerRepo)

			err = s.Patch()
			require.NoError(t, err)

			timerRepo.AssertNumberOfCalls(t, "Patch", 1)
		})

		t.Run("NotSelected", func(t *testing.T) {
			type timer struct {
				ID        int
				UpdatedAt *time.Time
			}

			c := newController(t)
			err = c.RegisterModels(timer{})
			require.NoError(t, err)

			repo, err := c.GetRepository(timer{})
			require.NoError(t, err)

			timerRepo, ok := repo.(*Repository)
			require.True(t, ok)

			s, err := NewC(c, &timer{ID: 3})
			require.NoError(t, err)

			updatedAt, ok := s.Struct().UpdatedAt()
			require.True(t, ok)

			timerRepo.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)
			timerRepo.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				v, ok := s.Value.(*timer)
				require.True(t, ok)

				_, ok = s.InFieldset(updatedAt)
				assert.True(t, ok)
				if assert.NotNil(t, v.UpdatedAt) {
					assert.True(t, time.Since(*v.UpdatedAt) < time.Second)
				}
			}).Return(nil)

			defer clearRepository(timerRepo)

			err = s.Patch()
			require.NoError(t, err)

			timerRepo.AssertNumberOfCalls(t, "Patch", 1)
		})
	})
}
