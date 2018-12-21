package models

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestCheckAttributes(t *testing.T) {
	clearMap()
	c.PrecomputeModels(&User{}, &Pet{})

	mStruct := c.MustGetModelStruct(&User{})

	errs := mStruct.checkAttributes("name", "surname")
	assertNotEmpty(t, errs)
	// fmt.Println(errs)
}

func TestCheckFields(t *testing.T) {
	clearMap()
	c.PrecomputeModels(&User{}, &Pet{})

	mStruct := c.MustGetModelStruct(&User{})

	errs := mStruct.checkFields("name", "pets")
	assertEmpty(t, errs)

	errs = mStruct.checkFields("pepsi")
	assertNotEmpty(t, errs)
	// fmt.Println(errs)
}

func TestGettersModelStruct(t *testing.T) {
	clearMap()
	c.PrecomputeModels(&User{}, &Pet{})

	mStruct := c.MustGetModelStruct(&Pet{})

	// Type
	assertTrue(t, mStruct.GetType() == reflect.TypeOf(Pet{}))

	// CollectionType
	assertTrue(t, mStruct.GetCollectionType() == "pets")

	primary := mStruct.GetPrimaryField()
	assertTrue(t, primary.fieldName() == "ID")
	assertTrue(t, primary.fieldType == Primary)

	// AttrField
	nameField := mStruct.GetAttributeField("name")
	assertNotNil(t, nameField)
	assertTrue(t, nameField.fieldName() == "Name")

	// RelationshipField
	relationshipField := mStruct.GetRelationshipField("owners")
	assertNotNil(t, relationshipField)
	assertTrue(t, relationshipField.fieldName() == "Owners")

	// Fields
	fields := mStruct.GetFields()
	assertNotEmpty(t, fields)
}

func TestComputeNestedIncludeCount(t *testing.T) {
	c.PrecomputeModels(&User{}, &Pet{})
	mUser := c.MustGetModelStruct(&User{})

	assertTrue(t, mUser.getMaxIncludedCount() == 2)

	clearMap()
	c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	mBlog := c.MustGetModelStruct(&Blog{})

	assertTrue(t, mBlog.getMaxIncludedCount() == 6)
	assertTrue(t, c.MustGetModelStruct(&Post{}).getMaxIncludedCount() == 2)

	clearMap()
}

func TestInitCheckFieldTypes(t *testing.T) {
	type invalidPrimary struct {
		ID float64 `jsonapi:"type=primary"`
	}
	clearMap()
	err := c.buildModelStruct(&invalidPrimary{}, c.Models)
	assertNil(t, err)

	mStruct := c.MustGetModelStruct(&invalidPrimary{})
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	// attribute of type chan
	type invalidAttribute struct {
		ID       int           `jsonapi:"type=primary"`
		ChanAttr chan (string) `jsonapi:"type=attr"`
	}
	err = c.buildModelStruct(&invalidAttribute{}, c.Models)
	assertNil(t, err)

	mStruct = c.MustGetModelStruct(&invalidAttribute{})

	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	//attribute of type func
	type invAttrFunc struct {
		ID       int          `jsonapi:"type=primary"`
		FuncAttr func(string) `jsonapi:"type=attr"`
	}
	err = c.buildModelStruct(&invAttrFunc{}, c.Models)
	assertNil(t, err)

	mStruct = c.MustGetModelStruct(&invAttrFunc{})
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	// relationship of type not a struct/ ptr struct / slice
	type invalidRelBasic struct {
		ID    int    `jsonapi:"type=primary"`
		Basic string `jsonapi:"type=relation"`
	}
	inv := invalidRelBasic{}
	mStruct = &ModelStruct{
		primary: &StructField{refStruct: reflect.StructField{Type: reflect.TypeOf(inv.ID)}},
		fields: []*StructField{{
			fieldType: RelationshipSingle,
			refStruct: reflect.StructField{Type: reflect.TypeOf(inv.Basic)}}},
	}

	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	mStruct = &ModelStruct{
		primary: &StructField{refStruct: reflect.StructField{Type: reflect.TypeOf(inv.ID)}},
		fields: []*StructField{{
			fieldType: RelationshipSingle,
			refStruct: reflect.StructField{Type: reflect.TypeOf([]*string{})}}},
	}
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	strVal := "val"
	mStruct = &ModelStruct{
		primary: &StructField{refStruct: reflect.StructField{Type: reflect.TypeOf(inv.ID)}},
		fields: []*StructField{{
			fieldType: RelationshipSingle,
			refStruct: reflect.StructField{Type: reflect.TypeOf(&strVal)}}},
	}

	err = mStruct.initCheckFieldTypes()
	assertError(t, err)
}

