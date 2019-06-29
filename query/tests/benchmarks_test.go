package tests

import (
	"github.com/neuronlabs/neuron/controller"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/repository"

	"github.com/neuronlabs/neuron/internal/query/scope"
)

import (
	// import and register mocks repository
	"github.com/neuronlabs/neuron/query/mocks"
)

type benchmarkType struct {
	ID int `neuron:"type=primary"`
}

// BenchmarkCastScope benchmarks the casting time from *scope.Scope into *query.Scope
func BenchmarkCastScope(b *testing.B) {
	b.StopTimer()
	c := newController(b)

	require.NoError(b, c.RegisterModels(&benchmarkType{}))

	mStruct := c.MustGetModelStruct(&benchmarkType{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = (*query.Scope)(scope.New(mStruct))
	}
}

// BenchmarkNoCastScope is a compare benchmark for the BenchmarkCastScope, that takes raw *scope.Scope.
func BenchmarkNoCastScope(b *testing.B) {
	b.StopTimer()

	c := newController(b)

	require.NoError(b, c.RegisterModels(&benchmarkType{}))

	mStruct := c.MustGetModelStruct(&benchmarkType{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = scope.New(mStruct)
	}
}

func BenchmarkScopeCreate(b *testing.B) {
	b.Helper()

	c := newController(b)
	require.NoError(b, c.RegisterModels(HasOneModel{}, HasManyModel{}, ForeignModel{}, Many2ManyModel{}, JoinModel{}, RelatedModel{}))

	b.Run("Related", func(b *testing.B) {
		b.Run("HasOne", func(b *testing.B) {
			c := (*controller.Controller)(c)

			repo, err := repository.GetRepository(c, HasOneModel{})
			require.NoError(b, err)

			hasOne, ok := repo.(*mocks.Repository)
			require.True(b, ok)

			defer clearRepository(hasOne)

			hasOne.On("Begin", mock.Anything, mock.Anything).Return(nil)
			hasOne.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				s := args[1].(*query.Scope)

				v := s.Value.(*HasOneModel)
				v.ID = 5
			}).Return(nil)
			hasOne.On("Commit", mock.Anything, mock.Anything).Return(nil)

			repo, err = repository.GetRepository(c, ForeignModel{})
			require.NoError(b, err)

			foreign, ok := repo.(*mocks.Repository)
			require.True(b, ok)

			defer clearRepository(foreign)

			foreign.On("Begin", mock.Anything, mock.Anything).Return(nil)
			foreign.On("List", mock.Anything, mock.Anything).Return(func(args mock.Arguments) {
				s := args[1].(*query.Scope)

				v := s.Value.(*[]*ForeignModel)
				(*v) = append((*v), &ForeignModel{ID: 1})
			}).Return(nil)

			foreign.On("Patch", mock.Anything, mock.Anything).Return(nil)
			foreign.On("Commit", mock.Anything, mock.Anything).Return(nil)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				model := &HasOneModel{HasOne: &ForeignModel{ID: 1}}
				s, err := query.NewC(c, model)
				require.NoError(b, err)

				err = s.Create()
				require.NoError(b, err)
			}
			b.StopTimer()
		})
	})
}
