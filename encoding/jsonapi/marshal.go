package jsonapi

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"reflect"
	"strconv"
	"time"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
	ctrl "github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/query"

	"github.com/neuronlabs/neuron-core/internal/controller"
	"github.com/neuronlabs/neuron-core/internal/models"
	"github.com/neuronlabs/neuron-core/internal/query/scope"
)

// Marshal marshals the provided value 'v' into the writer with the jsonapi encoding.
// Takes the default controller for the model mapping.
func Marshal(w io.Writer, v interface{}) error {
	return marshal((*controller.Controller)(ctrl.Default()), w, v)
}

// MarshalC marshals the provided value 'v' into the writer. It uses the 'c' controller
func MarshalC(c *ctrl.Controller, w io.Writer, v interface{}, option ...*MarshalOptions) error {
	return marshal((*controller.Controller)(c), w, v, option...)
}

// MarshalScope marshals the scope into the selceted writer for the given controller
func MarshalScope(w io.Writer, s *query.Scope, option ...*MarshalOptions) error {
	pl, err := marshalScope((*controller.Controller)(s.Controller()), (*scope.Scope)(s), option...)
	if err != nil {
		return err
	}

	return marshalPayload(w, pl)
}

// MarshalErrors writes a JSON API response using the given `[]error`.
//
// For more information on JSON API error payloads, see the spec here:
// http://jsonapi.org/format/#document-top-level
// and here: http://jsonapi.org/format/#error-objects.
func MarshalErrors(w io.Writer, errs ...*Error) error {
	if err := json.NewEncoder(w).Encode(&ErrorsPayload{Errors: errs}); err != nil {
		return err
	}
	return nil
}

// ErrorsPayload is a serializer struct for representing a valid JSON API errors payload.
type ErrorsPayload struct {
	JSONAPI map[string]interface{} `json:"jsonapi,omitempty"`
	Errors  []*Error               `json:"errors"`
}

// Marshal marshals provided value 'v' into writer 'w'
func marshal(c *controller.Controller, w io.Writer, v interface{}, option ...*MarshalOptions) error {
	if v == nil {
		// TODO: allow marshaling nil or empty values of given type.
		return errors.NewDet(class.EncodingMarshalNilValue, "nil value provided")
	}

	var o *MarshalOptions
	if len(option) > 0 {
		o = option[0]
	}

	// get the value reflection
	var isMany bool
	refVal := reflect.ValueOf(v)
	t := refVal.Type()

	// the only allowed values are pointer to struct or pointer to slice.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		return errors.NewDet(class.EncodingMarshalNonAddressable, "provided unaddressable value")
	}

	// check if value is a slice
	if t.Kind() == reflect.Slice {
		isMany = true
		t = t.Elem()

		// dereference until the type is not a pointer.
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	// the type kind should be a structure.
	if t.Kind() != reflect.Struct {
		return errors.NewDet(class.EncodingMarshalInput, "provided value is not a struct based")
	}

	mStruct := c.ModelMap().Get(t)
	if mStruct == nil {
		return errors.NewDetf(class.EncodingMarshalModelNotMapped, "model: '%s' is not registered.", t.Name())
	}

	var payload Payloader
	if isMany {
		nodes, err := visitManyNodes(c, refVal.Elem(), mStruct, o)
		if err != nil {
			log.Debug2f("visitManyNodes failed: %v", err)
			return err
		}
		payload = &ManyPayload{Data: nodes}
	} else {
		node, err := visitNode(c, refVal, mStruct, o)
		if err != nil {
			return err
		}
		payload = &SinglePayload{Data: node}
	}

	return marshalPayload(w, payload)
}

func marshalPayload(w io.Writer, payload Payloader) error {
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return err
	}
	return nil
}

