package scope_test

import (
	iScope "github.com/kucjac/jsonapi/internal/query/scope"
	"github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/jsonapi/query/scope/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

type benchmarkType struct {
	ID int `jsonapi:"type=primary"`
}

func BenchmarkCastScope(b *testing.B) {
	b.StopTimer()
	repo := &mocks.Repository{}
	c := newController(b, repo)

	require.NoError(b, c.RegisterModels(&benchmarkType{}))

	mStruct := c.MustGetModelStruct(&benchmarkType{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = (*scope.Scope)(iScope.New(mStruct))
	}
}

func BenchmarkNoCastScope(b *testing.B) {
	b.StopTimer()
	repo := &mocks.Repository{}
	c := newController(b, repo)

	require.NoError(b, c.RegisterModels(&benchmarkType{}))

	mStruct := c.MustGetModelStruct(&benchmarkType{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = iScope.New(mStruct)
	}
}
