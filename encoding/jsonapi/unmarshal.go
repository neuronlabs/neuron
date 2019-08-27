package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
	ctrl "github.com/neuronlabs/neuron-core/controller"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
	"github.com/neuronlabs/neuron-core/query"

	"github.com/neuronlabs/neuron-core/internal/controller"
	"github.com/neuronlabs/neuron-core/internal/models"
)

// Unmarshal unmarshals the incoming reader stream into provided value 'v'.
// The model of the value 'v' should already be registered in the default controller.
func Unmarshal(r io.Reader, v interface{}) error {
	_, err := unmarshal((*controller.Controller)(ctrl.Default()), r, v, false)
	return err
}

// UnmarshalC unmarshals the incoming reader stream 'r' into provided model
// 'v' assuming that it is already registered within the controller 'c'
func UnmarshalC(c *ctrl.Controller, r io.Reader, v interface{}) error {
	_, err := unmarshal((*controller.Controller)(c), r, v, false)
	return err
}

// UnmarshalErrors unmarshals the jsonapi errors from the provided input 'r'.
func UnmarshalErrors(r io.Reader) (*ErrorsPayload, error) {
	payload := &ErrorsPayload{}
	err := json.NewDecoder(r).Decode(payload)
	return payload, err
}

// func UnmarshalPayload(io.Reader, payload Payloader) error {

// 	return nil
// }

