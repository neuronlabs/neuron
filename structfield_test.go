package jsonapi

import (
	"reflect"
	"testing"
)

func TestStructFieldGetters(t *testing.T) {
	//having some model struct.
	clearMap()

	c.PrecomputeModels(&User{}, &Pet{})
	mStruct := c.MustGetModelStruct(&User{})

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

func TestGetDereferencedType(t *testing.T) {
	type someType struct {
		Field *string
	}
	ty := reflect.TypeOf(someType{})
	sField := &StructField{refStruct: ty.Field(0)}
	nty := sField.getDereferencedType()
	assertEqual(t, reflect.TypeOf(""), nty)
}
