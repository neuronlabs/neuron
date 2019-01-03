package models

import (
	"fmt"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/pkg/errors"
	"net/url"
	"reflect"
	"strings"
)

// FieldKind is an enum that defines the following field type (i.e. 'primary', 'attribute')
type FieldKind int

const (
	UnknownType FieldKind = iota
	// Primary is a 'primary' field
	KindPrimary

	// Attribute is an 'attribute' field
	KindAttribute

	// ClientID is id set by client
	KindClientID

	// RelationshipSingle is a 'relationship' with single object
	KindRelationshipSingle

	// RelationshipMultiple is a 'relationship' with multiple objects
	KindRelationshipMultiple

	// ForeignKey is the field type that is responsible for the relationships
	KindForeignKey

	// FilterKey is the field that is used only for special case filtering
	KindFilterKey

	KindNested
)

func (f FieldKind) String() string {
	switch f {
	case KindPrimary:
		return "Primary"
	case KindAttribute:
		return "Attribute"
	case KindClientID:
		return "ClientID"
	case KindRelationshipSingle, KindRelationshipMultiple:
		return "Relationship"
	case KindForeignKey:
		return "ForeignKey"
	case KindFilterKey:
		return "FilterKey"
	case KindNested:
		return "Nested"
	}

	return "Unknown"
}

type fieldFlag int

const (
	FDefault   fieldFlag = iota
	FOmitempty fieldFlag = 1 << (iota - 1)
	FIso8601
	FI18n
	FNoFilter
	FLanguage
	FHidden
	FSortable
	FClientID

	// field type
	FTime
	FMap
	FPtr
	FArray
	FSlice
	FBasePtr

	FNestedStruct
	FNestedField
)

// StructField represents a field structure with its json api parameters
// and model relationships.
type StructField struct {
	// model is the model struct that this field is part of.
	mStruct *ModelStruct

	// ApiName is jsonapi field name - representation for json
	apiName string

	// fieldKind
	fieldKind FieldKind

	reflectField reflect.StructField

	// isListRelated
	isListRelated bool

	relationship *Relationship

	// nested is the NestedStruct definition
	nested *NestedStruct

	fieldFlags fieldFlag
}

func NewStructField(
	refField reflect.StructField,
	mStruct *ModelStruct,
) *StructField {
	return &StructField{reflectField: refField, mStruct: mStruct}

}

// ApiName returns the structFields ApiName
func (s *StructField) ApiName() string {
	return s.apiName
}

// GetFieldIndex - gets the field index in the given model
func (s *StructField) FieldIndex() int {
	return s.getFieldIndex()
}

// FieldKind returns structFields kind
func (s *StructField) FieldKind() FieldKind {
	return s.fieldKind
}

// FieldName returns struct fields name
func (s *StructField) FieldName() string {
	return s.reflectField.Name
}

// FieldType returns field's reflect.Type
func (s *StructField) FieldType() reflect.Type {
	return s.reflectField.Type
}

// Nested returns nested field's structure
func (s *StructField) Nested() *NestedStruct {
	return s.nested
}

// ReflectField returns structs reflect.StructField
func (s *StructField) ReflectField() reflect.StructField {
	return s.reflectField
}

// Relationship returns StructField's Relationships
func (s *StructField) Relationship() *Relationship {
	return s.relationship
}

// SetRelationship sets the relationship value for the struct field
func (s *StructField) SetRelationship(rel *Relationship) {
	s.relationship = rel
}

// Struct returns fields modelstruct
func (s *StructField) Struct() *ModelStruct {
	return s.mStruct
}

// FieldRelationship
func FieldRelationship(s *StructField) *Relationship {
	return s.relationship
}

// FieldSetRelatinoship sets the relationship for the given field
func FieldSetRelationship(s *StructField, r *Relationship) {
	s.relationship = r
}

// Name returns the StructFields Golang Name
func (s *StructField) Name() string {
	return s.reflectField.Name
}

// IsRelationship checks if given field is a relationship
func (s *StructField) IsRelationship() bool {
	return s.isRelationship()
}

// FieldsStruct returns field's modelStruct
func FieldsStruct(s *StructField) *ModelStruct {
	return s.mStruct
}

