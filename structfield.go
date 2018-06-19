package jsonapi

import (
	"fmt"
	"reflect"
)

// FieldType is an enum that defines the following field type (i.e. 'primary', 'attribute')
type FieldType int

const (
	UnknownType FieldType = iota
	// Primary is a 'primary' field
	Primary

	// Attribute is an 'attribute' field
	Attribute

	// ClientID is id set by client
	ClientID

	// RelationshipSingle is a 'relationship' with single object
	RelationshipSingle

	// RelationshipMultiple is a 'relationship' with multiple objects
	RelationshipMultiple
)

func (f FieldType) String() string {
	switch f {
	case Primary:
		return "Primary"
	case Attribute:
		return "Attribute"
	case ClientID:
		return "ClientID"
	case RelationshipSingle, RelationshipMultiple:
		return "Relationship"
	}

	return "Unknown"
}

type fieldFlag int

const (
	fDefault   fieldFlag = iota
	fOmitempty fieldFlag = 1 << (iota - 1)
	fIso8601
	fI18n
	fTime
	fPtrTime
	fNoFilter
	fLanguage
	fHidden
	fSortable
)

// StructField represents a field structure with its json api parameters
// and model relationships.
type StructField struct {
	// model is the model struct that this field is part of.
	mStruct *ModelStruct

	// FieldName
	fieldName string

	// JSONAPIName is jsonapi field name - representation for json
	jsonAPIName string

	// fieldType
	fieldType FieldType

	// Given Field
	refStruct reflect.StructField

	// relatedModelType is a model type for the relationship
	relatedModelType reflect.Type
	relatedStruct    *ModelStruct

	// isListRelated
	isListRelated bool

	fieldFlags fieldFlag

	// omitempty, iso8601, isTime, i18n, isPtrTime, noFilter, isLanguage bool
}

// GetFieldIndex - gets the field index in the given model
func (s *StructField) GetFieldIndex() int {
	return s.getFieldIndex()
}

// GetFieldName - gets the field name for given model
func (s *StructField) GetFieldName() string {
	return s.fieldName
}

func (s *StructField) GetJSONAPIName() string {
	return s.jsonAPIName
}

// GetReflectStructField - gets the reflect.StructField for given field.
func (s *StructField) GetReflectStructField() reflect.StructField {
	return s.refStruct
}

// GetFieldType - gets the field's reflect.Type
func (s *StructField) GetFieldType() reflect.Type {
	return s.refStruct.Type
}

// GetRelatedModelType gets the reflect.Type of the related model
// used for relationship fields.
func (s *StructField) GetRelatedModelType() reflect.Type {
	return s.getRelatedModelType()
}

// GetFieldType gets the FieldType of the given struct field
func (s *StructField) GetFieldKind() FieldType {
	return s.fieldType
}

// IsRelationship checks if given field is a relationship
func (s *StructField) IsRelationship() bool {
	return s.isRelationship()
}

// GetRelatedModelStruct gets the ModelStruct of the field Type.
func (s *StructField) GetRelatedModelStruct() *ModelStruct {
	return s.relatedStruct
}

// I18n defines if the field is tagged as internationalization ready.
// This means that the value should be translated
func (s *StructField) I18n() bool {
	return s.isI18n()
}

func (s *StructField) IsPrimary() bool {
	return s.GetFieldKind() == Primary
}

func (s *StructField) isRelationship() bool {
	return s.fieldType == RelationshipMultiple || s.fieldType == RelationshipSingle
}

func (s *StructField) canBeSorted() bool {
	switch s.fieldType {
	case RelationshipSingle, RelationshipMultiple, Attribute:
		return true
	}
	return false
}

func (s *StructField) getRelatedModelType() reflect.Type {
	return s.relatedModelType
}

func (s *StructField) getDereferencedType() reflect.Type {
	var t reflect.Type = s.refStruct.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (s *StructField) getRelationshipPrimariyValues(fieldValue reflect.Value,
) (primaries reflect.Value, err error) {
	if s.fieldType == RelationshipSingle || s.fieldType == RelationshipMultiple {
		return s.relatedStruct.getPrimaryValues(fieldValue)
	}
	// error
	err = fmt.Errorf("Provided field: '%v' is not a relationship: '%s'", s.fieldName, s.fieldType.String())
	return
}

func (s *StructField) getFieldIndex() int {
	return s.refStruct.Index[0]
}

func (s *StructField) initCheckFieldType() error {
	fieldType := s.refStruct.Type
	switch s.fieldType {
	case Primary:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
		default:
			err := fmt.Errorf("Invalid primary field type: %s for the field: %s in model: %s.", fieldType, s.fieldName, s.mStruct.modelType.Name())
			return err
		}
	case Attribute:
		// almost any type
		switch fieldType.Kind() {
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Invalid:
			err := fmt.Errorf("Invalid attribute field type: %v for field: %s in model: %s", fieldType, s.fieldName, s.mStruct.modelType.Name())
			return err
		}
		if s.isLanguage() {
			if fieldType.Kind() != reflect.String {
				err := fmt.Errorf("Incorrect field type: %v for language field. The langtag field must be a string. Model: '%v'", fieldType, s.mStruct.modelType.Name())
				return err
			}
		}

	case RelationshipSingle, RelationshipMultiple:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.Struct:
		case reflect.Slice:
			fieldType = fieldType.Elem()
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() != reflect.Struct {
				err := fmt.Errorf("Invalid slice type: %v, for the relationship: %v", fieldType,
					s.jsonAPIName)
				return err
			}
		default:
			err := fmt.Errorf("Invalid field type: %v, for the relationship.: %v", fieldType, s.jsonAPIName)
			return err
		}
	}
	return nil
}

func (s *StructField) isOmitEmpty() bool {
	return s.fieldFlags&fOmitempty != 0
}

func (s *StructField) isIso8601() bool {
	return s.fieldFlags&fIso8601 != 0
}

func (s *StructField) isTime() bool {
	return s.fieldFlags&fTime != 0
}

func (s *StructField) isPtrTime() bool {
	return s.fieldFlags&fPtrTime != 0
}

func (s *StructField) isI18n() bool {
	return s.fieldFlags&fI18n != 0
}

func (s *StructField) isNoFilter() bool {
	return s.fieldFlags&fNoFilter != 0
}

func (s *StructField) isLanguage() bool {
	return s.fieldFlags&fLanguage != 0
}

func (s *StructField) isHidden() bool {
	return s.fieldFlags&fHidden != 0
}

func (s *StructField) isSortable() bool {
	return s.fieldFlags&fSortable != 0
}
