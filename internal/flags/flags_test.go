package flags

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

var key1 = uint(3)

// key2 := uint(5)
// key3 := uint(6)

func initContainers(length int) []*Container {
	containers := []*Container{}
	for i := 0; i < length; i++ {
		containers = append(containers, New())
	}

	return containers
}

func TestGetFirsts(t *testing.T) {

	// The list is the priority in getting the first element
	// so that if the first container has the value Get first should take
	// the value from it, otherwise, it should continue

	tests := map[string]func(*testing.T){
		"TakeFirst": func(t *testing.T) {
			c := initContainers(3)

			c[0].Set(key1, false)
			c[1].Set(key1, true)

			// the value should be false - because the first container has it set already
			v, ok := GetFirst(key1, c...)
			assert.Equal(t, c[0].value(key1), v)
			assert.True(t, ok)
		},
		"TakeLast": func(t *testing.T) {
			c := initContainers(3)

			c[2].Set(key1, true)

			v, ok := GetFirst(key1, c...)
			assert.True(t, ok)
			assert.Equal(t, c[2].value(key1), v)
		},
		"NotFound": func(t *testing.T) {
			c := initContainers(3)

			_, ok := GetFirst(key1, c...)
			assert.False(t, ok)
		},
	}

	for name, f := range tests {
		t.Run(name, f)
	}
}

func TestBasics(t *testing.T) {
	c := New()

	_, ok := c.Get(key1)
	assert.False(t, ok)

	c.set(key1, true)

	v, ok := c.Get(key1)
	assert.True(t, ok)
	assert.True(t, v)

	assert.True(t, c.IsSet(key1))

	assert.False(t, c.IsSet(uint(25)))
	assert.False(t, c.value(uint(25)), fmt.Sprintf("%0.64b", uint64(*c)))

	nc := New()

	assert.True(t, nc.SetFrom(key1, c))
	assert.Equal(t, c.value(key1), nc.value(key1))
}

func BenchmarkGetFirst(b *testing.B) {
	for i := 10; i <= 100000; i *= 10 {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for j := 1; j < b.N; j++ {
				c := initContainers(i)

				c[i-1].Set(key1, true)

				b.StartTimer()
				v, ok := GetFirst(key1, c...)
				assert.True(b, ok)
				assert.True(b, v)
				b.StopTimer()
			}
		})

	}
}
