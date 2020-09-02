package mapping

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNestedFields tests the nested field's definitions
func TestNestedFields(t *testing.T) {
	ms := testingModelMap(t)

	err := ms.RegisterModels(&ModelWithNested{})
	require.NoError(t, err)

	m, err := ms.ModelStruct(&ModelWithNested{})
	require.NoError(t, err)

	t.Run("ptr-composed", func(t *testing.T) {
		ptrField, ok := m.Attribute("ptr-composed")
		require.True(t, ok)

		assert.True(t, ptrField.isPtr())

		// ptr composed field must have a nested struct field
		require.NotNil(t, ptrField.nested)

		// ptrField in fact is not nested field. It contains nested subfields.
		// and it is a NestedStruct
		assert.False(t, ptrField.isNestedField())

		nested := ptrField.nested

		assert.Equal(t, nested.modelType, reflect.TypeOf(NestedAttribute{}))

		assert.Len(t, nested.fields, 7)

		// float field
		nestedField, ok := nested.fields["float"]
		if assert.True(t, ok) {
			assert.Equal(t, nestedField.structField.Kind(), KindNested)
			assert.Equal(t, reflect.Float64, nestedField.structField.reflectField.Type.Kind())
		}

		// nested int field
		nestedField, ok = nested.fields["int"]
		if assert.True(t, ok) {
			assert.Equal(t, nestedField.structField.Kind(), KindNested)
			assert.Equal(t, reflect.Int, nestedField.structField.reflectField.Type.Kind())
		}

		// int
		nestedField, ok = nested.fields["string"]
		if assert.True(t, ok) {
			assert.Equal(t, nestedField.structField.Kind(), KindNested)
			assert.Equal(t, reflect.String, nestedField.structField.reflectField.Type.Kind())
		}

		// slice
		nestedField, ok = nested.fields["slice"]
		if assert.True(t, ok) {
			assert.Equal(t, nestedField.structField.Kind(), KindNested)
			assert.Equal(t, reflect.Slice, nestedField.structField.reflectField.Type.Kind())
			assert.True(t, nestedField.structField.isSlice())
		}

		// inception
		t.Run("NestedInNested", func(t *testing.T) {
			nestedField, ok = nested.fields["inception"]
			if assert.True(t, ok) {
				nestedInNested := nestedField.structField.nested
				require.NotNil(t, nestedInNested)

				nStructFielder := nestedInNested.structField
				if assert.NotNil(t, nStructFielder) {
					assert.Equal(t, nestedField.structField, nStructFielder.Self())

					nestedFieldInterface, ok := nStructFielder.(nestedStructFielder)
					if assert.True(t, ok) {
						assert.Equal(t, nestedField, nestedFieldInterface.SelfNested())
					}
				}

				assert.Len(t, nestedInNested.fields, 2)

				subNestedField, ok := nestedInNested.fields["inception_first"]
				if assert.True(t, ok) {
					assert.Equal(t, reflect.Int, subNestedField.structField.reflectField.Type.Kind())
					assert.Equal(t, nestedInNested, subNestedField.root)
					assert.Equal(t, nestedField.structField.Kind(), KindNested)
				}

				subNestedField, ok = nestedInNested.fields["second"]
				if assert.True(t, ok) {
					assert.Equal(t, reflect.Float64, subNestedField.structField.reflectField.Type.Kind())
					assert.Equal(t, nestedField.structField.Kind(), KindNested)
				}
			}
		})

		t.Run("Tagged", func(t *testing.T) {
			nestedField, ok = nested.fields["float-tag"]
			if assert.True(t, ok) {
				assert.Equal(t, nestedField.structField.Kind(), KindNested)
				assert.Equal(t, reflect.Float64, nestedField.structField.reflectField.Type.Kind())
			}

			nestedField, ok = nested.fields["int-tag"]
			if assert.True(t, ok) {
				assert.Equal(t, nestedField.structField.Kind(), KindNested)
				assert.Equal(t, reflect.Int, nestedField.structField.reflectField.Type.Kind())
			}
		})
	})
}
