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

func marshalScope(scope *Scope) (Payloader, error) {
	scopeValue := reflect.ValueOf(scope.Value)
	switch scopeValue.Kind() {
	case reflect.Slice:
		valSlice, err := convertToSliceInterface(&scope.Value)
		if err != nil {
			return nil, err
		}
	case reflect.Ptr:

	}
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
				return nil, err
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
					return nil, err
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
					return nil, err
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
	return node, nil
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

func setNodePrimary(value reflect.Value, node *Node, field *StructField) (err error) {
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

func convertToSliceInterface(i *interface{}) ([]interface{}, error) {
	vals := reflect.ValueOf(*i)
	if vals.Kind() != reflect.Slice {
		return nil, ErrExpectedSlice
	}
	var response []interface{}
	for x := 0; x < vals.Len(); x++ {
		response = append(response, vals.Index(x).Interface())
	}
	return response, nil
}