func marshalScope(c *controller.Controller, sc *scope.Scope, o ...*MarshalOptions) (Payloader, error) {
	var (
		payload Payloader
		err     error
	)

	if sc.Value == nil {
		if sc.IsMany() {
			payload = &ManyPayload{Data: []*node{}}
		} else {
			payload = &SinglePayload{Data: nil}
		}

		return payload, nil
	}

	scopeValue := reflect.ValueOf(sc.Value)
	t := scopeValue.Type()
	if t.Kind() != reflect.Ptr {
		err = errors.NewDet(class.EncodingMarshalNonAddressable, "scope's value is non addressable")
		return nil, err
	}

	var option *MarshalOptions
	if len(o) > 0 {
		option = o[0]
	}

	switch t.Elem().Kind() {
	case reflect.Slice:
		payload, err = marshalScopeMany(c, sc, option)
	case reflect.Struct:
		payload, err = marshalScopeOne(c, sc, option)
	default:
		err = errors.NewDetf(class.EncodingMarshalInput, "invalid scope's value type: '%T'", sc.Value)
	}
	if err != nil {
		return nil, err
	}

	// try to unmarshal the includes
	included := []*node{}
	if err = marshalIncludes(c, sc, &included, option); err != nil {
		return nil, err
	}

	if len(included) != 0 {
		payload.setIncluded(included)
	}
	return payload, nil
}

func marshalIncludes(c *controller.Controller, rootScope *scope.Scope, included *[]*node, o *MarshalOptions) (err error) {
	for _, includedScope := range rootScope.IncludedScopes() {
		if err = marshalIncludedScope(c, includedScope, included, o); err != nil {
			return err
		}
	}
	return nil
}

func marshalIncludedScope(c *controller.Controller, includedScope *scope.Scope, included *[]*node, o *MarshalOptions) error {
	for _, elem := range includedScope.IncludedValues().Values() {
		if elem == nil {
			continue
		}

		node, err := visitScopeNode(c, elem, includedScope, o)
		if err != nil {
			return err
		}
		*included = append(*included, node)
	}
	return nil
}

// marshalNestedStructValue marshals the NestedStruct value as it was defined in the controller encoding
func marshalNestedStructValue(n *models.NestedStruct, v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result := reflect.New(models.NestedStructMarshalType(n))
	marshalValue := result.Elem()

	for _, nestedField := range models.NestedStructFields(n) {
		vField := v.FieldByIndex(nestedField.StructField().ReflectField().Index)
		mField := marshalValue.FieldByIndex(nestedField.StructField().ReflectField().Index)

		if nestedField.StructField().IsNestedStruct() {
			mField.Set(marshalNestedStructValue(nestedField.StructField().Nested(), vField))
		} else {
			mField.Set(vField)
		}
	}
	return marshalValue
}

func marshalScopeOne(c *controller.Controller, s *scope.Scope, o *MarshalOptions) (*SinglePayload, error) {
	n, err := visitScopeNode(c, s.Value, s, o)
	if err != nil {
		return nil, err
	}
	var links *Links
	if o.Link == RelationshipLink {
		links = &Links{
			"self":    path.Join(o.LinkURL, o.RootCollection, o.RootID, "relationships", o.RelatedField),
			"related": path.Join(o.LinkURL, o.RootCollection, o.RootID, o.RelatedField),
		}
	}

	return &SinglePayload{Data: n, Links: links}, nil
}

func marshalScopeMany(c *controller.Controller, s *scope.Scope, o *MarshalOptions) (*ManyPayload, error) {
	n, err := visitScopeManyNodes(c, s, o)
	if err != nil {
		return nil, err
	}

	var links *Links
	if o.Link == RelationshipLink {
		links = &Links{
			"self":    path.Join(o.LinkURL, o.RootCollection, o.RootID, "relationships", o.RelatedField),
			"related": path.Join(o.LinkURL, o.RootCollection, o.RootID, o.RelatedField),
		}
	}
	return &ManyPayload{Data: n, Links: links}, nil
}

func visitScopeManyNodes(c *controller.Controller, s *scope.Scope, o *MarshalOptions) ([]*node, error) {
	valInterface := reflect.ValueOf(s.Value).Elem().Interface()

	valSlice, err := convertToSliceInterface(&valInterface)
	if err != nil {
		return nil, err
	}
	nodes := []*node{}

	for _, value := range valSlice {
		node, err := visitScopeNode(c, value, s, o)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}
	return nodes, nil
}

