package scope

import (
	"reflect"

	"github.com/neuronlabs/neuron-core/errors"
	"github.com/neuronlabs/neuron-core/errors/class"
	"github.com/neuronlabs/neuron-core/log"

	"github.com/neuronlabs/neuron-core/internal/models"
	"github.com/neuronlabs/neuron-core/internal/namer/dialect"
)

// AddselectedFields adds provided fields into given Scope's SelectedFields.
func AddselectedFields(s *Scope, fields ...string) error {
	for _, addField := range fields {
		field := s.mStruct.FieldByName(addField)
		if field == nil {
			err := errors.New(class.QuerySelectedFieldsNotFound, "selected field not found")
			err.SetDetailf("Field: '%s' not found within model: %s", addField, s.mStruct.Collection())
			return err
		}

		s.selectedFields = append(s.selectedFields, field)

	}
	return nil
}

// AddSelectedFields adds the fields to the scope's selected fields.
func (s *Scope) AddSelectedFields(fields ...interface{}) error {
	return s.addToSelectedFields(fields...)
}

// AddSelectedField adds the selected field to the selected field's array.
func (s *Scope) AddSelectedField(field *models.StructField) {
	s.selectedFields = append(s.selectedFields, field)
}

// AreSelected iterates over all selected fields within given scope and returns
// boolean statement all of provided 'fields' are already selected.
// Returns an error if any of the 'fields' are not found within given model.
// or are of invalid type.
// 'field' may be in a following types/formats:
// 	- NeuronName - string
//	- Name - string
//	- models.StructField
func (s *Scope) AreSelected(fields ...interface{}) (bool, error) {
	if len(s.selectedFields) == 0 {
		return false, nil
	}

	selectedFields := make([]*models.StructField, len(s.selectedFields))
	copy(selectedFields, s.selectedFields)

	for _, field := range fields {
		isSelected, err := s.isSelected(field, &selectedFields, true)
		if err != nil {
			return false, err
		}

		if !isSelected {
			return false, nil
		}
	}

	return true, nil
}

// AutoSelectFields selects the fields automatically if none of the select field method were called.
func (s *Scope) AutoSelectFields() error {
	if s.selectedFields != nil {
		return nil
	}

	if s.Value == nil {
		return errors.New(class.QueryNoValue, "no value provided for scope")
	}

	defer func() {
		fieldsInflection := "field"
		if len(s.selectedFields) > 1 {
			fieldsInflection += "s"
		}
		log.Debug3f("SCOPE[%s][%s] Auto selected '%d' %s.", s.ID(), s.Struct().Collection(), len(s.selectedFields), fieldsInflection)
	}()

	v := reflect.ValueOf(s.Value).Elem()

	// check if the value is a struct
	if v.Kind() != reflect.Struct {
		return errors.New(class.QuerySelectedFieldsInvalidModel, "auto select fields model is not a single struct model")
	}

	for _, field := range s.mStruct.StructFields() {
		switch field.FieldKind() {
		case models.KindPrimary, models.KindAttribute, models.KindForeignKey, models.KindRelationshipSingle, models.KindRelationshipMultiple:
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
			s.selectedFields = append(s.selectedFields, field)
		}
	}

	return nil
}

// UnselectFields unselects provided fields.
func (s *Scope) UnselectFields(fields ...*models.StructField) error {
	return s.unselectFields(fields...)
}

// UnselectFieldIfSelected unselects provided field if it is selected.
func (s *Scope) UnselectFieldIfSelected(field *models.StructField) {
	for i := 0; i < len(s.selectedFields); i++ {
		if s.selectedFields[i] == field {
			s.selectedFields = append(s.selectedFields[:i], s.selectedFields[i+1:]...)
			i--
		}
	}
}

// DeleteselectedFields deletes the models.StructFields from the given scope Fieldset.
func DeleteselectedFields(s *Scope, fields ...*models.StructField) error {
	return s.unselectFields(fields...)
}

// IsSelected iterates over all selected fields within given scope and returns
// boolean statement if provided 'field' is already selected.
// Returns an error if the 'field' is not found within given model or it is of invalid type.
// 'field' may be of following types/formats:
// 	- NeuronName - string
//	- Name - string
//	- models.StructField
func (s *Scope) IsSelected(field interface{}) (bool, error) {
	selectedFields := &s.selectedFields
	return s.isSelected(field, selectedFields, false)
}

// NotSelectedFields lists all the fields that are not selected within the scope.
func (s *Scope) NotSelectedFields(foreignKeys ...bool) []*models.StructField {
	var notSelected []*models.StructField

	fields := s.Struct().Fields()

	if len(foreignKeys) > 0 && foreignKeys[0] {
		fields = append(fields, s.Struct().ForeignKeys()...)
	}

	for _, sf := range fields {
		var found bool
	inner:
		for _, selected := range s.selectedFields {
			if sf == selected {
				found = true
				break inner
			}
		}
		if !found {
			notSelected = append(notSelected, sf)
		}
	}
	return notSelected
}

