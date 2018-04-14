package jsonapi

import (
	"reflect"
	"strconv"
)

func setUintField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var uintValue uint64

	// Parse unsigned int
	uintValue, err = strconv.ParseUint(value, 10, bitSize)

	if err != nil {
		return err
	}

	// Set uint
	fieldValue.SetUint(uintValue)
	return nil
}

func setIntField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var intValue int64
	intValue, err = strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return err
	}

	// Set value if no error
	fieldValue.SetInt(intValue)
	return nil
}

func setFloatField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var floatValue float64

	// Parse float
	floatValue, err = strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}
	fieldValue.SetFloat(floatValue)
	return nil
}

func setBoolField(value string, fieldValue reflect.Value) (err error) {
	var boolValue bool
	// set default if empty
	if value == "" {
		value = "false"
	}
	boolValue, err = strconv.ParseBool(value)
	if err != nil {
		return err
	}
	fieldValue.SetBool(boolValue)
	return nil
}
