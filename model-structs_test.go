package jsonapi

import (
	"reflect"
	"testing"
)

func TestFieldByIndex(t *testing.T) {
	// having mapped model
	PrecomputeModels(&Timestamp{})

	mStruct := MustGetModelStruct(&Timestamp{})

	refType := reflect.TypeOf(Timestamp{})

	for i := 0; i < refType.NumField(); i++ {
		tField := refType.Field(i)
		sField, err := mStruct.FieldByIndex(i)
		if err != nil {
			t.Error(err)
		}

		if tField.Name != sField.refStruct.Name {
			t.Errorf("Fields names are not equal: %v != %v", tField.Name, sField.refStruct.Name)
		}
	}

	// getting out of range field
	_, err := mStruct.FieldByIndex(refType.NumField())
	if err == nil {
		t.Error("There should be no field")
	}

	clearMap()
	// User has one private field getting it from modelstruct by index should throw error
	PrecomputeModels(&User{}, &Pet{})

	mStruct = MustGetModelStruct(&User{})

	_, err = mStruct.FieldByIndex(0)
	if err == nil {
		t.Error("Index with private field should be nil. Thus an error should be returned")
	}
}

func TestCheckRelationships(t *testing.T) {
	clearMap()

	// let's have some modelstructs with relationships
	PrecomputeModels(&User{}, &Pet{})

	mStruct := MustGetModelStruct(&User{})

	errs, err := mStruct.checkRelationship("pets")
	assertNoError(t, err)
	assertEmpty(t, errs)

	errs, err = mStruct.checkRelationship("pets.owners")
	assertNilErrors(t, err)
	assertEmpty(t, errs)

	errs, err = mStruct.checkRelationship("pets.owners.imagined")
	assertNilErrors(t, err)
	assertNotEmpty(t, errs)
	// fmt.Println(errs)

	//Invalid relationship name

	errs, err = mStruct.checkRelationships("petse", "petsa.try")
	assertNilErrors(t, err)
	assertNotEmpty(t, errs)

	// if somehow the relationship is not mapped
	clearMap()
	buildModelStruct(&User{}, cacheModelMap)
	mStruct = MustGetModelStruct(&User{})

	_, err = mStruct.checkRelationships("pets.dog")
	assertError(t, err)
	clearMap()
}

func TestCheckAttributes(t *testing.T) {
	clearMap()
	PrecomputeModels(&User{}, &Pet{})

	mStruct := MustGetModelStruct(&User{})

	errs := mStruct.checkAttributes("name", "surname")
	assertNotEmpty(t, errs)
	// fmt.Println(errs)
}

func TestCheckFields(t *testing.T) {
	clearMap()
	PrecomputeModels(&User{}, &Pet{})

	mStruct := MustGetModelStruct(&User{})

	errs := mStruct.checkFields("name", "pets")
	assertEmpty(t, errs)

	errs = mStruct.checkFields("pepsi")
	assertNotEmpty(t, errs)
	// fmt.Println(errs)
}

func TestGettersModelStruct(t *testing.T) {
	clearMap()
	PrecomputeModels(&User{}, &Pet{})

	mStruct := MustGetModelStruct(&Pet{})

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

	v := reflect.ValueOf(&Pet{}).Elem()
	refType := v.Type()

	for i := 0; i < refType.NumField(); i++ {
		assertTrue(t, refType.Field(i).Name == fields[i].fieldName)
		assertTrue(t, refType.Field(i).Type == fields[i].refStruct.Type)
	}
}

func TestStructFieldGetters(t *testing.T) {
	//having some model struct.
	clearMap()

	PrecomputeModels(&User{}, &Pet{})
	mStruct := MustGetModelStruct(&User{})

	assertNotNil(t, mStruct)

	primary := mStruct.primary

	assertNotNil(t, primary)

	// check index
	v := reflect.ValueOf(&User{}).Elem()
	refType := v.Type()

	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		if primary.GetFieldName() == field.Name {
			assertTrue(t, primary.GetFieldIndex() == field.Index[0])
			assertTrue(t, primary.GetFieldType() == field.Type)
		}
	}

	// related for primary should be zero
	assertNil(t, primary.GetRelatedModelType())
}