func visitManyNodes(c *controller.Controller, v reflect.Value, mStruct *models.ModelStruct, o *MarshalOptions) ([]*node, error) {
	nodes := []*node{}

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.IsNil() {
			continue
		}

		node, err := visitNode(c, elem, mStruct, o)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}
	return nodes, nil
}

func visitNode(c *controller.Controller, value reflect.Value, mStruct *models.ModelStruct, o *MarshalOptions) (*node, error) {
	// check if any of the multiple nodes is not a struct
	if indirect := reflect.Indirect(value); indirect.Kind() != reflect.Struct {
		return nil, errors.NewDetf(class.EncodingMarshalInput, "one of the provided values is of invalid type: '%s'", indirect.Type().Name())
	}

	node := &node{Type: mStruct.Collection()}
	modelVal := value.Elem()

	primStruct := mStruct.PrimaryField()

	primIndex := primStruct.FieldIndex()
	primaryVal := modelVal.FieldByIndex(primIndex)

	var err error

	if !primStruct.IsHidden() && !primStruct.IsZeroValue(primaryVal.Interface()) {
		err = setNodePrimary(primaryVal, node)
		if err != nil {
			return nil, err
		}
	}

	// iterate over fields
	for _, field := range mStruct.Fields() {
		// Omit hidden fields
		if field.IsHidden() {
			continue
		}

		fieldValue := modelVal.FieldByIndex(field.ReflectField().Index)
		if field.IsOmitEmpty() {
			if reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(field.ReflectField().Type).Interface()) {
				continue
			}

		}

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

					if field.IsISO8601() {
						node.Attributes[field.NeuronName()] = t.UTC().Format(ISO8601TimeFormat)
					} else {
						node.Attributes[field.NeuronName()] = t.Unix()
					}

				} else {
					if fieldValue.IsNil() {
						if field.IsOmitEmpty() {
							continue
						}
						node.Attributes[field.NeuronName()] = nil
					} else {
						t := fieldValue.Interface().(*time.Time)

						if t.IsZero() && field.IsOmitEmpty() {
							continue
						}

						if field.IsISO8601() {
							node.Attributes[field.NeuronName()] = t.UTC().Format(ISO8601TimeFormat)
						} else {
							node.Attributes[field.NeuronName()] = t.Unix()
						}
					}
				}
			} else {
				if field.IsOmitEmpty() && field.IsPtr() && fieldValue.IsNil() {
					continue
				} else {
					emptyValue := reflect.Zero(fieldValue.Type())
					if field.IsOmitEmpty() && reflect.DeepEqual(fieldValue.Interface(), emptyValue.Interface()) {
						continue
					}
				}

				if field.IsNestedStruct() {
					node.Attributes[field.NeuronName()] = marshalNestedStructValue(field.Nested(), fieldValue).Interface()
					continue
				}

				strAttr, ok := fieldValue.Interface().(string)
				if ok {
					node.Attributes[field.NeuronName()] = strAttr
				} else {
					node.Attributes[field.NeuronName()] = fieldValue.Interface()
				}
			}
		case models.KindRelationshipMultiple, models.KindRelationshipSingle:
			isSlice := field.FieldKind() == models.KindRelationshipMultiple
			if field.IsOmitEmpty() && ((isSlice && fieldValue.Len() == 0) || (!isSlice && fieldValue.IsNil())) {
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			var relLinks *Links

			if o != nil && o.Link == DefaultLink {
				link := make(map[string]interface{})
				link["self"] = path.Join(o.LinkURL, mStruct.Collection(), node.ID, "relationships", field.NeuronName())
				link["related"] = path.Join(o.LinkURL, mStruct.Collection(), node.ID, field.NeuronName())
				links := Links(link)
				relLinks = &links
			} else if c.Config.EncodeLinks {
				link := make(map[string]interface{})
				link["self"] = fmt.Sprintf("%s/%s/relationships/%s", mStruct.Collection(), node.ID, field.NeuronName())
				link["related"] = fmt.Sprintf("%s/%s/%s", mStruct.Collection(), node.ID, field.NeuronName())
				links := Links(link)
				relLinks = &links
			}

			var relMeta *Meta
			if o != nil {
				if optionMeta, ok := o.RelationshipMeta[field.NeuronName()]; ok {
					relMeta = (*Meta)(&optionMeta)
				}
			}

			if isSlice {
				// get RelationshipManyNode
				relationship, err := visitRelationshipManyNode(c, fieldValue, primaryVal, field)
				if err != nil {
					return nil, err
				}

				relationship.Links = relLinks
				relationship.Meta = relMeta
				node.Relationships[field.NeuronName()] = relationship
			} else {
				// is to-one relationship
				if fieldValue.IsNil() {

					node.Relationships[field.NeuronName()] = &relationshipOneNode{Links: relLinks, Meta: relMeta}
					continue
				}
				relatedNode, err := visitRelationshipNode(c, fieldValue, primaryVal, field)
				if err != nil {
					return nil, err
				}
				relationship := &relationshipOneNode{
					Data:  relatedNode,
					Links: relLinks,
					Meta:  relMeta,
				}
				node.Relationships[field.NeuronName()] = relationship
			}
		}
	}

	if o != nil && o.Link == DefaultLink {
		links := make(map[string]interface{})
		links["self"] = path.Join(o.LinkURL, mStruct.Collection(), node.ID)
		linksObj := Links(links)
		node.Links = &(linksObj)
	} else if c.Config.EncodeLinks {
		links := make(map[string]interface{})
		links["self"] = fmt.Sprintf("%s/%s", mStruct.Collection(), node.ID)

		linksObj := Links(links)
		node.Links = &(linksObj)
	}

	return node, nil
}

