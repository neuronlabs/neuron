package mapping

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkField(b *testing.B) {
	b.StopTimer()
	m := prepareBench(b)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.FieldByName("Third")
	}
}

func BenchmarkFieldByName(b *testing.B) {
	b.StopTimer()
	m := prepareBench(b)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.FieldByName("Third")
	}
}

func BenchmarkAttribute(b *testing.B) {
	b.StopTimer()
	m := prepareBench(b)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Attribute("Third")
	}
}

func BenchmarkMapping(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			m := New(WithNamingConvention(SnakeCase))
			b.StartTimer()

			err := m.RegisterModels(&TModel{})
			require.NoError(b, err)
		}
	})

	b.Run("Relationship", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			m := New(WithNamingConvention(SnakeCase))
			b.StartTimer()

			err := m.RegisterModels(&Model1WithMany2Many{}, &Model2WithMany2Many{}, &JoinModel{}, &First{}, &Second{}, &FirstSeconds{})
			require.NoError(b, err)
		}
	})
}

func prepareBench(b *testing.B) *ModelStruct {
	b.Helper()

	mm := testingModelMap(b)
	err := mm.RegisterModels(&BenchModel{})
	require.NoError(b, err)

	mStruct, err := mm.ModelStruct(&BenchModel{})
	require.NoError(b, err)

	return mStruct
}