// SelectedFields return fields that were selected during unmarshaling.
func (s *Scope) SelectedFields() []*models.StructField {
	return s.selectedFields
}

func (s *Scope) addToSelectedFields(fields ...interface{}) error {
	var sfields []*models.StructField
	selectedFields := make(map[*models.StructField]struct{})

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
				return errors.Newf(class.QuerySelectedFieldsNotFound, "field '%s' not found within model: '%s'", f, s.mStruct.Type().Name())
			}
		case *models.StructField:
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
				return errors.Newf(class.QuerySelectedFieldsNotFound, "field '%s' not found within model: '%s'", f.Name(), s.mStruct.Type().Name())
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return errors.Newf(class.QuerySelectedFieldInvalid, "unknown selected field type: %T", f)
		}
	}

	// check if fields were not already selected
	for _, alreadySelected := range s.selectedFields {
		_, ok := selectedFields[alreadySelected]
		if ok {
			log.Errorf("Field: %s already set for the given scope.", alreadySelected.Name())
			return errors.Newf(class.QuerySelectedFieldAlreadyUsed, "field: '%s' were already selected", alreadySelected.Name())
		}
	}

	// add all fields to scope's selected fields
	s.selectedFields = append(s.selectedFields, sfields...)
	return nil
}

func (s *Scope) deleteSelectedField(index int) error {
	if index > len(s.selectedFields)-1 {
		return errors.Newf(class.InternalQuerySelectedField, "deleting selected field is out of range: %d", index).SetOperation("deleteSelectedField")
	}

	// copy the last element
	if index < len(s.selectedFields)-1 {
		s.selectedFields[index] = s.selectedFields[len(s.selectedFields)-1]
		s.selectedFields[len(s.selectedFields)-1] = nil
	}
	s.selectedFields = s.selectedFields[:len(s.selectedFields)-1]
	return nil
}

