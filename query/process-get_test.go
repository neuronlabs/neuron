package query

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron-core/controller"
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
		s, err := NewC((*controller.Controller)(c), &getter{})
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
		s, err := NewC((*controller.Controller)(c), &beforeGetter{})
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
		s, err := NewC((*controller.Controller)(c), &afterGetter{})
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
		s, err := NewC((*controller.Controller)(c), &HasManyModel{ID: 3})
		require.NoError(t, err)

		err = s.IncludeFields("has_many")
		require.NoError(t, err)

		err = s.SetFieldset(s.Struct().Primary())
		require.NoError(t, err)

		r, _ := s.Controller().GetRepository(s.Struct())
		repo := r.(*Repository)

		repo.On("Get", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			m, ok := s.Value.(*HasManyModel)
			require.True(t, ok)

			m.HasMany = []*ForeignModel{{ID: 3}, {ID: 4}}
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

			(*m) = append((*m), &ForeignModel{ID: 3}, &ForeignModel{ID: 4})
		}).Return(nil)

		// the second call is just a list call
		foreign.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			s, ok := args[1].(*Scope)
			require.True(t, ok)

			m, ok := s.Value.(*[]*ForeignModel)
			require.True(t, ok)

			(*m) = append((*m), &ForeignModel{ID: 3}, &ForeignModel{ID: 4})
		}).Return(nil)

		err = s.Get()
		require.NoError(t, err)

		foreignIncludes, err := s.IncludedModelValues(&ForeignModel{})
		require.NoError(t, err)

		_, ok = foreignIncludes[3]
		assert.True(t, ok)

		_, ok = foreignIncludes[4]
		assert.True(t, ok)

		assert.Len(t, foreignIncludes, 2)
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

			err = s.FilterField(NewFilter(deletedAt, OpNotNull))
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