func TestGetSortScopeCount(t *testing.T) {
	clearMap()
	PrecomputeModels(&User{}, &Pet{})

	mStruct := MustGetModelStruct(&User{})

	assertNotNil(t, mStruct)
}

func TestComputeNestedIncludeCount(t *testing.T) {
	PrecomputeModels(&User{}, &Pet{})
	mUser := MustGetModelStruct(&User{})

	assertTrue(t, mUser.getMaxIncludedCount() == 2)

	clearMap()
	PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	mBlog := MustGetModelStruct(&Blog{})

	assertTrue(t, mBlog.getMaxIncludedCount() == 6)
	assertTrue(t, MustGetModelStruct(&Post{}).getMaxIncludedCount() == 2)

	clearMap()
	PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	mBlog = MustGetModelStruct(&Blog{})
	// set some model struct to nil
	cacheModelMap.Set(reflect.TypeOf(Post{}), nil)

	assertPanic(t, func() { mBlog.initComputeNestedIncludedCount(0) })
}

func TestInitCheckFieldTypes(t *testing.T) {
	type invalidPrimary struct {
		ID float64 `jsonapi:"primary,invalids"`
	}
	clearMap()
	err := buildModelStruct(&invalidPrimary{}, cacheModelMap)
	assertNil(t, err)

	mStruct := MustGetModelStruct(&invalidPrimary{})
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	// attribute of type chan
	type invalidAttribute struct {
		ID       int           `jsonapi:"primary,invalidAttr"`
		ChanAttr chan (string) `jsonapi:"attr,channel"`
	}
	err = buildModelStruct(&invalidAttribute{}, cacheModelMap)
	assertNil(t, err)

	mStruct = MustGetModelStruct(&invalidAttribute{})

	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	//attribute of type func
	type invAttrFunc struct {
		ID       int          `jsonapi:"primary,invAttrFunc"`
		FuncAttr func(string) `jsonapi:"attr,func-attr"`
	}
	err = buildModelStruct(&invAttrFunc{}, cacheModelMap)
	assertNil(t, err)

	mStruct = MustGetModelStruct(&invAttrFunc{})
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	// relationship of type not a struct/ ptr struct / slice
	type invalidRelBasic struct {
		ID    int    `jsonapi:"primary,invRelBasic"`
		Basic string `jsonapi:"relation,basic"`
	}
	inv := invalidRelBasic{}
	mStruct = &ModelStruct{fields: []*StructField{{
		jsonAPIType: RelationshipSingle,
		refStruct:   reflect.StructField{Type: reflect.TypeOf(inv.Basic)}}},
	}
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	mStruct = &ModelStruct{fields: []*StructField{{
		jsonAPIType: RelationshipSingle,
		refStruct:   reflect.StructField{Type: reflect.TypeOf([]*string{})}}},
	}
	err = mStruct.initCheckFieldTypes()
	assertError(t, err)

	strVal := "val"
	mStruct = &ModelStruct{fields: []*StructField{{
		jsonAPIType: RelationshipSingle,
		refStruct:   reflect.StructField{Type: reflect.TypeOf(&strVal)}}},
	}

	err = mStruct.initCheckFieldTypes()
	assertError(t, err)
}

func TestStructSetModelURL(t *testing.T) {
	clearMap()
	err := PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	mStruct := MustGetModelStruct(&Blog{})
	err = mStruct.SetModelURL("Some/url")
	assertError(t, err)
}

func TestGetDereferencedType(t *testing.T) {
	type someType struct {
		Field *string
	}
	ty := reflect.TypeOf(someType{})
	sField := &StructField{refStruct: ty.Field(0)}
	nty := sField.getDereferencedType()
	assertEqual(t, reflect.TypeOf(""), nty)
}

func TestGetJSONAPIType(t *testing.T) {
	err := PrecomputeModels(&Blog{}, &Post{}, &Comment{})
	assertNil(t, err)

	mStruct := MustGetModelStruct(&Blog{})
	assertEqual(t, Primary, mStruct.primary.GetJSONAPIType())

}
