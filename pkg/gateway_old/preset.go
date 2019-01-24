package gateway

import (
	"errors"
	"github.com/kucjac/uni-db"
	"net/http"
	"reflect"
	"runtime/debug"
)

// GetPresetValues gets the values from the presetScope
func (h *Handler) GetPresetValues(
	presetScope *Scope,
	rw http.ResponseWriter,
) (values []interface{}, err error) {
	h.log.Debug("------Getting Preset Values-------")
	h.log.Debug("Preset Fieldset:")
	for field := range presetScope.Fieldset {
		h.log.Debug(field)
	}

	repo := h.GetRepositoryByType(presetScope.Struct.GetType())

	presetScope.NewValueMany()

	if errObj := h.HookBeforeReader(presetScope); errObj != nil {
		h.MarshalErrors(rw)
		err = newHandlerError(HErrAlreadyWritten, errObj.Error())
		return
	}

	err = repo.List(presetScope)
	if err != nil {
		if dbErr, ok := err.(*unidb.Error); ok {
			h.manageDBError(rw, dbErr)
			err = newHandlerError(HErrAlreadyWritten, dbErr.Message)
			return
		} else if errObj, ok := err.(*ErrorObject); ok {
			err = newHandlerError(HErrAlreadyWritten, errObj.Err.Error())
			h.MarshalErrors(rw)
			return
		}
	}
	v := reflect.ValueOf(presetScope.Value)
	for i := 0; i < v.Len(); i++ {
		h.log.Debugf("Value of presetscope: %+v at Index: %v", v.Index(i).Interface(), i)
	}

	if errObj := h.HookAfterReader(presetScope); errObj != nil {
		h.MarshalErrors(rw)
		err = newHandlerError(HErrAlreadyWritten, errObj.Error())
		return
	}

	scopeVal := reflect.ValueOf(presetScope.Value)
	if scopeVal.Len() == 0 {
		hErr := newHandlerError(HErrNoValues, "Provided resource does not exists.")
		err = hErr
		return
	}

	// set the primary values for the collection scope
	if err = presetScope.SetCollectionValues(); err != nil {
		hErr := newHandlerError(HErrInternal, err.Error())
		hErr.Scope = presetScope
		err = hErr
		return
	}

	for presetScope.NextIncludedField() {
		var field *IncludeField
		field, err = presetScope.CurrentIncludedField()
		if err != nil {
			hErr := newHandlerError(HErrInternal, err.Error())
			hErr.Scope = presetScope
			err = hErr
			return
		}

		var missing []interface{}
		missing, err = field.GetMissingPrimaries()
		if err != nil {
			hErr := newHandlerError(HErrInternal, err.Error())
			hErr.Field = field.StructField
			err = hErr
			return
		}
		if len(missing) == 0 {
			h.log.Debugf("Missing error")
			hErr := newHandlerError(HErrNoValues, "")
			err = hErr
			return
		}

		// Add the missing id filters
		field.Scope.SetIDFilters(missing...)

		if len(field.Scope.IncludedFields) != 0 {
			values, err = h.GetPresetValues(field.Scope, rw)
			if err != nil {
				return
			}
			return
		}
		return missing, nil
	}
	return presetScope.GetPrimaryFieldValues()

}

