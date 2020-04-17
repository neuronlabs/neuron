package query

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type beforeGetter struct {
	ID int `neuron:"type=primary"`
}

func (b *beforeGetter) BeforeGet(ctx context.Context, orm *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type afterGetter struct {
	ID int `neuron:"type=primary"`
}

func (a *afterGetter) AfterGet(ctx context.Context, orm *Scope) error {
	v := ctx.Value(testCtxKey)
	if v == nil {
		return errNotCalled
	}

	return nil
}

type getter struct {
	ID int `neuron:"type=primary"`
}

// TestGet tests the processor get method
func TestGet(t *testing.T) {
	c := newController(t)

	err := c.RegisterModels(&beforeGetter{}, &afterGetter{}, &getter{}, &ForeignModel{}, &HasManyModel{})
	require.NoError(t, err)

	t.Run("NoHooks", func(t *testing.T) {
		s, err := NewC(c, &getter{})
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*Repository)

		repo.On("Get", mock.Anything, mock.Anything).Return(nil)

		err = s.Get()
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything, mock.Anything)
		}
	})

	t.Run("BeforeGet", func(t *testing.T) {
		s, err := NewC(c, &beforeGetter{})
		require.NoError(t, err)

		require.NotNil(t, s.Value)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*Repository)

		repo.On("Get", mock.Anything, mock.Anything).Return(nil)

		err = s.GetContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything, mock.Anything)
		}
	})

	t.Run("AfterGet", func(t *testing.T) {
		s, err := NewC(c, &afterGetter{})
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())

		repo := r.(*Repository)

		repo.On("Get", mock.Anything, mock.Anything).Return(nil)

		err = s.GetContext(context.WithValue(context.Background(), testCtxKey, t))
		if assert.NoError(t, err) {
			repo.AssertCalled(t, "Get", mock.Anything, mock.Anything)
		}
	})

	t.Run("Included", func(t *testing.T) {
		s, err := NewC(c, &HasManyModel{ID: 3})
		require.NoError(t, err)

		err = s.Include("has_many")
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())
		repo := r.(*Repository)

		repo.On("Get", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			_, ok = s.Value.(*HasManyModel)
			require.True(t, ok)
		}).Return(nil)

		foreignRepo, err := s.Controller().GetRepository(&ForeignModel{})
		require.NoError(t, err)

		foreign, ok := foreignRepo.(*Repository)
		require.True(t, ok)

		// first is the reduce the foreign filters
		foreign.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			m, ok := s.Value.(*[]*ForeignModel)
			require.True(t, ok)

			*m = append(*m, &ForeignModel{ID: 3, ForeignKey: 3}, &ForeignModel{ID: 4, ForeignKey: 3})
		}).Return(nil)

		err = s.Get()
		require.NoError(t, err)
	})

	t.Run("Many2Many", func(t *testing.T) {
		c := newController(t)

		err = c.RegisterModels(Many2ManyModel{}, JoinModel{}, RelatedModel{})
		require.NoError(t, err)

		model := &Many2ManyModel{ID: 3}
		s, err := NewC(c, model)
		require.NoError(t, err)

		err = s.Include("Many2Many")
		require.NoError(t, err)

		repo, err := c.GetRepository(model)
		require.NoError(t, err)

		rp, ok := repo.(*Repository)
		require.True(t, ok)

		rp.On("Get", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			// Many2Many
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			modelValue, ok := s.Value.(*Many2ManyModel)
			require.True(t, ok)

			modelValue.ID = 3
		}).Return(nil)

		rp.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			// JoinTable
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			modelValues, ok := s.Value.(*[]*JoinModel)
			require.True(t, ok)

			*modelValues = append(*modelValues,
				&JoinModel{ForeignKey: 3, MtMForeignKey: 5},
				&JoinModel{ForeignKey: 3, MtMForeignKey: 6},
			)
		}).Return(nil)

		rp.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			// RelatedModel
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			modelValues, ok := s.Value.(*[]*RelatedModel)
			require.True(t, ok)

			*modelValues = append(*modelValues,
				&RelatedModel{ID: 5, FloatField: 230.2},
				&RelatedModel{ID: 6, FloatField: 1024.0},
			)
		}).Return(nil)

		err = s.Get()
		require.NoError(t, err)

		if assert.Len(t, model.Many2Many, 2) {
			assert.Equal(t, 5, model.Many2Many[0].ID)
			assert.Equal(t, 230.2, model.Many2Many[0].FloatField)
			assert.Equal(t, 6, model.Many2Many[1].ID)
			assert.Equal(t, 1024.0, model.Many2Many[1].FloatField)
		}
	})

	t.Run("BelongsTo", func(t *testing.T) {
		c := newController(t)

		err = c.RegisterModels(ForeignWithRelation{}, HasManyWithRelation{})
		require.NoError(t, err)

		model := &ForeignWithRelation{ID: 3}
		s, err := NewC(c, model)
		require.NoError(t, err)

		repo, err := c.GetRepository(model)
		require.NoError(t, err)

		rp, ok := repo.(*Repository)
		require.True(t, ok)

		err = s.Include("Relation")
		require.NoError(t, err)

		rp.On("Get", mock.Anything, mock.Anything).Once().
			Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				modelValue, ok := s.Value.(*ForeignWithRelation)
				require.NoError(t, err)

				modelValue.ForeignKey = 5
			}).
			Return(nil)

		rp.On("Get", mock.Anything, mock.Anything).Once().
			Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				hmModel, ok := s.Value.(*HasManyWithRelation)
				require.True(t, ok)
				hmModel.ID = 5
			}).Return(nil)

		rp.On("List", mock.Anything, mock.Anything).Once().
			Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				modelsValues, ok := s.Value.(*[]*ForeignWithRelation)
				require.NoError(t, err)

				*modelsValues = append(*modelsValues, &ForeignWithRelation{ID: 3, ForeignKey: 5}, &ForeignWithRelation{ID: 10, ForeignKey: 5})
			}).Return(nil)

		err = s.Get()
		require.NoError(t, err)
	})

	t.Run("TimeRelated", func(t *testing.T) {
		type timer struct {
			ID        int
			DeletedAt *time.Time
		}

		c := newController(t)
		err := c.RegisterModels(&timer{})
		require.NoError(t, err)

		repo, err := c.GetRepository(timer{})
		require.NoError(t, err)

		timerRepo, ok := repo.(*Repository)
		require.True(t, ok)

		mStruct, err := c.ModelStruct(timer{})
		require.NoError(t, err)

		deletedAt, hasDeletedAt := mStruct.DeletedAt()
		require.True(t, hasDeletedAt)

		t.Run("WithFilter", func(t *testing.T) {
			defer clearRepository(timerRepo)
			s, err := NewC(c, &timer{ID: 2})
			require.NoError(t, err)

			err = s.FilterField(NewFilterField(deletedAt, OpNotNull))
			require.NoError(t, err)

			timerRepo.On("Get", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				if assert.Len(t, s.AttributeFilters, 1) {
					af := s.AttributeFilters[0]
					if assert.Equal(t, deletedAt, af.StructField) {
						if assert.Len(t, af.Values, 1) {
							v := af.Values[0]
							assert.Equal(t, OpNotNull, v.Operator)
						}
					}
				}

				v := s.Value.(*timer)
				v.ID = 2
				tm := time.Now()
				v.DeletedAt = &tm
			}).Return(nil)

			err = s.Get()
			require.NoError(t, err)
		})

		t.Run("WithoutFilter", func(t *testing.T) {
			defer clearRepository(timerRepo)
			s, err := NewC(c, &timer{ID: 3})
			require.NoError(t, err)

			timerRepo.On("Get", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				if assert.Len(t, s.AttributeFilters, 1) {
					af := s.AttributeFilters[0]
					if assert.Equal(t, deletedAt, af.StructField) {
						if assert.Len(t, af.Values, 1) {
							v := af.Values[0]
							assert.Equal(t, OpIsNull, v.Operator)
						}
					}
				}

				v := s.Value.(*timer)
				v.ID = 3
			}).Return(nil)

			err = s.Get()
			require.NoError(t, err)
		})
	})
}
