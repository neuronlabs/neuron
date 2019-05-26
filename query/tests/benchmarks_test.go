package tests

import (
	iScope "github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/query"
	"github.com/stretchr/testify/require"
	"testing"
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
		_ = (*query.Scope)(iScope.New(mStruct))
	}
}

func BenchmarkNoCastScope(b *testing.B) {
	b.StopTimer()

	c := newController(b)

	require.NoError(b, c.RegisterModels(&benchmarkType{}))

	mStruct := c.MustGetModelStruct(&benchmarkType{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = iScope.New(mStruct)
	}
}
