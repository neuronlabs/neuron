package mapping

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/namer"
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
	type Model struct {
		ID          int
		CreatedTime time.Time  `neuron:"flags=created_at"`
		UpdatedTime *time.Time `neuron:"flags=updated_at"`
		DeletedTime *time.Time `neuron:"flags=deleted_at"`
		Number      int
		String      string
	}

	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			cfg := config.DefaultController()
			m := NewModelMap(namer.NamingSnake, cfg)
			b.StartTimer()

			err := m.RegisterModels(Model{})
			require.NoError(b, err)
		}
	})

	b.Run("Relationship", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			cfg := config.DefaultController()
			m := NewModelMap(namer.NamingSnake, cfg)
			b.StartTimer()

			err := m.RegisterModels(Model1WithMany2Many{}, Model2WithMany2Many{}, joinModel{}, First{}, Second{}, FirstSeconds{})
			require.NoError(b, err)
		}
	})
}

func prepareBench(b *testing.B) *ModelStruct {
	b.Helper()
	type Model struct {
		ID     int
		Name   string
		First  string
		Second int
		Third  string
	}

	mm := testingModelMap(b)
	err := mm.RegisterModels(Model{})
	require.NoError(b, err)

	mStruct, err := mm.GetModelStruct(Model{})
	require.NoError(b, err)

	return mStruct
}
