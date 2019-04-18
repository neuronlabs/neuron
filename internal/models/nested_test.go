package models

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "reflect"
	"testing"
)

type subNested struct {
	InceptionFirst  int
	InceptionSecond float64 `neuron:"name=second;flags=omitempty"`
}

type nestedAttribute struct {
	Float     float64
	Int       int
	String    string
	Slice     []int
	Inception subNested

	FloatTagged float64 `neuron:"name=float-tag"`
	IntTagged   int     `neuron:"name=int-tag"`
}

type modelWithNested struct {
	ID          int              `neuron:"type=primary"`
	PtrComposed *nestedAttribute `neuron:"type=attr;name=ptr-composed"`
}

func TestNestedFields(t *testing.T) {

	checkLogger()

	ms := testingSchemas(t)

	err := ms.RegisterModels(modelWithNested{})
	require.NoError(t, err)

	m, err := ms.GetModelStruct(modelWithNested{})
	require.NoError(t, err)
	// if assert.NoError(t, c.PrecomputeModels(&ModelWithNested{})) {
	// m := s .Get(reflect.TypeOf(ModelWithNested{}))
	if assert.NotNil(t, m) {

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

			assert.Equal(t, nested.modelType, reflect.TypeOf(nestedAttribute{}))

			var ctr int
			for nestedName, nestedField := range nested.fields {

				assert.Equal(t, ptrField, nestedField.attr())
				switch nestedName {
				case "float":
					t.Run("nestedFloat", func(t *testing.T) {
						assert.Equal(t, nestedField.FieldKind(), KindNested)
						assert.Equal(t, reflect.Float64, nestedField.reflectField.Type.Kind())
					})
					ctr += 1
				case "int":
					t.Run("nestedInt", func(t *testing.T) {
						assert.Equal(t, nestedField.FieldKind(), KindNested)
						assert.Equal(t, reflect.Int, nestedField.reflectField.Type.Kind())
					})
					ctr += 1
				case "string":
					t.Run("nestedString", func(t *testing.T) {
						assert.Equal(t, nestedField.FieldKind(), KindNested)
						assert.Equal(t, reflect.String, nestedField.reflectField.Type.Kind())
					})
					ctr += 1
				case "slice":
					t.Run("nestedInt", func(t *testing.T) {
						assert.Equal(t, nestedField.FieldKind(), KindNested)
						assert.Equal(t, reflect.Slice, nestedField.reflectField.Type.Kind())
						assert.True(t, nestedField.isSlice())
					})
					ctr += 1
				case "inception":
					t.Run("nestedInNested", func(t *testing.T) {
						nested := nestedField.nested
						require.NotNil(t, nested)

						nStructFielder := nested.structField
						if assert.NotNil(t, nStructFielder) {
							assert.Equal(t, nestedField.StructField, nStructFielder.Self())

							nestedFieldInterface, ok := nStructFielder.(NestedStructFielder)
							if assert.True(t, ok) {
								assert.Equal(t, nestedField, nestedFieldInterface.SelfNested())
							}
						}

						assert.Len(t, nested.fields, 2)
						var nctr int
						for nnName, nnField := range nested.fields {
							switch nnName {
							case "inception_first":
								t.Run("non-tagged", func(t *testing.T) {
									assert.Equal(t, reflect.Int, nnField.reflectField.Type.Kind())
									assert.Equal(t, nested, nnField.root)
									assert.Equal(t, nestedField.FieldKind(), KindNested)
								})
								nctr += 1
							case "second":
								nctr += 1
								t.Run("tagged", func(t *testing.T) {
									assert.Equal(t, reflect.Float64, nnField.reflectField.Type.Kind())
									assert.Equal(t, nestedField.FieldKind(), KindNested)
									assert.True(t, nnField.isOmitEmpty())

									assert.Equal(t, nested, nnField.root)
								})
							default:
								t.Logf("BadName for nested Field: %s", nnName)
							}
						}
						assert.Equal(t, len(nested.fields), nctr)
					})
					ctr += 1
				case "float-tag":
					t.Run("nestedFloatTagged", func(t *testing.T) {
						assert.Equal(t, nestedField.FieldKind(), KindNested)
						assert.Equal(t, reflect.Float64, nestedField.reflectField.Type.Kind())
					})
					ctr += 1
				case "int-tag":
					t.Run("nestedIntTagged", func(t *testing.T) {
						assert.Equal(t, nestedField.FieldKind(), KindNested)
						assert.Equal(t, reflect.Int, nestedField.reflectField.Type.Kind())
					})
					ctr += 1
				default:
					t.Fail()

				}
			}
			assert.Equal(t, len(nested.fields), ctr)
		})
	}

}