func (s *Scope) isSelected(field interface{}, selectedFields *[]*models.StructField, erease bool) (bool, error) {
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
	case *models.StructField:
		if typedField.Struct() != s.Struct() {
			return false, errors.New(class.QuerySelectedFieldsInvalidModel, "selected field's model doesn't match query model")
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

// selectedFieldValues gets the values for given
func (s *Scope) selectedFieldValues(dialectNamer dialect.FieldNamer) (map[string]interface{}, error) {
	if s.Value == nil {
		return nil, errors.New(class.QueryNoValue, "nil query value provided")
	}

	values := make(map[string]interface{})
	v := reflect.ValueOf(s.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, field := range s.selectedFields {
		fieldName := dialectNamer(field)
		// Skip empty fieldnames
		if fieldName == "" {
			continue
		}
		values[fieldName] = v.FieldByIndex(field.ReflectField().Index).Interface()
	}
	return values, nil
}

func (s *Scope) unselectFields(fields ...*models.StructField) error {
scopeFields:
	for i := 0; i < len(s.selectedFields); i++ {
		if len(fields) == 0 {
			break scopeFields
		}

		for j := 0; j < len(fields); j++ {
			if s.selectedFields[i] == fields[j] {
				fields = append(fields[:j], fields[j+1:]...)
				j--
				// Erease from Selected fields
				s.selectedFields = append(s.selectedFields[:i], s.selectedFields[i+1:]...)
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
		return errors.Newf(class.QuerySelectedFieldsNotSelected, "unselecting non selected fields: '%s'", notEreased)
	}
	return nil
}

/**

FIELDSET

*/

// AddToFieldset adds the fields into the fieldset
func (s *Scope) AddToFieldset(fields ...interface{}) error {
	if s.fieldset == nil {
		s.fieldset = map[string]*models.StructField{}
	}
	return s.addToFieldset(fields...)
}

// BuildFieldset builds the fieldset for the provided scope's string fields in an NeuronName form field1, field2.
func (s *Scope) BuildFieldset(fields ...string) []*errors.Error {
	var (
		errObj *errors.Error
		errs   []*errors.Error
	)
	// check if the length of the fields in the fieldset is not bigger then the fields number for the model.
	if len(fields) > s.mStruct.FieldCount() {
		errObj = errors.New(class.QueryFieldsetTooBig, "too many fields to set for the query")
		errObj.SetDetailf("Too many fields to set for the model: '%s'.", s.mStruct.Collection())
		errs = append(errs, errObj)
		return errs
	}

	prim := s.mStruct.PrimaryField()
	s.fieldset = map[string]*models.StructField{
		prim.Name(): prim,
	}

	for _, field := range fields {
		if field == "" {
			continue
		}

		sField, err := s.checkField(field)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		_, ok := s.fieldset[sField.NeuronName()]
		if ok {
			// duplicate
			errObj = errors.New(class.QueryFieldsetDuplicate, "unknown field provided")
			errObj.SetDetailf("Duplicated fieldset parameter: '%s' for: '%s' collection.", field, s.mStruct.Collection())
			errs = append(errs, errObj)
			if len(errs) > MaxPermissibleDuplicates {
				return errs
			}
			continue
		}
		s.fieldset[sField.NeuronName()] = sField

		if sField.IsRelationship() {
			r := sField.Relationship()
			if r != nil && r.Kind() == models.RelBelongsTo {
				if fk := r.ForeignKey(); fk != nil {
					s.fieldset[fk.NeuronName()] = fk
				}
			}
		}
	}
	return errs

}

// Fieldset returns current scope fieldset
func (s *Scope) Fieldset() []*models.StructField {
	var fs []*models.StructField
	for _, field := range s.fieldset {
		fs = append(fs, field)
	}
	return fs
}

// FieldsetDialectNames gets the fieldset names, named using provided DialetFieldNamer
func (s *Scope) FieldsetDialectNames(dialectNamer dialect.FieldNamer) []string {
	fieldNames := []string{}
	for _, field := range s.fieldset {
		dialectName := dialectNamer(field)
		if dialectName == "" {
			continue
		}
		fieldNames = append(fieldNames, dialectName)
	}
	return fieldNames
}

// FillFieldsetIfNotSet sets the fieldset to full if the fieldset is not set
func (s *Scope) FillFieldsetIfNotSet() {
	if s.fieldset == nil || (s.fieldset != nil && len(s.fieldset) == 0) {
		s.setAllFields()
	}
}

// InFieldset checks if the field is in Fieldset
func (s *Scope) InFieldset(field string) (*models.StructField, bool) {
	f, ok := s.fieldset[field]
	if !ok {
		for _, f := range s.fieldset {
			if f.Name() == field {
				return f, true
			}
		}
		return nil, false
	}
	return f, ok
}

// SetAllFields sets all fields in the fieldset
func (s *Scope) SetAllFields() {
	s.setAllFields()
}

// SetFieldsetNoCheck adds fields to the scope without checking if the fields are correct.
func (s *Scope) SetFieldsetNoCheck(fields ...*models.StructField) {
	for _, field := range fields {
		s.fieldset[field.NeuronName()] = field
	}
}

// SetEmptyFieldset sets the scope's fieldset to nil
func (s *Scope) SetEmptyFieldset() {
	s.fieldset = map[string]*models.StructField{}
}

// SetNilFieldset sets the fieldset to nil
func (s *Scope) SetNilFieldset() {
	s.fieldset = nil
}

// SetFields sets the fieldset from the provided fields
func (s *Scope) SetFields(fields ...interface{}) error {
	s.fieldset = map[string]*models.StructField{}
	s.addToFieldset(fields...)
	return nil
}

func (s *Scope) addToFieldset(fields ...interface{}) error {
	for _, field := range fields {
		var found bool
		switch f := field.(type) {
		case string:
			if "*" == f {
				for _, sField := range s.mStruct.Fields() {
					s.fieldset[sField.NeuronName()] = sField
				}
			} else {
				for _, sField := range s.mStruct.Fields() {
					if sField.NeuronName() == f || sField.Name() == f {
						s.fieldset[sField.NeuronName()] = sField
						found = true
						break
					}
				}
				if !found {
					log.Debugf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
					err := errors.New(class.QueryFieldsetUnknownField, "field not found in the model")
					err.SetDetailf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
					return err
				}
			}
		case *models.StructField:
			for _, sField := range s.mStruct.Fields() {
				if sField == f {
					s.fieldset[sField.NeuronName()] = f
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%v' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				err := errors.New(class.QueryFieldsetUnknownField, "field not found in the model")
				err.SetDetailf("Field: '%s' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				return err
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return errors.Newf(class.QueryFieldsetInvalid, "provided invalid field type: '%T'", f)
		}
	}
	return nil
}

func (s *Scope) setAllFields() {
	fieldset := map[string]*models.StructField{}

	for _, field := range s.mStruct.Fields() {
		fieldset[field.NeuronName()] = field
	}
	s.fieldset = fieldset
}

// IsPrimaryFieldSelected checks if the Scopes primary field is selected
func (s *Scope) IsPrimaryFieldSelected() bool {
	for _, field := range s.selectedFields {
		if field == s.mStruct.PrimaryField() {
			return true
		}
	}
	return false
}
