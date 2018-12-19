package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"strconv"
	"time"
)

const (
	unsuportedStructTagMsg = "Unsupported jsonapi tag annotation, %s"
)

var (
	IErrClientIDDisallowed = errors.New("Client id is disallowed for this model.")
	// IErrInvalidTime is returned when a struct has a time.Time type field, but
	// the JSON value was not a unix timestamp integer.
	IErrInvalidTime = errors.New("Only numbers can be parsed as dates, unix timestamps")
	// IErrInvalidISO8601 is returned when a struct has a time.Time type field and includes
	// "iso8601" in the tag spec, but the JSON value was not an ISO8601 timestamp string.
	IErrInvalidISO8601 = errors.New("Only strings can be parsed as dates, ISO8601 timestamps")
	// IErrUnknownFieldNumberType is returned when the JSON value was a float
	// (numeric) but the Struct field was a non numeric type (i.e. not int, uint,
	// float, etc)
	IErrUnknownFieldNumberType = errors.New("The struct field was not of a known number type")
	// IErrUnsupportedPtrType is returned when the Struct field was a pointer but
	// the JSON value was of a different type
	IErrUnsupportedPtrType = errors.New("Pointer type in struct is not supported")
	// IErrInvalidType is returned when the given type is incompatible with the expected type.
	IErrInvalidType = errors.New("Invalid type provided") // I wish we used punctuation.
	// ErrInvalidQuery

)

var (
	// IErrBadJSONAPIStructTag is returned when the Struct field's JSON API
	// annotation is invalid.
	IErrBadJSONAPIStructTag = errors.New("Bad jsonapi struct tag format")
	// IErrBadJSONAPIID is returned when the Struct JSON API annotated "id" field
	// was not a valid numeric type.
	IErrBadJSONAPIID = errors.New(
		"id should be either string, int(8,16,32,64) or uint(8,16,32,64)")
)

func (c *Controller) UnmarshalScopeMany(in io.Reader, model interface{}) (*Scope, error) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	mStruct := c.Models.Get(t)
	if mStruct == nil {
		return nil, IErrModelNotMapped
	}
	return c.unmarshalScope(in, mStruct, true, false)
}

func (c *Controller) UnmarshalScopeOne(
	in io.Reader,
	model interface{},
	addSelectedFields bool,
) (*Scope, error) {
	t := reflect.TypeOf(model)
	return c.unmarshalScopeOne(in, t, addSelectedFields)
}

func (c *Controller) unmarshalScopeOne(
	in io.Reader,
	t reflect.Type,
	addSelectedFields bool,
) (*Scope, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	mStruct := c.Models.Get(t)
	if mStruct == nil {
		return nil, IErrModelNotMapped
	}
	return c.unmarshalScope(in, mStruct, false, addSelectedFields)
}

func (c *Controller) Unmarshal(in io.Reader, v interface{}) error {
	_, err := c.unmarshal(in, v, false)
	return err
}

func (c *Controller) unmarshalScope(
	in io.Reader,
	mStruct *ModelStruct,
	useMany, usedFields bool,
) (*Scope, error) {

	var modelValue reflect.Value
	if useMany {
		modelValue = reflect.New(reflect.SliceOf(reflect.New(mStruct.modelType).Type()))
	} else {
		modelValue = reflect.New(mStruct.modelType)
	}

	v := modelValue.Interface()
	fields, err := c.unmarshal(in, v, usedFields)
	if err != nil {
		return nil, err
	}

	scope := newRootScope(mStruct)
	if useMany {

		scope.Value = reflect.ValueOf(v).Elem().Interface()
	} else {
		scope.Value = v
	}

	if usedFields && !useMany {
		for _, field := range fields {
			scope.SelectedFields = append(scope.SelectedFields, field)
		}

	}
	return scope, nil

}

