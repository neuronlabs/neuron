package jsonapi

import (
	"fmt"
	"reflect"
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
	jsonAPIType JSONAPIType

	// Given Field
	refStruct reflect.StructField

	// relatedModelType is a model type for the relationship
	relatedModelType reflect.Type
	relatedStruct    *ModelStruct

	// isListRelated
	isListRelated bool

	omitempty, iso8601, isTime, i18n, isPtrTime, noFilter bool
}

// GetFieldIndex - gets the field index in the given model
func (s *StructField) GetFieldIndex() int {
	return s.getFieldIndex()
}

// GetFieldName - gets the field name for given model
func (s *StructField) GetFieldName() string {
	return s.fieldName
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

// GetJSONAPIType gets the JSONAPIType of the given struct field
func (s *StructField) GetJSONAPIType() JSONAPIType {
	return s.jsonAPIType
}

// IsRelationship checks if given field is a relationship
func (s *StructField) IsRelationship() bool {
	return s.isRelationship()
}

// I18n defines if the field is tagged as internationalization ready.
// This means that the value should be translated
func (s *StructField) I18n() bool {
	return s.i18n
}

func (s *StructField) isRelationship() bool {
	return s.jsonAPIType == RelationshipMultiple || s.jsonAPIType == RelationshipSingle
}

func (s *StructField) canBeSorted() bool {
	switch s.jsonAPIType {
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

func (s *StructField) getFieldIndex() int {
	return s.refStruct.Index[0]
}

func (s *StructField) initCheckFieldType() error {
	fieldType := s.refStruct.Type
	switch s.jsonAPIType {
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
