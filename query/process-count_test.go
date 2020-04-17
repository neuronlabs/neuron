package query

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
)

type countModel1 struct {
	ID int `neuron:"type=primary"`
}

func (c *countModel1) BeforeCount(ctx context.Context, s *Scope) error {
	if err := s.FilterField(NewFilterField(s.Struct().Primary(), OpLessThan, 50)); err != nil {
		return err
	}
	return nil
}

var _ AfterCounter = &countModel2{}

type countModel2 struct {
	ID int `neuron:"type=primary"`
}

func (c *countModel2) AfterCount(ctx context.Context, s *Scope) error {
	v, ok := s.Value.(int64)
	if !ok {
		return errors.New(class.InternalQueryInvalidValue, "after counting value is not int64")
	}
	v += 50
	s.Value = v
	return nil
}

// TestCount tests the Count method.
func TestCount(t *testing.T) {
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

	c := newController(t)
	err := c.RegisterModels(foreignKeyModel{}, hasOneModel{}, hasManyModel{})
	require.NoError(t, err)

	t.Run("All", func(t *testing.T) {
		model := &foreignKeyModel{}
		s, err := NewC(c, model)
		require.NoError(t, err)

		repo, err := c.GetRepository(model)
		require.NoError(t, err)

		fkRepo, ok := repo.(*Repository)
		require.True(t, ok)

		defer clearRepository(fkRepo)

		count := int64(329)
		fkRepo.On("Count", mock.Anything, mock.Anything).Once().Return(count, nil)

		i, err := s.Count()
		require.NoError(t, err)

		assert.Equal(t, count, i)
	})

	t.Run("ForeignRelationships", func(t *testing.T) {
		t.Run("HasOne", func(t *testing.T) {
			model := &hasOneModel{}
			s, err := NewC(c, model)
			require.NoError(t, err)

			repo, err := c.GetRepository(model)
			require.NoError(t, err)

			hasOneRepo, ok := repo.(*Repository)
			require.True(t, ok)

			defer clearRepository(hasOneRepo)

			repo, err = c.GetRepository(foreignKeyModel{})
			require.NoError(t, err)

			fkRepo, ok := repo.(*Repository)
			require.True(t, ok)

			fkRepo.On("List", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				if assert.Len(t, s.PrimaryFilters, 1) {
					pf := s.PrimaryFilters[0]
					if assert.Len(t, pf.Values, 1) {
						v := pf.Values[0]
						assert.Equal(t, OpIn, v.Operator)
						assert.Equal(t, v.Values[0], 12)
					}
				}
				fkVal, ok := s.Value.(*[]*foreignKeyModel)
				require.True(t, ok)

				assert.Len(t, s.Fieldset, 1)
				assert.Contains(t, s.Fieldset, "fk")

				*fkVal = append(*fkVal, &foreignKeyModel{FK: 3}, &foreignKeyModel{FK: 12})
			}).Return(nil)

			count := int64(2)
			hasOneRepo.On("Count", mock.Anything, mock.Anything).Once().Return(count, nil)

			err = s.Filter("has_one.id IN", 12)
			require.NoError(t, err)

			i, err := s.Count()
			require.NoError(t, err)

			assert.Equal(t, count, i)
		})
	})

	t.Run("Hooks", func(t *testing.T) {
		t.Run("Before", func(t *testing.T) {
			c := newController(t)
			err := c.RegisterModels(&countModel1{})
			require.NoError(t, err)

			repo, err := c.GetRepository(countModel1{})
			require.NoError(t, err)

			rp, ok := repo.(*Repository)
			require.True(t, ok)

			count := int64(49)
			rp.On("Count", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				s, ok := args[1].(*Scope)
				require.True(t, ok)

				if assert.Len(t, s.PrimaryFilters, 1) {
					pf := s.PrimaryFilters[0]
					if assert.Len(t, pf.Values, 1) {
						v := pf.Values[0]
						assert.Equal(t, OpLessThan, v.Operator)
						assert.Equal(t, 50, v.Values[0])
					}
				}
			}).Return(count, nil)

			s, err := NewC(c, &countModel1{})
			require.NoError(t, err)

			// no primary functions added here.
			i, err := s.CountContext(context.Background())
			require.NoError(t, err)

			assert.Equal(t, count, i)
		})

		t.Run("After", func(t *testing.T) {
			c := newController(t)
			err := c.RegisterModels(&countModel2{})
			require.NoError(t, err)

			repo, err := c.GetRepository(countModel2{})
			require.NoError(t, err)

			rp, ok := repo.(*Repository)
			require.True(t, ok)

			count := int64(49)
			rp.On("Count", mock.Anything, mock.Anything).Once().Return(count, nil)

			s, err := NewC(c, &countModel2{})
			require.NoError(t, err)

			// no primary functions added here.
			i, err := s.CountContext(context.Background())
			require.NoError(t, err)

			// the AfterCount method adds 50 to the count.
			assert.Equal(t, count+50, i)
		})
	})
}
