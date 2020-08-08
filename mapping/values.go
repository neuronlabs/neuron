package mapping

import (
	"math"
	"reflect"
)

const (
	// MaxUint defines maximum uint value for given machine.
	MaxUint = ^uint(0)
	// MaxInt defines maximum int value for given machine.
	MaxInt = int(MaxUint >> 1)
)

// IntegerBitSize is the integer bit size for given machine.
var IntegerBitSize int

func init() {
	if MaxInt == math.MaxInt32 {
		IntegerBitSize = 32
	} else {
		IntegerBitSize = 64
	}
}

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
