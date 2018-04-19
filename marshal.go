package jsonapi

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func MarshalScope(scope *Scope) error {

	return nil
}

func marshalScopeOne(scope *Scope) (*OnePayload, error) {

	return nil, nil
}

func marshalScopeMany(scope *Scope) (*ManyPayload, error) {
	return nil, nil
}

func visitScopeNode(scope *Scope) (*Node, error) {
	node := &Node{Type: scope.Struct.collectionType}
	modelVal := reflect.ValueOf(scope.Value).Elem()

	for _, field := range scope.Fields {

		fieldValue := modelVal.Field(field.getFieldIndex())

		switch field.jsonAPIType {
		case Primary:
			err := setNodePrimary(fieldValue, node, field)
			if err != nil {
				return err
			}
		case Attribute:
			if node.Attributes == nil {
				node.Attributes = make(map[string]interface{})
			}

			if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				t := fieldValue.Interface().(time.Time)

				if t.IsZero() {
					continue
				}

				if field.iso8601 {
					node.Attributes[field.jsonAPIName] = t.UTC().Format(iso8601TimeFormat)
				} else {
					node.Attributes[field.jsonAPIName] = t.Unix()
				}
			} else if fieldValue.Type() == reflect.TypeOf(new(time.Time)) {
				if fieldValue.IsNil() {
					if field.omitempty {
						continue
					}
					node.Attributes[field.jsonAPIName] = nil
				} else {
					t := fieldValue.Interface().(*time.Time)

					if t.IsZero() && field.omitempty {
						continue
					}

					if field.iso8601 {
						node.Attributes[field.jsonAPIName] = t.UTC().Format(iso8601TimeFormat)
					} else {
						node.Attributes[field.jsonAPIName] = t.Unix()
					}
				}
			} else {
				emptyValue := reflect.Zero(fieldValue.Type())
				if field.omitempty && reflect.
					DeepEqual(fieldValue.Interface(), emptyValue.Interface()) {
					continue
				}

				strAttr, ok := fieldValue.Interface().(string)
				if ok {
					node.Attributes[field.jsonAPIName] = strAttr
				} else {
					node.Attributes[field.jsonAPIName] = fieldValue.Interface()
				}
			}
		case ClientID:
			clientID := fieldValue.String()
			if clientID != "" {
				node.ClientID = clientID
			}
		case RelationshipMultiple, RelationshipSingle:
			var isSlice bool = field.jsonAPIType == RelationshipMultiple
			if field.omitempty &&
				(isSlice && fieldValue.Len() < 1 || !isSlice && fieldValue.IsNil()) {
				continue
			}
			// how to handle links?
			var relLinks *Links
			if linkableModel, ok := scope.Value.(RelationshipLinkable); ok {
				relLinks = linkableModel.JSONAPIRelationshipLinks(field.jsonAPIName)
			}

			var relMeta *Meta
			if metableModel, ok := scope.Value.(Metable); ok {
				relMeta = metableModel.JSONAPIMeta()
			}
			if isSlice {
				// get RelationshipManyNode
				relationship, err := visitRelationshipManyNode(fieldValue, field)
				if err != nil {
					return
				}
				relationship.Links = relLinks
				relationship.Meta = relMeta
				node.Relationships[field.jsonAPIName] = relationship
			} else {
				// is to-one relationship
				if fieldValue.IsNil() {
					node.Relationships[field.jsonAPIName] = &RelationshipOneNode{Data: nil}
					continue
				}
				relatedNode, err := visitRelationshipNode(fieldValue, field)
				if err != nil {
					return
				}
				relationship := &RelationshipOneNode{
					Data:  relatedNode,
					Links: relLinks,
					Meta:  relMeta,
				}
				node.Relationships[field.jsonAPIName] = relationship
			}
		}
	}
	return node, err
}

func visitRelationshipManyNode(manyValue reflect.Value, field *StructField,
) (*RelationshipManyNode, error) {
	nodes := []*Node{}

	for i := 0; i < manyValue.Len(); i++ {
		elemValue := manyValue.Index(i)

		node, err := visitRelationshipNode(elemValue, field)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)

	}
	return &RelationshipManyNode{Data: nodes}, nil
}