// FieldIsZeroValue checks if the provided field has Zero value.
func FieldIsZeroValue(s *StructField, fieldValue interface{}) bool {
	return reflect.DeepEqual(fieldValue, reflect.Zero(s.reflectField.Type).Interface())
}

// FieldRelatedModelType gets the relationship's model type
// Returns nil if the structfield is not a relationship
func FieldsRelatedModelType(s *StructField) reflect.Type {
	return s.getRelatedModelType()
}

// FieldsRelatedModelStruct gets the ModelStruct of the field Type.
// Returns nil if the structField is not a relationship
func FieldsRelatedModelStruct(s *StructField) *ModelStruct {
	if s.relationship == nil {
		return nil
	}
	return s.relationship.mStruct
}

// FieldInitCheckFieldType initializes StructField type
func FieldInitCheckFieldType(s *StructField) error {
	return s.initCheckFieldType()
}

// // IsMap checks if given field is of type map
// func (s *StructField) IsMap() bool {
// 	return s.isMap()
// }

// IsPrimary checks if the field is the primary field type
func (s *StructField) IsPrimary() bool {
	return s.fieldKind == KindPrimary
}

// FieldBaseType returns the base 'reflect.Type' for the provided field.
// The base is the lowest possible dereference of the field's type.
func FieldBaseType(s *StructField) reflect.Type {
	return s.baseFieldType()
}

// FieldsNested gets NestedStruct in the field
func FieldsNested(s *StructField) *NestedStruct {
	return s.nested
}

// FieldSetNested sets the structFields nested field
func FieldSetNested(s *StructField, nested *NestedStruct) {
	s.nested = nested
}

// FieldsSetApiName sets the field's apiName
func FieldsSetApiName(s *StructField, apiName string) {
	s.apiName = apiName
}

// FieldSetFlag sets the provided flag for the structField
func FieldSetFlag(s *StructField, flag fieldFlag) {
	s.fieldFlags = s.fieldFlags | flag
}

// FeildSetFieldKind sets the field's kind
func FieldSetFieldKind(s *StructField, fieldKind FieldKind) {
	s.fieldKind = fieldKind
}

// FieldSetRelatedType sets the related type for the provided structField
func FieldSetRelatedType(sField *StructField) error {

	modelType := sField.reflectField.Type
	// get error function
	getError := func() error {
		return fmt.Errorf("Incorrect relationship type provided. The Only allowable types are structs, pointers or slices. This type is: %v", modelType)
	}

	switch modelType.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Struct:
	default:
		err := getError()
		return err
	}
	for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
		if modelType.Kind() == reflect.Slice {
			sField.fieldKind = KindRelationshipMultiple
		}
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		err := getError()
		return err
	}

	if sField.fieldKind == UnknownType {
		sField.fieldKind = KindRelationshipSingle
	}
	if sField.relationship == nil {
		sField.relationship = &Relationship{}
	}

	sField.relationship.modelType = modelType

	return nil
}

func (s *StructField) SetRelatedModel(relModel *ModelStruct) {
	if s.relationship == nil {
		s.relationship = &Relationship{
			modelType: relModel.Type(),
		}
	}

	s.relationship.mStruct = relModel
}

// // CanBeSorted returns if the struct field can be sorted
// func (s *StructField) CanBeSorted() bool {
// 	return s.canBeSorted()
// }

// FieldTagValues gets field's tag values
func FieldTagValues(s *StructField, tag string) (url.Values, error) {
	return s.getTagValues(tag)
}

// TagValues returns the url.Values for the specific tag
func (s *StructField) TagValues(tag string) (url.Values, error) {
	return s.getTagValues(tag)
}

func (s *StructField) isRelationship() bool {
	return s.fieldKind == KindRelationshipMultiple || s.fieldKind == KindRelationshipSingle
}

// baseFieldType is the field's base dereferenced type
func (s *StructField) baseFieldType() reflect.Type {
	var elem reflect.Type = s.reflectField.Type

	for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Slice ||
		elem.Kind() == reflect.Array || elem.Kind() == reflect.Map {

		elem = elem.Elem()
	}
	return elem
}

