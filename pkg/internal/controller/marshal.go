package controller

import (
	"encoding/json"
	aerrors "github.com/kucjac/jsonapi/pkg/errors"
	"github.com/kucjac/jsonapi/pkg/internal"
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/internal/query/scope"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/pkg/errors"

	"fmt"
	"github.com/kucjac/jsonapi/pkg/flags"
	"io"
	"reflect"
	"strconv"
	"time"
)

// Marshal marshals provided value 'v' into writer 'w'
func (c *Controller) Marshal(w io.Writer, v interface{}) error {
	if v == nil {
		return internal.IErrNilValue
	}

	var isMany bool
	refVal := reflect.ValueOf(v)
	t := refVal.Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		return internal.IErrUnsupportedPtrType
	}

	if t.Kind() == reflect.Slice {
		isMany = true
		t = t.Elem()
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	if t.Kind() != reflect.Struct {
		return internal.IErrUnexpectedType
	}

	var schemaName string
	schemaNamer, ok := reflect.New(t).Interface().(models.SchemaNamer)
	if ok {
		schemaName = schemaNamer.SchemaName()
	} else {
		schemaName = c.schemas.DefaultSchema().Name
	}

	schema, ok := c.schemas.Schema(schemaName)
	if !ok {
		return errors.Errorf("Model: %s with schema: '%s' not found:", t.Name(), schemaName)
	}

	mStruct := schema.Model(t)
	if mStruct == nil {
		return errors.Errorf("Model: '%s' not registered.", t.Name())
	}

	var payload Payloader
	if isMany {
		nodes, err := c.visitManyNodes(refVal.Elem(), mStruct)
		if err != nil {
			log.Debugf("visitManyNodes failed: %v", err)
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

func MarshalScope(scope *scope.Scope, controller *Controller) (payloader Payloader, err error) {
	return marshalScope(scope, controller)
}

// MarshalScope marshals given scope into jsonapi format.
func (c *Controller) MarshalScope(scope *scope.Scope) (payloader Payloader, err error) {
	return marshalScope(scope, c)
}

func marshalScope(sc *scope.Scope, controller *Controller) (payloader Payloader, err error) {
	if sc.Value == nil && sc.Kind() >= scope.RelationshipKind {
		/** TO DO:  Build paths */

		if sc.IsMany() {

			payloader = &ManyPayload{Data: []*Node{}}
		} else {
			payloader = &OnePayload{Data: nil}
		}
		return
	}

	scopeValue := reflect.ValueOf(sc.Value)
	t := scopeValue.Type()
	if t.Kind() != reflect.Ptr {
		log.Debugf("Not a pointer")
		err = internal.IErrUnexpectedType
		return
	}
	switch t.Elem().Kind() {
	case reflect.Slice:
		payloader, err = marshalScopeMany(sc, controller)
	case reflect.Struct:
		payloader, err = marshalScopeOne(sc, controller)
	default:
		err = internal.IErrUnexpectedType
	}
	if err != nil {
		return
	}

	included := []*Node{}
	if err = marshalIncludes(sc, &included, controller); err != nil {
		return
	}

	if len(included) != 0 {
		payloader.setIncluded(included)
	}
	return
}

func marshalIncludes(
	rootScope *scope.Scope,
	included *[]*Node,
	controller *Controller,
) (err error) {
	for _, includedScope := range rootScope.IncludedScopes() {
		if err = marshalIncludedScope(includedScope, included, controller); err != nil {
			return err
		}
	}
	return nil
}

func marshalIncludedScope(
	includedScope *scope.Scope,
	included *[]*Node,
	controller *Controller,
) (err error) {
	for _, elem := range includedScope.IncludedValues().Values() {
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

// marshalNestedStructValue marshals the NestedStruct value as it was defined in the controller encoding
func marshalNestedStructValue(n *models.NestedStruct, v reflect.Value) reflect.Value {

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result := reflect.New(models.NestedStructMarshalType(n))
	marshalValue := result.Elem()

	for _, nestedField := range models.NestedStructFields(n) {
		vField := v.FieldByIndex(nestedField.ReflectField().Index)
		mField := marshalValue.FieldByIndex(nestedField.ReflectField().Index)
		if models.FieldIsNestedStruct(nestedField.StructField) {
			mField.Set(marshalNestedStructValue(nestedField.Nested(), vField))
		} else {
			mField.Set(vField)
		}
	}

	return marshalValue
}

//
func unmarshalNestedStructValue(c *Controller, n *models.NestedStruct, value interface{}) (reflect.Value, error) {
	mp, ok := value.(map[string]interface{})
	if !ok {
		err := aerrors.ErrInvalidJSONFieldValue.Copy()
		err.Detail = fmt.Sprintf("Invalid field value for the subfield within attribute: '%s'", models.NestedStructAttr(n).ApiName())
		return reflect.Value{}, err
	}

	result := reflect.New(n.Type())
	resElem := result.Elem()
	for mpName, mpVal := range mp {
		nestedField, ok := models.NestedStructFields(n)[mpName]
		if !ok {
			if !c.StrictUnmarshalMode {
				continue
			}
			err := aerrors.ErrInvalidJSONFieldValue.Copy()
			err.Detail = fmt.Sprintf("No subfield named: '%s' within attr: '%s'", mpName, models.NestedStructAttr(n).ApiName())
			return reflect.Value{}, err
		}

		fieldValue := resElem.FieldByIndex(nestedField.ReflectField().Index)

		err := c.unmarshalAttrFieldValue(nestedField.StructField, fieldValue, mpVal)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	if models.FieldIsBasePtr(n.StructField().Self()) {
		log.Debugf("NestedStruct: '%v' isBasePtr. Attr: '%s'", result.Type(), n.Attr().Name())
		return result, nil
	}
	log.Debugf("NestedStruct: '%v' isNotBasePtr. Attr: '%s'", resElem.Type(), n.Attr().Name())
	return resElem, nil
}

func marshalScopeOne(scope *scope.Scope, controller *Controller) (*OnePayload, error) {
	node, err := visitScopeNode(scope.Value, scope, controller)
	if err != nil {
		return nil, err
	}
	return &OnePayload{Data: node}, nil
}

func marshalScopeMany(scope *scope.Scope, controller *Controller) (*ManyPayload, error) {
	nodes, err := visitScopeManyNodes(scope, controller)
	if err != nil {
		return nil, err
	}
	return &ManyPayload{Data: nodes}, nil
}

func visitScopeManyNodes(scope *scope.Scope, controller *Controller,
) ([]*Node, error) {
	valInterface := reflect.ValueOf(scope.Value).Elem().Interface()
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

func (c *Controller) visitManyNodes(v reflect.Value, mStruct *models.ModelStruct) ([]*Node, error) {
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
	mStruct *models.ModelStruct,
) (*Node, error) {

	if reflect.Indirect(value).Kind() != reflect.Struct {
		return nil, internal.IErrUnexpectedType
	}

	valInt := value.Interface()

	node := &Node{Type: mStruct.Collection()}
	modelVal := value.Elem()

	primStruct := mStruct.PrimaryField()

	primIndex := primStruct.FieldIndex()
	primaryVal := modelVal.FieldByIndex(primIndex)

	var err error

	if !models.FieldIsHidden(primStruct) && !models.FieldIsZeroValue(primStruct, primaryVal.Interface()) {
		err = setNodePrimary(primaryVal, node)
		if err != nil {
			return nil, err
		}
	}

	// iterate over fields
	for _, field := range mStruct.Fields() {

		// Omit hidden fields
		if models.FieldIsHidden(field) {
			continue
		}

		fieldValue := modelVal.FieldByIndex(field.ReflectField().Index)
		if models.FieldIsOmitEmpty(field) {
			if reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(field.ReflectField().Type).Interface()) {
				continue
			}

		}

		switch field.FieldKind() {
		case models.KindAttribute:
			if node.Attributes == nil {
				node.Attributes = make(map[string]interface{})
			}

			if models.FieldIsTime(field) {
				if !models.FieldIsBasePtr(field) {
					t := fieldValue.Interface().(time.Time)

					if t.IsZero() {
						continue
					}

					if models.FieldIsIso8601(field) {
						node.Attributes[field.ApiName()] = t.UTC().Format(internal.Iso8601TimeFormat)
					} else {
						node.Attributes[field.ApiName()] = t.Unix()
					}

				} else {
					if fieldValue.IsNil() {
						if models.FieldIsOmitEmpty(field) {
							continue
						}
						node.Attributes[field.ApiName()] = nil
					} else {
						t := fieldValue.Interface().(*time.Time)

						if t.IsZero() && models.FieldIsOmitEmpty(field) {
							continue
						}

						if models.FieldIsIso8601(field) {
							node.Attributes[field.ApiName()] = t.UTC().Format(internal.Iso8601TimeFormat)
						} else {
							node.Attributes[field.ApiName()] = t.Unix()
						}
					}
				}
			} else {
				if models.FieldIsOmitEmpty(field) && models.FieldIsPtr(field) && fieldValue.IsNil() {
					continue
				} else {
					emptyValue := reflect.Zero(fieldValue.Type())
					if models.FieldIsOmitEmpty(field) && reflect.
						DeepEqual(fieldValue.Interface(), emptyValue.Interface()) {
						continue
					}
				}

				if models.FieldIsNestedStruct(field) {
					node.Attributes[field.ApiName()] = marshalNestedStructValue(field.Nested(), fieldValue).Interface()
					continue
				}

				strAttr, ok := fieldValue.Interface().(string)
				if ok {
					node.Attributes[field.ApiName()] = strAttr
				} else {
					node.Attributes[field.ApiName()] = fieldValue.Interface()
				}
			}
		case models.KindRelationshipMultiple, models.KindRelationshipSingle:

			var isSlice bool = field.FieldKind() == models.KindRelationshipMultiple
			if models.FieldIsOmitEmpty(field) &&
				(isSlice && fieldValue.Len() == 0 || !isSlice && fieldValue.IsNil()) {
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			// how to handle links?
			var relLinks *Links

			if linkableModel, ok := valInt.(RelationshipLinkable); ok {
				relLinks = linkableModel.JSONAPIRelationshipLinks(field.ApiName())
			} else if value, ok := c.Flags.Get(flags.UseLinks); ok && value {
				link := make(map[string]interface{})
				link["self"] = fmt.Sprintf("%s/%s/%s/relationships/%s", mStruct.SchemaName(), mStruct.Collection(), node.ID, field.ApiName())
				link["related"] = fmt.Sprintf("%s/%s/%s/%s", mStruct.SchemaName(), mStruct.Collection(), node.ID, field.ApiName())
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
				node.Relationships[field.ApiName()] = relationship
			} else {
				// is to-one relationship
				if fieldValue.IsNil() {

					node.Relationships[field.ApiName()] = &RelationshipOneNode{Links: relLinks, Meta: relMeta}
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
				node.Relationships[field.ApiName()] = relationship
			}
		}
	}

	if linkable, ok := valInt.(Linkable); ok {
		node.Links = linkable.JSONAPILinks()
	} else if value, ok := c.Flags.Get(flags.UseLinks); ok && value {
		links := make(map[string]interface{})
		links["self"] = fmt.Sprintf("%s/%s/%s", mStruct.SchemaName(), mStruct.Collection(), node.ID)

		linksObj := Links(links)
		node.Links = &(linksObj)
	}

	return node, nil
}

func visitScopeNode(value interface{}, sc *scope.Scope, controller *Controller,
) (*Node, error) {

	if reflect.Indirect(reflect.ValueOf(value)).Kind() != reflect.Struct {
		return nil, internal.IErrUnexpectedType
	}
	node := &Node{Type: sc.Struct().Collection()}
	modelVal := reflect.ValueOf(value).Elem()

	// set primary

	primStruct := sc.Struct().PrimaryField()

	primIndex := primStruct.FieldIndex()
	primaryVal := modelVal.FieldByIndex(primIndex)

	var err error
	if !primStruct.IsHidden() && !models.FieldIsZeroValue(primStruct, primaryVal.Interface()) {
		err = setNodePrimary(primaryVal, node)
		if err != nil {
			return nil, err
		}
	}

	for _, field := range sc.GetModelsRootScope(sc.Struct()).Fieldset() {
		if field.IsHidden() {
			continue
		}

		fieldValue := modelVal.FieldByIndex(field.FieldIndex())
		switch field.FieldKind() {
		case models.KindAttribute:
			if node.Attributes == nil {
				node.Attributes = make(map[string]interface{})
			}

			if field.IsTime() {

				if !field.IsBasePtr() {
					t := fieldValue.Interface().(time.Time)

					if t.IsZero() {
						continue
					}

					if field.IsIso8601() {
						node.Attributes[field.ApiName()] = t.UTC().Format(internal.Iso8601TimeFormat)
					} else {
						node.Attributes[field.ApiName()] = t.Unix()
					}
				} else {
					if fieldValue.IsNil() {
						if field.IsOmitEmpty() {
							continue
						}
						node.Attributes[field.ApiName()] = nil
					} else {
						t := fieldValue.Interface().(*time.Time)

						if t.IsZero() && field.IsOmitEmpty() {
							continue
						}

						if field.IsIso8601() {
							node.Attributes[field.ApiName()] = t.UTC().Format(internal.Iso8601TimeFormat)
						} else {
							node.Attributes[field.ApiName()] = t.Unix()
						}
					}
				}
			} else {
				emptyValue := reflect.Zero(fieldValue.Type())
				if field.IsOmitEmpty() && reflect.
					DeepEqual(fieldValue.Interface(), emptyValue.Interface()) {
					continue
				}

				strAttr, ok := fieldValue.Interface().(string)
				if ok {
					node.Attributes[field.ApiName()] = strAttr
				} else {
					node.Attributes[field.ApiName()] = fieldValue.Interface()
				}
			}
		case models.KindRelationshipMultiple, models.KindRelationshipSingle:

			var isSlice bool = field.FieldKind() == models.KindRelationshipMultiple
			if field.IsOmitEmpty() &&
				(isSlice && fieldValue.Len() == 0 || !isSlice && fieldValue.IsNil()) {
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			// how to handle links?
			var relLinks *Links

			if linkableModel, ok := sc.Value.(RelationshipLinkable); ok {
				relLinks = linkableModel.JSONAPIRelationshipLinks(field.ApiName())
			} else if value, ok := sc.Flags().Get(flags.UseLinks); ok && value {
				link := make(map[string]interface{})
				link["self"] = fmt.Sprintf("%s/%s/%s/relationships/%s", sc.Struct().SchemaName(), sc.Struct().Collection(), node.ID, field.ApiName())
				link["related"] = fmt.Sprintf("%s/%s/%s/%s", sc.Struct().SchemaName(), sc.Struct().Collection(), node.ID, field.ApiName())
				links := Links(link)
				relLinks = &links
			}

			var relMeta *Meta
			if metableModel, ok := sc.Value.(Metable); ok {
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
				node.Relationships[field.ApiName()] = relationship
			} else {
				// is to-one relationship
				if fieldValue.IsNil() {

					node.Relationships[field.ApiName()] = &RelationshipOneNode{Links: relLinks, Meta: relMeta}
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
				node.Relationships[field.ApiName()] = relationship
			}
		}
	}

	if linkable, ok := sc.Value.(Linkable); ok {
		node.Links = linkable.JSONAPILinks()
	} else if value, ok := sc.Flags().Get(flags.UseLinks); ok && value {
		links := make(map[string]interface{})
		var self string
		switch sc.Kind() {
		case scope.RootKind, scope.IncludedKind:
			self = fmt.Sprintf("%s/%s/%s", sc.Struct().SchemaName(), sc.Struct().Collection(), node.ID)
		case scope.RelatedKind:
			rootScope := sc.GetModelsRootScope(sc.Struct())
			if rootScope == nil || len(rootScope.IncludedFields()) == 0 {
				err = fmt.Errorf("Invalid scope provided as related scope. Scope value type: '%s'", sc.Struct().Type())
				return nil, err
			}

			relatedName := rootScope.IncludedFields()[0].ApiName()
			self = fmt.Sprintf("%s/%s/%s/%s",
				rootScope.Struct().SchemaName(),
				sc.Struct().Collection(),
				node.ID,
				relatedName,
			)
		case scope.RelationshipKind:
			rootScope := sc.GetModelsRootScope(sc.Struct())
			if rootScope == nil || len(rootScope.IncludedFields()) == 0 {
				err = fmt.Errorf("Invalid scope provided as related scope. Scope value type: '%s'", sc.Struct().Type())
				return nil, err
			}
			relatedName := rootScope.IncludedFields()[0].ApiName()
			self = fmt.Sprintf("%s/%s/%s/relationships/%s",
				rootScope.Struct().SchemaName(),
				sc.Struct().Collection(),
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
	field *models.StructField,
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
	field *models.StructField,
	controller *Controller,
) (*Node, error) {
	mStruct := models.FieldsRelatedModelStruct(field)
	prim := mStruct.PrimaryField()
	node := &Node{Type: mStruct.Collection()}

	index := prim.FieldIndex()

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	nodeValue := value.FieldByIndex(index)
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
		return nil, internal.IErrExpectedSlice
	}
	var response []interface{}
	for x := 0; x < vals.Len(); x++ {
		response = append(response, vals.Index(x).Interface())
	}
	return response, nil
}
