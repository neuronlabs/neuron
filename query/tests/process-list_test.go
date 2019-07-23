package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/query"
	"github.com/neuronlabs/neuron-core/query/filters"
	"github.com/neuronlabs/neuron-core/query/mocks"
)

type beforeLister struct {
	ID int `neuron:"type=primary"`
}

func (b *beforeLister) BeforeList(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type afterLister struct {
	ID int `neuron:"type=primary"`
}

func (a *afterLister) AfterList(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type lister struct {
	ID int `neuron:"type=primary"`
}

// TestList tests the list method for the processor.
func TestList(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&beforeLister{}, &afterLister{}, &lister{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		v := []*lister{}
		s, err := query.NewC((*controller.Controller)(c), &v)
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("List", mock.Anything, mock.Anything).Return(nil)

		err = s.ListContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		v := []*beforeLister{}
		s, err := query.NewC((*controller.Controller)(c), &v)
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("List", mock.Anything, mock.Anything).Return(nil)

		err = s.ListContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		v := []*afterLister{}
		s, err := query.NewC((*controller.Controller)(c), &v)
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("List", mock.Anything, mock.Anything).Return(nil)

		err = s.ListContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything, mock.Anything)
		}
	})

}

type relationModel struct {
	ID       int           `neuron:"type=primary"`
	Relation *relatedModel `neuron:"type=relation;foreign=FK"`
	FK       int           `neuron:"type=foreign"`
}
type relatedModel struct {
	ID       int            `neuron:"type=primary"`
	Relation *relationModel `neuron:"type=relation;foreign=FK"`
	SomeAttr string         `neuron:"type=attr"`
}

type multiRelatedModel struct {
	ID        int              `neuron:"type=primary"`
	Relations []*relationModel `neuron:"type=relation;foreign=FK"`
}

