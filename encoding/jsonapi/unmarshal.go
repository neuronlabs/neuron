package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	ctrl "github.com/neuronlabs/neuron/controller"
	aerrors "github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/models"
	iscope "github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/query"

	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"strconv"
	"time"
)

const (
	unsuportedStructTagMsg = "Unsupported jsonapi tag annotation, %s"
)

// Unmarshal unmarshals the incoming reader stream into provided
// assuming that the value 'v' is already registered in the default controller
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

// UnmarshalSingleScopeC unmarshals the value from the reader and creates new scope
func UnmarshalSingleScopeC(c *ctrl.Controller, r io.Reader, m *mapping.ModelStruct) (*query.Scope, error) {
	s, err := unmarshalScopeOne((*controller.Controller)(c), r, (*models.ModelStruct)(m), true)
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalSingleScope unmarshals the value from the reader and creates the scope for the default controller
func UnmarshalSingleScope(r io.Reader, m *mapping.ModelStruct) (*query.Scope, error) {
	s, err := unmarshalScopeOne((controller.Default()), r, (*models.ModelStruct)(m), true)
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalManyScope unmarshals the scope of multiple values for the default controller and given model struct
func UnmarshalManyScope(r io.Reader, m *mapping.ModelStruct) (*query.Scope, error) {
	s, err := unmarshalScopeMany(controller.Default(), r, (*models.ModelStruct)(m))
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalManyScopeC unmarshals the scope of multiple values for the given controller and  model struct
func UnmarshalManyScopeC(c *ctrl.Controller, r io.Reader, m *mapping.ModelStruct) (*query.Scope, error) {
	s, err := unmarshalScopeMany((*controller.Controller)(c), r, (*models.ModelStruct)(m))
	if err != nil {
		return nil, err
	}
	return (*query.Scope)(s), nil
}

// UnmarshalWithSelected unmarshals the value from io.Reader and returns the selected fields if the value is a single
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

// UnmarshalWithSelectedC unmarshals the value from io.Reader and returns the selected fields if the value is a single
func UnmarshalWithSelectedC(c *controller.Controller, r io.Reader, v interface{}) ([]*mapping.StructField, error) {
	fields, err := unmarshal(c, r, v, true)
	if err != nil {
		return nil, err
	}

	mappingFields := make([]*mapping.StructField, len(fields))

	for i, field := range fields {
		mappingFields[i] = (*mapping.StructField)(field)
	}

	return mappingFields, nil
}

func unmarshalScopeMany(c *controller.Controller, in io.Reader, model interface{}) (*iscope.Scope, error) {
	mStruct, err := c.ModelSchemas().GetModelStruct(model)
	if err != nil {
		return nil, err
	}
	return unmarshalScope(c, in, mStruct, true, false)
}

func unmarshalScopeOne(
	c *controller.Controller,
	in io.Reader,
	model interface{},
	addSelectedFields bool,
) (*iscope.Scope, error) {
	mStruct, ok := model.(*models.ModelStruct)
	if !ok {

		var err error
		mStruct, err = c.ModelSchemas().GetModelStruct(model)
		if err != nil {
			return nil, err
		}
	}

	return unmarshalScope(c, in, mStruct, false, addSelectedFields)
}

// UnmarshalWithRegister unmarshals the reader with the possibility to register provided model
// into the schemas
func unmarshalWithRegister(c *controller.Controller, in io.Reader, v interface{}) error {
	return nil
}

func unmarshalScope(
	c *controller.Controller,
	in io.Reader,
	mStruct *models.ModelStruct,
	useMany, usedFields bool,
) (*iscope.Scope, error) {

	var modelValue reflect.Value
	if useMany {
		modelValue = reflect.New(reflect.SliceOf(reflect.New(mStruct.Type()).Type()))
	} else {
		modelValue = reflect.New(mStruct.Type())
	}

	v := modelValue.Interface()
	fields, err := unmarshal(c, in, v, usedFields)
	if err != nil {
		return nil, err
	}

	sc := iscope.NewRootScope(mStruct)
	if useMany {
		sc.Value = v
	} else {
		sc.Value = v
	}

	sc.Store[internal.ControllerCtxKey] = c

	if usedFields && !useMany {
		for _, field := range fields {
			sc.AddSelectedField(field)
		}

	}
	return sc, nil

}

func unmarshal(
	c *controller.Controller,
	in io.Reader,
	model interface{},
	usedFields bool,
) ([]*models.StructField, error) {

	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return nil, internal.ErrUnexpectedType
	}

	mStruct, err := c.ModelSchemas().GetModelStruct(model)
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

		if one.Data.Type != mStruct.Collection() {
			log.Debugf("Unmarshaled data of collection: '%s' for model's collection: '%s'", one.Data.Type, mStruct.Collection())
			errObj := aerrors.ErrInvalidResourceName.Copy()
			errObj.Detail = fmt.Sprintf("The specified collection: '%s' is not recognized by the server.", one.Data.Type)
			return nil, errObj
		}

		if one.Included != nil {
			log.Debug("Payload contains Included values.")
			includedMap := make(map[string]*node)
			for _, included := range one.Included {
				key := fmt.Sprintf("%s,%s", included.Type, included.ID)
				includedMap[key] = included
			}
			err = unmarshalNode(c, one.Data, reflect.ValueOf(model), &includedMap)
		} else {
			err = unmarshalNode(c, one.Data, reflect.ValueOf(model), nil)
		}
		switch err {
		case internal.ErrUnknownFieldNumberType, internal.ErrInvalidTime, internal.ErrInvalidISO8601,
			internal.ErrInvalidType, internal.ErrBadJSONAPIID:
			errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
			errObj.Detail = err.Error()
			errObj.Err = err
			return nil, errObj
		default:
			if uErr, ok := err.(*json.UnmarshalFieldError); ok {
				errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid type for field: %v. This field is of type: %s.", uErr.Key, uErr.Type.String())
				errObj.Err = err
				return nil, errObj
			}
		}
		return fields, err
	case reflect.Slice:

		t = t.Elem().Elem()
		if t.Kind() != reflect.Ptr {
			return nil, internal.ErrUnexpectedType
		}
		mStruct, err := c.ModelSchemas().ModelByType(t.Elem())
		if mStruct == nil {
			return nil, internal.ErrModelNotMapped
		}

		many, err := unmarshalManyPayload(c, in)
		if err != nil {
			return nil, err
		}

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
	}
	return nil, nil
}

func unmarshalPayload(c *controller.Controller, in io.Reader) (*onePayload, error) {
	payload := new(onePayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, unmarshalHandleDecodeError(c, err)
	}

	if payload.Data == nil {
		errObj := aerrors.ErrInvalidJSONDocument.Copy()
		errObj.Detail = "Specified request contains no data."
		return nil, errObj
	}
	return payload, nil
}

func unmarshalManyPayload(c *controller.Controller, in io.Reader) (*manyPayload, error) {
	payload := new(manyPayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, unmarshalHandleDecodeError(c, err)
	}
	if payload.Data == nil {
		errObj := aerrors.ErrInvalidJSONDocument.Copy()
		errObj.Detail = "Specified request contains no data."
		return nil, errObj
	}
	return payload, nil
}

func unmarshalHandleDecodeError(c *controller.Controller, er error) error {
	if serr, ok := er.(*json.SyntaxError); ok {

		errObj := aerrors.ErrInvalidJSONDocument.Copy()
		errObj.Detail = fmt.Sprintf("Syntax Error: %s. At offset: %d.", er.Error(), serr.Offset)
		return errObj
	} else if uErr, ok := er.(*json.UnmarshalTypeError); ok {
		if uErr.Type == reflect.TypeOf(onePayload{}) ||
			uErr.Type == reflect.TypeOf(manyPayload{}) {
			errObj := aerrors.ErrInvalidJSONDocument.Copy()
			errObj.Detail = fmt.Sprintln("Invalid JSON document syntax.")
			return errObj
		} else {
			errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
			var fieldType string
			switch uErr.Field {
			case "id", "type", "client-id":
				fieldType = uErr.Type.String()
			case "relationships", "attributes", "links", "meta":
				fieldType = "object"
			}
			errObj.Detail = fmt.
				Sprintf("Invalid type for: '%s' field. It must be of '%s' type but is: '%v'",
					uErr.Field, fieldType, uErr.Value)
			return errObj
		}
	} else if er == io.EOF {
		errObj := aerrors.ErrInvalidJSONDocument.Copy()
		errObj.Detail = fmt.Sprint("Provided document is empty.")
		errObj.Err = er
		return errObj
	} else {
		log.Errorf("Unknown error occured while decoding the payload. %v", er)
		return er
	}

	return nil
}

func unmarshalNode(
	c *controller.Controller,
	data *node,
	modelValue reflect.Value,
	included *map[string]*node,
) (err error) {
	var schema *models.Schema
	schema, err = c.ModelSchemas().SchemaByType(modelValue.Type())
	if err != nil {
		log.Debugf("getSchemaByType failed: %v", err)
		return
	}

	mStruct := schema.ModelByCollection(data.Type)
	if mStruct == nil {
		log.Debugf("Invalid collection type: %s", data.Type)
		errObj := aerrors.ErrInvalidResourceName.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid collection name: '%s'", data.Type)
		err = errObj
		return
	}

	modelValue = modelValue.Elem()
	modelType := modelValue.Type()

	if modelType != mStruct.Type() {
		errObj := aerrors.ErrInvalidJSONDocument.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid collection name for given model: '%s'", data.Type)
		err = errObj
		return
	}

	// add primary key
	if data.ID != "" {
		primary := modelValue.FieldByIndex(mStruct.PrimaryField().ReflectField().Index)
		if err = unmarshalIDField(primary, data.ID); err != nil {
			return
		}
	}

	// if data.ClientID != "" {
	// 	if mStruct.clientID != nil {
	// 		clientID := modelValue.FieldByIndex(mStruct.clientID.refStruct.Index)
	// 		if err = unmarshalIDField(clientID, data.ClientID); err != nil {
	// 			return
	// 		}
	// 	} else {
	// 		err = ErrClientIDDisallowed
	// 		return
	// 	}
	// }
	if data.Attributes != nil {

		// Iterate over the data attributes
		for attrName, attrValue := range data.Attributes {
			modelAttr, ok := mStruct.Attribute(attrName)
			if !ok || (ok && modelAttr.IsHidden()) {
				if c.StrictUnmarshalMode {
					errObj := aerrors.ErrInvalidJSONDocument.Copy()
					errObj.Detail = fmt.Sprintf("The server doesn't allow unknown field names. Provided unknown field name: '%s'.", attrName)
					err = errObj
					return
				} else {
					continue
				}
			}

			fieldValue := modelValue.FieldByIndex(modelAttr.ReflectField().Index)

			err = unmarshalAttrFieldValue(c, modelAttr, fieldValue, attrValue)
			if err != nil {
				return
			}
		}

	}

	if data.Relationships != nil {
		for relName, relValue := range data.Relationships {
			modelRel, ok := mStruct.RelationshipField(relName)
			if !ok || (ok && modelRel.IsHidden()) {
				if c.StrictUnmarshalMode {
					errObj := aerrors.ErrInvalidJSONDocument.Copy()
					errObj.Detail = fmt.Sprintf("The server doesn't allow unknown field names. Provided unknown field name: '%s'.", relName)
					err = errObj
					return
				} else {
					continue
				}
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
					errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}
				err = json.NewDecoder(buf).Decode(relationship)
				if err != nil {
					log.Debugf("Controller.UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}

				data := relationship.Data

				models := reflect.New(fieldValue.Type()).Elem()

				for _, n := range data {
					m := reflect.New(fieldValue.Type().Elem().Elem())
					if err = unmarshalNode(
						c,
						fullNode(n, included),
						m,
						included,
					); err != nil {
						log.Debugf("unmarshalNode.RelationshipMany - unmarshalNode failed. %v", err)
						return
					}

					models = reflect.Append(models, m)
				}
				fieldValue.Set(models)
			} else if modelRel.FieldKind() == models.KindRelationshipSingle {
				relationship := new(relationshipOneNode)
				buf := bytes.NewBuffer(nil)

				if err = json.NewEncoder(buf).Encode(relValue); err != nil {
					log.Debugf("Controller.UnmarshalNode.relationshipSingle json.Encode failed. %v", err)
					errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}

				if err = json.NewDecoder(buf).Decode(relationship); err != nil {
					log.Debugf("Controller.UnmarshalNode.RelationshipSingel json.Decode failed. %v", err)
					errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}

				if relationship.Data == nil {
					continue
				}

				m := reflect.New(fieldValue.Type().Elem())
				if err := unmarshalNode(
					c,
					fullNode(relationship.Data, included),
					m,
					included,
				); err != nil {
					return err
				}

				fieldValue.Set(m)
			}
		}
	}
	return
}

func unmarshalAttrFieldValue(
	c *controller.Controller,
	modelAttr *models.StructField,
	fieldValue reflect.Value,
	attrValue interface{},
) (err error) {

	v := reflect.ValueOf(attrValue)

	fieldType := modelAttr.ReflectField()

	baseType := models.FieldBaseType(modelAttr)

	if modelAttr.IsSlice() || modelAttr.IsArray() {
		var sliceValue reflect.Value
		sliceValue, err = unmarshalSliceValue(c, modelAttr, v, fieldType.Type, baseType)
		if err != nil {
			return
		}
		fieldValue.Set(sliceValue)
	} else if modelAttr.IsMap() {
		var mapValue reflect.Value
		mapValue, err = unmarshalMapValue(c, modelAttr, v, fieldType.Type, baseType)
		if err != nil {
			return
		}
		fieldValue.Set(mapValue)
	} else {
		var resultValue reflect.Value
		resultValue, err = unmarshalSingleFieldValue(c, modelAttr, v, baseType)
		if err != nil {
			return
		}

		fieldValue.Set(resultValue)
	}

	return
}

func unmarshalMapValue(
	c *controller.Controller,
	modelAttr *models.StructField,
	v reflect.Value,
	currentType reflect.Type,
	baseType reflect.Type,
) (mapValue reflect.Value, err error) {
	mpType := currentType

	var isPtr bool
	if mpType.Kind() == reflect.Ptr {
		isPtr = true
		mpType = mpType.Elem()
	}

	if mpType.Kind() != reflect.Map {
		err = errors.Errorf("Invalid currentType: '%s'. The field should be a Map. StructField: '%s'", currentType.String(), modelAttr.Name())
		return
	}

	if isPtr {
		if v.IsNil() {
			mapValue = v
			return
		}
	}

	if v.Kind() != reflect.Map {
		errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
		errObj.Detail = fmt.Sprintf("Field: '%s' should contain a value of map type.", modelAttr.ApiName())
		err = errObj
		return
	}

	mapValue = reflect.New(currentType).Elem()
	mapValue.Set(reflect.MakeMap(currentType))

	// Get the map values type
	nestedType := mpType.Elem()

	if nestedType.Kind() == reflect.Ptr {
		nestedType = nestedType.Elem()
	}

	for _, key := range v.MapKeys() {
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
				return
			}
		default:
			log.Debugf("Map ModelAttr: '%s' has basePtr: '%v'", modelAttr.Name(), modelAttr.IsBasePtr())
			nestedValue, err = unmarshalSingleFieldValue(c, modelAttr, value, baseType)
			if err != nil {
				return
			}

			log.Debugf("Map ModelAttr: '%s' nestedValue: '%v'", nestedValue.Type().String())

		}
		mapValue.SetMapIndex(key, nestedValue)
	}
	return
}

