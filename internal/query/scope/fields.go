package scope

import (
	"fmt"
	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/namer/dialect"
	"github.com/neuronlabs/neuron/log"
	"github.com/pkg/errors"
	"reflect"
)

/**

SELECTED FIELDS

*/

// AddselectedFields adds provided fields into given Scope's selectedFields Container
func AddselectedFields(s *Scope, fields ...string) error {
	for _, addField := range fields {

		field := models.StructFieldByName(s.mStruct, addField)
		if field == nil {
			return errors.Errorf("Field: '%s' not found within model: %s", addField, s.mStruct.Collection())
		}
		s.selectedFields = append(s.selectedFields, field)

	}
	return nil
}

// AddToSelectedFields adds the fields to the scope's selected fields
func (s *Scope) AddToSelectedFields(fields ...interface{}) error {
	return s.addToSelectedFields(fields...)
}

// AddSelectedField adds the selected field to the selected field's array
func (s *Scope) AddSelectedField(field *models.StructField) {
	s.selectedFields = append(s.selectedFields, field)
}

// AutoSelectFields selects the fields automatically if none of the select field method were called
func (s *Scope) AutoSelectFields() error {
	if s.selectedFields != nil {
		return nil
	}

	if s.Value == nil {
		return ErrNoValue
	}

	v := reflect.ValueOf(s.Value).Elem()

	// check if the value is a struct
	if v.Kind() != reflect.Struct {
		return internal.ErrInvalidType
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

// UnselectFields unselects provided fields
func (s *Scope) UnselectFields(fields ...*models.StructField) error {
	return unselectFields(s, fields...)
}

// DeleteselectedFields deletes the models.StructFields from the given scope Fieldset
func DeleteselectedFields(s *Scope, fields ...*models.StructField) error {
	return unselectFields(s, fields...)
}

func unselectFields(s *Scope, fields ...*models.StructField) error {

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
			notEreased += field.ApiName() + " "
		}
		return errors.Errorf("The following fields were not in the Selected fields scope: '%v'", notEreased)
	}

	return nil
}

// IsPrimaryFieldSelected checks if the Scopes primary field is selected
func (s *Scope) IsPrimaryFieldSelected() bool {
	for _, field := range s.selectedFields {
		if field == models.StructPrimary(s.mStruct) {
			return true
		}
	}
	return false
}

// NotSelectedFields lists all the fields that are not selected within the scope
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

// SelectedFields return fields that were selected during unmarshaling
func (s *Scope) SelectedFields() []*models.StructField {
	return s.selectedFields
}

