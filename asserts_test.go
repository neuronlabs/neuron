package jsonapi

import (
	"bytes"
	"reflect"
	"testing"
)

func assertPanic(t *testing.T, testFunc func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The function: %v should panic.", testFunc)
		}
	}()
	testFunc()
}

func assertError(t *testing.T, err error) {
	if err == nil {
		t.Error("Provided error is nil.")
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func assertErrors(t *testing.T, errs ...error) {
	for _, err := range errs {
		if err == nil {
			t.Error("Provided errors are nil")
		}
	}
}

func assertErrorObjects(t *testing.T, errs ...*ErrorObject) {
	for _, err := range errs {
		if err == nil {
			t.Error("Provided nil error.")
		}
	}
}

func assertNotNil(t *testing.T, obj interface{}) {
	if isNil(obj) {
		t.Errorf("Provided obj: '%v' should not be nil.", obj)
	}
}

func assertNil(t *testing.T, obj interface{}) {
	if !isNil(obj) {
		t.Errorf("Provided obj: %v is not nil.", obj)
	}
}

func assertNilErrors(t *testing.T, errs ...error) {
	for _, err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
}

func assertTrue(t *testing.T, value bool) {
	if !value {
		t.Error("Provided value is not true.")
	}
}

func assertFalse(t *testing.T, value bool) {
	if value {
		t.Error("Provided value is not false")
	}
}

func assertEmpty(t *testing.T, obj interface{}) {
	if !isEmpty(obj) {
		t.Error("Object is not empty")
	}
}

func assertNotEmpty(t *testing.T, obj interface{}) {
	if isEmpty(obj) {
		t.Error("Object is empty.")
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	if !areEqual(expected, actual) {
		t.Errorf("Objects: %v and %v  are not equal.", expected, actual)
	}
}

func assertNotEqual(t *testing.T, expected, actual interface{}) {
	if areEqual(expected, actual) {
		t.Errorf("Objects: %s and %s are equal.", expected, actual)
	}
}

func isEmpty(obj interface{}) bool {
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		v = v.Elem()
		return isEmpty(v.Interface())
	default:
		zero := reflect.Zero(v.Type())
		return reflect.DeepEqual(obj, zero)
	}
}

func isNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	v := reflect.ValueOf(obj)
	k := v.Kind()
	if k >= reflect.Chan && k <= reflect.Slice && v.IsNil() {
		return true
	}
	return false
}

func areEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	if exp, ok := expected.([]byte); ok {
		act, ok := actual.([]byte)
		if !ok {
			return false
		} else if exp == nil || act == nil {
			return exp == nil && act == nil
		}
		return bytes.Equal(exp, act)
	}
	return reflect.DeepEqual(expected, actual)
}
