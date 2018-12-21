package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kucjac/jsonapi/pkg/flags"
	"io"
	"reflect"
	"strconv"
	"time"
)

func (c *Controller) Marshal(w io.Writer, v interface{}) error {
	if v == nil {
		return IErrNilValue
	}

	var isMany bool
	refVal := reflect.ValueOf(v)
	t := refVal.Type()

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Slice {
		isMany = true
		t = t.Elem()
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	if t.Kind() != reflect.Struct {
		return IErrUnexpectedType
	}

	mStruct := c.Models.Get(t)
	if mStruct == nil {
		return IErrModelNotMapped
	}

	var payload Payloader
	if isMany {
		nodes, err := c.visitManyNodes(refVal, mStruct)
		if err != nil {
			return err
		}
		payload = &ManyPayload{Data: nodes}
	} else {
		node, err := c.visistNode(refVal, mStruct)
		if err != nil {
			return err
		}
		payload = &OnePayload{Data: node}
	}

	return MarshalPayload(w, payload)
}

func MarshalPayload(w io.Writer, payload Payloader) error {
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return err
	}
	return nil
}

func MarshalScope(scope *Scope, controller *Controller) (payloader Payloader, err error) {
	return marshalScope(scope, controller)
}

func marshalScope(scope *Scope, controller *Controller) (payloader Payloader, err error) {
	if scope.Value == nil && scope.kind >= relationshipKind {
		/** TO DO:  Build paths */

		if scope.IsMany {

			payloader = &ManyPayload{Data: []*Node{}}
		} else {
			payloader = &OnePayload{Data: nil}
		}
		return
	}

	scopeValue := reflect.ValueOf(scope.Value)
	switch scopeValue.Kind() {
	case reflect.Slice:
		payloader, err = marshalScopeMany(scope, controller)
	case reflect.Ptr:
		payloader, err = marshalScopeOne(scope, controller)
	default:
		err = IErrUnexpectedType
	}
	if err != nil {
		return
	}

	included := []*Node{}
	if err = marshalIncludes(scope, &included, controller); err != nil {
		return
	}

	if len(included) != 0 {
		payloader.setIncluded(included)
	}
	return
}

func marshalIncludes(
	rootScope *Scope,
	included *[]*Node,
	controller *Controller,
) (err error) {
	for _, includedScope := range rootScope.IncludedScopes {
		if err = marshalIncludedScope(includedScope, included, controller); err != nil {
			return err
		}
	}
	return nil
}

