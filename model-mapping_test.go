package jsonapi

import (
	"reflect"
	"testing"
)

func TestPrecomputeModels(t *testing.T) {
	// input valid models
	validModels := []interface{}{&User{}, &Pet{}, &Driver{}, &Car{}, &WithPointer{}, &Blog{}, &Post{}, &Comment{}}
	err := PrecomputeModels(validModels...)
	if err != nil {
		t.Error(err)
	}
	clearMap()

	// if somehow map is nil
	cacheModelMap = nil
	err = PrecomputeModels(&Timestamp{})
	if err != nil {
		t.Error(err)
	}

	// if one of the relationship is not precomputed
	clearMap()
	// User has relationship with Pet
	err = PrecomputeModels(&User{})
	if err == nil {
		t.Error("The User is related to Pets and so that should be an error")
	}
	clearMap()

	// if one of the models is invalid
	err = PrecomputeModels(&Timestamp{}, &BadModel{})
	if err == nil {
		t.Error("BadModel should not be accepted in precomputation.")
	}

	// provided Struct type to precompute models
	err = PrecomputeModels(Timestamp{})
	if err == nil {
		t.Error("A pointer to the model should be provided.")
	}

	// provided ptr to basic type
	basic := "value"
	err = PrecomputeModels(&basic)
	if err == nil {
		t.Error("Only structs should be accepted!")
	}

	// provided slice
	err = PrecomputeModels(&[]*Timestamp{})
	if err == nil {
		t.Error("Slice should not be accepted in precomputedModels")
	}

	// if no tagged fields are provided an error would be thrown
	err = PrecomputeModels(&ModelNonTagged{})
	if err == nil {
		t.Error("Non tagged models are not allowed.")
	}
	clearMap()

	// models without primary are not allowed.
	err = PrecomputeModels(&NoPrimaryModel{})
	if err == nil {
		t.Error("No primary field provided.")
	}
	clearMap()
}

func TestGetModelStruct(t *testing.T) {
	// MustGetModelStruct
	// if the model is not in the cache map
	clearMap()
	assertPanic(t, func() {
		MustGetModelStruct(Timestamp{})
	})

	cacheModelMap.Set(reflect.TypeOf(Timestamp{}), &ModelStruct{})
	mStruct := MustGetModelStruct(Timestamp{})
	if mStruct == nil {
		t.Error("The model struct shoud not be nil.")
	}

	// GetModelStruct
	// providing ptr should return mStruct
	var err error
	_, err = GetModelStruct(&Timestamp{})
	if err != nil {
		t.Error(err)
	}

	// nil model
	_, err = GetModelStruct(nil)
	if err == nil {
		t.Error(err)
	}
}

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

func TestGetRelatedType(t *testing.T) {
	// Providing type struct
	var refType, relType reflect.Type
	var err error
	refType = reflect.TypeOf(Car{})
	relType, err = getRelatedType(refType)
	if err != nil {
		t.Error(err)
	}

	if refType != relType {
		t.Error("These should be the same type")
	}

	// Ptr type
	refType = reflect.TypeOf(&Car{})
	relType, err = getRelatedType(refType)
	if err != nil {
		t.Error(err)
	}

	if relType != refType.Elem() {
		t.Error("Pointer type should be the same")
	}

	//Slice of ptr
	refType = reflect.TypeOf([]*Car{})
	relType, err = getRelatedType(relType)
	if err != nil {
		t.Error(err)
	}
	if relType != refType.Elem().Elem() {
		t.Error("Error in slice of ptr")
	}

	// Slice
	refType = reflect.TypeOf([]Car{})
	relType, err = getRelatedType(relType)
	if err != nil {
		t.Error(err)
	}

	if relType != refType.Elem() {
		t.Error("Error in slice")
	}

	// basic type
	refType = reflect.TypeOf("string")
	_, err = getRelatedType(refType)
	if err == nil {
		t.Error("String should throw error")
	}

	// ptr to basic type
	var b string = "String"
	refType = reflect.TypeOf(&b)
	_, err = getRelatedType(refType)
	if err == nil {
		t.Error("Ptr to string should throw error")
	}
}

func clearMap() {
	cacheModelMap = newModelMap()
}