func (s *StructField) canBeSorted() bool {
	switch s.fieldKind {
	case KindRelationshipSingle, KindRelationshipMultiple, KindAttribute:
		return true
	}
	return false
}

func (s *StructField) getRelatedModelType() reflect.Type {
	if s.relationship == nil {
		return nil
	}
	return s.relationship.modelType
}

// GetDereferencedType returns structField dereferenced type
func (s *StructField) GetDereferencedType() reflect.Type {
	return s.getDereferencedType()
}

func (s *StructField) getDereferencedType() reflect.Type {
	var t reflect.Type = s.reflectField.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (s *StructField) getFieldIndex() int {
	return s.reflectField.Index[0]
}

func (s *StructField) getTagValues(tag string) (url.Values, error) {
	mp := url.Values{}
	seperated := strings.Split(tag, internal.AnnotationTagSeperator)
	for _, option := range seperated {
		i := strings.IndexRune(option, internal.AnnotationTagEqual)
		if i == -1 {
			return nil, errors.Errorf("No annotation tag equal found for tag value: %v", option)
		}
		key := option[:i]
		var values []string
		if i != len(option)-1 {
			values = strings.Split(option[i+1:], internal.AnnotationSeperator)
		}
		mp[key] = values
	}
	return mp, nil
}

func (s *StructField) initCheckFieldType() error {
	fieldType := s.reflectField.Type
	switch s.fieldKind {
	case KindPrimary:
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		switch fieldType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
		default:
			err := fmt.Errorf("Invalid primary field type: %s for the field: %s in model: %s.", fieldType, s.fieldName(), s.mStruct.modelType.Name())
			return err
		}
	case KindAttribute:
		// almost any type
		switch fieldType.Kind() {
		case reflect.Interface, reflect.Chan, reflect.Func, reflect.Invalid:
			err := fmt.Errorf("Invalid attribute field type: %v for field: %s in model: %s", fieldType, s.Name(), s.mStruct.modelType.Name())
			return err
		}
		if s.isLanguage() {
			if fieldType.Kind() != reflect.String {
				err := fmt.Errorf("Incorrect field type: %v for language field. The langtag field must be a string. Model: '%v'", fieldType, s.mStruct.modelType.Name())
				return err
			}
		}

	case KindRelationshipSingle, KindRelationshipMultiple:
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
					s.apiName)
				return err
			}
		default:
			err := fmt.Errorf("Invalid field type: %v, for the relationship.: %v", fieldType, s.apiName)
			return err
		}
	}
	return nil
}

// FieldAllowClientID checks if the given field allow ClientID
func FieldAllowClientID(s *StructField) bool {
	return s.allowClientID()
}

func (s *StructField) allowClientID() bool {
	return s.fieldFlags&FClientID != 0
}

func (s *StructField) fieldName() string {
	return s.reflectField.Name
}

// FieldIsOmitEmpty checks wether the field uses OmitEmpty flag
func FieldIsOmitEmpty(s *StructField) bool {
	return s.isOmitEmpty()
}

// IsOmitEmpty checks if the given field has a omitempty flag
func (s *StructField) IsOmitEmpty() bool {
	return s.isOmitEmpty()
}

func (s *StructField) isOmitEmpty() bool {
	return s.fieldFlags&FOmitempty != 0
}

// FieldIsIso8601 checks wether the field uses FIso8601 flag
func FieldIsIso8601(s *StructField) bool {
	return s.isIso8601()
}

// IsIso8601 checks wether the field uses FIso8601 flag
func (s *StructField) IsIso8601() bool {
	return s.isIso8601()
}

func (s *StructField) isIso8601() bool {
	return s.fieldFlags&FIso8601 != 0
}

// FieldIsTime checks wether the field uses time flag
func FieldIsTime(s *StructField) bool {
	return s.isTime()
}

// IsTime checks wether the field uses time flag
func (s *StructField) IsTime() bool {
	return s.isTime()
}

func (s *StructField) isTime() bool {
	return s.fieldFlags&FTime != 0
}

// FieldIsPtrTime checks wether the field is a base ptr time flag
func FieldIsPtrTime(s *StructField) bool {
	return s.isPtrTime()
}

