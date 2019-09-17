package mapping

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// PrimaryValues gets the primary values from the provided 'value' reflection.
// The reflection might be a slice or a single value.
// Returns the slice of interfaces containing primary field values.
func PrimaryValues(m *ModelStruct, value reflect.Value) (primaries []interface{}, err error) {
	if value.IsNil() {
		err = errors.NewDet(class.QueryNoValue, "nil value provided")
		return nil, err
	}

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	zero := reflect.Zero(m.primary.ReflectField().Type).Interface()

	appendPrimary := func(primary reflect.Value) bool {
		if primary.IsValid() {
			pv := primary.Interface()
			if !reflect.DeepEqual(pv, zero) {
				primaries = append(primaries, pv)
			}
			return true
		}
		return false
	}

	primaryIndex := m.primary.getFieldIndex()
	switch value.Type().Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() != reflect.Ptr {
			err = errors.NewDet(class.QueryValueType, "provided invalid value - the slice doesn't contain pointers to models")
			return nil, err
		}

		// create slice of values
		for i := 0; i < value.Len(); i++ {
			single := value.Index(i)
			if single.IsNil() {
				continue
			}
			single = single.Elem()
			primaryValue := single.FieldByIndex(primaryIndex)
			appendPrimary(primaryValue)
		}
	case reflect.Struct:
		primaryValue := value.FieldByIndex(primaryIndex)
		if !appendPrimary(primaryValue) {
			err = errors.NewDetf(class.QueryValueType, "provided invalid value for model: %v", m.Type())
			return nil, err
		}
	default:
		err = errors.NewDet(class.QueryValueType, "unknown value type")
		return nil, err
	}
	return primaries, nil
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

// NewReflectValueMany creates the *[]*m.Type reflect.Value.
func NewReflectValueMany(m *ModelStruct) reflect.Value {
	return reflect.New(reflect.SliceOf(reflect.PtrTo(m.Type())))
}

func setPrimaryField(value string, fieldValue reflect.Value) errors.DetailedError {
	var err errors.DetailedError
	// if the id field is of string type set it to the strValue
	t := fieldValue.Type()

	switch t.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int:
		err = setIntField(value, fieldValue, 64)
	case reflect.Int16:
		err = setIntField(value, fieldValue, 16)
	case reflect.Int32:
		err = setIntField(value, fieldValue, 32)
	case reflect.Int64:
		err = setIntField(value, fieldValue, 64)
	case reflect.Uint:
		err = setUintField(value, fieldValue, 64)
	case reflect.Uint16:
		err = setUintField(value, fieldValue, 16)
	case reflect.Uint32:
		err = setUintField(value, fieldValue, 32)
	case reflect.Uint64:
		err = setUintField(value, fieldValue, 64)
	default:
		err = errors.NewDetf(class.InternalQueryFilter, "model primary invalid format type: '%s'", t.Name())
	}
	return err
}

func setAttributeField(value string, fieldValue reflect.Value) errors.DetailedError {
	var err errors.DetailedError
	// the attribute can be:
	t := fieldValue.Type()

	switch t.Kind() {
	case reflect.Int:
		err = setIntField(value, fieldValue, 64)
	case reflect.Int8:
		err = setIntField(value, fieldValue, 8)
	case reflect.Int16:
		err = setIntField(value, fieldValue, 16)
	case reflect.Int32:
		err = setIntField(value, fieldValue, 32)
	case reflect.Int64:
		err = setIntField(value, fieldValue, 64)
	case reflect.Uint:
		err = setUintField(value, fieldValue, 64)
	case reflect.Uint16:
		err = setUintField(value, fieldValue, 16)
	case reflect.Uint32:
		err = setUintField(value, fieldValue, 32)
	case reflect.Uint64:
		err = setUintField(value, fieldValue, 64)
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Bool:
		err = setBoolField(value, fieldValue)
	case reflect.Float32:
		err = setFloatField(value, fieldValue, 32)
	case reflect.Float64:
		err = setFloatField(value, fieldValue, 64)
	case reflect.Struct:
		// check if it is time
		if _, ok := fieldValue.Elem().Interface().(time.Time); ok {
			// TODO: set the time field
			err = setTimeField(value, fieldValue)
		} else {
			// TODO: set the nested attribute struct
			err = errors.NewDet(class.QueryFilterValue, "filtering over nested structure is not supported yet")
			err.SetDetails("Filtering over nested structures is not supported yet.")
		}
	default:
		log.Debug("Filtering over unsupported type: '%s'", t.Name())

		err = errors.NewDet(class.QueryFilterValue, "Unsupported field value")
	}
	return err
}

func setTimeField(value string, fieldValue reflect.Value) errors.DetailedError {
	err := errors.NewDet(class.QueryFilterValue, "filtering over time field is not supported yet")
	err.SetDetails("Filtering over time fields is not supported yet.")
	return err
}

func setUintField(value string, fieldValue reflect.Value, bitSize int) errors.DetailedError {
	// Parse unsigned int
	uintValue, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid unsigned integer value.")
		return err
	}

	// Set uint
	fieldValue.SetUint(uintValue)
	return nil
}

func setIntField(value string, fieldValue reflect.Value, bitSize int) errors.DetailedError {
	intValue, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid integer value.")
		return err
	}

	// Set value if no error
	fieldValue.SetInt(intValue)
	return nil
}

func setFloatField(value string, fieldValue reflect.Value, bitSize int) errors.DetailedError {
	// Parse float
	floatValue, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid float value.")
		return err
	}
	fieldValue.SetFloat(floatValue)

	return nil
}

func setBoolField(value string, fieldValue reflect.Value) errors.DetailedError {
	// set default if empty
	if value == "" {
		return nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		err := errors.NewDet(class.QueryFilterValue, "invalid unsinged integer value")
		err.SetDetailsf("Invalid boolean value.")
		return err
	}

	fieldValue.SetBool(boolValue)
	return nil
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
		*values = append(*values, strconv.FormatUint(uint64(v), 10))
	case int:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int8:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int16:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int32:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
	case int64:
		*values = append(*values, strconv.FormatInt(int64(v), 10))
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