func unmarshalSliceValue(
	c *controller.Controller,
	modelAttr *models.StructField, // modelAttr is the attribute StructField
	v reflect.Value, // v is the current field value
	currentType reflect.Type,
	baseType reflect.Type, // baseType is the reflect.Type that is the base of the slice
) (fieldValue reflect.Value, err error) {

	var (
		isPtr   bool
		capSize int
	)

	slType := currentType

	if slType.Kind() == reflect.Ptr {
		isPtr = true
		slType = slType.Elem()
	}

	if slType.Kind() != reflect.Array && slType.Kind() != reflect.Slice {
		err = errors.Errorf("Invalid currenType provided into 'unmarshalSliceValue' %v", slType.String())
		log.Errorf("Attribute: %v, Err: %v", modelAttr.Name(), err)
		return
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
			log.Debugf("Interface value elemed: %v, type: %v", v, v.Type())
		}
		errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
		errObj.Detail = fmt.Sprintf("Field: '%s' the slice value is not a slice.", modelAttr.ApiName())
		err = errObj
		return
	}

	// check the lengths - it does matter if the provided type is an array
	if capSize != -1 {
		if v.Len() > capSize {
			errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
			errObj.Detail = fmt.Sprintf("Field: '%s' the slice value is too long. The maximum capacity is: '%d'", modelAttr.ApiName(), capSize)
			err = errObj
			return
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
	var nestedType reflect.Type = slType
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
				return
			}

			// otherwise this must be a single value
		} else {
			nestedValue, err = unmarshalSingleFieldValue(c, modelAttr, vElem, baseType)
			if err != nil {
				log.Debugf("NestedSlice -> unmarshalSingleFieldValue failed. %v. Velem: %v", err)
				return
			}
		}

		// if the value was an Array set it's value at index
		if fieldValue.Type().Kind() == reflect.Array {
			fieldValue.Index(i).Set(nestedValue)

			// otherwise the value is a slice append new value
		} else {
			fieldValue.Set(reflect.Append(fieldValue, nestedValue))
		}
	}

	// if the value was a Ptr type it's address has to be returned
	if isPtr {
		fieldValue = fieldValue.Addr()
	}

	return
}

