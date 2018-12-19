package jsonapi

import (
	"fmt"
	"reflect"
)

// NestedStruct is the field StructField that is composed from different abstraction
// then the basic data types.
// It may contain multiple fields *NestedFields
type NestedStruct struct {
	// structField is the reference to it's root structfield
	structField structFielder

	// modelType is the NestedStruct's model type
	modelType reflect.Type

	// fields - nestedStruct may have it's nested fields
	// fields may not be a relationship field
	fields map[string]*NestedField

	// marshal type
	marshalType reflect.Type
}

// MarshalValue marshals the NestedStruct value as it was defined in the controller encoding
func (n *NestedStruct) MarshalValue(v reflect.Value) reflect.Value {

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result := reflect.New(n.marshalType)
	marshalValue := result.Elem()

	for _, nestedField := range n.fields {
		vField := v.FieldByIndex(nestedField.refStruct.Index)
		mField := marshalValue.FieldByIndex(nestedField.refStruct.Index)
		if nestedField.isNestedStruct() {
			mField.Set(nestedField.nested.MarshalValue(vField))
		} else {
			mField.Set(vField)
		}
	}

	return marshalValue
}

func (n *NestedStruct) UnmarshalValue(c *Controller, value interface{}) (reflect.Value, error) {
	mp, ok := value.(map[string]interface{})
	if !ok {
		err := ErrInvalidJSONFieldValue.Copy()
		err.Detail = fmt.Sprintf("Invalid field value for the subfield within attribute: '%s'", n.attr().jsonAPIName)
		return reflect.Value{}, err
	}

	result := reflect.New(n.modelType)
	resElem := result.Elem()
	for mpName, mpVal := range mp {
		nestedField, ok := n.fields[mpName]
		if !ok {
			if !c.StrictUnmarshalMode {
				continue
			}
			err := ErrInvalidJSONFieldValue.Copy()
			err.Detail = fmt.Sprintf("No subfield named: '%s' within attr: '%s'", mpName, n.attr().jsonAPIName)
			return reflect.Value{}, err
		}

		fieldValue := resElem.FieldByIndex(nestedField.refStruct.Index)

		err := c.unmarshalAttrFieldValue(nestedField.StructField, fieldValue, mpVal)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	if n.structField.self().isBasePtr() {
		c.log().Debugf("NestedStruct: '%v' isBasePtr. Attr: '%s'", result.Type(), n.attr().Name())
		return result, nil
	}
	c.log().Debugf("NestedStruct: '%v' isNotBasePtr. Attr: '%s'", resElem.Type(), n.attr().Name())
	return resElem, nil
}

func (n *NestedStruct) attr() *StructField {
	var attr *StructField
	sFielder := n.structField
	for {
		if nested, ok := sFielder.(nestedStructFielder); ok {
			sFielder = nested.selfNested().root.structField
		} else {
			attr = sFielder.self()
			break
		}
	}
	return attr
}

// NestedField is the field within the NestedStruct
type NestedField struct {
	*StructField

	// root defines the NestedField Struct Model
	root *NestedStruct
}

func (n *NestedField) selfNested() *NestedField {
	return n
}

func (n *NestedField) attr() *StructField {
	var attr *StructField
	sFielder := n.root.structField
	for {
		if nested, ok := sFielder.(nestedStructFielder); ok {
			sFielder = nested.selfNested().root.structField
		} else {
			attr = sFielder.self()
			break
		}
	}
	return attr
}

type structFielder interface {
	self() *StructField
}

type nestedStructFielder interface {
	structFielder
	selfNested() *NestedField
}