// UnmarshalSingleScopeC unmarshals the value from the reader and creates new scope.
// Provided model argument may be a value of the Model i.e. &Model{} or a mapping.ModelStruct.
func UnmarshalSingleScopeC(c *ctrl.Controller, r io.Reader, model interface{}) (*query.Scope, error) {
	s, err := unmarshalScopeOne((*controller.Controller)(c), r, model, true)
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalSingleScope unmarshals the value from the reader and creates the scope for the default controller.
// Provided model argument may be a value of the Model i.e. &Model{} or a mapping.ModelStruct.
func UnmarshalSingleScope(r io.Reader, model interface{}) (*query.Scope, error) {
	s, err := unmarshalScopeOne((controller.Default()), r, model, true)
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalManyScope unmarshals the scope of multiple values for the default controller and given model struct.
// Provided model argument may be a value of the slice of models i.e. &[]*Model{} or a mapping.ModelStruct.
func UnmarshalManyScope(r io.Reader, model interface{}) (*query.Scope, error) {
	s, err := unmarshalScopeMany(controller.Default(), r, model)
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalManyScopeC unmarshals the scope of multiple values for the given controller and  model struct.
// Provided model argument may be a value of the slice of models i.e. &[]*Model{} or a mapping.ModelStruct.
func UnmarshalManyScopeC(c *ctrl.Controller, r io.Reader, model interface{}) (*query.Scope, error) {
	s, err := unmarshalScopeMany((*controller.Controller)(c), r, model)
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalWithSelected unmarshals the value from io.Reader and returns the selected fields if the value is a single.
func UnmarshalWithSelected(r io.Reader, v interface{}) ([]*mapping.StructField, error) {
	fields, err := unmarshal((*controller.Controller)(ctrl.Default()), r, v, true)
	if err != nil {
		return nil, err
	}

	mappingFields := make([]*mapping.StructField, len(fields))

	for i, field := range fields {
		mappingFields[i] = (*mapping.StructField)(field)
	}

	return mappingFields, nil
}

// UnmarshalWithSelectedC unmarshals the value from io.Reader and returns the selected fields if the value is a single.
func UnmarshalWithSelectedC(c *ctrl.Controller, r io.Reader, v interface{}) ([]*mapping.StructField, error) {
	fields, err := unmarshal((*controller.Controller)(c), r, v, true)
	if err != nil {
		return nil, err
	}

	mappingFields := make([]*mapping.StructField, len(fields))

	for i, field := range fields {
		mappingFields[i] = (*mapping.StructField)(field)
	}

	return mappingFields, nil
}

func unmarshalScopeMany(c *controller.Controller, in io.Reader, model interface{}) (*query.Scope, error) {
	return unmarshalScope(c, in, model, true, false)
}

func unmarshalScopeOne(c *controller.Controller, in io.Reader, model interface{}, addSelectedFields bool) (*query.Scope, error) {
	return unmarshalScope(c, in, model, false, addSelectedFields)
}

func unmarshalScope(c *controller.Controller, in io.Reader, model interface{}, useMany, usedFields bool) (*query.Scope, error) {
	var (
		mStruct  *models.ModelStruct
		hasValue bool
	)

	switch m := model.(type) {
	case *models.ModelStruct:
		mStruct = m
	case *mapping.ModelStruct:
		mStruct = (*models.ModelStruct)(m)
	default:
		var err error
		mStruct, err = c.ModelMap().GetModelStruct(model)
		if err != nil {
			return nil, err
		}
		hasValue = true
	}

	var scopeValue interface{}
	if hasValue {
		t := reflect.TypeOf(model)
		if t.Kind() != reflect.Ptr {
			return nil, errors.NewDet(class.EncodingUnmarshalInvalidOutput, "output value is not addressable")
		}
		t = t.Elem()
		switch t.Kind() {
		case reflect.Struct:
			if useMany {
				scopeValue = reflect.New(reflect.SliceOf(reflect.New(mStruct.Type()).Type())).Interface()
			} else {
				scopeValue = model
			}
		case reflect.Slice:
			if !useMany {
				scopeValue = reflect.New(mStruct.Type()).Interface()
			} else {
				scopeValue = model
			}
		}
	} else {
		if useMany {
			scopeValue = reflect.New(reflect.SliceOf(reflect.New(mStruct.Type()).Type())).Interface()
		} else {
			scopeValue = reflect.New(mStruct.Type()).Interface()
		}
	}

	fields, err := unmarshal(c, in, scopeValue, usedFields)
	if err != nil {
		return nil, err
	}

	s, err := query.NewC((*ctrl.Controller)(c), scopeValue)
	if err != nil {
		return nil, err
	}

	if usedFields && !useMany {
		for _, field := range fields {
			if err = s.AppendSelectedFields(field); err != nil {
				return nil, err
			}
		}
	}
	return s, nil

}

func unmarshal(c *controller.Controller, in io.Reader, model interface{}, usedFields bool) ([]*models.StructField, error) {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.EncodingUnmarshalInvalidOutput, "invalid output 'model'")
	}

	mStruct, err := c.ModelMap().GetModelStruct(model)
	if err != nil {
		return nil, err
	}

	switch t.Elem().Kind() {
	case reflect.Struct:
		one, err := unmarshalPayload(c, in)
		if err != nil {
			return nil, err
		}

		var fields []*models.StructField
		if usedFields {
			if one.Data.ID != "" {
				fields = append(fields, mStruct.PrimaryField())
			}

			for attr := range one.Data.Attributes {
				field, ok := mStruct.Attribute(attr)
				if ok {
					fields = append(fields, field)
				}
			}

			for relation := range one.Data.Relationships {
				field, ok := mStruct.RelationshipField(relation)
				if ok {
					fields = append(fields, field)
				}
			}
		}

		// check if the collection types match
		if one.Data.Type != mStruct.Collection() {
			log.Debug2f("Unmarshaled data of collection: '%s' for model's collection: '%s'", one.Data.Type, mStruct.Collection())
			err := errors.NewDet(class.EncodingUnmarshalCollection, "unmarshaling collection doesn't match the output value type")
			err.SetDetailsf("One of the input collection: '%s' doesn't match the root collection: '%s'", one.Data.Type, mStruct.Collection())
			return nil, err
		}

		if one.Included != nil {
			log.Debug2("Payload contains Included values.")
			includedMap := make(map[string]*node)
			for _, included := range one.Included {
				key := fmt.Sprintf("%s,%s", included.Type, included.ID)
				includedMap[key] = included
			}
			err = unmarshalNode(c, one.Data, reflect.ValueOf(model), &includedMap)
			if err != nil {
				return nil, err
			}
		} else {
			err = unmarshalNode(c, one.Data, reflect.ValueOf(model), nil)
			if err != nil {
				return nil, err
			}
		}
		return fields, nil
	case reflect.Slice:
		t = t.Elem().Elem()
		if t.Kind() != reflect.Ptr {
			return nil, errors.NewDet(class.EncodingUnmarshalInvalidType, "provided invalid input data type")
		}

		mStruct := c.ModelMap().Get(t.Elem())
		if mStruct == nil {
			return nil, err
		}

		many, err := unmarshalManyPayload(c, in)
		if err != nil {
			return nil, err
		}

		// get new slice instance
		models := reflect.New(reflect.SliceOf(reflect.New(mStruct.Type()).Type())).Elem()

		// will be populated from the "data"
		includedMap := map[string]*node{} // will be populate from the "included"

		if many.Included != nil {
			for _, included := range many.Included {
				key := fmt.Sprintf("%s,%s", included.Type, included.ID)
				includedMap[key] = included
			}
		}

		for _, data := range many.Data {
			model := reflect.New(mStruct.Type())
			err = unmarshalNode(c, data, model, &includedMap)
			if err != nil {
				return nil, err
			}
			models = reflect.Append(models, model)
		}
		input := reflect.ValueOf(model)
		input.Elem().Set(models)
		return nil, nil
	default:
		return nil, errors.NewDet(class.EncodingUnmarshalInvalidType, "provided invalid input data type")
	}
}

func unmarshalPayload(c *controller.Controller, in io.Reader) (*SinglePayload, error) {
	payload := new(SinglePayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, unmarshalHandleDecodeError(c, err)
	}

	if payload.Data == nil {
		err := errors.NewDet(class.EncodingUnmarshalNoData, "no data found from the reader")
		err.SetDetails("Specified request contains no data.")
		return nil, err
	}
	return payload, nil
}

func unmarshalManyPayload(c *controller.Controller, in io.Reader) (*ManyPayload, error) {
	payload := new(ManyPayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, unmarshalHandleDecodeError(c, err)
	}

	// if the data is nil throw no error data.
	if payload.Data == nil {
		err := errors.NewDet(class.EncodingUnmarshalNoData, "no data found from the reader")
		err.SetDetails("Specified request contains no data.")
		return nil, err
	}

	return payload, nil
}

func unmarshalHandleDecodeError(c *controller.Controller, err error) error {
	// handle the incoming error
	switch e := err.(type) {
	case errors.DetailedError:
		return err
	case *json.SyntaxError:
		err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "syntax error")
		err.SetDetailsf("Document syntax error: '%s'. At data offset: '%d'", e.Error(), e.Offset)
		return err
	case *json.UnmarshalTypeError:
		if e.Type == reflect.TypeOf(SinglePayload{}) || e.Type == reflect.TypeOf(ManyPayload{}) {
			err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "invalid jsonapi document syntax")
			return err
		}
		err := errors.NewDet(class.EncodingUnmarshalInvalidType, "invalid field type")

		var fieldType string
		switch e.Field {
		case "id", "type", "client-id":
			fieldType = e.Type.String()
		case "relationships", "attributes", "links", "meta":
			fieldType = "object"
		}
		err.SetDetailsf("Invalid type for: '%s' field. Required type '%s' but is: '%v'", e.Field, fieldType, e.Value)
		return err
	default:
		if e == io.EOF || e == io.ErrUnexpectedEOF {
			err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "reader io.EOF occurred")
			err.SetDetailsf("invalid document syntax")
			return err
		}
		err := errors.NewDetf(class.EncodingUnmarshal, "unknown unmarshal error: %s", e.Error())
		return err
	}
}