// unmarshalSingleFieldValue gets the
func unmarshalSingleFieldValue(
	c *controller.Controller,
	modelAttr *models.StructField,
	v reflect.Value, // v is the incoming value to unmarshal
	baseType reflect.Type, // fieldType is the base Type of the model field
) (reflect.Value, error) {

	// check if the model attribute field is based with pointer
	if modelAttr.IsBasePtr() {
		log.Debugf("IsBasePtr: %v", modelAttr.Name())
		if !v.IsValid() {
			return reflect.Zero(reflect.PtrTo(baseType)), nil
		}
		if v.Interface() == nil {
			return reflect.Zero(reflect.PtrTo(baseType)), nil
		}
		log.Debugf("Value: '%v' is valid. ", v)
	} else {
		log.Debugf("Is not basePtr: %v", modelAttr.Name())
		if !v.IsValid() {
			return reflect.Value{}, aerrors.ErrInvalidJSONFieldValue.Copy().WithDetail(fmt.Sprintf("Field: '%s'", modelAttr.ApiName()))
		}
		log.Debugf("Is Valid: %v", v)

	}

	if v.Kind() == reflect.Interface && v.IsValid() {
		log.Debugf("Value is interface kind. %v", v)
		v = v.Elem()

		// if the value was an uinitialized interface{} then v.Elem
		// would be invalid.
		if !v.IsValid() {
			return reflect.Value{}, aerrors.ErrInvalidJSONFieldValue.Copy().WithDetail(fmt.Sprintf("Field: %v'", modelAttr.ApiName()))
		}

	}

	if modelAttr.IsTime() {
		if modelAttr.IsIso8601() {
			var tm string

			// on ISO8601 the incoming value must be a string
			if v.Kind() == reflect.String {
				tm = v.Interface().(string)
			} else {
				return reflect.Value{}, internal.ErrInvalidISO8601
			}

			// parse the string time with iso formatting
			t, err := time.Parse(internal.Iso8601TimeFormat, tm)
			if err != nil {
				return reflect.Value{}, internal.ErrInvalidISO8601
			}

			if modelAttr.IsBasePtr() {
				return reflect.ValueOf(&t), nil
			} else {
				return reflect.ValueOf(t), nil
			}
		} else {
			// if the time is not iso8601 it must be an integer
			var at int64

			// the integer may be a float64 or an int
			if v.Kind() == reflect.Float64 {
				at = int64(v.Interface().(float64))
			} else if v.Kind() == reflect.Int || v.Kind() == reflect.Int64 {
				at = v.Int()
			} else {
				log.Debugf("Invalid time format: %v", v.Kind().String())
				return reflect.Value{}, internal.ErrInvalidTime
			}

			t := time.Unix(at, 0)

			if modelAttr.IsBasePtr() {
				return reflect.ValueOf(&t), nil
			} else {
				return reflect.ValueOf(t), nil
			}
		}
	} else if modelAttr.IsNestedStruct() {
		value := v.Interface()
		return unmarshalNestedStructValue(c, modelAttr.Nested(), value)

		// if the incoming value v is of kind float64
	} else if v.Kind() == reflect.Float64 {
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
			return reflect.Value{}, internal.ErrUnknownFieldNumberType
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
				errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
				errObj.Detail = fmt.Sprintf("Field: '%s'. Provided value: '%d' is not an unsigned integer.", modelAttr.ApiName(), intValue)
				return reflect.Value{}, errObj
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
			return reflect.Value{}, internal.ErrUnknownFieldNumberType
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
		errObj := aerrors.ErrInvalidJSONFieldValue.Copy()
		errObj.Detail = fmt.Sprintf("The value: '%v' for the field: '%s' is not allowed", value, modelAttr.ApiName())
		return errObj
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
		unmErr := new(json.UnmarshalFieldError)
		unmErr.Field = modelAttr.ReflectField()
		unmErr.Type = baseType
		unmErr.Key = modelAttr.ApiName()
		if c != nil {
			log.Debugf("Error in unmarshal field: %v\n. Value Type: %v. Value: %v", unmErr, v.Type().String(), value)
		}
		return reflect.Value{}, unmErr
	}

	log.Debugf("AttrField: '%s' equal kind: %v", modelAttr.ApiName(), baseType.String())

	if modelAttr.IsBasePtr() {
		return concreteVal.Addr(), nil
	}

	return concreteVal, nil
}

var timeType reflect.Type = reflect.TypeOf(time.Time{})

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
		return internal.ErrBadJSONAPIID

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
		return internal.ErrBadJSONAPIID

	}
	assign(fieldValue, idValue)
	return nil
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
