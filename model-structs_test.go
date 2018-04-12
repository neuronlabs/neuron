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
	if err != nil {
		t.Error(err)
	}
	if len(errs) != 0 {
		t.Error(errs)
	}

}

func TestCheckAttributes(t *testing.T) {

}

func TestCheckFields(t *testing.T) {

}
