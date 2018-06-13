package jsonapi

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type assertOptions int

const (
	_ assertOptions = iota
	failNow
	printPanic
)

func assertPanic(t *testing.T, testFunc func(), options ...assertOptions) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The function: %v should panic. \n%s", reflect.TypeOf(testFunc).Name(), traceInfo(3))
			treatOptions(t, options...)
		} else {
			for _, option := range options {
				if option == printPanic {
					t.Log(r)
				}
			}
		}

	}()
	testFunc()
}

func assertNoPanic(t *testing.T, testFunc func(), options ...assertOptions) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The function should not panic.'%s'. \n%s", r, traceInfo(3))

			treatOptions(t, options...)
		}
	}()
	testFunc()
}

func assertError(t *testing.T, err error, options ...assertOptions) {
	if err == nil {
		t.Errorf("Provided error is nil. \n%s", traceInfo(2))
		treatOptions(t, options...)
	}
}

func assertNoError(t *testing.T, err error, options ...assertOptions) {
	if err != nil {
		t.Errorf("Given error is not nil: %s. \n%s", err, traceInfo(2))
		treatOptions(t, options...)
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
			t.Errorf("Provided nil error.\n%s", traceInfo(2))
		}
	}
}

func assertNotNil(t *testing.T, obj interface{}, options ...assertOptions) {
	if isNil(obj) {
		t.Errorf("Provided obj: '%v' should not be nil.\n%s", obj, traceInfo(2))
		treatOptions(t, options...)
	}
}

func assertNil(t *testing.T, obj interface{}, options ...assertOptions) {
	if !isNil(obj) {
		t.Errorf("Provided obj: %v is not nil.\n%s", obj, traceInfo(2))
		treatOptions(t, options...)
	}
}

func assertNilErrors(t *testing.T, errs ...error) {
	for _, err := range errs {
		if err != nil {
			t.Errorf("Given errors are not nil: %s. \n%s", err, traceInfo(2))
		}
	}
}

func assertTrue(t *testing.T, value bool, options ...assertOptions) {
	if !value {
		t.Errorf("Provided value is not true. \n%sv", traceInfo(2))
		treatOptions(t, options...)
	}
}

func assertFalse(t *testing.T, value bool, options ...assertOptions) {
	if value {
		t.Errorf("Provided value is not false. \n%v", traceInfo(2))
		treatOptions(t, options...)
	}
}

func assertEmpty(t *testing.T, obj interface{}, options ...assertOptions) {
	if !isEmpty(obj) {
		t.Errorf("Object is not empty. \n%v", traceInfo(2))
		treatOptions(t, options...)
	}

}

func assertNotEmpty(t *testing.T, obj interface{}, options ...assertOptions) {
	if isEmpty(obj) {
		t.Errorf("Object is empty. \n%v", traceInfo(2))
		treatOptions(t, options...)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}, options ...assertOptions) {
	if !areEqual(expected, actual) {
		t.Errorf("Objects: %v and %v  are not equal. \n%s", expected, actual, traceInfo(2))
		treatOptions(t, options...)
	}
}

func assertNotEqual(t *testing.T, expected, actual interface{}, options ...assertOptions) {
	if areEqual(expected, actual) {
		t.Errorf("Objects: %v and %v are equal. \n%s", expected, actual, traceInfo(2))
		treatOptions(t, options...)
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

func traceInfo(depth int) string {
	pc, file, line, ok := runtime.Caller(depth)
	if ok {
		fnc := runtime.FuncForPC(pc)
		split := strings.Split(fnc.Name(), "/")
		info := fmt.Sprintf("%s \n f:\t%s:%d\n", split[len(split)-1], file, line)
		return info
	}
	return ""

}

func treatOptions(t *testing.T, options ...assertOptions) {
	if len(options) == 0 {
		return
	}

	if options[0] == failNow {
		t.FailNow()
	}
}
