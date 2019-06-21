package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron/internal/query/scope"
)

import (
	// import and register mocks repository
	_ "github.com/neuronlabs/neuron/query/mocks"
)

type benchmarkType struct {
	ID int `neuron:"type=primary"`
}

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