func unmarshalNode(
	c *controller.Controller,
	data *node,
	modelValue reflect.Value,
	included *map[string]*node,
) error {
	var err error

	mStruct := c.ModelMap().GetByCollection(data.Type)
	if mStruct == nil {
		log.Debugf("Invalid collection type: %s", data.Type)
		err := errors.NewDet(class.EncodingUnmarshalCollection, "unmarshaling invalid collection name")
		err.SetDetailsf("Provided unsupported/unknown collection: '%s'", data.Type)
		return err
	}

	modelValue = modelValue.Elem()
	modelType := modelValue.Type()

	if modelType != mStruct.Type() {
		err := errors.NewDet(class.EncodingUnmarshalCollection, "unmarshaling collection name doesn't match the root struct")
		err.SetDetailsf("Unmarshaling collection: '%s' doesn't match root collection:'%s'", data.Type, mStruct.Collection())
		return err
	}

	// add primary key
	if data.ID != "" {
		primary := modelValue.FieldByIndex(mStruct.PrimaryField().ReflectField().Index)
		if err = unmarshalIDField(primary, data.ID); err != nil {
			return err
		}
	}

	if data.Attributes != nil {
		// Iterate over the data attributes
		for attrName, attrValue := range data.Attributes {
			modelAttr, ok := mStruct.Attribute(attrName)
			if !ok || (ok && modelAttr.IsHidden()) {
				if c.Config.StrictUnmarshalMode {
					err := errors.NewDet(class.EncodingUnmarshalUnknownField, "unknown field name")
					err.SetDetailsf("Provided unknown field name: '%s', for the collection: '%s'.", attrName, data.Type)
					return err
				}
				continue
			}

			fieldValue := modelValue.FieldByIndex(modelAttr.ReflectField().Index)
			err = unmarshalAttrFieldValue(c, modelAttr, fieldValue, attrValue)
			if err != nil {
				return err
			}
		}

	}

	if data.Relationships != nil {
		for relName, relValue := range data.Relationships {
			modelRel, ok := mStruct.RelationshipField(relName)
			if !ok || (ok && modelRel.IsHidden()) {
				if c.Config.StrictUnmarshalMode {
					err := errors.NewDet(class.EncodingUnmarshalUnknownField, "unknown field name")
					err.SetDetailsf("Provided unknown field name: '%s', for the collection: '%s'.", relName, data.Type)
					return err
				}
				continue
			}

			fieldValue := modelValue.FieldByIndex(modelRel.ReflectField().Index)
			// unmarshal the relationship
			if modelRel.FieldKind() == models.KindRelationshipMultiple {
				// to-many relationship
				relationship := new(relationshipManyNode)

				buf := bytes.NewBuffer(nil)
				err = json.NewEncoder(buf).Encode(data.Relationships[relName])
				if err != nil {
					log.Debugf("Controller.UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "invalid relationship format")
					err.SetDetailsf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				err = json.NewDecoder(buf).Decode(relationship)
				if err != nil {
					log.Debugf("Controller.UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "invalid relationship format")
					err.SetDetailsf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				data := relationship.Data
				models := reflect.New(fieldValue.Type()).Elem()

				for _, n := range data {
					m := reflect.New(fieldValue.Type().Elem().Elem())
					if err = unmarshalNode(c, fullNode(n, included), m, included); err != nil {
						log.Debugf("unmarshalNode.RelationshipMany - unmarshalNode failed. %v", err)
						return err
					}

					models = reflect.Append(models, m)
				}
				fieldValue.Set(models)
			} else if modelRel.FieldKind() == models.KindRelationshipSingle {
				relationship := new(relationshipOneNode)
				buf := bytes.NewBuffer(nil)

				if err = json.NewEncoder(buf).Encode(relValue); err != nil {
					log.Debugf("Controller.UnmarshalNode.relationshipSingle json.Encode failed. %v", err)
					err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "invalid relationship format")
					err.SetDetailsf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				if err = json.NewDecoder(buf).Decode(relationship); err != nil {
					log.Debugf("Controller.UnmarshalNode.RelationshipSingel json.Decode failed. %v", err)
					err := errors.NewDet(class.EncodingUnmarshalInvalidFormat, "invalid relationship format")
					err.SetDetailsf("The value for the relationship: '%s' is of invalid form.", relName)
					return err
				}

				if relationship.Data == nil {
					continue
				}

				m := reflect.New(fieldValue.Type().Elem())
				if err := unmarshalNode(c, fullNode(relationship.Data, included), m, included); err != nil {
					return err
				}

				fieldValue.Set(m)
			}
		}
	}
	return nil
}

func unmarshalAttrFieldValue(
	c *controller.Controller,
	modelAttr *models.StructField,
	fieldValue reflect.Value,
	attrValue interface{},
) (err error) {
	v := reflect.ValueOf(attrValue)
	fieldType := modelAttr.ReflectField()
	baseType := modelAttr.BaseType()

	if modelAttr.IsSlice() || modelAttr.IsArray() {
		var sliceValue reflect.Value
		log.Debug2f("Field Value: %s, modelAttr: %s", fieldValue.Type(), modelAttr.NeuronName())
		sliceValue, err = unmarshalSliceValue(c, modelAttr, v, fieldType.Type, baseType)
		if err != nil {
			return err
		}

		log.Debug2f("IsValid: %v", fieldValue.IsValid())
		log.Debug2f("IsValid: %v", sliceValue.IsValid())

		fieldValue.Set(sliceValue)
	} else if modelAttr.IsMap() {
		var mapValue reflect.Value
		mapValue, err = unmarshalMapValue(c, modelAttr, v, fieldType.Type, baseType)
		if err != nil {
			return err
		}
		fieldValue.Set(mapValue)
	} else {
		var resultValue reflect.Value
		resultValue, err = unmarshalSingleFieldValue(c, modelAttr, v, baseType)
		if err != nil {
			return err
		}

		fieldValue.Set(resultValue)
	}
	return nil
}

func unmarshalMapValue(
	c *controller.Controller,
	modelAttr *models.StructField,
	v reflect.Value,
	currentType reflect.Type,
	baseType reflect.Type,
) (reflect.Value, error) {
	var (
		mapValue reflect.Value
		err      error
		isPtr    bool
	)

	mpType := currentType
	if mpType.Kind() == reflect.Ptr {
		isPtr = true
		mpType = mpType.Elem()
	}

	if mpType.Kind() != reflect.Map {
		log.Errorf("Invalid currentType: '%s'. The field should be a Map. StructField: '%s'", currentType.String(), modelAttr.Name())
		err := errors.NewDet(class.InternalEncodingModelFieldType, "the map field type should be a map.")
		return reflect.Value{}, err
	}

	if isPtr {
		if v.IsNil() {
			mapValue = v
			return mapValue, nil
		}
	}

	if v.Kind() != reflect.Map {
		err := errors.NewDet(class.EncodingUnmarshalInvalidType, "map field is of invalid type")
		err.SetDetailsf("Field: '%s' should contain a value of object/map type.", modelAttr.NeuronName())
		return reflect.Value{}, err
	}

	mapValue = reflect.New(currentType).Elem()
	mapValue.Set(reflect.MakeMap(currentType))

	// Get the map values type
	nestedType := mpType.Elem()
	if nestedType.Kind() == reflect.Ptr {
		nestedType = nestedType.Elem()
	}

	for _, key := range v.MapKeys() {
		log.Debugf("Map Key: %v", key.Interface())
		value := v.MapIndex(key)

		if value.Kind() == reflect.Interface && value.IsValid() {
			value = value.Elem()
		}

		log.Debugf("Value: %v. Kind: %v", value, value.Kind())

		var nestedValue reflect.Value
		switch nestedType.Kind() {
		case reflect.Slice, reflect.Array:
			log.Debugf("Nested Slice within map field: %v", modelAttr.Name())
			nestedValue, err = unmarshalSliceValue(c, modelAttr, value, mpType.Elem(), baseType)
			if err != nil {
				return reflect.Value{}, err
			}
		default:
			log.Debugf("Map ModelAttr: '%s' has basePtr: '%v'", modelAttr.Name(), modelAttr.IsBasePtr())
			nestedValue, err = unmarshalSingleFieldValue(c, modelAttr, value, baseType)
			if err != nil {
				return reflect.Value{}, err
			}

			log.Debugf("Map ModelAttr: '%s' nestedValue: '%v'", nestedValue.Type().String())
		}
		mapValue.SetMapIndex(key, nestedValue)
	}

	return mapValue, nil
}

func unmarshalSliceValue(
	c *controller.Controller,
	modelAttr *models.StructField, // modelAttr is the attribute StructField
	v reflect.Value, // v is the current field value
	currentType reflect.Type,
	baseType reflect.Type, // baseType is the reflect.Type that is the base of the slice
) (reflect.Value, error) {
	var (
		fieldValue reflect.Value
		isPtr      bool
		capSize    int
		err        error
	)

	slType := currentType
	if slType.Kind() == reflect.Ptr {
		isPtr = true
		slType = slType.Elem()
	}

	if slType.Kind() != reflect.Array && slType.Kind() != reflect.Slice {
		err := errors.NewDet(class.EncodingUnmarshalInvalidType, "slice field should be an array or slice")
		err.SetDetailsf("Field: '%s' should be an array", modelAttr.NeuronName())
		log.Errorf("Attribute: %v, Err: %v", modelAttr.Name(), err)
		return reflect.Value{}, err
	}

	if slType.Kind() == reflect.Array {
		capSize = currentType.Len()
	} else {
		capSize = -1
	}

	if v.Kind() == reflect.Interface && v.IsValid() {
		v = v.Elem()
	}

	// check what is the current value kind
	// v Kind can't be an array -> the json unmarshaller won't unmarshal it into an array
	if v.Kind() != reflect.Slice {
		log.Debugf("Value within the slice is not of slice type. %v", v.Kind().String())
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		err := errors.NewDet(class.EncodingUnmarshalFieldValue, "slice value is not a slice")
		err.SetDetailsf("Field: '%s' the slice value is not a slice.", modelAttr.NeuronName())
		return reflect.Value{}, err
	}

	// check the lengths - it does matter if the provided type is an array
	if capSize != -1 {
		if v.Len() > capSize {
			err := errors.NewDet(class.EncodingUnmarshalValueOutOfRange, "field value length is out of the possible size")
			err.SetDetailsf("Field: '%s' the slice value is too long. The maximum capacity is: '%d'", modelAttr.NeuronName(), capSize)
			return reflect.Value{}, err
		}
	}

	// Create the field value
	if currentType.Kind() == reflect.Ptr {
		fieldValue = reflect.New(currentType.Elem()).Elem()
	} else {
		fieldValue = reflect.New(currentType).Elem()
	}

	// check if current type contains other slices
	slType = slType.Elem()

	var hasNestedSlice bool
	nestedType := slType
	if slType.Kind() == reflect.Ptr {
		nestedType = nestedType.Elem()
	}

	if nestedType.Kind() == reflect.Slice || nestedType.Kind() == reflect.Array {
		hasNestedSlice = true
	}

	for i := 0; i < v.Len(); i++ {
		vElem := v.Index(i)
		var nestedValue reflect.Value

		// if the value is a nestedSlice unmarshal it into the slice
		if hasNestedSlice {
			nestedValue, err = unmarshalSliceValue(c, modelAttr, vElem, slType, baseType)
			if err != nil {
				log.Debug("Nested Slice Value failed: %v", err)
				return reflect.Value{}, err
			}
			// otherwise this must be a single value
		} else {
			nestedValue, err = unmarshalSingleFieldValue(c, modelAttr, vElem, baseType)
			if err != nil {
				log.Debugf("NestedSlice -> unmarshalSingleFieldValue failed. %v. Velem: %v", err)
				return reflect.Value{}, err
			}
		}

		// if the value was an Array set it's value at index
		if fieldValue.Type().Kind() == reflect.Array {
			fieldValue.Index(i).Set(nestedValue)
		} else {
			// otherwise the value is a slice append new value
			fieldValue.Set(reflect.Append(fieldValue, nestedValue))
		}
	}

	// if the value was a Ptr type it's address has to be returned
	if isPtr {
		fieldValue = fieldValue.Addr()
	}

	return fieldValue, nil
}

// unmarshalSingleFieldValue gets the single field value from the unmarshaled data.
func unmarshalSingleFieldValue(
	c *controller.Controller,
	modelAttr *models.StructField,
	// v is the incoming value to unmarshal
	v reflect.Value,
	// fieldType is the base Type of the model field
	baseType reflect.Type,
) (reflect.Value, error) {
	// check if the model attribute field is based with pointer

	if modelAttr.IsBasePtr() {
		if !v.IsValid() {
			return reflect.Zero(reflect.PtrTo(baseType)), nil
		}
		if v.Interface() == nil {
			return reflect.Zero(reflect.PtrTo(baseType)), nil
		}
	}

	if v.Kind() == reflect.Interface && v.IsValid() {
		v = v.Elem()
		// if the value was an uinitialized interface{} then v.Elem
		// would be invalid.
	}

	if !v.IsValid() {
		err := errors.NewDet(class.EncodingUnmarshalFieldValue, "invalid field value")
		err.SetDetailsf("Field: %v' has invalid value.", modelAttr.NeuronName())
		return reflect.Value{}, err
	}

	if modelAttr.IsTime() {
		if modelAttr.IsISO8601() {
			// on ISO8601 the incoming value must be a string
			var tm string
			if v.Kind() == reflect.String {
				tm = v.Interface().(string)
			} else {
				err := errors.NewDet(class.EncodingUnmarshalInvalidTime, "invalid ISO8601 time field")
				err.SetDetailsf("Time field: '%s' has invalid formatting.", modelAttr.NeuronName())
				return reflect.Value{}, err
			}

			// parse the string time with iso formatting
			t, err := time.Parse(ISO8601TimeFormat, tm)
			if err != nil {
				err := errors.NewDet(class.EncodingUnmarshalInvalidTime, "invalid ISO8601 time field")
				err.SetDetailsf("Time field: '%s' has invalid formatting.", modelAttr.NeuronName())
				return reflect.Value{}, err
			}

			if modelAttr.IsBasePtr() {
				return reflect.ValueOf(&t), nil
			}
			return reflect.ValueOf(t), nil
		}

		// if the time is not iso8601 it must be an integer - unix time
		var at int64

		// the integer may be a float64 or an int
		switch v.Kind() {
		case reflect.Float64:
			at = int64(v.Interface().(float64))
		case reflect.Int, reflect.Int64:
			at = v.Int()
		case reflect.Uint, reflect.Uint64:
			at = int64(v.Uint())
		default:
			err := errors.NewDet(class.EncodingUnmarshalInvalidTime, "invalid time field format")
			err.SetDetailsf("Time field: '%s' has invalid formatting.", modelAttr.NeuronName())
			return reflect.Value{}, err
		}

		t := time.Unix(at, 0)
		if modelAttr.IsBasePtr() {
			return reflect.ValueOf(&t), nil
		}
		return reflect.ValueOf(t), nil
	} else if modelAttr.IsNestedStruct() {
		value := v.Interface()
		return unmarshalNestedStructValue(c, modelAttr.Nested(), value)
	} else if v.Kind() == reflect.Float64 {
		// if the incoming value v is of kind float64
		// get the float value from the incoming 'v'
		floatValue := v.Interface().(float64)

		// The field may or may not be a pointer to a numeric; the kind var
		// will not contain a pointer type
		var numericValue reflect.Value
		switch baseType.Kind() {
		case reflect.Int:
			n := int(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int8:
			n := int8(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int16:
			n := int16(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int32:
			n := int32(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int64:
			n := int64(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Uint:
			n := uint(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Uint8:
			n := uint8(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Uint16:
			n := uint16(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Uint32:
			n := uint32(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Uint64:
			n := uint64(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Float32:
			n := float32(floatValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Float64:
			n := floatValue
			numericValue = reflect.ValueOf(&n)
		default:
			log.Debugf("Unknown field number type: '%v'", baseType.String())
			err := errors.NewDet(class.EncodingUnmarshalInvalidType, "invalid field type")
			err.SetDetailsf("Field: '%s' has invalid type provided.", modelAttr.NeuronName())
			return reflect.Value{}, err
		}

		// if the field was ptr
		if modelAttr.IsBasePtr() {
			return numericValue, nil
		}

		return numericValue.Elem(), nil

		// somehow map values are unmarshaled as int, not as float64, even though UseNumber is not
		//set
	} else if v.Kind() == reflect.Int {
		// get the float value from the incoming 'v'
		intValue := v.Interface().(int)

		// The field may or may not be a pointer to a numeric; the kind var
		// will not contain a pointer type
		var numericValue reflect.Value
		switch baseType.Kind() {
		case reflect.Int:
			n := int(intValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int8:
			n := int8(intValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int16:
			n := int16(intValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int32:
			n := int32(intValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Int64:
			n := int64(intValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if intValue < 0 {
				err := errors.NewDet(class.EncodingUnmarshalInvalidType, "invalid field value type")
				err.SetDetailsf("Field: '%s'. Provided value: '%d' is not an unsigned integer.", modelAttr.NeuronName(), intValue)
				return reflect.Value{}, err
			}

			switch baseType.Kind() {
			case reflect.Uint:
				n := uint(intValue)
				numericValue = reflect.ValueOf(&n)
			case reflect.Uint8:
				n := uint8(intValue)
				numericValue = reflect.ValueOf(&n)
			case reflect.Uint16:
				n := uint16(intValue)
				numericValue = reflect.ValueOf(&n)
			case reflect.Uint32:
				n := uint32(intValue)
				numericValue = reflect.ValueOf(&n)
			case reflect.Uint64:
				n := uint64(intValue)
				numericValue = reflect.ValueOf(&n)
			}
		case reflect.Float32:
			n := float32(intValue)
			numericValue = reflect.ValueOf(&n)
		case reflect.Float64:
			n := float64(intValue)
			numericValue = reflect.ValueOf(&n)
		default:
			log.Debugf("Unknown field number type: '%v'", baseType.String())
			err := errors.NewDet(class.EncodingUnmarshalInvalidType, "invalid numerical field type")
			err.SetDetailsf("Field: '%s' has invalid type.", modelAttr.NeuronName())
			return reflect.Value{}, err
		}

		// if the field was ptr
		if modelAttr.IsBasePtr() {
			return numericValue, nil
		}
		return numericValue.Elem(), nil
	}

	var concreteVal reflect.Value
	value := v.Interface()
	valueNotAllowed := func() error {
		err := errors.NewDet(class.EncodingUnmarshalFieldValue, "invalid field value")
		err.SetDetailsf("The value: '%v' for the field: '%s' is not allowed", value, modelAttr.NeuronName())
		return err
	}

	switch tv := value.(type) {
	case string:
		if baseType.Kind() == reflect.String {
			ptrVal := &tv
			concreteVal = reflect.ValueOf(ptrVal).Elem()
		} else {
			return reflect.Value{}, valueNotAllowed()
		}
	case int:
		// switch the base type if it fits
		switch baseType.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetUint(uint64(tv))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetInt(int64(tv))
		case reflect.Float32, reflect.Float64:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetFloat(float64(tv))
		default:
			return reflect.Value{}, valueNotAllowed()
		}
	case int64:
		// switch the base type if it fits
		switch baseType.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetUint(uint64(tv))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetInt(int64(tv))
		case reflect.Float32, reflect.Float64:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetFloat(float64(tv))
		default:
			return reflect.Value{}, valueNotAllowed()
		}
	case float64:
		// switch the base type if it fits
		switch baseType.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetUint(uint64(tv))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetInt(int64(tv))
		case reflect.Float32, reflect.Float64:
			concreteVal = reflect.New(baseType).Elem()
			concreteVal.SetFloat(float64(tv))
		default:
			return reflect.Value{}, valueNotAllowed()
		}
	case bool:
		if baseType.Kind() == reflect.Bool {
			ptrVal := &tv
			concreteVal = reflect.ValueOf(ptrVal).Elem()
		} else {
			return reflect.Value{}, valueNotAllowed()
		}
	case complex128:
		if baseType.Kind() == reflect.Complex128 {
			ptrVal := &tv
			concreteVal = reflect.ValueOf(ptrVal).Elem()
		} else {
			return reflect.Value{}, valueNotAllowed()
		}

	case complex64:
		if baseType.Kind() == reflect.Complex64 {
			ptrVal := &tv
			concreteVal = reflect.ValueOf(ptrVal).Elem()
		} else {
			return reflect.Value{}, valueNotAllowed()
		}
	default:
		// As a final catch-all, ensure types line up to avoid a runtime panic.
		log.Debugf("Invalid Value: %v with Kind: %v, should be: '%v' for field %v", v, v.Kind(), baseType.String(), modelAttr.Name())
		err := errors.NewDet(class.ModelFieldType, "invalid field type")
		err.SetDetailsf("Field: '%s' has invalid field type.", modelAttr.NeuronName())
		return reflect.Value{}, err
	}
	if modelAttr.IsBasePtr() {
		return concreteVal.Addr(), nil
	}
	return concreteVal, nil
}

func unmarshalIDField(fieldValue reflect.Value, dataValue string) error {
	v := reflect.ValueOf(dataValue)

	var kind reflect.Kind
	if fieldValue.Kind() == reflect.Ptr {
		kind = fieldValue.Type().Elem().Kind()
	} else {
		kind = fieldValue.Type().Kind()
	}

	if kind == reflect.String {
		assign(fieldValue, v)
		return nil
	}

	// Value was not a string... only other supported type was a numeric,
	// which would have been sent as a float value.
	floatValue, err := strconv.ParseFloat(dataValue, 64)
	if err != nil {
		// Could not convert the value in the "id" attr to a float
		return errors.NewDet(class.EncodingUnmarshalInvalidID, "invalid id field format")
	}

	// Convert the numeric float to one of the supported ID numeric types
	// (int[8,16,32,64] or uint[8,16,32,64])
	var idValue reflect.Value
	switch kind {
	case reflect.Int:
		n := int(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Int8:
		n := int8(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Int16:
		n := int16(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Int32:
		n := int32(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Int64:
		n := int64(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Uint:
		n := uint(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Uint8:
		n := uint8(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Uint16:
		n := uint16(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Uint32:
		n := uint32(floatValue)
		idValue = reflect.ValueOf(&n)
	case reflect.Uint64:
		n := uint64(floatValue)
		idValue = reflect.ValueOf(&n)
	default:
		// We had a JSON float (numeric), but our field was not one of the
		// allowed numeric types
		err := errors.NewDet(class.EncodingUnmarshalInvalidID, "unmarshaling invalid primary field type")
		err.SetDetailsf("Invalid primary field value: '%s'", dataValue)
		return err
	}
	assign(fieldValue, idValue)
	return nil
}

func unmarshalNestedStructValue(c *controller.Controller, n *models.NestedStruct, value interface{}) (reflect.Value, error) {
	mp, ok := value.(map[string]interface{})
	if !ok {
		err := errors.NewDet(class.EncodingUnmarshalFieldValue, "invalid field value")
		err.SetDetailsf("Invalid field value for the subfield within attribute: '%s'", n.Attr().NeuronName())
		return reflect.Value{}, err
	}

	result := reflect.New(n.Type())
	resElem := result.Elem()
	for mpName, mpVal := range mp {
		nestedField, ok := models.NestedStructFields(n)[mpName]
		if !ok {
			if !c.Config.StrictUnmarshalMode {
				continue
			}
			err := errors.NewDet(class.EncodingUnmarshalUnknownField, "nested field not found")
			err.SetDetailsf("No subfield named: '%s' within attr: '%s'", mpName, n.Attr().NeuronName())
			return reflect.Value{}, err
		}

		fieldValue := resElem.FieldByIndex(nestedField.StructField().ReflectField().Index)

		err := unmarshalAttrFieldValue(c, nestedField.StructField(), fieldValue, mpVal)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	if n.StructField().Self().IsBasePtr() {
		log.Debugf("NestedStruct: '%v' isBasePtr. Attr: '%s'", result.Type(), n.Attr().Name())
		return result, nil
	}
	log.Debugf("NestedStruct: '%v' isNotBasePtr. Attr: '%s'", resElem.Type(), n.Attr().Name())
	return resElem, nil
}

func fullNode(n *node, included *map[string]*node) *node {
	includedKey := fmt.Sprintf("%s,%s", n.Type, n.ID)

	if included != nil && (*included)[includedKey] != nil {
		return (*included)[includedKey]
	}

	return n
}

// assign will take the value specified and assign it to the field; if
// field is expecting a ptr assign will assign a ptr.
func assign(field, value reflect.Value) {
	if field.Kind() == reflect.Ptr {
		field.Set(value)
	} else {
		field.Set(reflect.Indirect(value))
	}
}
