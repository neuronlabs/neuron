package query

import (
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

// IsAutoSelected returns boolean if the scope has auto selected fields.
func (s *Scope) IsAutoSelected() bool {
	return s.autoSelectedFields
}

// IsPrimarySelected checks if the primary field is selected.
func (s *Scope) IsPrimarySelected() bool {
	return s.isPrimarySelected()
}

// IsSelected iterates over all selected fields within given scope and returns
// boolean statement if provided 'field' is already selected.
// Returns an error if the 'field' is not found within given model or it is of invalid type.
// 'field' may be of following types/formats:
// 	- NeuronName - string
//	- Name - string
//	- mapping.StructField
func (s *Scope) IsSelected(field interface{}) (bool, error) {
	selectedFields := &s.SelectedFields
	return s.isSelected(field, selectedFields, false)
}

// NotSelectedFields returns fields that are not selected in the query.
func (s *Scope) NotSelectedFields(withForeigns ...bool) (notSelected []*mapping.StructField) {
	return s.notSelectedFields(withForeigns...)
}

// SelectField selects the field by the name.
// Selected fields are used in the patching process.
// By default selected fields are all non zero valued fields in the struct.
func (s *Scope) SelectField(name string) error {
	field, ok := s.Struct().FieldByName(name)
	if !ok {
		log.Debug("Field not found: '%s'", name)
		return errors.NewDetf(class.QuerySelectedFieldsNotFound, "field: '%s' not found", name)
	}
	s.SelectedFields = append(s.SelectedFields, field)
	return nil
}

// SelectFields clears current scope selected fields and set it to
// the provided 'fields'. The fields might be a string (NeuronName, Name)
// or the *mapping.StructField.
func (s *Scope) SelectFields(fields ...interface{}) error {
	return s.selectFields(true, fields...)
}

// autoSelectFields selects the fields automatically if none of the select field method were called.
func (s *Scope) autoSelectFields() error {
	if s.SelectedFields != nil {
		return nil
	}

	if s.Value == nil {
		return errors.NewDet(class.QueryNoValue, "no value provided for scope")
	}

	if log.Level() == log.LDEBUG3 {
		defer func() {
			fieldsInflection := "field"
			if len(s.SelectedFields) > 1 {
				fieldsInflection += "s"
			}
			log.Debug3f("SCOPE[%s][%s] Auto selected '%d' %s.", s.ID(), s.Struct().Collection(), len(s.SelectedFields), fieldsInflection)
		}()
	}

	s.autoSelectedFields = true
	v := reflect.ValueOf(s.Value).Elem()

	// check if the value is a struct
	if v.Kind() != reflect.Struct {
		return errors.NewDet(class.QuerySelectedFieldsInvalidModel, "auto select fields model is not a single struct model")
	}

	for _, field := range s.mStruct.StructFields() {
		switch field.FieldKind() {
		case mapping.KindPrimary, mapping.KindAttribute, mapping.KindForeignKey, mapping.KindRelationshipSingle, mapping.KindRelationshipMultiple:
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
			s.SelectedFields = append(s.SelectedFields, field)
		}
	}
	return nil
}

func (s *Scope) isPrimarySelected() bool {
	for _, field := range s.SelectedFields {
		if field == s.mStruct.Primary() {
			return true
		}
	}
	return false
}

func (s *Scope) isSelected(field interface{}, selectedFields *[]*mapping.StructField, erease bool) (bool, error) {
	switch typedField := field.(type) {
	case string:
		for i, selectedField := range *selectedFields {
			if selectedField.NeuronName() == typedField || selectedField.Name() == typedField {
				if erease {
					(*selectedFields) = append((*selectedFields)[:i], (*selectedFields)[i+1:]...)
				}
				return true, nil
			}
		}
	case *mapping.StructField:
		if typedField.Struct() != s.Struct() {
			return false, errors.NewDet(class.QuerySelectedFieldsInvalidModel, "selected field's model doesn't match query model")
		}
		for i := 0; i < len(*selectedFields); i++ {
			selectedField := (*selectedFields)[i]
			if typedField == selectedField {
				if erease {
					(*selectedFields) = append((*selectedFields)[:i], (*selectedFields)[i+1:]...)
				}
				return true, nil
			}
		}
	}
	return false, nil
}

// notSelectedFields lists all the fields that are not selected within the scope.
func (s *Scope) notSelectedFields(foreignKeys ...bool) []*mapping.StructField {
	var notSelected []*mapping.StructField

	fields := s.Struct().Fields()
	if len(foreignKeys) > 0 && foreignKeys[0] {
		fields = append(fields, s.Struct().ForeignKeys()...)
	}

	for _, sf := range fields {
		var found bool
		for _, selected := range s.SelectedFields {
			if sf == selected {
				found = true
				break
			}
		}
		if !found {
			notSelected = append(notSelected, sf)
		}
	}
	return notSelected
}

func (s *Scope) selectFields(initContainer bool, fields ...interface{}) error {
	var sfields []*mapping.StructField
	selectedFields := make(map[*mapping.StructField]struct{})

	for _, field := range fields {
		var found bool

		switch f := field.(type) {
		case string:
			for _, sField := range s.mStruct.Fields() {
				if sField.NeuronName() == f || sField.Name() == f {
					selectedFields[sField] = struct{}{}
					sfields = append(sfields, sField)
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
				return errors.NewDetf(class.QuerySelectedFieldsNotFound, "field '%s' not found within model: '%s'", f, s.mStruct.Type().Name())
			}
		case *mapping.StructField:
			for _, sField := range s.mStruct.Fields() {
				if sField == f {
					found = true
					selectedFields[sField] = struct{}{}
					sfields = append(sfields, sField)
					break
				}
			}
			if !found {
				log.Debugf("Field: '%v' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				return errors.NewDetf(class.QuerySelectedFieldsNotFound, "field '%s' not found within model: '%s'", f.Name(), s.mStruct.Type().Name())
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return errors.NewDetf(class.QuerySelectedFieldInvalid, "unknown selected field type: %T", f)
		}
	}

	if initContainer {
		s.SelectedFields = sfields
		return nil
	}

	// check if fields were not already selected
	for _, alreadySelected := range s.SelectedFields {
		_, ok := selectedFields[alreadySelected]
		if ok {
			log.Errorf("Field: %s already set for the given scope.", alreadySelected.Name())
			return errors.NewDetf(class.QuerySelectedFieldAlreadyUsed, "field: '%s' were already selected", alreadySelected.Name())
		}
	}

	// add all fields to scope's selected fields
	s.SelectedFields = append(s.SelectedFields, sfields...)
	return nil
}

func (s *Scope) unselectFieldIfSelected(field *mapping.StructField) {
	for i := 0; i < len(s.SelectedFields); i++ {
		if s.SelectedFields[i] == field {
			s.SelectedFields = append(s.SelectedFields[:i], s.SelectedFields[i+1:]...)
			i--
		}
	}
}

func (s *Scope) unselectFields(fields ...*mapping.StructField) error {
scopeFields:
	for i := 0; i < len(s.SelectedFields); i++ {
		if len(fields) == 0 {
			break scopeFields
		}

		for j := 0; j < len(fields); j++ {
			if s.SelectedFields[i] == fields[j] {
				fields = append(fields[:j], fields[j+1:]...)
				j--
				// Erease from Selected fields
				s.SelectedFields = append(s.SelectedFields[:i], s.SelectedFields[i+1:]...)
				i--
				continue scopeFields
			}
		}
	}

	if len(fields) > 0 {
		var notEreased string
		for _, field := range fields {
			notEreased += field.NeuronName() + " "
		}
		return errors.NewDetf(class.QuerySelectedFieldsNotSelected, "unselecting non selected fields: '%s'", notEreased)
	}
	return nil
}
