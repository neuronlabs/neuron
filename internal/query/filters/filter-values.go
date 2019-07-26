package filters

import (
	"reflect"
	"strconv"
	"time"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
)

// OpValuePair are the values used within the provided FilterField
// It contains the Values which is a slice of provided values for given 'Operator'
type OpValuePair struct {
	Values   []interface{}
	operator *Operator
}

// Operator gets the operator for OpValuePair
func (o *OpValuePair) Operator() *Operator {
	return o.operator
}

// SetOperator sets the operator for given opvalue pair
func (o *OpValuePair) SetOperator(op *Operator) {
	o.operator = op
}

// CopyOpValuePair copies operator value pair
func CopyOpValuePair(o *OpValuePair) *OpValuePair {
	return o.copy()
}

// NewOpValuePair creates new operator value pair
func NewOpValuePair(o *Operator, values ...interface{}) *OpValuePair {
	op := &OpValuePair{operator: o, Values: values}
	return op
}

func (o *OpValuePair) copy() *OpValuePair {
	fv := &OpValuePair{operator: o.operator}
	fv.Values = make([]interface{}, len(o.Values))
	copy(fv.Values, o.Values)
	return fv
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

		err = errors.NewDet(class.QueryFilterValue, "filtering over nested structure is not supported yet")
		err.SetDetails("Filtering over nested structures is not supported yet.")
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
		value = "false"
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
