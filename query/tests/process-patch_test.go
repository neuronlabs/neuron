package tests

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/errors/class"
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
	ID   int    `neuron:"type=primary"`
	Attr string `neuron:"type=attr"`
}

func (b *testBeforePatcher) BeforePatch(ctx context.Context, s *query.Scope) error {
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

func (a *testAfterPatcher) AfterPatch(ctx context.Context, s *query.Scope) error {
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
		s, err := query.NewC((*controller.Controller)(c), &testPatcher{ID: 5})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)

		err = s.Patch()
		if assert.Error(t, err) {
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
		s, err := query.NewC((*controller.Controller)(c), &testBeforePatcher{ID: 1, Attr: "MustBeSomething"})
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("Patch", mock.Anything, mock.Anything).Once().Return(nil)

		require.NoError(t, s.PatchContext(context.WithValue(context.Background(), testCtxKey, t)))

	})

	t.Run("HookAfter", func(t *testing.T) {
		s, err := query.NewC((*controller.Controller)(c), &testAfterPatcher{ID: 2, Attr: "MustBeSomething"})
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

			defer clearRepository(repo)

			// Begin define
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on patchTMRelations")
			}).Once().Return(nil)

			_, err = s.Begin()
			require.NoError(t, err)

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)

			repo.On("List", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
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

			defer clearRepository(repo2)

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
			defer clearRepository(repo)

			// Begin the transaction
			repo.On("Begin", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
				log.Debug("Begin on patchTMRelations")
			}).Return(nil)

			_, err = s.Begin()
			require.NoError(t, err)

			repo.AssertCalled(t, "Begin", mock.Anything, mock.Anything)
			repo.On("List", mock.Anything, mock.Anything).Once().Return(nil)

			model, err := c.GetModelStruct(&patchTMRelated{})
			require.NoError(t, err)

			m2Repo, err := repository.GetRepository(s.Controller(), (*mapping.ModelStruct)(model))
			require.NoError(t, err)

			repo2, ok := m2Repo.(*mocks.Repository)
			require.True(t, ok)

			defer clearRepository(repo2)

			// Begin the transaction on subquery
			repo2.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)
			repo2.On("Patch", mock.Anything, mock.Anything).Once().Return(stdErrors.New("Some error"))

			err = s.Patch()
			require.Error(t, err)

			// Rollback the result
			repo2.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)
			repo.On("Rollback", mock.Anything, mock.Anything).Once().Return(nil)

			err = s.Rollback()
			require.NoError(t, err)

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

				defer clearRepository(repo)

				foreignRepo, err := repository.GetRepository((*controller.Controller)(c), model.HasOne)
				require.NoError(t, err)

				frepo, ok := foreignRepo.(*mocks.Repository)
				require.True(t, ok)

				defer clearRepository(frepo)

				repo.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s, ok := args[1].(*query.Scope)
					require.True(t, ok)

					primaries := s.PrimaryFilters()

					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 1) {
								assert.Contains(t, fv.Values, model.ID)
							}
						}
					}

					if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fieldSet[0], s.Struct().Primary())
					}

					sv, ok := s.Value.(*[]*HasOneModel)
					require.True(t, ok)

					(*sv) = append((*sv), &HasOneModel{ID: model.ID})

				}).Return(nil)

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

					isSelected, err := s.IsSelected("ForeignKey")
					require.NoError(t, err)
					assert.True(t, isSelected)
				}).Return(nil)

				require.NoError(t, s.Patch())

				repo.AssertNotCalled(t, "Patch", mock.Anything, mock.Anything)
				frepo.AssertCalled(t, "Patch", mock.Anything, mock.Anything)
			})
		})

		t.Run("HasMany", func(t *testing.T) {
			t.Run("NonEmpty", func(t *testing.T) {
				model := &HasManyModel{
					ID: 3,
					HasMany: []*ForeignModel{
						&ForeignModel{ID: 1},
						&ForeignModel{ID: 5},
					},
				}

				s, err := query.NewC((*controller.Controller)(c), model)
				require.NoError(t, err)

				mr, err := repository.GetRepository(c, model)
				require.NoError(t, err)

				hasMany := mr.(*mocks.Repository)
				hasMany.Calls = []mock.Call{}

				// patch the 'hasMany' model.
				hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)
					primaries := s.PrimaryFilters()

					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 1) {
								assert.Contains(t, fv.Values, model.ID)
							}
						}
					}

					if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fieldSet[0], s.Struct().Primary())
					}

					sv, ok := s.Value.(*[]*HasManyModel)
					require.True(t, ok)

					(*sv) = append((*sv), &HasManyModel{ID: model.ID})
				}).Return(nil)

				fr, err := repository.GetRepository(c, &ForeignModel{})
				require.NoError(t, err)

				foreignModel := fr.(*mocks.Repository)
				foreignModel.Calls = []mock.Call{}

				foreignModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

				// get the foreign relationships with the foreign key equal to the primaries of the root
				foreignModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					foreignKeys := s.ForeignFilters()
					if assert.Len(t, foreignKeys, 1) {
						single := foreignKeys[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 1) {
								assert.Contains(t, fv.Values, 3)
							}
						}
					}

					if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fieldSet[0], s.Struct().Primary())
					}

					sv := s.Value.(*[]*ForeignModel)
					(*sv) = append((*sv), &ForeignModel{ID: 4}, &ForeignModel{ID: 7})
				}).Return(nil)

				// clear the relationship foreign keys
				foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					primaries := s.PrimaryFilters()
					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 2) {
								assert.Contains(t, fv.Values, 4)
								assert.Contains(t, fv.Values, 7)
							}
						}
					}

					isFKSelected, err := s.IsSelected("ForeignKey")
					require.NoError(t, err)

					assert.True(t, isFKSelected)

					sv := s.Value.(*ForeignModel)
					assert.Equal(t, 0, sv.ForeignKey)
				}).Return(nil)

				// patch model's with provided id's and set their's foreign keys into 'root' primary
				foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					primaries := s.PrimaryFilters()
					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 2) {
								assert.Contains(t, fv.Values, 1)
								assert.Contains(t, fv.Values, 5)
							}
						}
					}

					isFKSelected, err := s.IsSelected("ForeignKey")
					require.NoError(t, err)

					assert.True(t, isFKSelected)

					sv, ok := s.Value.(*ForeignModel)
					require.True(t, ok)

					assert.Equal(t, model.ID, sv.ForeignKey)
				}).Return(nil)

				foreignModel.Mock.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				err = s.Patch()
				require.NoError(t, err)

				// one call to hasMany - check if exists
				hasMany.AssertNumberOfCalls(t, "List", 1)

				// one call to foreign List - get the primary id's - a part of clear scope
				foreignModel.AssertNumberOfCalls(t, "List", 1)

				// 1. clear scope, 2. patch scope
				foreignModel.AssertNumberOfCalls(t, "Patch", 2)
			})

			t.Run("Clear", func(t *testing.T) {
				model := &HasManyModel{
					ID:      3,
					HasMany: []*ForeignModel{},
				}

				s, err := query.NewC((*controller.Controller)(c), model)
				require.NoError(t, err)

				mr, err := repository.GetRepository(c, model)
				require.NoError(t, err)

				hasMany := mr.(*mocks.Repository)
				hasMany.Calls = []mock.Call{}

				// list the 'hasMany' model primaries - check if exists.
				hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)
					primaries := s.PrimaryFilters()

					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 1) {
								assert.Contains(t, fv.Values, model.ID)
							}
						}
					}

					if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fieldSet[0], s.Struct().Primary())
					}

					sv, ok := s.Value.(*[]*HasManyModel)
					require.True(t, ok)

					(*sv) = append((*sv), &HasManyModel{ID: model.ID})
				}).Return(nil)

				fr, err := repository.GetRepository(c, &ForeignModel{})
				require.NoError(t, err)

				foreignModel := fr.(*mocks.Repository)
				foreignModel.Calls = []mock.Call{}

				foreignModel.On("Begin", mock.Anything, mock.Anything).Once().Return(nil)

				// get the foreign relationships with the foreign key equal to the primaries of the root
				foreignModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					foreignKeys := s.ForeignFilters()
					if assert.Len(t, foreignKeys, 1) {
						single := foreignKeys[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 1) {
								assert.Contains(t, fv.Values, model.ID)
							}
						}
					}

					if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
						assert.Equal(t, fieldSet[0], s.Struct().Primary())
					}

					sv := s.Value.(*[]*ForeignModel)
					(*sv) = append((*sv), &ForeignModel{ID: 4}, &ForeignModel{ID: 7})
				}).Return(nil)

				// clear the relationship foreign keys
				foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					primaries := s.PrimaryFilters()
					if assert.Len(t, primaries, 1) {
						single := primaries[0]

						if assert.Len(t, single.Values(), 1) {
							fv := single.Values()[0]

							assert.Equal(t, filters.OpIn, fv.Operator())
							if assert.Len(t, fv.Values, 2) {
								assert.Contains(t, fv.Values, 4)
								assert.Contains(t, fv.Values, 7)
							}
						}
					}

					isFKSelected, err := s.IsSelected("ForeignKey")
					require.NoError(t, err)

					assert.True(t, isFKSelected)

					sv := s.Value.(*ForeignModel)
					assert.Equal(t, 0, sv.ForeignKey)
				}).Return(nil)

				foreignModel.On("Commit", mock.Anything, mock.Anything).Once().Return(nil)

				err = s.Patch()
				require.NoError(t, err)

				// one call to hasMany - check if exists
				hasMany.AssertNumberOfCalls(t, "List", 1)

				// one call to foreign List - get the primary id's - a part of clear scope
				foreignModel.AssertNumberOfCalls(t, "List", 1)

				// 1. clear scope
				foreignModel.AssertNumberOfCalls(t, "Patch", 1)
			})

			t.Run("Error", func(t *testing.T) {
				// no primary found - list when no attributes were selected
				t.Run("NoRootFound", func(t *testing.T) {
					model := &HasManyModel{
						ID:      3,
						HasMany: []*ForeignModel{},
					}

					s, err := query.NewC((*controller.Controller)(c), model)
					require.NoError(t, err)

					mr, err := repository.GetRepository(c, model)
					require.NoError(t, err)

					hasMany := mr.(*mocks.Repository)
					hasMany.Calls = []mock.Call{}

					// list the 'hasMany' model primaries - check if exists.
					hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
						s := args[1].(*query.Scope)
						primaries := s.PrimaryFilters()

						if assert.Len(t, primaries, 1) {
							single := primaries[0]

							if assert.Len(t, single.Values(), 1) {
								fv := single.Values()[0]

								assert.Equal(t, filters.OpIn, fv.Operator())
								if assert.Len(t, fv.Values, 1) {
									assert.Contains(t, fv.Values, model.ID)
								}
							}
						}

						if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
							assert.Equal(t, fieldSet[0], s.Struct().Primary())
						}

						_, ok := s.Value.(*[]*HasManyModel)
						require.True(t, ok)
					}).Return(errors.New(class.QueryValueNoResult, "no results"))

					err = s.Patch()
					assert.Error(t, err)
				})

				// no related found
				t.Run("NoRelatedFound", func(t *testing.T) {
					model := &HasManyModel{
						ID: 3,
						HasMany: []*ForeignModel{
							&ForeignModel{ID: 6},
							&ForeignModel{ID: 7},
						},
					}

					s, err := query.NewC((*controller.Controller)(c), model)
					require.NoError(t, err)

					mr, err := repository.GetRepository(c, model)
					require.NoError(t, err)

					hasMany := mr.(*mocks.Repository)
					hasMany.Calls = []mock.Call{}

					// list the 'hasMany' model primaries - check if exists.
					hasMany.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
						s := args[1].(*query.Scope)
						primaries := s.PrimaryFilters()

						if assert.Len(t, primaries, 1) {
							single := primaries[0]

							if assert.Len(t, single.Values(), 1) {
								fv := single.Values()[0]

								assert.Equal(t, filters.OpIn, fv.Operator())
								if assert.Len(t, fv.Values, 1) {
									assert.Contains(t, fv.Values, model.ID)
								}
							}
						}

						if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
							assert.Equal(t, fieldSet[0], s.Struct().Primary())
						}

						sv, ok := s.Value.(*[]*HasManyModel)
						require.True(t, ok)

						(*sv) = append((*sv), &HasManyModel{ID: model.ID})
					}).Return(nil)

					fr, err := repository.GetRepository(c, &ForeignModel{})
					require.NoError(t, err)

					foreignModel := fr.(*mocks.Repository)
					foreignModel.Calls = []mock.Call{}

					// get the foreign relationships with the foreign key equal to the primaries of the root
					foreignModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
						s := args[1].(*query.Scope)

						foreignKeys := s.ForeignFilters()
						if assert.Len(t, foreignKeys, 1) {
							single := foreignKeys[0]

							if assert.Len(t, single.Values(), 1) {
								fv := single.Values()[0]

								assert.Equal(t, filters.OpIn, fv.Operator())
								if assert.Len(t, fv.Values, 1) {
									assert.Contains(t, fv.Values, model.ID)
								}
							}
						}

						if fieldSet := s.Fieldset(); assert.Len(t, fieldSet, 1) {
							assert.Equal(t, fieldSet[0], s.Struct().Primary())
						}

						sv, ok := s.Value.(*[]*ForeignModel)
						assert.True(t, ok)

						(*sv) = append((*sv), &ForeignModel{ID: 11})
					}).Return(nil)

					// clear the relationship foreign keys
					foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
						s := args[1].(*query.Scope)

						primaries := s.PrimaryFilters()
						if assert.Len(t, primaries, 1) {
							single := primaries[0]

							if assert.Len(t, single.Values(), 1) {
								fv := single.Values()[0]

								assert.Equal(t, filters.OpIn, fv.Operator())
								if assert.Len(t, fv.Values, 1) {
									assert.Contains(t, fv.Values, 11)
								}
							}
						}

						isFKSelected, err := s.IsSelected("ForeignKey")
						require.NoError(t, err)

						assert.True(t, isFKSelected)

						sv := s.Value.(*ForeignModel)
						assert.Equal(t, 0, sv.ForeignKey)
					}).Return(nil)

					// patch model's with provided id's and set their's foreign keys into 'root' primary
					foreignModel.On("Patch", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
						s := args[1].(*query.Scope)

						primaries := s.PrimaryFilters()
						if assert.Len(t, primaries, 1) {
							single := primaries[0]

							if assert.Len(t, single.Values(), 1) {
								fv := single.Values()[0]

								assert.Equal(t, filters.OpIn, fv.Operator())
								if assert.Len(t, fv.Values, 2) {
									assert.Contains(t, fv.Values, 6)
									assert.Contains(t, fv.Values, 7)
								}
							}
						}

						isFKSelected, err := s.IsSelected("ForeignKey")
						require.NoError(t, err)

						assert.True(t, isFKSelected)

						sv, ok := s.Value.(*ForeignModel)
						require.True(t, ok)

						assert.Equal(t, model.ID, sv.ForeignKey)
					}).Return(errors.New(class.QueryValueNoResult, "no results"))

					err = s.Patch()
					assert.Error(t, err)
				})

				t.Run("Other", func(t *testing.T) {

				})
			})
		})

		t.Run("Many2Many", func(t *testing.T) {
			// TODO: patch many2many tests
		})
	})
}