// PresetScopeValue presets provided values for given scope.
// The fieldFilter points where the value should be set within given scope.
// The scope value should not be nil
func (h *Handler) PresetScopeValue(
	scope *Scope,
	fieldFilter *FilterField,
	values ...interface{},
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Errorf("Panic within preset scope value. %v. %v", r, string(debug.Stack()))
			err = IErrPresetInvalidScope
		}
	}()

	if scope.Value == nil {
		return IErrScopeNoValue
	}

	if len(values) == 0 {
		return IErrPresetNoValues
	}

	v := reflect.ValueOf(scope.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		if v.Kind() != reflect.Struct {
			return IErrPresetInvalidScope
		}
	}

	fIndex := fieldFilter.GetFieldIndex()
	field := v.Field(fIndex)

	if len(fieldFilter.Nested) != 0 {
		switch fieldFilter.GetFieldKind() {

		case RelationshipSingle:
			relIndex := fieldFilter.Nested[0].GetFieldIndex()
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			relatedField := field.Elem().Field(relIndex)
			switch relatedField.Kind() {
			case reflect.Slice:
				refValues := reflect.ValueOf(values)
				relatedField = reflect.AppendSlice(relatedField, refValues)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				relatedField.Set(reflect.ValueOf(values[0]))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				relatedField.Set(reflect.ValueOf(values[0]))
			case reflect.Float32, reflect.Float64:
				relatedField.Set(reflect.ValueOf(values[0]))
			case reflect.String:
				relatedField.Set(reflect.ValueOf(values[0]))
			case reflect.Struct:
				relatedField.Set(reflect.ValueOf(values[0]))
			case reflect.Ptr:
				relatedField.Set(reflect.ValueOf(values[0]))
			}

			if len(values) > 1 {
				h.log.Errorf("Provided values length is greatern than 1 but the field is not of slice type. Presetting only the first value. Field: '%s'.", field.String())
			}

		case RelationshipMultiple:
			fieldElemType := fieldFilter.GetRelatedModelType()
			relatedIndex := fieldFilter.Nested[0].GetFieldIndex()
			for _, value := range values {
				refValue := reflect.ValueOf(value)
				fieldElem := reflect.New(fieldElemType)
				relatedField := fieldElem.Field(relatedIndex)
				switch relatedField.Kind() {
				case reflect.Slice:
					relatedField = reflect.Append(relatedField, refValue)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					relatedField.Set(refValue)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					relatedField.Set(refValue)
				case reflect.Float32, reflect.Float64:
					relatedField.Set(refValue)
				case reflect.String:
					relatedField.Set(refValue)
				case reflect.Struct:
					relatedField.Set(refValue)
				case reflect.Ptr:
					relatedField.Set(refValue)
				}

				field = reflect.Append(field, relatedField)
			}
		default:
			h.log.Error("Relationship filter is of invalid type. Model: '%s'. Field: '%s'", scope.Struct.GetType().Name(), field.String())
			return IErrPresetInvalidScope
		}

	} else {

		switch field.Kind() {
		case reflect.Slice:
			refValues := reflect.ValueOf(values)
			field = reflect.AppendSlice(field, refValues)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.Set(reflect.ValueOf(values[0]))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.Set(reflect.ValueOf(values[0]))
		case reflect.Float32, reflect.Float64:
			field.Set(reflect.ValueOf(values[0]))
		case reflect.String:
			field.Set(reflect.ValueOf(values[0]))
		case reflect.Struct:
			field.Set(reflect.ValueOf(values[0]))
		case reflect.Ptr:
			field.Set(reflect.ValueOf(values[0]))
		}

		if len(values) > 1 {
			h.log.Warningf("Provided values length is greatern than 1 but the field is not of slice type. Presetting only the first value. Field: '%s'.", field.String())
		}
	}
	return nil
}

func (h *Handler) SetPresetFilters(
	scope *Scope,
	model *ModelHandler,
	req *http.Request,
	rw http.ResponseWriter,
	filters ...*PresetFilter,
) (ok bool) {
	for _, filter := range filters {
		if value := req.Context().Value(filter.Key); value != nil {
			if err := h.SetPresetFilterValues(filter.FilterField, value); err != nil {
				h.log.Errorf("Cannot set preset filter values. Model: %v, Filterfield: %v, Value: %v, Path: %v. Error: %v", model.ModelType.Name(), filter.GetFieldName(), value, req.URL.Path, err)
				h.MarshalInternalError(rw)
				return
			}
		}
		if err := scope.AddFilterField(filter.FilterField); err != nil {
			h.log.Errorf("Cannot add filter field. Path: %v, Error: %v", req.URL.Path, err)
			h.MarshalInternalError(rw)
			return
		}
	}
	return true
}

// SetPresetValues sets provided values for given filterfield.
// If the filterfield does not contain values or subfield values the function returns error.
func (h *Handler) SetPresetFilterValues(
	filter *FilterField,
	values ...interface{},
) error {
	if len(filter.Values) != 0 {
		fv := filter.Values[0]
		(*fv).Values = values
		return nil
	} else if len(filter.Nested) != 0 {
		if len(filter.Nested[0].Values) != 0 {
			fv := filter.Nested[0].Values[0]
			(*fv).Values = values
			return nil
		}
	}
	return errors.New("Invalid preset filter.")
}

/**

PRIVATE

*/

func (h *Handler) getPresetFilter(
	key interface{},
	presetScope *Scope,
	req *http.Request,
	model *ModelHandler,
) bool {
	return h.getPrecheckFilter(key, presetScope, req, model)
}

func (h *Handler) getPrecheckFilter(
	key interface{},
	precheckScope *Scope,
	req *http.Request,
	model *ModelHandler,
) (exists bool) {
	h.log.Debugf("Key value: %+v", key)
	precheckValue := req.Context().Value(key)
	if precheckValue == nil {
		h.log.Warningf("Empty preset value in precheckpair for model: %s", model.ModelType.Name())
		return
	}

	precheckFilter, ok := precheckValue.(*FilterField)
	if !ok {
		presetFilter, ok := precheckValue.(*PresetFilter)
		if !ok {
			h.log.Warningf("PrecheckValue is not a FilterField. Model: %v, endpoint: CREATE", model.ModelType.Name())
			return
		}
		precheckFilter = presetFilter.FilterField
	}
	if !h.addPresetFilterToPresetScope(precheckScope, precheckFilter) {
		return
	}
	return true
}