func (c *Controller) unmarshal(
	in io.Reader,
	model interface{},
	usedFields bool,
) ([]*StructField, error) {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return nil, IErrUnexpectedType
	}

	switch t.Elem().Kind() {
	case reflect.Struct:

		mStruct := c.Models.Get(t.Elem())
		if mStruct == nil {
			return nil, IErrModelNotMapped
		}

		one, err := c.unmarshalPayload(in)
		if err != nil {
			return nil, err
		}
		var fields []*StructField
		if usedFields {
			if one.Data.ID != "" {
				fields = append(fields, mStruct.primary)
			}

			for attr := range one.Data.Attributes {
				field, ok := mStruct.attributes[attr]
				if ok {
					fields = append(fields, field)
				}
			}

			for relation := range one.Data.Relationships {
				field, ok := mStruct.relationships[relation]
				if ok {
					fields = append(fields, field)
				}
			}
		}

		if one.Data.Type != mStruct.collectionType {
			c.log().Debugf("Unmarshaled data of collection: '%s' for model's collection: '%s'", one.Data.Type, mStruct.collectionType)
			errObj := ErrInvalidResourceName.Copy()
			errObj.Detail = fmt.Sprintf("The specified collection: '%s' is not recognized by the server.", one.Data.Type)
			return nil, errObj
		}

		if one.Included != nil {
			c.log().Debug("Payload contains Included values.")
			includedMap := make(map[string]*Node)
			for _, included := range one.Included {
				key := fmt.Sprintf("%s,%s", included.Type, included.ID)
				includedMap[key] = included
			}
			err = c.unmarshalNode(one.Data, reflect.ValueOf(model), &includedMap)
		} else {
			err = c.unmarshalNode(one.Data, reflect.ValueOf(model), nil)
		}
		switch err {
		case IErrUnknownFieldNumberType, IErrInvalidTime, IErrInvalidISO8601,
			IErrInvalidType, IErrBadJSONAPIID:
			errObj := ErrInvalidJSONFieldValue.Copy()
			errObj.Detail = err.Error()
			errObj.Err = err
			return nil, errObj
		default:
			if uErr, ok := err.(*json.UnmarshalFieldError); ok {
				errObj := ErrInvalidJSONFieldValue.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid type for field: %v. This field is of type: %s.", uErr.Key, uErr.Type.String())
				errObj.Err = err
				return nil, errObj
			}
		}
		return fields, err
	case reflect.Slice:
		t = t.Elem().Elem()
		if t.Kind() != reflect.Ptr {
			return nil, IErrUnexpectedType
		}
		t = t.Elem()
		mStruct := c.Models.Get(t)
		if mStruct == nil {
			return nil, IErrModelNotMapped
		}

		many, err := c.unmarshalManyPayload(in)
		if err != nil {
			return nil, err
		}

		models := reflect.New(reflect.SliceOf(reflect.New(mStruct.modelType).Type())).Elem()

		// will be populated from the "data"
		includedMap := map[string]*Node{} // will be populate from the "included"

		if many.Included != nil {
			for _, included := range many.Included {
				key := fmt.Sprintf("%s,%s", included.Type, included.ID)
				includedMap[key] = included
			}
		}

		for _, data := range many.Data {
			model := reflect.New(mStruct.modelType)
			err = c.unmarshalNode(data, model, &includedMap)
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

func (c *Controller) unmarshalPayload(in io.Reader) (*OnePayload, error) {
	payload := new(OnePayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, c.unmarshalHandleDecodeError(err)
	}
	if payload.Data == nil {
		errObj := ErrInvalidJSONDocument.Copy()
		errObj.Detail = "Specified request contains no data."
		return nil, errObj
	}
	return payload, nil
}

func (c *Controller) unmarshalManyPayload(in io.Reader) (*ManyPayload, error) {
	payload := new(ManyPayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, c.unmarshalHandleDecodeError(err)
	}
	if payload.Data == nil {
		errObj := ErrInvalidJSONDocument.Copy()
		errObj.Detail = "Specified request contains no data."
		return nil, errObj
	}
	return payload, nil
}

func (c *Controller) unmarshalHandleDecodeError(er error) error {
	if serr, ok := er.(*json.SyntaxError); ok {
		errObj := ErrInvalidJSONDocument.Copy()
		errObj.Detail = fmt.Sprintf("Syntax Error: %s. At offset: %d.", er.Error(), serr.Offset)
		return errObj
	} else if uErr, ok := er.(*json.UnmarshalTypeError); ok {
		if uErr.Type == reflect.TypeOf(OnePayload{}) ||
			uErr.Type == reflect.TypeOf(ManyPayload{}) {
			errObj := ErrInvalidJSONDocument.Copy()
			errObj.Detail = fmt.Sprintln("Invalid JSON document syntax.")
			return errObj
		} else {
			errObj := ErrInvalidJSONFieldValue.Copy()
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
		errObj := ErrInvalidJSONDocument.Copy()
		errObj.Detail = fmt.Sprint("Provided document is empty.")
		errObj.Err = er
		return errObj
	} else {
		c.log().Errorf("Unknown error occured while decoding the payload. %v", er)
		return er
	}

	return nil
}

func (c *Controller) unmarshalNode(
	data *Node,
	modelValue reflect.Value,
	included *map[string]*Node,
) (err error) {
	mStruct := c.Models.GetByCollection(data.Type)
	if mStruct == nil {
		c.log().Debugf("Invalid collection type: %s", data.Type)
		errObj := ErrInvalidResourceName.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid collection name: '%s'", data.Type)
		err = errObj
		return
	}
	modelValue = modelValue.Elem()
	modelType := modelValue.Type()

	if modelType != mStruct.modelType {
		errObj := ErrInvalidJSONDocument.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid collection name for given model: '%s'", data.Type)
		err = errObj
		return
	}

	// add primary key
	if data.ID != "" {
		primary := modelValue.FieldByIndex(mStruct.primary.refStruct.Index)
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
	// 		err = IErrClientIDDisallowed
	// 		return
	// 	}
	// }
	if data.Attributes != nil {

		// Iterate over the data attributes
		for attrName, attrValue := range data.Attributes {
			modelAttr, ok := mStruct.attributes[attrName]
			if !ok || (ok && modelAttr.isHidden()) {
				if c.StrictUnmarshalMode {
					errObj := ErrInvalidJSONDocument.Copy()
					errObj.Detail = fmt.Sprintf("The server doesn't allow unknown field names. Provided unknown field name: '%s'.", attrName)
					err = errObj
					return
				} else {
					continue
				}
			}

			fieldValue := modelValue.FieldByIndex(modelAttr.refStruct.Index)

			err = c.unmarshalAttrFieldValue(modelAttr, fieldValue, attrValue)
			if err != nil {
				return
			}
		}

	}

	if data.Relationships != nil {
		for relName, relValue := range data.Relationships {
			modelRel, ok := mStruct.relationships[relName]
			if !ok || (ok && modelRel.isHidden()) {
				if c.StrictUnmarshalMode {
					errObj := ErrInvalidJSONDocument.Copy()
					errObj.Detail = fmt.Sprintf("The server doesn't allow unknown field names. Provided unknown field name: '%s'.", relName)
					err = errObj
					return
				} else {
					continue
				}
			}

			fieldValue := modelValue.FieldByIndex(modelRel.refStruct.Index)
			// unmarshal the relationship
			if modelRel.fieldType == RelationshipMultiple {
				// to-many relationship
				relationship := new(RelationshipManyNode)

				buf := bytes.NewBuffer(nil)

				err = json.NewEncoder(buf).Encode(data.Relationships[relName])
				if err != nil {
					c.log().Debugf("Controller.UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					errObj := ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}
				err = json.NewDecoder(buf).Decode(relationship)
				if err != nil {
					c.log().Debugf("Controller.UnmarshalNode.relationshipMultiple json.Encode failed. %v", err)
					errObj := ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}

				data := relationship.Data

				models := reflect.New(fieldValue.Type()).Elem()

				for _, n := range data {
					m := reflect.New(fieldValue.Type().Elem().Elem())
					if err = c.unmarshalNode(
						fullNode(n, included),
						m,
						included,
					); err != nil {
						c.log().Debugf("unmarshalNode.RelationshipMany - unmarshalNode failed. %v", err)
						return
					}

					models = reflect.Append(models, m)
				}
				fieldValue.Set(models)
			} else if modelRel.fieldType == RelationshipSingle {
				relationship := new(RelationshipOneNode)
				buf := bytes.NewBuffer(nil)

				if err = json.NewEncoder(buf).Encode(relValue); err != nil {
					c.log().Debugf("Controller.UnmarshalNode.relationshipSingle json.Encode failed. %v", err)
					errObj := ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}

				if err = json.NewDecoder(buf).Decode(relationship); err != nil {
					c.log().Debugf("Controller.UnmarshalNode.RelationshipSingel json.Decode failed. %v", err)
					errObj := ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value for the relationship: '%s' is of invalid form.", relName)
					errObj.Err = err
					err = errObj
					return
				}

				if relationship.Data == nil {
					continue
				}

				m := reflect.New(fieldValue.Type().Elem())
				if err := c.unmarshalNode(
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

func (c *Controller) unmarshalAttrFieldValue(
	modelAttr *StructField,
	fieldValue reflect.Value,
	attrValue interface{},
) (err error) {

	v := reflect.ValueOf(attrValue)
	fieldType := modelAttr.refStruct

	baseType := modelAttr.baseFieldType()

	if modelAttr.isSlice() || modelAttr.isArray() {
		var sliceValue reflect.Value
		sliceValue, err = c.unmarshalSliceValue(modelAttr, v, fieldType.Type, baseType)
		if err != nil {
			return
		}
		fieldValue.Set(sliceValue)
	} else if modelAttr.isMap() {
		var mapValue reflect.Value
		mapValue, err = c.unmarshalMapValue(modelAttr, v, fieldType.Type, baseType)
		if err != nil {
			return
		}
		fieldValue.Set(mapValue)
	} else {
		var resultValue reflect.Value
		resultValue, err = c.unmarshalSingleFieldValue(modelAttr, v, baseType)
		if err != nil {
			return
		}

		fieldValue.Set(resultValue)
	}

	return
}

func (c *Controller) unmarshalMapValue(
	modelAttr *StructField,
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
		errObj := ErrInvalidJSONFieldValue.Copy()
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

		c.log().Debugf("Value: %v. Kind: %v", value, value.Kind())

		var nestedValue reflect.Value
		switch nestedType.Kind() {
		case reflect.Slice, reflect.Array:
			c.log().Debugf("Nested Slice within map field: %v", modelAttr.Name())
			nestedValue, err = c.unmarshalSliceValue(modelAttr, value, mpType.Elem(), baseType)
			if err != nil {
				return
			}
		default:
			c.log().Debugf("Map ModelAttr: '%s' has basePtr: '%v'", modelAttr.Name(), modelAttr.isBasePtr())
			nestedValue, err = c.unmarshalSingleFieldValue(modelAttr, value, baseType)
			if err != nil {
				return
			}

			c.log().Debugf("Map ModelAttr: '%s' nestedValue: '%v'", nestedValue.Type().String())

		}
		mapValue.SetMapIndex(key, nestedValue)
	}
	return
}

func (c *Controller) unmarshalSliceValue(
	modelAttr *StructField, // modelAttr is the attribute StructField
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
		c.log().Errorf("Attribute: %v, Err: %v", modelAttr.Name(), err)
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
		c.log().Debugf("Value within the slice is not of slice type. %v", v.Kind().String())
		if v.Kind() == reflect.Interface {
			v = v.Elem()
			c.log().Debugf("Interface value elemed: %v, type: %v", v, v.Type())
		}
		errObj := ErrInvalidJSONFieldValue.Copy()
		errObj.Detail = fmt.Sprintf("Field: '%s' the slice value is not a slice.", modelAttr.ApiName())
		err = errObj
		return
	}

	// check the lengths - it does matter if the provided type is an array
	if capSize != -1 {
		if v.Len() > capSize {
			errObj := ErrInvalidJSONFieldValue.Copy()
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
			nestedValue, err = c.unmarshalSliceValue(modelAttr, vElem, slType, baseType)
			if err != nil {
				c.log().Debug("Nested Slice Value failed: %v", err)
				return
			}

			// otherwise this must be a single value
		} else {
			nestedValue, err = c.unmarshalSingleFieldValue(modelAttr, vElem, baseType)
			if err != nil {
				c.log().Debugf("NestedSlice -> unmarshalSingleFieldValue failed. %v. Velem: %v", err)
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
func (c *Controller) unmarshalSingleFieldValue(
	modelAttr *StructField,
	v reflect.Value, // v is the incoming value to unmarshal
	baseType reflect.Type, // fieldType is the base Type of the model field
) (reflect.Value, error) {

	// check if the model attribute field is based with pointer
	if modelAttr.isBasePtr() {
		c.log().Debugf("IsBasePtr: %v", modelAttr.Name())
		if !v.IsValid() {
			return reflect.Zero(reflect.PtrTo(baseType)), nil
		}
		if v.Interface() == nil {
			return reflect.Zero(reflect.PtrTo(baseType)), nil
		}
		c.log().Debugf("Value: '%v' is valid. ", v)
	} else {
		c.log().Debugf("Is not basePtr: %v", modelAttr.Name())
		if !v.IsValid() {
			return reflect.Value{}, ErrInvalidJSONFieldValue.Copy().WithDetail(fmt.Sprintf("Field: '%s'", modelAttr.ApiName()))
		}
		c.log().Debugf("Is Valid: %v", v)

	}

	if v.Kind() == reflect.Interface && v.IsValid() {
		c.log().Debugf("Value is interface kind. %v", v)
		v = v.Elem()

		// if the value was an uinitialized interface{} then v.Elem
		// would be invalid.
		if !v.IsValid() {
			return reflect.Value{}, ErrInvalidJSONFieldValue.Copy().WithDetail(fmt.Sprintf("Field: %v'", modelAttr.ApiName()))
		}

	}

	if modelAttr.isTime() {
		if modelAttr.isIso8601() {
			var tm string

			// on ISO8601 the incoming value must be a string
			if v.Kind() == reflect.String {
				tm = v.Interface().(string)
			} else {
				return reflect.Value{}, IErrInvalidISO8601
			}

			// parse the string time with iso formatting
			t, err := time.Parse(iso8601TimeFormat, tm)
			if err != nil {
				return reflect.Value{}, IErrInvalidISO8601
			}

			if modelAttr.isBasePtr() {
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
				c.log().Debugf("Invalid time format: %v", v.Kind().String())
				return reflect.Value{}, IErrInvalidTime
			}

			t := time.Unix(at, 0)

			if modelAttr.isBasePtr() {
				return reflect.ValueOf(&t), nil
			} else {
				return reflect.ValueOf(t), nil
			}
		}
	} else if modelAttr.isNestedStruct() {
		value := v.Interface()
		return modelAttr.nested.UnmarshalValue(c, value)
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
			c.log().Debugf("Unknown field number type: '%v'", baseType.String())
			return reflect.Value{}, IErrUnknownFieldNumberType
		}

		// if the field was ptr
		if modelAttr.isBasePtr() {
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
				errObj := ErrInvalidJSONFieldValue.Copy()
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
			c.log().Debugf("Unknown field number type: '%v'", baseType.String())
			return reflect.Value{}, IErrUnknownFieldNumberType
		}

		// if the field was ptr
		if modelAttr.isBasePtr() {
			return numericValue, nil
		}

		return numericValue.Elem(), nil
	}

	var concreteVal reflect.Value

	value := v.Interface()
	valueNotAllowed := func() error {
		errObj := ErrInvalidJSONFieldValue.Copy()
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
		c.log().Debugf("Invalid Value: %v with Kind: %v, should be: '%v' for field %v", v, v.Kind(), baseType.String(), modelAttr.Name())
		unmErr := new(json.UnmarshalFieldError)
		unmErr.Field = modelAttr.refStruct
		unmErr.Type = baseType
		unmErr.Key = modelAttr.ApiName()
		if c != nil {
			c.log().Debugf("Error in unmarshal field: %v\n. Value Type: %v. Value: %v", unmErr, v.Type().String(), value)
		}
		return reflect.Value{}, unmErr
	}

	c.log().Debugf("AttrField: '%s' equal kind: %v", modelAttr.ApiName(), baseType.String())

	if modelAttr.isBasePtr() {
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
		return IErrBadJSONAPIID

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
		return IErrBadJSONAPIID

	}
	assign(fieldValue, idValue)
	return nil
}

func fullNode(n *Node, included *map[string]*Node) *Node {
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
