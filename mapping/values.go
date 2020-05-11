package mapping

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/neuronlabs/neuron/log"
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

// StringValues gets the string values from the provided struct field, 'value'.
func StringValues(value interface{}, svalues *[]string) []string {
	if svalues == nil {
		values := []string{}
		svalues = &values
	}
	if value == nil {
		*svalues = append(*svalues, "null")
	}
	stringValue(value, svalues)
	return *svalues
}

func stringValue(val interface{}, values *[]string) {
	switch v := val.(type) {
	case string:
		*values = append(*values, v)
	case uint:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint8:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint16:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint32:
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case uint64:
		*values = append(*values, strconv.FormatUint(v, 10))
	case int:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int8:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int16:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int32:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int64:
		*values = append(*values, strconv.FormatInt(v, 10))
	case float32:
		*values = append(*values, strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		*values = append(*values, strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		*values = append(*values, strconv.FormatBool(v))
	case time.Time:
		*values = append(*values, strconv.FormatInt(v.Unix(), 10))
	case *time.Time:
		*values = append(*values, strconv.FormatInt(v.Unix(), 10))
	case fmt.Stringer:
		*values = append(*values, v.String())
	case []string:
		for _, s := range v {
			stringValue(s, values)
		}
	case []interface{}:
		for _, s := range v {
			stringValue(s, values)
		}
	case []time.Time:
		for _, s := range v {
			stringValue(s, values)
		}
	case []*time.Time:
		for _, s := range v {
			stringValue(s, values)
		}
	case []int64:
		for _, s := range v {
			stringValue(s, values)
		}
	case []int32:
		for _, s := range v {
			stringValue(s, values)
		}
	case []int16:
		for _, s := range v {
			stringValue(s, values)
		}
	case []int8:
		for _, s := range v {
			stringValue(s, values)
		}
	case []uint8:
		for _, s := range v {
			stringValue(s, values)
		}
	case []uint16:
		for _, s := range v {
			stringValue(s, values)
		}
	case []uint32:
		for _, s := range v {
			stringValue(s, values)
		}
	case []uint64:
		for _, s := range v {
			stringValue(s, values)
		}
	case []uint:
		for _, s := range v {
			stringValue(s, values)
		}
	case []fmt.Stringer:
		for _, s := range v {
			stringValue(s, values)
		}
	default:
		log.Debugf("Filter parse query. Invalid or unsupported value type: %t", v)
	}
}
