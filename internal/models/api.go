package models

import (
	"github.com/kucjac/jsonapi/flags"
	"reflect"
)

// StructAllFields return all model's fields
func StructAllFields(m *ModelStruct) []*StructField {
	return m.fields
}

// StructAppendField appends the field to the fieldset for the given model struct
func StructAppendField(m *ModelStruct, field *StructField) {
	m.fields = append(m.fields, field)
}

// StructAttr returns attribute for the provided structField
func StructAttr(m *ModelStruct, attr string) (*StructField, bool) {
	s, ok := m.attributes[attr]
	return s, ok
}

// StructCollectionUrlIndex returns index for the collectionUrl
func StructCollectionUrlIndex(m *ModelStruct) int {
	return m.collectionURLIndex
}

// StructFieldByName returns field for provided name
// It matches both name and apiName
func StructFieldByName(m *ModelStruct, name string) *StructField {
	for _, field := range m.fields {
		if field.apiName == name || field.Name() == name {
			return field
		}
	}
	return nil
}

// Flags return preset flags for the given model
func StructFlags(m *ModelStruct) *flags.Container {
	return m.flags
}

// StructSetAttr sets the attribute field for the provided model
func StructSetAttr(m *ModelStruct, attr *StructField) {
	m.attributes[attr.apiName] = attr
}

// StructSetRelField sets the relationship field for the model
func StructSetRelField(m *ModelStruct, relField *StructField) {
	m.relationships[relField.apiName] = relField
}

// StructSetForeignKey sets the foreign key for the model struct
func StructSetForeignKey(m *ModelStruct, fk *StructField) {
	m.foreignKeys[fk.apiName] = fk
}

// StructSetForeignKey sets the filter key for the model struct
func StructSetFilterKey(m *ModelStruct, fk *StructField) {
	m.filterKeys[fk.apiName] = fk
}

// StructSetPrimary
func StructSetPrimary(m *ModelStruct, primary *StructField) {
	m.primary = primary
}

// StructSetType
func StructSetType(m *ModelStruct, tp reflect.Type) {
	m.modelType = tp
}

// StructIsEqual checks if ModelStructs are equal
func StructIsEqual(first, second *ModelStruct) bool {
	return first.id == second.id
}

// StructSetUrl sets the ModelStruct's url value
func StructSetUrl(m *ModelStruct, url string) error {
	return m.setModelURL(url)
}

// StructPrimary gets the primary field for the given model
func StructPrimary(m *ModelStruct) *StructField {
	return m.primary
}

// StructI18n returns i18n struct field for given model.
func StructAppendI18n(m *ModelStruct, f *StructField) {
	m.i18n = append(m.i18n, f)
}

// StructSetLanguage sets the language field for given model
func StructSetLanguage(m *ModelStruct, f *StructField) {
	f.fieldFlags = f.fieldFlags | FLanguage
	m.language = f
}

func StructSetBelongsToForeigns(m *ModelStruct, v reflect.Value) error {
	return m.setBelongsToForeigns(v)
}

// StructRelField returns ModelsStruct relationship field if exists
func StructRelField(m *ModelStruct, relField string) (*StructField, bool) {
	s, ok := m.relationships[relField]
	return s, ok
}

// StructForeignKeyField returns ModelStruct foreign key field if exists
func StructForeignKeyField(m *ModelStruct, fk string) (*StructField, bool) {
	s, ok := m.foreignKeys[fk]
	if !ok {
		// If no APIname provided, check if this isn't the struct field name
		for _, field := range m.foreignKeys {
			if field.Name() == fk {
				return field, true
			}
		}
	}
	return s, ok
}

// StructFilterKeyField returns ModelStruct filterKey
func StructFilterKeyField(m *ModelStruct, fk string) (*StructField, bool) {
	s, ok := m.filterKeys[fk]
	return s, ok
}

// StructSetBelongsToRelationWithFields sets the value to the Relationship value
func StructSetBelongsToRelationWithFields(m *ModelStruct, v reflect.Value, fields ...*StructField) error {
	return m.setBelongsToRelationWithFields(v, fields...)
}