func TestStructSetModelURL(t *testing.T) {
	clearMap()
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	mStruct := c.MustGetModelStruct(&Blog{})
	err = mStruct.SetModelURL("Some/url")
	assertError(t, err)
}

func TestGetFieldKind(t *testing.T) {
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	mStruct := c.MustGetModelStruct(&Blog{})
	assertEqual(t, Primary, mStruct.primary.GetFieldKind())

}

func TestAttrArray(t *testing.T) {

	type arr struct {
		model interface{}
		f     func(*testing.T, error)
	}

	type AttrArrStruct struct {
		ID  int       `jsonapi:"type=primary"`
		Arr []*string `jsonapi:"type=attr"`
	}

	type ArrayModel struct {
		ID  int       `jsonapi:"type=primary"`
		Arr [2]string `jsonapi:"type=attr"`
	}

	type SliceInt struct {
		ID int   `jsonapi:"type=primary"`
		Sl []int `jsonapi:"type=attr"`
	}

	type ArrInt struct {
		ID  int    `jsonapi:"type=primary"`
		Arr [2]int `jsonapi:"type=attr"`
	}

	type NestedStruct struct {
		Name string
	}

	type SliceStruct struct {
		ID int             `jsonapi:"type=primary"`
		Sl []*NestedStruct `jsonapi:"type=attr"`
	}

	type PtrSlicePtrStruct struct {
		ID int              `jsonapi:"type=primary"`
		Sl *[]*NestedStruct `jsonapi:"type=attr"`
	}

	type DoubleSlice struct {
		ID int     `jsonapi:"type=primary"`
		Sl [][]int `jsonapi:"type=attr"`
	}

	b := bytes.NewBuffer(nil)

	tests := map[string]arr{
		"StringPtr": {
			model: &AttrArrStruct{},
			f: func(t *testing.T, err error) {
				require.NoError(t, err)

				scope, err := c.NewScope(&AttrArrStruct{})
				require.NoError(t, err)

				scope.Value = &AttrArrStruct{ID: 1, Arr: make([]*string, 7)}
				p, err := MarshalScope(scope, c)
				assert.NoError(t, err)

				err = MarshalPayload(b, p)
				assertNil(t, err)

				c.log().Debug(b.String())
			},
		},
		"StringArray": {
			model: &ArrayModel{},
			f: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"IntSlice": {
			model: &SliceInt{},
			f: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"StructSlice": {
			model: &SliceStruct{},
			f: func(t *testing.T, err error) {
				require.NoError(t, err)
				mStruct, err := c.GetModelStruct(&SliceStruct{})
				if assert.NoError(t, err) {
					slField, ok := mStruct.Attribute("sl")
					if assert.True(t, ok) {
						assert.False(t, slField.isPtr())
						assert.True(t, slField.isSlice())
						assert.True(t, slField.isBasePtr())
						assert.True(t, slField.isNestedStruct())
					}
				}
			},
		},
		"IntArray": {
			model: &ArrInt{},
			f: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"PtrSlicePtrType": {
			model: &PtrSlicePtrStruct{},
			f: func(t *testing.T, err error) {
				require.NoError(t, err)
				var mStruct *ModelStruct
				mStruct, err = c.GetModelStruct(&PtrSlicePtrStruct{})
				if assert.NoError(t, err) {
					slField, ok := mStruct.Attribute("sl")
					if assert.True(t, ok) {
						assert.True(t, slField.isPtr())
						assert.True(t, slField.isSlice())
						assert.True(t, slField.isBasePtr())
						assert.True(t, slField.isNestedStruct())
					}
				}
			},
		},
		"DoubleSlice": {
			model: &DoubleSlice{},
			f: func(t *testing.T, err error) {
				require.NoError(t, err)
				mStruct, err := c.GetModelStruct(&DoubleSlice{})
				if assert.NoError(t, err) {
					slField, ok := mStruct.Attribute("sl")
					if assert.True(t, ok) {
						assert.False(t, slField.isPtr())
						assert.True(t, slField.isSlice())
						assert.False(t, slField.isBasePtr())
						assert.False(t, slField.isNestedStruct())
					}
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			clearMap()
			err := c.PrecomputeModels(test.model)
			test.f(t, err)
		})
	}

}