func (s *Scope) addToSelectedFields(fields ...interface{}) error {
	var selectedFields map[*models.StructField]struct{} = make(map[*models.StructField]struct{})

	var sfields []*models.StructField

	for _, field := range fields {
		var found bool
		switch f := field.(type) {
		case string:

			for _, sField := range models.StructAllFields(s.mStruct) {
				if sField.ApiName() == f || sField.Name() == f {
					selectedFields[sField] = struct{}{}
					sfields = append(sfields, sField)
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
				return internal.ErrFieldNotFound
			}

		case *models.StructField:
			for _, sField := range models.StructAllFields(s.mStruct) {
				if sField == f {
					found = true
					selectedFields[sField] = struct{}{}
					sfields = append(sfields, sField)
					break
				}
			}
			if !found {
				log.Debugf("Field: '%v' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				return internal.ErrFieldNotFound
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return internal.ErrInvalidType
		}
	}

	// check if fields were not already selected
	for _, alreadySelected := range s.selectedFields {
		_, ok := selectedFields[alreadySelected]
		if ok {
			log.Errorf("Field: %s already set for the given scope.", alreadySelected.Name())
			return internal.ErrFieldAlreadySelected
		}
	}

	// add all fields to scope's selected fields
	s.selectedFields = append(s.selectedFields, sfields...)

	return nil
}

func (s *Scope) deleteSelectedField(index int) error {
	if index > len(s.selectedFields)-1 {
		return errors.Errorf("Index out of range: %v", index)
	}

	// copy the last element
	if index < len(s.selectedFields)-1 {
		s.selectedFields[index] = s.selectedFields[len(s.selectedFields)-1]
		s.selectedFields[len(s.selectedFields)-1] = nil
	}
	s.selectedFields = s.selectedFields[:len(s.selectedFields)-1]
	return nil
}

// selectedFieldValues gets the values for given
func (s *Scope) selectedFieldValues(dialectNamer dialect.FieldNamer) (
	values map[string]interface{}, err error,
) {
	if s.Value == nil {
		err = internal.ErrNilValue
		return
	}

	values = map[string]interface{}{}

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
	return
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

// BuildFieldset builds the fieldset for the provided scope fields[collection] = field1, field2
func (s *Scope) BuildFieldset(fields ...string) (errs []*aerrors.ApiError) {
	var (
		errObj *aerrors.ApiError
	)

	if len(fields) > s.mStruct.FieldCount() {
		errObj = aerrors.ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Too many fields to set.")
		errs = append(errs, errObj)
		return
	}

	prim := models.StructPrimary(s.mStruct)
	s.fieldset = map[string]*models.StructField{
		prim.Name(): prim,
	}

	for _, field := range fields {
		if field == "" {
			continue
		}

		sField, err := s.checkField(field)
		if err != nil {
			if field == "id" {
				err = aerrors.ErrInvalidQueryParameter.Copy()
				err.Detail = "Invalid fields parameter. 'id' is not a field - it is primary key."
			}
			errs = append(errs, err)
			continue
		}

		_, ok := s.fieldset[sField.ApiName()]
		if ok {
			// duplicate
			errObj = aerrors.ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Duplicated fieldset parameter: '%s' for: '%s' collection.", field, s.mStruct.Collection())
			errs = append(errs, errObj)
			if len(errs) > MaxPermissibleDuplicates {
				return
			}
			continue
		}
		s.fieldset[sField.ApiName()] = sField

		if sField.IsRelationship() {

			r := models.FieldRelationship(sField)
			if r != nil && models.RelationshipGetKind(r) == models.RelBelongsTo {
				if fk := models.RelationshipForeignKey(r); fk != nil {
					s.fieldset[fk.ApiName()] = fk
				}
			}
		}
	}

	return

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
	return f, ok
}

// SetAllFields sets all fields in the fieldset
func (s *Scope) SetAllFields() {
	s.setAllFields()
}

// SetFieldsetNoCheck adds fields to the scope without checking if the fields are correct.
func (s *Scope) SetFieldsetNoCheck(fields ...*models.StructField) {
	for _, field := range fields {
		s.fieldset[field.ApiName()] = field
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

func (s *Scope) setAllFields() {
	fieldset := map[string]*models.StructField{}

	for _, field := range models.StructAllFields(s.mStruct) {
		fieldset[field.ApiName()] = field
	}
	s.fieldset = fieldset
}

func (s *Scope) addToFieldset(fields ...interface{}) error {
	for _, field := range fields {
		var found bool
		switch f := field.(type) {
		case string:

			if "*" == f {
				for _, sField := range s.mStruct.Fields() {
					s.fieldset[sField.ApiName()] = sField
				}
			} else {
				for _, sField := range models.StructAllFields(s.mStruct) {
					if sField.ApiName() == f || sField.Name() == f {
						s.fieldset[sField.ApiName()] = sField
						found = true
						break
					}
				}
				if !found {
					log.Debugf("Field: '%s' not found for model:'%s'", f, s.mStruct.Type().Name())
					return internal.ErrFieldNotFound
				}
			}

		case *models.StructField:
			for _, sField := range models.StructAllFields(s.mStruct) {
				if sField == f {
					s.fieldset[sField.ApiName()] = f
					found = true
					break
				}
			}
			if !found {
				log.Debugf("Field: '%v' not found for model:'%s'", f.Name(), s.mStruct.Type().Name())
				return internal.ErrFieldNotFound
			}
		default:
			log.Debugf("Unknown field type: %v", reflect.TypeOf(f))
			return internal.ErrInvalidType
		}
	}
	return nil
}