func visitRelationshipNode(value reflect.Value, field *StructField) (*Node, error) {
	mStruct := field.mStruct
	prim := mStruct.primary
	node := &Node{Type: mStruct.collectionType}

	index := prim.getFieldIndex()

	nodeValue := value.Index(index)
	err := setNodePrimary(nodeValue, node, field)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func setNodePrimary(value reflect.Value, node *Node, field *StructField) error {
	v := value
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	switch field.getDereferencedType().Kind() {
	case reflect.String:
		node.ID = v.Interface().(string)
	case reflect.Int:
		node.ID = strconv.FormatInt(int64(v.Interface().(int)), 10)
	case reflect.Int8:
		node.ID = strconv.FormatInt(int64(v.Interface().(int8)), 10)
	case reflect.Int16:
		node.ID = strconv.FormatInt(int64(v.Interface().(int16)), 10)
	case reflect.Int32:
		node.ID = strconv.FormatInt(int64(v.Interface().(int32)), 10)
	case reflect.Int64:
		node.ID = strconv.FormatInt(v.Interface().(int64), 10)
	case reflect.Uint:
		node.ID = strconv.FormatUint(uint64(v.Interface().(uint)), 10)
	case reflect.Uint8:
		node.ID = strconv.FormatUint(uint64(v.Interface().(uint8)), 10)
	case reflect.Uint16:
		node.ID = strconv.FormatUint(uint64(v.Interface().(uint16)), 10)
	case reflect.Uint32:
		node.ID = strconv.FormatUint(uint64(v.Interface().(uint32)), 10)
	case reflect.Uint64:
		node.ID = strconv.FormatUint(v.Interface().(uint64), 10)
	default:
		err = fmt.Errorf("Invalid primary field type: %v.", field.getDereferencedType())
		return err
	}
	return nil
}

func setPrimaryField(value string, fieldValue reflect.Value) (err error) {
	// if the id field is of string type set it to the strValue
	t := fieldValue.Type()

	switch t.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int:
		err = setIntField(value, fieldValue, 64)
	case reflect.Int16:
		err = setIntField(value, fieldValue, 16)
	case reflect.Int32:
		err = setIntField(value, fieldValue, 32)
	case reflect.Int64:
		err = setIntField(value, fieldValue, 64)
	case reflect.Uint:
		err = setUintField(value, fieldValue, 64)
	case reflect.Uint16:
		err = setUintField(value, fieldValue, 16)
	case reflect.Uint32:
		err = setUintField(value, fieldValue, 32)
	case reflect.Uint64:
		err = setUintField(value, fieldValue, 64)
	default:
		// should never happen - model checked at precomputation.
		/**

		TO DO:

		Panic - recover
		for internals

		*/
		err = fmt.Errorf("Internal error. Invalid model primary field format: %v", t)
	}
	return
}

func setAttributeField(value string, fieldValue reflect.Value) (err error) {
	// the attribute can be:
	t := fieldValue.Type()
	switch t.Kind() {
	case reflect.Int:
		err = setIntField(value, fieldValue, 64)
	case reflect.Int8:
		err = setIntField(value, fieldValue, 8)
	case reflect.Int16:
		err = setIntField(value, fieldValue, 16)
	case reflect.Int32:
		err = setIntField(value, fieldValue, 32)
	case reflect.Int64:
		err = setIntField(value, fieldValue, 64)
	case reflect.Uint:
		err = setUintField(value, fieldValue, 64)
	case reflect.Uint16:
		err = setUintField(value, fieldValue, 16)
	case reflect.Uint32:
		err = setUintField(value, fieldValue, 32)
	case reflect.Uint64:
		err = setUintField(value, fieldValue, 64)
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Bool:
		err = setBoolField(value, fieldValue)
	case reflect.Float32:
		err = setFloatField(value, fieldValue, 32)
	case reflect.Float64:
		err = setFloatField(value, fieldValue, 64)
	case reflect.Struct:
		// check if it is time

		if _, ok := fieldValue.Elem().Interface().(time.Time); ok {
			// it is time
		} else {
			// structs are not allowed as attribute
			err = fmt.Errorf("The struct is not allowed as an attribute. FieldName: '%s'",
				t.Name())
		}
	default:
		// unknown field
		err = fmt.Errorf("Unsupported field type as an attribute: '%s'.", t.Name())
	}
	return
}

func setTimeField(value string, fieldValue reflect.Value) (err error) {
	return
}

func setUintField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var uintValue uint64

	// Parse unsigned int
	uintValue, err = strconv.ParseUint(value, 10, bitSize)

	if err != nil {
		return err
	}

	// Set uint
	fieldValue.SetUint(uintValue)
	return nil
}

func setIntField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var intValue int64
	intValue, err = strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return err
	}

	// Set value if no error
	fieldValue.SetInt(intValue)
	return nil
}

func setFloatField(value string, fieldValue reflect.Value, bitSize int) (err error) {
	var floatValue float64

	// Parse float
	floatValue, err = strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}
	fieldValue.SetFloat(floatValue)
	return nil
}

func setBoolField(value string, fieldValue reflect.Value) (err error) {
	var boolValue bool
	// set default if empty
	if value == "" {
		value = "false"
	}
	boolValue, err = strconv.ParseBool(value)
	if err != nil {
		return err
	}
	fieldValue.SetBool(boolValue)
	return nil
}
