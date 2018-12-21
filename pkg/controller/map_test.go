package controller

import (
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestErrorBuildingFunctions(t *testing.T) {
	var err error
	err = errNoRelationship("Collection", "Included name")
	if err == nil {
		t.Error("An error should be returned")
	}
	// fmt.Println(err)

	err = errNoRelationshipInModel(reflect.TypeOf(Driver{}), reflect.TypeOf(Car{}), "drivers")
	if err == nil {
		t.Error("an error should be returned")
	}

	// fmt.Println(err)

	err = errNoModelMappedForRel(reflect.TypeOf(Car{}), reflect.TypeOf(Driver{}), "owner")
	if err == nil {
		t.Error("An error should be returned")
	}
	// fmt.Println(err)
}

func TestGetSliceElemType(t *testing.T) {
	// having slice of ptr
	var refType, elemType reflect.Type
	var err error

	refType = reflect.TypeOf([]*Car{})
	elemType, err = getSliceElemType(refType)
	if err != nil {
		t.Error(err)
	}
	if elemType != refType.Elem().Elem() {
		t.Errorf("Invalid elem type: %v", elemType)
	}

	refType = reflect.TypeOf(&[]*Car{})
	elemType, err = getSliceElemType(refType)
	if err != nil {
		t.Error(err)
	}

	if elemType != refType.Elem().Elem().Elem() {
		t.Errorf("Invalid elem type: %v", elemType)
	}

	refType = reflect.TypeOf(Car{})
	_, err = getSliceElemType(refType)
	if err == nil {
		t.Error("Provided struct not a slice.")
	}
}

func TestSetRelatedType(t *testing.T) {
	// Providing type struct
	var err error
	var sField *StructField = &StructField{}

	sField.refStruct.Type = reflect.TypeOf(Car{})
	err = setRelatedType(sField)
	if err != nil {
		t.Error(err)
	}

	if sField.relatedModelType != sField.refStruct.Type {
		t.Error("These should be the same type")
	}

	// Ptr type
	sField.refStruct.Type = reflect.TypeOf(&Car{})
	err = setRelatedType(sField)
	if err != nil {
		t.Error(err)
	}

	if sField.relatedModelType != sField.refStruct.Type.Elem() {
		t.Error("Pointer type should be the same")
	}

	//Slice of ptr
	sField.refStruct.Type = reflect.TypeOf([]*Car{})
	err = setRelatedType(sField)
	if err != nil {
		t.Error(err)
	}
	if sField.relatedModelType != sField.refStruct.Type.Elem().Elem() {
		t.Error("Error in slice of ptr")
	}

	// Slice
	sField.refStruct.Type = reflect.TypeOf([]Car{})
	err = setRelatedType(sField)
	if err != nil {
		t.Error(err)
	}

	if sField.relatedModelType != sField.refStruct.Type.Elem() {
		t.Error("Error in slice")
	}

	// basic type
	sField.refStruct.Type = reflect.TypeOf("string")
	err = setRelatedType(sField)
	if err == nil {
		t.Error("String should throw error")
	}

	// ptr to basic type
	var b string = "String"
	sField.refStruct.Type = reflect.TypeOf(&b)
	err = setRelatedType(sField)
	if err == nil {
		t.Error("Ptr to string should throw error")
	}
}

func TestI18nModels(t *testing.T) {
	model := &Modeli18n{}
	clearMap()
	err := c.buildModelStruct(model, c.Models)
	assertNil(t, err)

	mStruct := c.Models.Get(reflect.TypeOf(Modeli18n{}))
	assertNotNil(t, mStruct)
	assertNotNil(t, mStruct.language)

	assertEqual(t, 1, len(mStruct.i18n))
	assertTrue(t, mStruct.i18n[0].I18n())
}

func TestMappingModel(t *testing.T) {
	t.Run("WithMaps", func(t *testing.T) {

		type NestedModel struct {
			SomeField   string
			TaggedField int `jsonapi:"name=tagged"`
		}

		type MapInt struct {
			ID  int            `jsonapi:"type=primary"`
			Map map[string]int `jsonapi:"type=attr"`
		}

		type MapPtrInt struct {
			ID  int             `jsonapi:"type=primary"`
			Map map[string]*int `jsonapi:"type=attr"`
		}

		type MapSliceInt struct {
			ID  int              `jsonapi:"type=primary"`
			Map map[string][]int `jsonapi:"type=attr"`
		}

		type MapSlicePtrInt struct {
			ID  int               `jsonapi:"type=primary"`
			Map map[string][]*int `jsonapi:"type=attr"`
		}

		type MapPtrSliceInt struct {
			ID  int            `jsonapi:"type=primary"`
			Map map[string]int `jsonapi:"type=attr"`
		}

		type MapNested struct {
			ID  int                    `jsonapi:"type=primary"`
			Map map[string]NestedModel `jsonapi:"type=attr"`
		}

		type MapPtrNested struct {
			ID  int                     `jsonapi:"type=primary"`
			Map map[string]*NestedModel `jsonapi:"type=attr"`
		}

		type MapSliceNested struct {
			ID  int                      `jsonapi:"type=primary"`
			Map map[string][]NestedModel `jsonapi:"type=attr"`
		}

		type MapPtrSliceNested struct {
			ID  int                       `jsonapi:"type=primary"`
			Map map[string]*[]NestedModel `jsonapi:"type=attr"`
		}

		type MapSlicePtrNested struct {
			ID  int                       `jsonapi:"type=primary"`
			Map map[string][]*NestedModel `jsonapi:"type=attr"`
		}

		type MapTime struct {
			ID  int                  `jsonapi:"type=primary"`
			Map map[string]time.Time `jsonapi:"type=attr"`
		}

		type MapPtrTime struct {
			ID  int                   `jsonapi:"type=primary"`
			Map map[string]*time.Time `jsonapi:"type=attr"`
		}

		type MapSliceTime struct {
			ID  int                    `jsonapi:"type=primary"`
			Map map[string][]time.Time `jsonapi:"type=attr"`
		}

		type MapSlicePtrTime struct {
			ID  int                     `jsonapi:"type=primary"`
			Map map[string][]*time.Time `jsonapi:"type=attr"`
		}

		type maptest struct {
			model interface{}
			f     func(t *testing.T, model *ModelStruct, field *StructField, err error)
		}

		// strCreator := func(tp, attrKey, attrValue string) string {
		// 	return fmt.Sprintf(`{"data":{"type":"%s","attributes":{"%s":{"%s":%s}}}}`)
		// }

		tests := map[string]maptest{
			"Int": {
				model: &MapInt{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.False(t, field.isBasePtr())
					assert.False(t, field.isNestedStruct())
				},
			},
			"PtrInt": {
				model: &MapPtrInt{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.True(t, field.isBasePtr())
					assert.False(t, field.isNestedStruct())
				},
			},
			"SliceInt": {
				model: &MapSliceInt{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.False(t, field.isBasePtr())
					assert.False(t, field.isArray())
					assert.False(t, field.isSlice())
					assert.False(t, field.isNestedStruct())
				},
			},
			"SlicePtrInt": {
				model: &MapSlicePtrInt{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.True(t, field.isBasePtr())
					assert.False(t, field.isArray())
					assert.False(t, field.isSlice())
					assert.False(t, field.isNestedStruct())
				},
			},
			"PtrSliceInt": {
				model: &MapPtrSliceInt{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.False(t, field.isBasePtr())
					assert.False(t, field.isArray())
					assert.False(t, field.isSlice())
					assert.False(t, field.isNestedStruct())
				},
			},
			"Nested": {
				model: &MapNested{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.True(t, field.isNestedStruct())
				},
			},
			"PtrNested": {
				model: &MapPtrNested{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.True(t, field.isBasePtr())
					assert.True(t, field.isNestedStruct())
				},
			},
			"SliceNested": {
				model: &MapSliceNested{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.False(t, field.isBasePtr())
					assert.True(t, field.isNestedStruct())
				},
			},

			"PtrSliceNested": {
				model: &MapPtrSliceNested{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.False(t, field.isBasePtr())
					assert.True(t, field.isNestedStruct())
				},
			},

			"SlicePtrNested": {
				model: &MapSlicePtrNested{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.True(t, field.isBasePtr())
					assert.True(t, field.isNestedStruct())
				},
			},

			"Time": {
				model: &MapTime{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.False(t, field.isBasePtr())
					assert.False(t, field.isNestedStruct())
					assert.True(t, field.isTime())
				},
			},
			"PtrTime": {
				model: &MapPtrTime{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.True(t, field.isBasePtr())
					assert.False(t, field.isNestedStruct())
					assert.True(t, field.isTime())
				},
			},
			"SliceTime": {
				model: &MapSliceTime{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.False(t, field.isBasePtr())
					assert.False(t, field.isNestedStruct())
					assert.True(t, field.isTime())
				},
			},
			"SlicePtrTime": {
				model: &MapSlicePtrTime{},
				f: func(t *testing.T, model *ModelStruct, field *StructField, err error) {
					require.NoError(t, err)
					assert.True(t, field.isMap())
					assert.True(t, field.isBasePtr())
					assert.False(t, field.isNestedStruct())
					assert.True(t, field.isTime())
				},
			},
		}

		for name, test := range tests {
			t.Run("Base"+name, func(t *testing.T) {
				clearMap()
				require.NoError(t, c.PrecomputeModels(test.model))
				model, err := c.GetModelStruct(test.model)
				require.NoError(t, err)

				field, ok := model.Attribute("map")
				require.True(t, ok)
				test.f(t, model, field, err)
			})
		}

	})
}

func clearMap() {
	c.Models = newModelMap()
	if *debugFlag {
		basic := c.log().(*unilogger.BasicLogger)
		basic.SetLevel(unilogger.DEBUG)
		c.logger = basic
	}
}
