package jsonapi

import (
	"fmt"
	"reflect"
)

type HookBeforeCreator interface {
	JSONAPIBeforeCreate(scope *Scope) error
}

type HookAfterCreator interface {
	JSONAPIAfterCreate(scope *Scope) error
}

type HookBeforeReader interface {
	JSONAPIBeforeRead(scope *Scope) error
}

type HookAfterReader interface {
	JSONAPIAfterRead(scope *Scope) error
}

type HookBeforePatcher interface {
	JSONAPIBeforePatch(scope *Scope) error
}

type HookAfterPatcher interface {
	JSONAPIAfterPatch(scope *Scope) error
}

type HookBeforeDeleter interface {
	JSONAPIBeforeDelete(scope *Scope) error
}

type HookAfterDeleter interface {
	JSONAPIAfterDelete(scope *Scope) error
}

func (h *JSONAPIHandler) HookBeforeReader(scope *Scope) *ErrorObject {
	if scope.Value == nil {
		h.log.Errorf("Provided nil value to HookBeforeReader. Model: %v", scope.Struct.GetType().Name())
		return ErrInternalError.Copy()
	}

	beforeReader := reflect.TypeOf((*HookBeforeReader)(nil)).Elem()
	t := reflect.New(scope.Struct.GetType()).Type()
	if !t.Implements(beforeReader) {
		h.log.Debugf("The %v does not implement: %v.", t, beforeReader)
		return nil
	}

	v := reflect.ValueOf(scope.Value)
	switch v.Kind() {
	case reflect.Ptr:
		beforeReader, ok := scope.Value.(HookBeforeReader)
		if ok {
			if err := beforeReader.JSONAPIBeforeRead(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					return errObj
				}
				h.log.Errorf("Unknown error type in JSONAPIBeforeRead. Error: %v", err)
				return ErrInternalError.Copy()
			}
		}
	case reflect.Slice:

		for i := 0; i < v.Len(); i++ {

			single := v.Index(i).Interface()
			h.log.Debugf("Elem: %v", single)

			beforeReader, ok := single.(HookBeforeReader)
			if ok {
				h.log.Debugf("Implements: %v", beforeReader)
				if err := beforeReader.JSONAPIBeforeRead(scope); err != nil {
					if errObj, ok := err.(*ErrorObject); ok {
						return errObj
					}
					h.log.Errorf("Unknown error type in JSONAPIBeforeRead. Error: %v", err)
					return ErrInternalError.Copy()
				}
			} else {
				h.log.Debug("Does not implement elem inside slice!!!")
			}
			v.Index(i).Set(reflect.ValueOf(single))
		}
	default:
		h.log.Errorf("Provided invalid value for HookBeforeReader. Model: %v. Value: %v", scope.Struct.GetType().Name(), v)
		return ErrInternalError.Copy()
	}

	return nil
}

func (h *JSONAPIHandler) HookAfterReader(scope *Scope) *ErrorObject {
	if scope.Value == nil {
		h.log.Errorf("Provided nil value after HookAfterReader. Model: %v", scope.Struct.GetType().Name())
		return ErrInternalError.Copy()
	}

	afterReader := reflect.TypeOf((*HookAfterReader)(nil)).Elem()
	if !reflect.New(scope.Struct.GetType()).Type().Implements(afterReader) {
		h.log.Debugf("'%v' does not implement After Reader: %v", reflect.New(scope.Struct.GetType()).Type(), afterReader)
		return nil
	}

	v := reflect.ValueOf(scope.Value)
	switch v.Kind() {
	case reflect.Ptr:
		afterReader, ok := scope.Value.(HookAfterReader)
		if ok {
			if err := afterReader.JSONAPIAfterRead(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					return errObj
				}
				h.log.Error(err)
				errObj := ErrInternalError.Copy()
				errObj.Detail = fmt.Sprintf("Error while using HookAfterReader for single value. Model %v", scope.Struct.GetType().Name())
				return errObj
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i).Interface()

			afterReader, ok := single.(HookAfterReader)
			if ok {
				if err := afterReader.JSONAPIAfterRead(scope); err != nil {
					if errObj, ok := err.(*ErrorObject); ok {
						return errObj
					}
					h.log.Error(err)
					errObj := ErrInternalError.Copy()
					errObj.Detail = fmt.Sprintf("Error while using HookAfterReader for single value. Model %v", scope.Struct.GetType().Name())
					return errObj
				}
			}
			v.Index(i).Set(reflect.ValueOf(single))
		}
	default:
		h.log.Errorf("Provided invalid value for HookAfterReader. Model: %v. Value: %v", scope.Struct.GetType().Name(), v)
		return ErrInternalError.Copy()
	}
	return nil

}