// TestListRelationshipFilters tests the lists function with the relationship filters
func TestListRelationshipFilters(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&relationModel{}, &relatedModel{}, &multiRelatedModel{})
	require.NoError(t, err)

	t.Run("BelongsTo", func(t *testing.T) {
		t.Run("MixedFilters", func(t *testing.T) {
			s, err := query.NewC((*controller.Controller)(c), &[]*relationModel{})
			require.NoError(t, err)

			require.NoError(t, s.Filter("[relation_models][relation][some_attr][$eq]", "test-value"))

			repoRoot, err := s.Controller().GetRepository(&relationModel{})
			require.NoError(t, err)

			repoRelated, err := s.Controller().GetRepository(&relatedModel{})
			require.NoError(t, err)

			rr := repoRoot.(*mocks.Repository)

			// there should be one query on root repo
			rr.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)

				foreignFilters := s.ForeignFilters()
				if assert.Len(t, foreignFilters, 1) {
					if assert.Len(t, foreignFilters[0].Values(), 1) {
						if assert.Len(t, foreignFilters[0].Values()[0].Values, 2) {
							assert.Equal(t, foreignFilters[0].Values()[0].Values[0], 4)
							assert.Equal(t, foreignFilters[0].Values()[0].Values[1], 3)
							assert.Equal(t, foreignFilters[0].Values()[0].Operator(), filters.OpIn)
						}
					}
				}
			}).Return(nil)

			rl := repoRelated.(*mocks.Repository)
			rl.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)

				attrFilters := s.AttributeFilters()
				if assert.NotEmpty(t, attrFilters) {
					if assert.Len(t, attrFilters[0].Values(), 1) {
						if assert.Len(t, attrFilters[0].Values()[0].Values, 1) {
							assert.Equal(t, attrFilters[0].Values()[0].Values[0], "test-value")
							assert.Equal(t, attrFilters[0].Values()[0].Operator(), filters.OpEqual)
						}
					}
				}
				if assert.Len(t, s.Fieldset(), 1) {
					_, ok := s.InFieldset("id")
					assert.True(t, ok)
				}

				s.Value = &([]*relationModel{{ID: 4}, {ID: 3}})
			}).Return(nil)

			require.NotPanics(t, func() {
				require.NoError(t, s.List())
			})
		})

		t.Run("OnlyPrimes", func(t *testing.T) {
			s, err := query.NewC((*controller.Controller)(c), &([]*relationModel{}))
			require.NoError(t, err)

			require.NoError(t, s.Filter("[relation_models][relation][id][$eq]", 1))

			repoRoot, err := s.Controller().GetRepository(&relationModel{})
			require.NoError(t, err)

			rr := repoRoot.(*mocks.Repository)

			// there should be one query on root repo
			rr.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				if assert.Len(t, args, 2) {
					s := args[1].(*query.Scope)

					ff := s.ForeignFilters()
					if assert.NotEmpty(t, ff) {
						if assert.Len(t, ff[0].Values(), 1) {
							if assert.Len(t, ff[0].Values()[0].Values, 1) {
								assert.Equal(t, ff[0].Values()[0].Values[0], 1)
								assert.Equal(t, ff[0].Values()[0].Operator(), filters.OpEqual)
							}
						}
					}
				}
			}).Return(nil)

			require.NotPanics(t, func() {
				require.NoError(t, s.List())
			})
		})
	})

	t.Run("HasOne", func(t *testing.T) {
		relatedValues := []*relatedModel{}
		s, err := query.NewC((*controller.Controller)(c), &relatedValues)
		require.NoError(t, err)

		require.NoError(t, s.Filter("[related_models][relation][id][$eq]", "1"))

		relatedRepo, err := s.Controller().GetRepository(&relatedModel{})
		require.NoError(t, err)

		relationRepo, err := s.Controller().GetRepository(&relationModel{})
		require.NoError(t, err)

		relation := relationRepo.(*mocks.Repository)
		relation.On("List", mock.Anything, mock.Anything).Once().Run(
			func(args mock.Arguments) {
				s := args[1].(*query.Scope)
				primaryFilters := s.PrimaryFilters()
				if assert.Len(t, primaryFilters, 1) {
					values := primaryFilters[0].Values()
					if assert.Len(t, values, 1) {
						assert.Equal(t, filters.OpEqual, values[0].Operator())
						assert.Equal(t, "1", values[0].Values[0])
					}
				}
				_, ok := s.InFieldset("FK")
				assert.True(t, ok)

				values, ok := s.Value.(*[]*relationModel)
				require.True(t, ok)

				(*values) = append((*values), &relationModel{ID: 1, FK: 5})
			}).Return(nil)

		related := relatedRepo.(*mocks.Repository)
		related.On("List", mock.Anything, mock.Anything).Once().Run(
			func(args mock.Arguments) {
				s := args[1].(*query.Scope)
				primaries := s.PrimaryFilters()
				if assert.Len(t, primaries, 1) {
					values := primaries[0].Values()
					if assert.Len(t, values, 1) {
						op := values[0]
						assert.Equal(t, filters.OpIn, op.Operator())
						assert.Equal(t, 5, op.Values[0])
					}
				}
				values := s.Value.(*[]*relatedModel)
				(*values) = append((*values), &relatedModel{ID: 5, SomeAttr: "Attr"})
			}).Return(nil)

		// after list
		relation.On("List", mock.Anything, mock.Anything).Once().Run(
			func(args mock.Arguments) {
				s := args[1].(*query.Scope)
				foreignFilters := s.ForeignFilters()
				if assert.Len(t, foreignFilters, 1) {
					values := foreignFilters[0].Values()
					if assert.Len(t, values, 1) {
						assert.Equal(t, filters.OpIn, values[0].Operator())
						assert.Equal(t, 5, values[0].Values[0])
					}
				}

				_, ok := s.InFieldset("FK")
				assert.True(t, ok)

				values, ok := s.Value.(*[]*relationModel)
				require.True(t, ok)

				(*values) = append((*values), &relationModel{ID: 1, FK: 5})
			}).Return(nil)

		err = s.List()
		require.NoError(t, err)

		if assert.Len(t, relatedValues, 1) {
			v := relatedValues[0]
			assert.Equal(t, 5, v.ID)
			if assert.NotNil(t, v.Relation) {
				assert.Equal(t, 1, v.Relation.ID)
			}
		}
	})

	t.Run("HasMany", func(t *testing.T) {
		t.Run("WithRelated", func(t *testing.T) {
			// create the multiRelatedModel scope.
			relatedValues := []*multiRelatedModel{}
			s, err := query.NewC((*controller.Controller)(c), &relatedValues)
			require.NoError(t, err)

			require.NoError(t, s.Filter("[multi_related_models][relations][id][$in]", "1", "2"))

			relationRepo, err := s.Controller().GetRepository(&relationModel{})
			require.NoError(t, err)

			// handle initial filter model list.
			relation := relationRepo.(*mocks.Repository)
			relation.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)
				primaries := s.PrimaryFilters()

				if assert.Len(t, primaries, 1) {
					op := primaries[0].Values()
					if assert.Len(t, op, 1) {
						assert.Equal(t, filters.OpIn, op[0].Operator())
						values := op[0].Values
						if assert.Len(t, values, 2) {
							assert.Equal(t, values[0], "1")
							assert.Equal(t, values[1], "2")
						}
					}
				}

				v, ok := s.Value.(*[]*relationModel)
				require.True(t, ok)

				(*v) = append((*v), []*relationModel{{ID: 1, FK: 3}, {ID: 2, FK: 5}, {ID: 3, FK: 3}}...)
			}).Return(nil)

			relatedRepo, err := s.Controller().GetRepository(&multiRelatedModel{})
			require.NoError(t, err)

			multi := relatedRepo.(*mocks.Repository)

			multi.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)
				primaries := s.PrimaryFilters()

				if assert.Len(t, primaries, 1) {
					op := primaries[0].Values()
					if assert.Len(t, op, 1) {
						assert.Equal(t, filters.OpIn, op[0].Operator())
						values := op[0].Values

						assert.Len(t, values, 2)
						assert.Contains(t, values, 3)
						assert.Contains(t, values, 5)
					}
				}

				v, ok := s.Value.(*[]*multiRelatedModel)
				require.True(t, ok)

				(*v) = append((*v), []*multiRelatedModel{{ID: 3}, {ID: 5}}...)
			}).Return(nil)

			// handle after getting list
			relation.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)
				foreigns := s.ForeignFilters()

				if assert.Len(t, foreigns, 1) {
					op := foreigns[0].Values()
					if assert.Len(t, op, 1) {
						assert.Equal(t, filters.OpIn, op[0].Operator())
						values := op[0].Values
						if assert.Len(t, values, 2) {
							assert.Equal(t, values[0], 3)
							assert.Equal(t, values[1], 5)
						}
					}
				}

				v, ok := s.Value.(*[]*relationModel)
				require.True(t, ok)

				(*v) = append((*v), []*relationModel{{ID: 1, FK: 3}, {ID: 2, FK: 5}, {ID: 3, FK: 3}}...)
			}).Return(nil)

			err = s.List()
			require.NoError(t, err)

			if assert.Len(t, relatedValues, 2) {
				if assert.Equal(t, 3, relatedValues[0].ID) {
					relations := relatedValues[0].Relations
					if assert.Len(t, relations, 2) {
						for _, relation := range relations {
							switch relation.ID {
							case 1, 3:
							default:
								t.Errorf("invalid relaiton id: %v", relation.ID)
							}
						}

					}
				}

				if assert.Equal(t, 5, relatedValues[1].ID) {
					relations := relatedValues[1].Relations
					if assert.Len(t, relations, 1) {
						for _, relation := range relations {
							switch relation.ID {
							case 2:
							default:
								t.Errorf("invalid relaiton id: %v", relation.ID)
							}
						}
					}
				}
			}
		})

		t.Run("NoFilterResult", func(t *testing.T) {
			s, err := query.NewC((*controller.Controller)(c), &[]*multiRelatedModel{})
			require.NoError(t, err)

			require.NoError(t, s.Filter("[multi_related_models][relations][id][$in]", "1", "2"))

			relationRepo, err := s.Controller().GetRepository(&relationModel{})
			require.NoError(t, err)

			// handle initial filter model list.
			relation := relationRepo.(*mocks.Repository)
			relation.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)
				primaries := s.PrimaryFilters()

				if assert.Len(t, primaries, 1) {
					op := primaries[0].Values()
					if assert.Len(t, op, 1) {
						assert.Equal(t, filters.OpIn, op[0].Operator())
						values := op[0].Values
						if assert.Len(t, values, 2) {
							assert.Equal(t, values[0], "1")
							assert.Equal(t, values[1], "2")
						}
					}
				}
			}).Return(nil)

			err = s.List()
			require.Error(t, err)

			e, ok := err.(*errors.Error)
			require.True(t, ok)

			assert.Equal(t, class.QueryValueNoResult, e.Class)
		})
	})

	t.Run("Many2Many", func(t *testing.T) {
		t.Run("OnlyPrimaries", func(t *testing.T) {
			c := newController(t)
			err := c.RegisterModels(Many2ManyModel{}, JoinModel{}, RelatedModel{})
			require.NoError(t, err)

			scopeValue := []*Many2ManyModel{}

			s, err := query.NewC((*controller.Controller)(c), &scopeValue)
			require.NoError(t, err)

			err = s.Filter("[many_2_many_models][many_2_many][id][$in]", "1", "2", "4")
			require.NoError(t, err)

			m2mRepo, err := c.GetRepository(s.Struct())
			require.NoError(t, err)

			m2m := m2mRepo.(*mocks.Repository)

			jmRepo, err := c.GetRepository(JoinModel{})
			require.NoError(t, err)

			joinModel := jmRepo.(*mocks.Repository)

			// get initial list of join model filter
			joinModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)

				foreignFilters := s.ForeignFilters()
				if assert.Len(t, foreignFilters, 1) {
					mtmFK, ok := s.Struct().ForeignKey("MtMForeignKey")
					if assert.True(t, ok) {
						field := foreignFilters[0].StructField()
						assert.Equal(t, field, mtmFK)
					}

					opValues := foreignFilters[0].Values()
					if assert.Len(t, opValues, 1) {
						assert.Equal(t, filters.OpIn, opValues[0].Operator())
						if assert.Len(t, opValues[0].Values, 3) {
							assert.Contains(t, opValues[0].Values, "1")
							assert.Contains(t, opValues[0].Values, "2")
							assert.Contains(t, opValues[0].Values, "4")
						}
					}
				}

				jmValues, ok := s.Value.(*[]*JoinModel)
				if assert.True(t, ok) {
					(*jmValues) = append((*jmValues),
						&JoinModel{ForeignKey: 4, MtMForeignKey: 1},
						&JoinModel{ForeignKey: 8, MtMForeignKey: 2},
						&JoinModel{ForeignKey: 16, MtMForeignKey: 4},
					)
				}
			}).Return(nil)

			m2m.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)

				primaries := s.PrimaryFilters()
				if assert.Len(t, primaries, 1) {
					opValues := primaries[0].Values()

					if assert.Len(t, opValues, 1) {
						values := opValues[0]

						assert.Equal(t, filters.OpIn, values.Operator())
						if assert.Len(t, values.Values, 3) {
							assert.Contains(t, values.Values, 4)
							assert.Contains(t, values.Values, 8)
							assert.Contains(t, values.Values, 16)
						}
					}
				}

				v, ok := s.Value.(*[]*Many2ManyModel)
				if assert.True(t, ok) {
					(*v) = append((*v),
						&Many2ManyModel{ID: 4},
						&Many2ManyModel{ID: 8},
						&Many2ManyModel{ID: 16},
					)
				}
			}).Return(nil)

			// Now again query the join table
			joinModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)

				foreignKeys := s.ForeignFilters()
				if assert.Len(t, foreignKeys, 1) {
					foreignKey, ok := s.Struct().ForeignKey("ForeignKey")
					if assert.True(t, ok) {
						assert.Equal(t, foreignKey, foreignKeys[0].StructField())
					}

					opValues := foreignKeys[0].Values()
					if assert.Len(t, opValues, 1) {
						assert.Equal(t, filters.OpIn, opValues[0].Operator())

						values := opValues[0].Values
						if assert.Len(t, values, 3) {
							assert.Contains(t, values, 4)
							assert.Contains(t, values, 8)
							assert.Contains(t, values, 16)
						}
					}
				}

				v, ok := s.Value.(*[]*JoinModel)
				if assert.True(t, ok) {
					(*v) = append((*v),
						&JoinModel{ForeignKey: 4, MtMForeignKey: 1},
						&JoinModel{ForeignKey: 8, MtMForeignKey: 2},
						&JoinModel{ForeignKey: 16, MtMForeignKey: 4},
					)
				}
			}).Return(nil)

			err = s.List()
			assert.NoError(t, err)

			if assert.Equal(t, 4, scopeValue[0].ID) {
				if assert.Len(t, scopeValue[0].Many2Many, 1) {
					mtm := scopeValue[0].Many2Many[0]
					assert.Equal(t, 1, mtm.ID)
				}
			}

			if assert.Equal(t, 8, scopeValue[1].ID) {
				if assert.Len(t, scopeValue[1].Many2Many, 1) {
					mtm := scopeValue[1].Many2Many[0]
					assert.Equal(t, 2, mtm.ID)
				}
			}

			if assert.Equal(t, 16, scopeValue[2].ID) {
				if assert.Len(t, scopeValue[2].Many2Many, 1) {
					mtm := scopeValue[2].Many2Many[0]
					assert.Equal(t, 4, mtm.ID)
				}
			}

			m2m.AssertNumberOfCalls(t, "List", 1)
			joinModel.AssertNumberOfCalls(t, "List", 2)
		})

		t.Run("RelatedFilter", func(t *testing.T) {
			t.Run("RelatedExists", func(t *testing.T) {
				c := newController(t)
				err := c.RegisterModels(Many2ManyModel{}, JoinModel{}, RelatedModel{})
				require.NoError(t, err)

				scopeValue := []*Many2ManyModel{}

				s, err := query.NewC((*controller.Controller)(c), &scopeValue)
				require.NoError(t, err)

				err = s.Filter("[many_2_many_models][many_2_many][float_field][$gt]", "1.2415")
				require.NoError(t, err)

				m2mRepo, err := c.GetRepository(s.Struct())
				require.NoError(t, err)

				m2m := m2mRepo.(*mocks.Repository)

				jmRepo, err := c.GetRepository(JoinModel{})
				require.NoError(t, err)

				joinModel := jmRepo.(*mocks.Repository)

				relatedRepo, err := c.GetRepository(RelatedModel{})
				require.NoError(t, err)

				related := relatedRepo.(*mocks.Repository)

				// at first find related repository filter
				related.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					attrFilters := s.AttributeFilters()
					if assert.Len(t, attrFilters, 1) {
						attrF := attrFilters[0]

						shouldBeField, ok := s.Struct().Attr("FloatField")
						require.True(t, ok)

						assert.Equal(t, shouldBeField, attrF.StructField())

						opValues := attrF.Values()
						if assert.Len(t, opValues, 1) {
							opV := opValues[0]

							assert.Equal(t, filters.OpGreaterThan, opV.Operator())
							if assert.Len(t, opV.Values, 1) {
								assert.Equal(t, opV.Values[0], "1.2415")
							}
						}
					}

					fieldSet := s.Fieldset()
					assert.Len(t, fieldSet, 1)
					_, isPrimary := s.InFieldset("ID")
					assert.True(t, isPrimary)

					sv := s.Value.(*[]*RelatedModel)

					(*sv) = append((*sv), &RelatedModel{ID: 5})
				}).Return(nil)

				// now the join model should be queried by the mtmForeignKey equal to the relatedScope primary.
				joinModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					foreignFilters := s.ForeignFilters()
					if assert.Len(t, foreignFilters, 1) {
						foreignField, ok := s.Struct().ForeignKey("MtMForeignKey")
						require.True(t, ok)

						assert.Equal(t, foreignField, foreignFilters[0].StructField())

						if assert.Len(t, foreignFilters[0].Values(), 1) {
							op := foreignFilters[0].Values()[0]

							assert.Equal(t, filters.OpIn, op.Operator())
							if assert.Len(t, op.Values, 1) {
								assert.Equal(t, 5, op.Values[0])
							}
						}
					}

					sv := s.Value.(*[]*JoinModel)
					(*sv) = append((*sv),
						&JoinModel{ID: 1, ForeignKey: 2, MtMForeignKey: 5},
						&JoinModel{ID: 3, ForeignKey: 4, MtMForeignKey: 5},
					)
				}).Return(nil)

				m2m.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					primaries := s.PrimaryFilters()
					if assert.Len(t, primaries, 1) {
						pmValues := primaries[0].Values()
						if assert.Len(t, pmValues, 1) {
							op := pmValues[0]

							assert.Equal(t, filters.OpIn, op.Operator())
							if assert.Len(t, op.Values, 2) {
								assert.Contains(t, op.Values, 2)
								assert.Contains(t, op.Values, 4)
							}
						}
					}

					sv := s.Value.(*[]*Many2ManyModel)
					(*sv) = append((*sv), &Many2ManyModel{ID: 2}, &Many2ManyModel{ID: 4})
				}).Return(nil)

				// now the join model should be queried by the mtmForeignKey equal to the relatedScope primary.
				joinModel.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					foreignFilters := s.ForeignFilters()
					if assert.Len(t, foreignFilters, 1) {
						foreignField, ok := s.Struct().ForeignKey("ForeignKey")
						require.True(t, ok)

						assert.Equal(t, foreignField, foreignFilters[0].StructField())

						if assert.Len(t, foreignFilters[0].Values(), 1) {
							op := foreignFilters[0].Values()[0]

							assert.Equal(t, filters.OpIn, op.Operator())
							if assert.Len(t, op.Values, 2) {
								assert.Contains(t, op.Values, 2)
								assert.Contains(t, op.Values, 4)
							}
						}
					}

					_, ok := s.InFieldset("ForeignKey")
					assert.True(t, ok)
					_, ok = s.InFieldset("MtMForeignKey")
					assert.True(t, ok)

					sv := s.Value.(*[]*JoinModel)
					(*sv) = append((*sv),
						&JoinModel{ForeignKey: 2, MtMForeignKey: 5},
						&JoinModel{ForeignKey: 4, MtMForeignKey: 5},
					)
				}).Return(nil)

				err = s.List()
				require.NoError(t, err)

				related.AssertNumberOfCalls(t, "List", 1)
				joinModel.AssertNumberOfCalls(t, "List", 2)
				m2m.AssertNumberOfCalls(t, "List", 1)
			})

			t.Run("RelatedNotExists", func(t *testing.T) {
				c := newController(t)
				err := c.RegisterModels(Many2ManyModel{}, JoinModel{}, RelatedModel{})
				require.NoError(t, err)

				scopeValue := []*Many2ManyModel{}

				s, err := query.NewC((*controller.Controller)(c), &scopeValue)
				require.NoError(t, err)

				err = s.Filter("[many_2_many_models][many_2_many][float_field][$gt]", "1.2415")
				require.NoError(t, err)

				relatedRepo, err := c.GetRepository(RelatedModel{})
				require.NoError(t, err)

				related := relatedRepo.(*mocks.Repository)

				// at first find related repository filter
				related.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					s := args[1].(*query.Scope)

					attrFilters := s.AttributeFilters()
					if assert.Len(t, attrFilters, 1) {
						attrF := attrFilters[0]

						shouldBeField, ok := s.Struct().Attr("FloatField")
						require.True(t, ok)

						assert.Equal(t, shouldBeField, attrF.StructField())

						opValues := attrF.Values()
						if assert.Len(t, opValues, 1) {
							opV := opValues[0]

							assert.Equal(t, filters.OpGreaterThan, opV.Operator())
							if assert.Len(t, opV.Values, 1) {
								assert.Equal(t, opV.Values[0], "1.2415")
							}
						}
					}

					fieldSet := s.Fieldset()
					assert.Len(t, fieldSet, 1)
					_, isPrimary := s.InFieldset("ID")
					assert.True(t, isPrimary)
				}).Return(errors.New(class.QueryValueNoResult, "no results"))

				require.NotPanics(t, func() {
					err = s.List()
					require.Error(t, err)
				})
			})
		})
	})

	t.Run("Include", func(t *testing.T) {
		t.Run("InFieldset", func(t *testing.T) {
			relatedValues := []*relatedModel{}

			s, err := query.NewC((*controller.Controller)(c), &relatedValues)
			require.NoError(t, err)

			err = s.IncludeFields("relation")
			require.NoError(t, err)

			relatedRepo, err := c.GetRepository(s.Struct())
			require.NoError(t, err)

			relatedModelRepo, ok := relatedRepo.(*mocks.Repository)
			require.True(t, ok)

			// List the related model values
			relatedModelRepo.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*query.Scope)
				require.True(t, ok)

				values, ok := s.Value.(*[]*relatedModel)
				require.True(t, ok)

				(*values) = append((*values), &relatedModel{ID: 3})
				(*values) = append((*values), &relatedModel{ID: 4})

			}).Return(nil)

			relField, ok := s.Struct().RelationField("Relation")
			require.True(t, ok)

			// get included values
			relationRepo, err := c.GetRepository(relField.Relationship().ModelStruct())
			require.NoError(t, err)

			relationModelRepo, ok := relationRepo.(*mocks.Repository)
			require.True(t, ok)

			// get related models
			relationModelRepo.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*query.Scope)
				require.True(t, ok)

				values, ok := s.Value.(*[]*relationModel)
				require.True(t, ok)

				// check primary filter values

				(*values) = append((*values), &relationModel{ID: 45, FK: 3})
				(*values) = append((*values), &relationModel{ID: 56, FK: 4})
			}).Return(nil)

			firstIncluded := &relationModel{ID: 45, FK: 3}
			secondIncluded := &relationModel{ID: 56, FK: 4}

			// get included models
			relationModelRepo.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*query.Scope)
				require.True(t, ok)

				values, ok := s.Value.(*[]*relationModel)
				require.True(t, ok)

				// check primary filter values

				(*values) = append((*values), firstIncluded)
				(*values) = append((*values), secondIncluded)
			}).Return(nil)

			err = s.List()
			require.NoError(t, err)

			values, err := s.IncludedModelValues(&relationModel{})
			require.NoError(t, err)

			rValues, ok := values.(*[]*relationModel)
			require.True(t, ok, "%T", values)

			if assert.Len(t, *rValues, 2) {
				assert.Contains(t, *rValues, firstIncluded)
				assert.Contains(t, *rValues, secondIncluded)
			}
		})
	})
}
