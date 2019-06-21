package class

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClass tests the error classification system.
func TestClass(t *testing.T) {
	majors.reset()
	defer registerClasses()
	defer majors.reset()

	t.Run("RegisterMajor", func(t *testing.T) {
		defer majors.reset()

		m, err := RegisterMajor("TestingMajor")
		require.NoError(t, err)

		assert.True(t, m.InBounds())
		assert.Equal(t, uint8(1), uint8(m))

		m, err = RegisterMajor("TestingMajor2")
		require.NoError(t, err)

		assert.True(t, m.InBounds())
		assert.Equal(t, uint8(2), uint8(m))
	})

	t.Run("DuplicatedMajor", func(t *testing.T) {
		defer majors.reset()

		_, err := RegisterMajor("TestingMajor")
		require.NoError(t, err)

		_, err = RegisterMajor("TestingMajor")
		require.Error(t, err)
	})

	t.Run("RegisterMinor", func(t *testing.T) {
		defer majors.reset()

		m, err := RegisterMajor("TestingMajor")
		require.NoError(t, err)

		minorName := "TestingMinor"

		minor, err := m.RegisterMinor(minorName)
		require.NoError(t, err)

		assert.Equal(t, minor.Name(), minorName)

		_, err = m.RegisterMinor(minorName)
		require.Error(t, err)

		t.Run("RegisterIndex", func(t *testing.T) {
			indexName := "TestingIndex"

			index, err := minor.RegisterIndex(indexName)
			require.NoError(t, err)

			assert.True(t, index.valid(), "%d", index.value)

			_, err = minor.RegisterIndex(indexName)
			require.Error(t, err)

			assert.Equal(t, indexName, index.Name())
		})
	})

	t.Run("NewClass", func(t *testing.T) {
		defer majors.reset()

		major, err := RegisterMajor("TestingMajor")
		require.NoError(t, err)

		minor, err := major.RegisterMinor("TestingMinor")
		require.NoError(t, err)

		index, err := minor.RegisterIndex("TestingIndex")
		require.NoError(t, err)

		class, err := NewClass(index)
		require.NoError(t, err)

		assert.Equal(t, Class(1<<(32-majorBitSize)|1<<(32-minorBitSize-majorBitSize)|1), class)
		assert.Equal(t, "00000010000000001000000000000001", fmt.Sprintf("%032b", class))

		assert.Equal(t, index.Value(), class.Index().Value())
		assert.Equal(t, minor.Value(), class.Minor().Value())
		assert.Equal(t, major, class.Major())
	})

	t.Run("String", func(t *testing.T) {
		defer majors.reset()

		major, err := RegisterMajor("Major")
		require.NoError(t, err)

		minor, err := major.RegisterMinor("Minor")
		require.NoError(t, err)

		t.Logf("Minor: %v", minor.Value())

		index, err := minor.RegisterIndex("Index")
		require.NoError(t, err)

		class, err := NewClass(index)
		require.NoError(t, err)

		name := class.String()
		assert.Equal(t, "MajorMinorIndex", name)
	})
}