func marshalIncludedScope(
	includedScope *Scope,
	included *[]*Node,
	controller *Controller,
) (err error) {
	for _, elem := range includedScope.IncludedValues.values {
		if elem == nil {
			continue
		}
		node, err := visitScopeNode(elem, includedScope, controller)
		if err != nil {
			return err
		}
		*included = append(*included, node)
	}
	return
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
	valInterface := reflect.ValueOf(scope.Value).Interface()
	valSlice, err := convertToSliceInterface(&valInterface)
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

func (c *Controller) visitManyNodes(v reflect.Value, mStruct *ModelStruct) ([]*Node, error) {
	nodes := []*Node{}

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.IsNil() {
			continue
		}
		node, err := c.visistNode(elem, mStruct)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (c *Controller) visistNode(
	value reflect.Value,
	mStruct *ModelStruct,
) (*Node, error) {

	if reflect.Indirect(value).Kind() != reflect.Struct {
		return nil, IErrUnexpectedType
	}

	valInt := value.Interface()

	node := &Node{Type: mStruct.collectionType}
	modelVal := value.Elem()

	primStruct := mStruct.primary

	primIndex := primStruct.getFieldIndex()
	primaryVal := modelVal.Field(primIndex)

	var err error
	if !primStruct.isHidden() && !primStruct.IsZeroValue(primaryVal.Interface()) {
		err = setNodePrimary(primaryVal, node)
		if err != nil {
			return nil, err
		}
	}

	for _, field := range mStruct.fields {

		// Omit hidden fields
		if field.isHidden() {
			continue
		}

		fieldValue := modelVal.FieldByIndex(field.refStruct.Index)
		if field.isOmitEmpty() {
			if reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(field.refStruct.Type).Interface()) {
				continue
			}

		}

		switch field.fieldType {
		case Attribute:
			if node.Attributes == nil {
				node.Attributes = make(map[string]interface{})
			}

			if field.isTime() {
				if !field.isBasePtr() {
					t := fieldValue.Interface().(time.Time)

					if t.IsZero() {
						continue
					}

					if field.isIso8601() {
						node.Attributes[field.jsonAPIName] = t.UTC().Format(iso8601TimeFormat)
					} else {
						node.Attributes[field.jsonAPIName] = t.Unix()
					}

				} else {
					if fieldValue.IsNil() {
						if field.isOmitEmpty() {
							continue
						}
						node.Attributes[field.jsonAPIName] = nil
					} else {
						t := fieldValue.Interface().(*time.Time)

						if t.IsZero() && field.isOmitEmpty() {
							continue
						}

						if field.isIso8601() {
							node.Attributes[field.jsonAPIName] = t.UTC().Format(iso8601TimeFormat)
						} else {
							node.Attributes[field.jsonAPIName] = t.Unix()
						}
					}
				}
			} else {
				if field.isOmitEmpty() && field.isPtr() && fieldValue.IsNil() {
					continue
				} else {
					emptyValue := reflect.Zero(fieldValue.Type())
					if field.isOmitEmpty() && reflect.
						DeepEqual(fieldValue.Interface(), emptyValue.Interface()) {
						continue
					}
				}

				if field.isNestedStruct() {
					node.Attributes[field.jsonAPIName] = field.nested.MarshalValue(fieldValue).Interface()
					continue
				}

				strAttr, ok := fieldValue.Interface().(string)
				if ok {
					node.Attributes[field.jsonAPIName] = strAttr
				} else {
					node.Attributes[field.jsonAPIName] = fieldValue.Interface()
				}
			}
		case RelationshipMultiple, RelationshipSingle:

			var isSlice bool = field.fieldType == RelationshipMultiple
			if field.isOmitEmpty() &&
				(isSlice && fieldValue.Len() == 0 || !isSlice && fieldValue.IsNil()) {
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			// how to handle links?
			var relLinks *Links

			if linkableModel, ok := valInt.(RelationshipLinkable); ok {
				relLinks = linkableModel.JSONAPIRelationshipLinks(field.jsonAPIName)
			} else if value, ok := c.Flags.Get(flags.UseLinks); ok && value {
				link := make(map[string]interface{})
				link["self"] = fmt.Sprintf("%s/%s/%s/relationships/%s", c.APIURLBase, mStruct.collectionType, node.ID, field.jsonAPIName)
				link["related"] = fmt.Sprintf("%s/%s/%s/%s", c.APIURLBase, mStruct.collectionType, node.ID, field.jsonAPIName)
				links := Links(link)
				relLinks = &links
			}

			var relMeta *Meta
			if metableModel, ok := valInt.(Metable); ok {
				relMeta = metableModel.JSONAPIMeta()
			}
			if isSlice {
				// get RelationshipManyNode
				relationship, err := visitRelationshipManyNode(fieldValue, primaryVal, field, c)
				if err != nil {
					return nil, err
				}

				relationship.Links = relLinks
				relationship.Meta = relMeta
				node.Relationships[field.jsonAPIName] = relationship
			} else {
				// is to-one relationship
				if fieldValue.IsNil() {

					node.Relationships[field.jsonAPIName] = &RelationshipOneNode{Links: relLinks, Meta: relMeta}
					continue
				}
				relatedNode, err := visitRelationshipNode(fieldValue, primaryVal, field, c)
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

	if linkable, ok := valInt.(Linkable); ok {
		node.Links = linkable.JSONAPILinks()
	} else if value, ok := c.Flags.Get(flags.UseLinks); ok && value {
		links := make(map[string]interface{})
		links["self"] = fmt.Sprintf("%s/%s/%s", c.APIURLBase, mStruct.collectionType, node.ID)

		linksObj := Links(links)
		node.Links = &(linksObj)
	}

	return node, nil
}

func visitScopeNode(value interface{}, scope *Scope, controller *Controller,
) (*Node, error) {

	if reflect.Indirect(reflect.ValueOf(value)).Kind() != reflect.Struct {
		return nil, IErrUnexpectedType
	}
	node := &Node{Type: scope.Struct.collectionType}
	modelVal := reflect.ValueOf(value).Elem()

	// set primary

	primStruct := scope.Struct.primary

	primIndex := primStruct.getFieldIndex()
	primaryVal := modelVal.Field(primIndex)

	var err error
	if !primStruct.isHidden() && !primStruct.IsZeroValue(primaryVal.Interface()) {
		err = setNodePrimary(primaryVal, node)
		if err != nil {
			return nil, err
		}
	}

	for _, field := range scope.getModelsRootScope(scope.Struct).Fieldset {
		if field.isHidden() {
			continue
		}

		fieldValue := modelVal.Field(field.getFieldIndex())
		switch field.fieldType {
		case Attribute:
			if node.Attributes == nil {
				node.Attributes = make(map[string]interface{})
			}

			if field.isTime() {

				if !field.isBasePtr() {
					t := fieldValue.Interface().(time.Time)

					if t.IsZero() {
						continue
					}

					if field.isIso8601() {
						node.Attributes[field.jsonAPIName] = t.UTC().Format(iso8601TimeFormat)
					} else {
						node.Attributes[field.jsonAPIName] = t.Unix()
					}
				} else {
					if fieldValue.IsNil() {
						if field.isOmitEmpty() {
							continue
						}
						node.Attributes[field.jsonAPIName] = nil
					} else {
						t := fieldValue.Interface().(*time.Time)

						if t.IsZero() && field.isOmitEmpty() {
							continue
						}

						if field.isIso8601() {
							node.Attributes[field.jsonAPIName] = t.UTC().Format(iso8601TimeFormat)
						} else {
							node.Attributes[field.jsonAPIName] = t.Unix()
						}
					}
				}
			} else {
				emptyValue := reflect.Zero(fieldValue.Type())
				if field.isOmitEmpty() && reflect.
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
		case RelationshipMultiple, RelationshipSingle:

			var isSlice bool = field.fieldType == RelationshipMultiple
			if field.isOmitEmpty() &&
				(isSlice && fieldValue.Len() == 0 || !isSlice && fieldValue.IsNil()) {
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			// how to handle links?
			var relLinks *Links

			if linkableModel, ok := scope.Value.(RelationshipLinkable); ok {
				relLinks = linkableModel.JSONAPIRelationshipLinks(field.jsonAPIName)
			} else if value, ok := scope.Flags().Get(flags.UseLinks); ok && value {
				link := make(map[string]interface{})
				link["self"] = fmt.Sprintf("%s/%s/%s/relationships/%s", controller.APIURLBase, scope.Struct.collectionType, node.ID, field.jsonAPIName)
				link["related"] = fmt.Sprintf("%s/%s/%s/%s", controller.APIURLBase, scope.Struct.collectionType, node.ID, field.jsonAPIName)
				links := Links(link)
				relLinks = &links
			}

			var relMeta *Meta
			if metableModel, ok := scope.Value.(Metable); ok {
				relMeta = metableModel.JSONAPIMeta()
			}
			if isSlice {
				// get RelationshipManyNode
				relationship, err := visitRelationshipManyNode(fieldValue, primaryVal, field, controller)
				if err != nil {
					return nil, err
				}

				relationship.Links = relLinks
				relationship.Meta = relMeta
				node.Relationships[field.jsonAPIName] = relationship
			} else {
				// is to-one relationship
				if fieldValue.IsNil() {

					node.Relationships[field.jsonAPIName] = &RelationshipOneNode{Links: relLinks, Meta: relMeta}
					continue
				}
				relatedNode, err := visitRelationshipNode(fieldValue, primaryVal, field, controller)
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

	if linkable, ok := scope.Value.(Linkable); ok {
		node.Links = linkable.JSONAPILinks()
	} else if value, ok := scope.Flags().Get(flags.UseLinks); ok && value {
		links := make(map[string]interface{})
		var self string
		switch scope.kind {
		case rootKind, includedKind:
			self = fmt.Sprintf("%s/%s/%s", controller.APIURLBase, scope.Struct.collectionType, node.ID)
		case relatedKind:
			if scope.rootScope == nil || len(scope.rootScope.IncludedFields) == 0 {
				err = fmt.Errorf("Invalid scope provided as related scope. Scope value type: '%s'", scope.Struct.GetType())
				return nil, err
			}
			relatedName := scope.rootScope.IncludedFields[0].jsonAPIName
			self = fmt.Sprintf("%s/%s/%s/%s",
				controller.APIURLBase,
				scope.Struct.collectionType,
				node.ID,
				relatedName,
			)
		case relationshipKind:
			if scope.rootScope == nil || len(scope.rootScope.IncludedFields) == 0 {
				err = fmt.Errorf("Invalid scope provided as related scope. Scope value type: '%s'", scope.Struct.GetType())
				return nil, err
			}
			relatedName := scope.rootScope.IncludedFields[0].jsonAPIName
			self = fmt.Sprintf("%s/%s/%s/relationships/%s",
				controller.APIURLBase,
				scope.Struct.collectionType,
				node.ID,
				relatedName,
			)
		}
		links["self"] = self
		linksObj := Links(links)
		node.Links = &(linksObj)
	}
	return node, nil
}

func visitRelationshipManyNode(
	manyValue, rootID reflect.Value,
	field *StructField,
	controller *Controller,
) (*RelationshipManyNode, error) {
	nodes := []*Node{}

	for i := 0; i < manyValue.Len(); i++ {
		elemValue := manyValue.Index(i)
		node, err := visitRelationshipNode(elemValue, rootID, field, controller)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)

	}

	return &RelationshipManyNode{Data: nodes}, nil
}

func visitRelationshipNode(
	value, rootID reflect.Value,
	field *StructField,
	controller *Controller,
) (*Node, error) {
	mStruct := field.relatedStruct
	prim := mStruct.primary
	node := &Node{Type: mStruct.collectionType}

	index := prim.getFieldIndex()

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	nodeValue := value.Field(index)
	err := setNodePrimary(nodeValue, node)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func setNodePrimary(value reflect.Value, node *Node) (err error) {
	v := value
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	switch v.Kind() {
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
		err = fmt.Errorf("Invalid primary field type: %v.", v.Type())
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