func (s *StructField) isPtrTime() bool {
	return (s.fieldFlags&FTime != 0) && (s.fieldFlags&FPtr != 0)
}

// FieldIsI18n checks wether the field has a i18n flag
func FieldIsI18n(s *StructField) bool {
	return s.isI18n()
}

// IsI18n returns flag if the struct fields is an i18n field
func (s *StructField) IsI18n() bool {
	return s.isI18n()
}

func (s *StructField) isI18n() bool {
	return s.fieldFlags&FI18n != 0
}

// FieldIsNoFilter checks wether the field uses no filter flag
func FieldIsNoFilter(s *StructField) bool {
	return s.isNoFilter()
}
func (s *StructField) isNoFilter() bool {
	return s.fieldFlags&FNoFilter != 0
}

// FieldIsFlag checks wether the field has a language flag
func FieldIsFlag(s *StructField) bool {
	return s.isLanguage()
}

// IsLangugage checks wether the field is a language type field
func (s *StructField) IsLanguage() bool {
	return s.isLanguage()
}

func (s *StructField) isLanguage() bool {
	return s.fieldFlags&FLanguage != 0
}

// FieldIsHidden checks if the field has a hidden flag
func FieldIsHidden(s *StructField) bool {
	return s.isHidden()
}

func (s *StructField) isHidden() bool {
	return s.fieldFlags&FHidden != 0
}

// IsHidden checks if the field has a hidden flag
func (s *StructField) IsHidden() bool {
	return s.isHidden()
}

// FieldIsSortable checks if the field has a sortable flag
func FieldIsSortable(s *StructField) bool {
	return s.isSortable()
}

func (s *StructField) isSortable() bool {
	return s.fieldFlags&FSortable != 0
}

// FieldIsMap checks if the field has a fMap flag
func FieldIsMap(s *StructField) bool {
	return s.isMap()
}

// IsMap checks if the field has a fMap flag
func (s *StructField) IsMap() bool {
	return s.isMap()
}

func (s *StructField) isMap() bool {
	return s.fieldFlags&FMap != 0
}

// FieldIsSlice checks if the field is a slice based
func FieldIsSlice(s *StructField) bool {
	return s.isSlice()
}

// IsSlice checks if the field is a slice based
func (s *StructField) IsSlice() bool {
	return s.isSlice()
}

func (s *StructField) isSlice() bool {
	return s.fieldFlags&FSlice != 0
}

// FieldIsArray checks if the field is an array
func FieldIsArray(s *StructField) bool {
	return s.isArray()
}

// IsArray checks if the field is an array
func (s *StructField) IsArray() bool {
	return s.isArray()
}

func (s *StructField) isArray() bool {
	return s.fieldFlags&FArray != 0
}

// FieldIsPtr checks if the field is a pointer
func FieldIsPtr(s *StructField) bool {
	return s.isPtr()
}

// IsPtr checks if the field is a pointer
func (s *StructField) IsPtr() bool {
	return s.isPtr()
}

func (s *StructField) isPtr() bool {
	return s.fieldFlags&FPtr != 0
}

// FieldIsBasePtr checks if the field has a pointer type in the base
func FieldIsBasePtr(s *StructField) bool {
	return s.isBasePtr()
}

// IsBasePtr checks if the field has a pointer type in the base
func (s *StructField) IsBasePtr() bool {
	return s.isBasePtr()
}

func (s *StructField) isBasePtr() bool {
	return s.fieldFlags&FBasePtr != 0
}

// FieldIsNestedStruct checks if the field is a nested structure
func FieldIsNestedStruct(s *StructField) bool {
	return s.isNestedStruct()
}

//IsNestedStruct checks if the field is a nested structure
func (s *StructField) IsNestedStruct() bool {
	return s.nested != nil
}

func (s *StructField) isNestedStruct() bool {
	return s.nested != nil
}

// FieldIsNestedField checks if the field is not defined within ModelStruct
func FieldIsNestedField(s *StructField) bool {
	return s.isNestedField()
}

// isNested means that the struct field is not defined within the ModelStruct
func (s *StructField) isNestedField() bool {
	return s.fieldFlags&FNestedField != 0
}

func (s *StructField) Self() *StructField {
	return s
}
