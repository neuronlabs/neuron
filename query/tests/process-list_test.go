package tests

import (
	"context"
	ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/query/filters"
	"github.com/neuronlabs/neuron/repository"
	"github.com/neuronlabs/uni-logger"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type beforeLister struct {
	ID int `neuron:"type=primary"`
}

func (b *beforeLister) HBeforeList(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type afterLister struct {
	ID int `neuron:"type=primary"`
}

func (a *afterLister) HAfterList(ctx context.Context, s *query.Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type lister struct {
	ID int `neuron:"type=primary"`
}

func TestList(t *testing.T) {
	if testing.Verbose() {
		err := log.SetLevel(unilogger.DEBUG)
		require.NoError(t, err)
	}

	c := newController(t)

	err := c.RegisterModels(&beforeLister{}, &afterLister{}, &lister{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		v := []*lister{}
		s, err := query.NewC((*ctrl.Controller)(c), &v)
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("List", mock.Anything, mock.Anything).Return(nil)

		err = s.ListContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookBefore", func(t *testing.T) {
		v := []*beforeLister{}
		s, err := query.NewC((*ctrl.Controller)(c), &v)
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

		repo := r.(*mocks.Repository)

		repo.On("List", mock.Anything, mock.Anything).Return(nil)

		err = s.ListContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "List", mock.Anything, mock.Anything)
		}
	})

	t.Run("HookAfter", func(t *testing.T) {
		v := []*afterLister{}
		s, err := query.NewC((*ctrl.Controller)(c), &v)
		require.NoError(t, err)

		r, _ := repository.GetRepository(s.Controller(), s.Struct())

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

// TestListRelationFilters tests the lists function with the relationship filters
func TestListRelationFilters(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&relationModel{}, &relatedModel{}, &multiRelatedModel{})
	require.NoError(t, err)

	t.Run("BelongsTo", func(t *testing.T) {

		t.Run("MixedFilters", func(t *testing.T) {

			s, err := query.NewC((*ctrl.Controller)(c), &[]*relationModel{})
			require.NoError(t, err)

			require.NoError(t, s.AddStringFilter("[relation_models][relation][some_attr][$eq]", "test-value"))

			repoRoot, err := repository.GetRepository(s.Controller(), &relationModel{})
			require.NoError(t, err)

			repoRelated, err := repository.GetRepository(s.Controller(), &relatedModel{})
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
			s, err := query.NewC((*ctrl.Controller)(c), &([]*relationModel{}))
			require.NoError(t, err)

			require.NoError(t, s.AddStringFilter("[relation_models][relation][id][$eq]", 1))

			repoRoot, err := repository.GetRepository(s.Controller(), &relationModel{})
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
}
