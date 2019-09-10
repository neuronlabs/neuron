package query

import (
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

// InFieldset checks if the provided field is in the scope's fieldset.
func (s *Scope) InFieldset(field interface{}) (*mapping.StructField, bool) {
	switch fv := field.(type) {
	case string:
		f, ok := s.Fieldset[fv]
		if !ok {
			for _, f := range s.Fieldset {
				if f.Name() == field {
					return f, true
				}
			}
		}
		return f, ok
	case *mapping.StructField:
		for _, f := range s.Fieldset {
			if f == fv {
				return f, true
			}
		}
		return nil, false
	default:
		log.Warningf("Unknown field type: %v - %v", reflect.TypeOf(field), field)
		return nil, false
	}
}

func (s *Scope) addToFieldset(fields ...interface{}) error {
	for _, field := range fields {
		var found bool
		switch f := field.(type) {
		case string:
			if "*" == f {
				for _, sField := range s.mStruct.Fields() {
					s.Fieldset[sField.NeuronName()] = sField
				}
			} else {
				for _, sField := range s.mStruct.Fields() {
					if sField.NeuronName() == f || sField.Name() == f {
						s.Fieldset[sField.NeuronName()] = sField
						found = true
						break
					}
				}
				if !found {
					log.Debugf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
					err := errors.NewDet(class.QueryFieldsetUnknownField, "field not found in the model")
					err.SetDetailsf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
					return err
				}
			}
		case *mapping.StructField:
			for _, sField := range s.mStruct.Fields() {
				if sField == f {
					s.Fieldset[sField.NeuronName()] = f
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%v' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				err := errors.NewDet(class.QueryFieldsetUnknownField, "field not found in the model")
				err.SetDetailsf("Field: '%s' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				return err
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return errors.NewDetf(class.QueryFieldsetInvalid, "provided invalid field type: '%T'", f)
		}
	}
	return nil
}

// autoSelectFields selects the fields automatically if none of the select field method were called.
func (s *Scope) autoSelectFields() error {
	if len(s.Fieldset) != 0 {
		return nil
	}

	if s.Value == nil {
		return errors.NewDet(class.QueryNoValue, "no value provided for scope")
	}

	if log.Level() == log.LDEBUG3 {
		defer func() {
			fieldsInflection := "field"
			if len(s.Fieldset) > 1 {
				fieldsInflection += "s"
			}
			log.Debug3f("SCOPE[%s][%s] Auto selected '%d' %s.", s.ID(), s.Struct().Collection(), len(s.Fieldset), fieldsInflection)
		}()
	}

	v := reflect.ValueOf(s.Value).Elem()

	// check if the value is a struct
	if v.Kind() != reflect.Struct {
		return errors.NewDet(class.QuerySelectedFieldsInvalidModel, "auto select fields model is not a single struct model")
	}

	for _, field := range s.mStruct.Fields() {
		tp := field.ReflectField().Type

		fieldValue := v.FieldByIndex(field.ReflectField().Index)
		switch tp.Kind() {
		case reflect.Map, reflect.Slice, reflect.Ptr:
			if fieldValue.IsNil() {
				continue
			}
		default:
			if reflect.DeepEqual(reflect.Zero(tp).Interface(), fieldValue.Interface()) {
				continue
			}
		}
		s.Fieldset[field.NeuronName()] = field
	}
	return nil
}

// fillFieldsetIfNotSet sets the fieldset to full if the fieldset is not set
func (s *Scope) fillFieldsetIfNotSet() {
	if s.Fieldset == nil || len(s.Fieldset) == 0 {
		s.setAllFields()
	}
}

func (s *Scope) isDefaultFieldset() bool {
	return len(s.Fieldset) == len(s.Struct().Fields())
}

func (s *Scope) setAllFields() {
	fieldset := map[string]*mapping.StructField{}
	for _, field := range s.mStruct.Fields() {
		fieldset[field.NeuronName()] = field
	}
	s.Fieldset = fieldset
}

// setFields sets the fieldset from the provided fields
func (s *Scope) setFields(fields ...interface{}) error {
	s.Fieldset = map[string]*mapping.StructField{}
	return s.addToFieldset(fields...)
}

func (s *Scope) setFieldsetNoCheck(fields ...*mapping.StructField) {
	for _, field := range fields {
		s.Fieldset[field.NeuronName()] = field
	}
}
