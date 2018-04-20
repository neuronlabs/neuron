package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
)

var (
	IErrUnexpectedType = errors.
				New("models should be a struct pointer or slice of struct pointers")
	IErrExpectedSlice = errors.New("models should be a slice of struct pointers")
)

func MarshalPayload(w io.Writer, payload Payloader) error {
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return err
	}
	return nil
}

func marshalScope(scope *Scope, controller *Controller) (payloader Payloader, err error) {
	scopeValue := reflect.ValueOf(scope.Value)
	switch scopeValue.Kind() {
	case reflect.Slice:
		payloader, err = marshalScopeMany(scope, controller)
	case reflect.Ptr:
		payloader, err = marshalScopeOne(scope, controller)
	}
	if err != nil {
		return
	}

	for _, subscope := range scope.SubScopes {
		err = marshalSubScope(subscope, payloader.getIncluded(), controller)
		if err != nil {
			return
		}
	}
	return
}

func marshalSubScope(scope *Scope, included *[]*Node, controller *Controller) error {
	// get this
	scopeValue := reflect.ValueOf(scope.Value)
	switch scopeValue.Kind() {
	case reflect.Slice:
		nodes, err := visitScopeManyNodes(scope, controller)
		if err != nil {
			return err
		}
		*included = append(*included, nodes...)
	case reflect.Ptr:
		node, err := visitScopeNode(scope.Value, scope, controller)
		if err != nil {
			return err
		}
		*included = append(*included, node)
	}
	// iterate over subscopes and marshalsubscopes
	for _, sub := range scope.SubScopes {
		err := marshalSubScope(sub, included, controller)
		if err != nil {
			return err
		}
	}

	return nil
}

func marshalScopeOne(scope *Scope, controller *Controller) (*OnePayload, error) {
	node, err := visitScopeNode(scope.Value, scope, controller)
	if err != nil {
		return nil, err
	}
	return &OnePayload{Data: node}, nil
}

func marshalScopeMany(scope *Scope, controller *Controller) (*ManyPayload, error) {
	nodes, err := visitScopeManyNodes(scope, controller)
	if err != nil {
		return nil, err
	}
	return &ManyPayload{Data: nodes}, nil
}

func visitScopeManyNodes(scope *Scope, controller *Controller,
) ([]*Node, error) {
	valSlice, err := convertToSliceInterface(&scope.Value)
	if err != nil {
		return nil, err
	}
	nodes := []*Node{}

	for _, value := range valSlice {
		node, err := visitScopeNode(value, scope, controller)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}
	return nodes, nil
}

func visitScopeNode(value interface{}, scope *Scope, controller *Controller,
) (*Node, error) {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() != reflect.Struct {
		return nil, IErrUnexpectedType
	}
	node := &Node{Type: scope.Struct.collectionType}
	modelVal := reflect.ValueOf(value).Elem()

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
				relationship, err := visitRelationshipManyNode(fieldValue, field, controller)
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
				relatedNode, err := visitRelationshipNode(fieldValue, field, controller)
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

func visitRelationshipManyNode(manyValue reflect.Value, idval reflect.Value, field *StructField, controller *Controller,
) (*RelationshipManyNode, error) {
	nodes := []*Node{}

	for i := 0; i < manyValue.Len(); i++ {
		elemValue := manyValue.Index(i)

		node, err := visitRelationshipNode(elemValue, field, controller)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)

	}
	return &RelationshipManyNode{Data: nodes}, nil
}

func visitRelationshipNode(value reflect.Value, field *StructField, controller *Controller,
) (*Node, error) {
	mStruct := field.mStruct
	prim := mStruct.primary
	node := &Node{Type: mStruct.collectionType}

	index := prim.getFieldIndex()

	nodeValue := value.Index(index)
	err := setNodePrimary(nodeValue, node, field)
	if err != nil {
		return nil, err
	}

	controller.APIURLBase
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
		return nil, IErrExpectedSlice
	}
	var response []interface{}
	for x := 0; x < vals.Len(); x++ {
		response = append(response, vals.Index(x).Interface())
	}
	return response, nil
}