func visitScopeNode(c *controller.Controller, value interface{}, sc *scope.Scope, o *MarshalOptions) (*node, error) {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() != reflect.Struct {
		return nil, errors.NewDet(class.EncodingMarshalInput, "one of the provided values is of invalid type")
	}

	node := &node{Type: sc.Struct().Collection()}
	modelVal := reflect.ValueOf(value).Elem()

	// set primary
	primStruct := sc.Struct().PrimaryField()
	primIndex := primStruct.FieldIndex()
	primaryVal := modelVal.FieldByIndex(primIndex)

	if !primStruct.IsHidden() && !primStruct.IsZeroValue(primaryVal.Interface()) {
		err := setNodePrimary(primaryVal, node)
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

					if field.IsISO8601() {
						node.Attributes[field.NeuronName()] = t.UTC().Format(ISO8601TimeFormat)
					} else {
						node.Attributes[field.NeuronName()] = t.Unix()
					}
				} else {
					if fieldValue.IsNil() {
						if field.IsOmitEmpty() {
							continue
						}
						node.Attributes[field.NeuronName()] = nil
					} else {
						t := fieldValue.Interface().(*time.Time)

						if t.IsZero() && field.IsOmitEmpty() {
							continue
						}

						if field.IsISO8601() {
							node.Attributes[field.NeuronName()] = t.UTC().Format(ISO8601TimeFormat)
						} else {
							node.Attributes[field.NeuronName()] = t.Unix()
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
					node.Attributes[field.NeuronName()] = strAttr
				} else {
					node.Attributes[field.NeuronName()] = fieldValue.Interface()
				}
			}
		case models.KindRelationshipMultiple, models.KindRelationshipSingle:
			isSlice := field.FieldKind() == models.KindRelationshipMultiple
			if field.IsOmitEmpty() &&
				(isSlice && fieldValue.Len() == 0 || !isSlice && fieldValue.IsNil()) {
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]interface{})
			}

			// how to handle links?
			var relLinks *Links

			if o != nil && o.Link == DefaultLink {
				link := make(map[string]interface{})
				link["self"] = path.Join(o.LinkURL, sc.Struct().Collection(), node.ID, "relationships", field.NeuronName())
				link["related"] = path.Join(o.LinkURL, sc.Struct().Collection(), node.ID, field.NeuronName())
				links := Links(link)
				relLinks = &links
			} else if value, ok := sc.StoreGet(encodeLinksCtxKey); ok {
				if encodeLinks, ok := value.(bool); ok && encodeLinks {
					link := make(map[string]interface{})
					link["self"] = fmt.Sprintf("%s/%s/relationships/%s", sc.Struct().Collection(), node.ID, field.NeuronName())
					link["related"] = fmt.Sprintf("%s/%s/%s", sc.Struct().Collection(), node.ID, field.NeuronName())
					links := Links(link)
					relLinks = &links
				}
			} else if c.Config.EncodeLinks {
				link := make(map[string]interface{})
				link["self"] = fmt.Sprintf("%s/%s/relationships/%s", sc.Struct().Collection(), node.ID, field.NeuronName())
				link["related"] = fmt.Sprintf("%s/%s/%s", sc.Struct().Collection(), node.ID, field.NeuronName())
				links := Links(link)
				relLinks = &links
			}

			var relMeta *Meta
			if o != nil {
				if optionMeta, ok := o.RelationshipMeta[field.NeuronName()]; ok {
					relMeta = (*Meta)(&optionMeta)
				}
			}

			if isSlice {
				// get RelationshipManyNode
				relationship, err := visitRelationshipManyNode(c, fieldValue, primaryVal, field)
				if err != nil {
					return nil, err
				}

				relationship.Links = relLinks
				relationship.Meta = relMeta
				node.Relationships[field.NeuronName()] = relationship
			} else {
				// is to-one relationship
				if fieldValue.IsNil() {
					node.Relationships[field.NeuronName()] = &relationshipOneNode{Links: relLinks, Meta: relMeta}
					continue
				}

				relatedNode, err := visitRelationshipNode(c, fieldValue, primaryVal, field)
				if err != nil {
					return nil, err
				}

				relationship := &relationshipOneNode{
					Data:  relatedNode,
					Links: relLinks,
					Meta:  relMeta,
				}
				node.Relationships[field.NeuronName()] = relationship
			}
		}
	}

	if o != nil && o.Link == DefaultLink {
		links := make(map[string]interface{})
		links["self"] = path.Join(o.LinkURL, sc.Struct().Collection(), node.ID)
		linksObj := Links(links)
		node.Links = &(linksObj)
	} else if value, ok := sc.StoreGet(encodeLinksCtxKey); ok {
		if encodeLinks, ok := value.(bool); ok && encodeLinks {
			links := make(map[string]interface{})

			links["self"] = path.Join(sc.Struct().Collection(), node.ID)
			linksObj := Links(links)
			node.Links = &(linksObj)
		}
	} else if c.Config.EncodeLinks {
		links := make(map[string]interface{})

		links["self"] = path.Join(sc.Struct().Collection(), node.ID)
		linksObj := Links(links)
		node.Links = &(linksObj)
	}
	return node, nil
}

