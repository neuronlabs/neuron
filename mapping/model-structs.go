package mapping

import (
	"github.com/kucjac/jsonapi/internal/models"
	"reflect"
)

// import (
// 	"fmt"
// 	apiErrors "github.com/kucjac/jsonapi/errors"
// 	"github.com/kucjac/jsonapi/internal"
// 	"github.com/pkg/errors"
// 	"reflect"
// 	"strings"
// )

type ModelStruct models.ModelStruct

// Attr returns the attribute for the provided ModelStruct
// If the attribute doesn't exists
func (m *ModelStruct) Attr(attr string) (*StructField, bool) {

	s, ok := models.StructAttr((*models.ModelStruct)(m), attr)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), true
}

// RelationField gets the relationship field for the provided string
// If the relationship field doesn't exists returns nil and false
func (m *ModelStruct) RelationField(rel string) (*StructField, bool) {
	s, ok := models.StructRelField((*models.ModelStruct)(m), rel)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), true
}

// ForeignKey returns model's foreign key field if exists
func (m *ModelStruct) ForeignKey(fk string) (*StructField, bool) {
	s, ok := models.StructForeignKeyField((*models.ModelStruct)(m), fk)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), ok
}

// Primary returns model's primary field
func (m *ModelStruct) Primary() *StructField {
	p := models.StructPrimary((*models.ModelStruct)(m))
	if p == nil {
		return nil
	}
	return (*StructField)(p)
}

// FilterKey returns model's filter key if exists
func (m *ModelStruct) FilterKey(fk string) (*StructField, bool) {
	s, ok := models.StructFilterKeyField((*models.ModelStruct)(m), fk)
	if !ok {
		return nil, ok
	}
	return (*StructField)(s), ok
}

func (m *ModelStruct) FieldByName(name string) (*StructField, bool) {
	field := models.StructFieldByName((*models.ModelStruct)(m), name)
	if field == nil {
		return nil, false
	}
	return (*StructField)(field), true
}

// Field
func (m *ModelStruct) Fields() (fields []*StructField) {
	for _, field := range models.StructAllFields((*models.ModelStruct)(m)) {
		fields = append(fields, (*StructField)(field))
	}
	return
}

// StoreSet sets into the store the value 'value' for given 'key'
func (s *ModelStruct) StoreSet(key string, value interface{}) {
	(*models.ModelStruct)(s).StoreSet(key, value)
}

// StoreGet gets the value from the store at the key: 'key'.
func (s *ModelStruct) StoreGet(key string) (interface{}, bool) {
	return (*models.ModelStruct)(s).StoreGet(key)
}

// StructFields return all struct fields used by the model
func (m *ModelStruct) StructFields() []*StructField {

	// init StructField
	var mFields []*StructField

	fields := (*models.ModelStruct)(m).StructFields()
	for _, f := range fields {
		mFields = append(mFields, (*StructField)(f))
	}

	return mFields

}

// LanguageField returns model's language field
func (m *ModelStruct) LanguageField() *StructField {
	lf := models.StructLanguage((*models.ModelStruct)(m))
	if lf == nil {
		return nil
	}
	return (*StructField)(lf)
}

func (m *ModelStruct) toModels() *models.ModelStruct {
	return (*models.ModelStruct)(m)
}

// Type returns model struct type
func (m *ModelStruct) Type() reflect.Type {
	return (*models.ModelStruct)(m).Type()
}

// Collection returns model's collection
func (m *ModelStruct) Collection() string {
	return (*models.ModelStruct)(m).Collection()
}
