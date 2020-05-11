package mapping

import (
	"reflect"
)

// NewValueSingle creates and returns new value for the given model type.
func NewValueSingle(m *ModelStruct) interface{} {
	return m.newReflectValueSingle().Interface()
}

// NewValueMany creates and returns a model's new slice of pointers to values.
func NewValueMany(m *ModelStruct) interface{} {
	return NewReflectValueMany(m).Interface()
}

// NewReflectValueSingle creates and returns a model's new single value.
func NewReflectValueSingle(m *ModelStruct) reflect.Value {
	return m.newReflectValueSingle()
}

// NewReflectValueMany creates the *[]*m.Type reflect.Models.
func NewReflectValueMany(m *ModelStruct) reflect.Value {
	return reflect.New(reflect.SliceOf(reflect.PtrTo(m.Type())))
}