func visitRelationshipManyNode(c *controller.Controller, manyValue, rootID reflect.Value, field *models.StructField) (*relationshipManyNode, error) {
	nodes := []*node{}

	for i := 0; i < manyValue.Len(); i++ {
		elemValue := manyValue.Index(i)
		node, err := visitRelationshipNode(c, elemValue, rootID, field)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)

	}

	return &relationshipManyNode{Data: nodes}, nil
}

func visitRelationshipNode(c *controller.Controller, value, rootID reflect.Value, field *models.StructField) (*node, error) {
	mStruct := field.RelatedModelStruct()
	prim := mStruct.PrimaryField()
	node := &node{Type: mStruct.Collection()}

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

func setNodePrimary(value reflect.Value, node *node) error {
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
		return errors.NewDetf(class.InternalEncodingUnsupportedID, "unsupported primary field type: '%s'", v.Type().Name())
	}
	return nil
}

func convertToSliceInterface(i *interface{}) ([]interface{}, error) {
	vals := reflect.ValueOf(*i)
	if vals.Kind() != reflect.Slice {
		return nil, errors.NewDet(class.InternalEncodingValue, "value is not a slice")
	}
	var response []interface{}
	for x := 0; x < vals.Len(); x++ {
		response = append(response, vals.Index(x).Interface())
	}
	return response, nil
}
