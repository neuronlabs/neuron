package models

import (
	"reflect"
)

// NestedStruct is the field StructField that is composed from different abstraction
// then the basic data types.
// It may contain multiple fields *NestedFields.
type NestedStruct struct {
	// structField is the reference to it's root structfield
	structField StructFielder

	// modelType is the NestedStruct's model type
	modelType reflect.Type

	// fields - nestedStruct may have it's nested fields
	// fields may not be a relationship field
	fields map[string]*NestedField

	// marshal type
	marshalType reflect.Type
}

// Attr returns nested struct related attribute field
func (n *NestedStruct) Attr() *StructField {
	return n.attr()
}

// Type returns nested struct's reflect.Type
func (n *NestedStruct) Type() reflect.Type {
	return n.modelType
}

// Fields return nested fields for the given structure
func (n *NestedStruct) Fields() map[string]*NestedField {
	return n.fields
}

// StructField returns nested structs related struct field
func (n *NestedStruct) StructField() StructFielder {
	return n.structField
}

// NewNestedStruct returns new nested structure
func NewNestedStruct(t reflect.Type, structField StructFielder) *NestedStruct {
	return &NestedStruct{structField: structField, modelType: t, fields: map[string]*NestedField{}}
}

// NestedStructAttr returns related attribute to the provided nested struct
func NestedStructAttr(n *NestedStruct) *StructField {
	return n.attr()
}

// NestedStructSetSubfield sets the subfield for the nestedStructr
func NestedStructSetSubfield(s *NestedStruct, n *NestedField) {
	s.fields[n.structField.neuronName] = n
}

func (n *NestedStruct) attr() *StructField {
	var attr *StructField
	sFielder := n.structField
	for {
		if nested, ok := sFielder.(nestedStructFielder); ok {
			sFielder = nested.SelfNested().root.structField
		} else {
			attr = sFielder.Self()
			break
		}
	}
	return attr
}

// NestedStructType returns the reflect.Type of the nestedStruct
func NestedStructType(n *NestedStruct) reflect.Type {
	return n.modelType
}

// NestedStructSubField returns the NestedStruct subfield if exists.
func NestedStructSubField(n *NestedStruct, field string) (*NestedField, bool) {
	f, ok := n.fields[field]
	if !ok {
		return nil, ok
	}
	return f, ok
}

// NestedStructFields gets the nested struct fields
func NestedStructFields(n *NestedStruct) map[string]*NestedField {
	return n.fields
}

// NestedStructSetMarshalType sets the nested structs marshal type
func NestedStructSetMarshalType(n *NestedStruct, mType reflect.Type) {
	n.marshalType = mType
}

// NestedStructMarshalType returns the marshal type for the provided nested struct
func NestedStructMarshalType(n *NestedStruct) reflect.Type {
	return n.marshalType
}

// NestedField is the field within the NestedStruct
type NestedField struct {
	structField *StructField

	// root defines the NestedField Struct Model
	root *NestedStruct
}

// NewNestedField returns New NestedField
func NewNestedField(
	root *NestedStruct,
	structFielder StructFielder,
	nField reflect.StructField,
) *NestedField {
	nestedField := &NestedField{
		structField: &StructField{
			mStruct:      structFielder.Self().mStruct,
			reflectField: nField,
			fieldKind:    KindNested,
			fieldFlags:   FDefault | FNestedField,
		},
		root: root,
	}
	return nestedField
}

// NestedFieldAttr returns nested Fields Attribute
func NestedFieldAttr(n *NestedField) *StructField {
	return n.attr()
}

// NestedFieldRoot returns the root of the NestedField
func NestedFieldRoot(n *NestedField) *NestedStruct {
	return n.root
}

// StructField returns the structField
func (n *NestedField) StructField() *StructField {
	return n.structField
}

// Self is the relation to it's struct field
func (n *NestedField) Self() *StructField {
	return n.structField.Self()
}

// SelfNested returns the pointer to itself
func (n *NestedField) SelfNested() *NestedField {
	return n
}

func (n *NestedField) attr() *StructField {
	var attr *StructField
	sFielder := n.root.structField
	for {
		if nested, ok := sFielder.(nestedStructFielder); ok {
			sFielder = nested.SelfNested().root.structField
		} else {
			attr = sFielder.Self()
			break
		}
	}
	return attr
}

// StructFielder is the interfaces used for getting the pointer to itself
type StructFielder interface {
	Self() *StructField
}

type nestedStructFielder interface {
	StructFielder
	SelfNested() *NestedField
}
