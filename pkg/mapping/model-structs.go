package mapping

import (
	"github.com/kucjac/jsonapi/pkg/internal/models"
)

// import (
// 	"fmt"
// 	apiErrors "github.com/kucjac/jsonapi/errors"
// 	"github.com/kucjac/jsonapi/internal"
// 	"github.com/pkg/errors"
// 	"reflect"
// 	"strings"
// )

type ModelStruct struct {
	*models.ModelStruct
}

// Attr returns the attribute for the provided ModelStruct
// If the attribute doesn't exists
func (m *ModelStruct) Attr(attr string) (*StructField, bool) {
	s, ok := models.StructAttr(m.ModelStruct, attr)
	if !ok {
		return nil, ok
	}
	return &StructField{StructField: s}, true
}

// RelationField gets the relationship field for the provided string
// If the relationship field doesn't exists returns nil and false
func (m *ModelStruct) RelationField(rel string) (*StructField, bool) {
	s, ok := models.StructRelField(m.ModelStruct, rel)
	if !ok {
		return nil, ok
	}
	return &StructField{StructField: s}, true
}

// ForeignKey returns model's foreign key field if exists
func (m *ModelStruct) ForeignKey(fk string) (*StructField, bool) {
	s, ok := models.StructForeignKeyField(m.ModelStruct, fk)
	if !ok {
		return nil, ok
	}
	return &StructField{s}, ok
}

// Primary returns model's primary field
func (m *ModelStruct) Primary() *StructField {
	p := models.StructPrimary(m.ModelStruct)
	if p == nil {
		return nil
	}
	return &StructField{p}
}

// FilterKey returns model's filter key if exists
func (m *ModelStruct) FilterKey(fk string) (*StructField, bool) {
	s, ok := models.StructFilterKeyField(m.ModelStruct, fk)
	if !ok {
		return nil, ok
	}
	return &StructField{s}, ok
}

func (m *ModelStruct) FieldByName(name string) (*StructField, bool) {
	field := models.StructFieldByName(m.ModelStruct, name)
	if field == nil {
		return nil, false
	}
	return &StructField{field}, true
}

func (m *ModelStruct) Fields() (fields []*StructField) {
	for _, field := range models.StructAllFields(m.ModelStruct) {
		fields = append(fields, &StructField{field})
	}
	return
}

// LanguageField returns model's language field
func (m *ModelStruct) LanguageField() *StructField {
	lf := models.StructLanguage(m.ModelStruct)
	if lf == nil {
		return nil
	}
	return &StructField{lf}
}
