package jsonapi

import (
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
	assertTrue(t, primary.fieldName == "ID")
	assertTrue(t, primary.jsonAPIType == Primary)

	// AttrField
	nameField := mStruct.GetAttributeField("name")
	assertNotNil(t, nameField)
	assertTrue(t, nameField.fieldName == "Name")

	// RelationshipField
	relationshipField := mStruct.GetRelationshipField("owners")
	assertNotNil(t, relationshipField)
	assertTrue(t, relationshipField.fieldName == "Owners")

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
		ID float64 `jsonapi:"primary,invalids"`
	}
	clearMap()
	err := buildModelStruct(&invalidPrimary{}, c.Models)
	assertNil(t, err)

	mStruct := c.MustGetModelStruct(&invalidPrimary{})
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	// attribute of type chan
	type invalidAttribute struct {
		ID       int           `jsonapi:"primary,invalidAttr"`
		ChanAttr chan (string) `jsonapi:"attr,channel"`
	}
	err = buildModelStruct(&invalidAttribute{}, c.Models)
	assertNil(t, err)

	mStruct = c.MustGetModelStruct(&invalidAttribute{})

	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	//attribute of type func
	type invAttrFunc struct {
		ID       int          `jsonapi:"primary,invAttrFunc"`
		FuncAttr func(string) `jsonapi:"attr,func-attr"`
	}
	err = buildModelStruct(&invAttrFunc{}, c.Models)
	assertNil(t, err)

	mStruct = c.MustGetModelStruct(&invAttrFunc{})
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	// relationship of type not a struct/ ptr struct / slice
	type invalidRelBasic struct {
		ID    int    `jsonapi:"primary,invRelBasic"`
		Basic string `jsonapi:"relation,basic"`
	}
	inv := invalidRelBasic{}
	mStruct = &ModelStruct{
		primary: &StructField{refStruct: reflect.StructField{Type: reflect.TypeOf(inv.ID)}},
		fields: []*StructField{{
			jsonAPIType: RelationshipSingle,
			refStruct:   reflect.StructField{Type: reflect.TypeOf(inv.Basic)}}},
	}

	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	mStruct = &ModelStruct{
		primary: &StructField{refStruct: reflect.StructField{Type: reflect.TypeOf(inv.ID)}},
		fields: []*StructField{{
			jsonAPIType: RelationshipSingle,
			refStruct:   reflect.StructField{Type: reflect.TypeOf([]*string{})}}},
	}
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	strVal := "val"
	mStruct = &ModelStruct{
		primary: &StructField{refStruct: reflect.StructField{Type: reflect.TypeOf(inv.ID)}},
		fields: []*StructField{{
			jsonAPIType: RelationshipSingle,
			refStruct:   reflect.StructField{Type: reflect.TypeOf(&strVal)}}},
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

func TestGetJSONAPIType(t *testing.T) {
	err := c.PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	mStruct := c.MustGetModelStruct(&Blog{})
	assertEqual(t, Primary, mStruct.primary.GetJSONAPIType())

}
