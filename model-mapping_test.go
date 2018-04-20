package jsonapi

import (
	"reflect"
	"testing"
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
		t.Error("Invalid elem type: %v", elemType)
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

// func TestSetModelURL(t *testing.T) {
// 	assertNil(t, PrecomputeModels(&Blog{}, &Post{}, &Comment{}))

// 	assertError(t, SetModelURL(&Blog{}, "/invalid/url"))
// 	assertError(t, SetModelURL(&User{}, "/doesnt/matter/"))

// }

func clearMap() {
	c.Models = newModelMap()
}
