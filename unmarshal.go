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
			fieldType := modelAttr.refStruct

			// v is the incoming value
			v := reflect.ValueOf(attrValue)

			// time
			if modelAttr.isTime() || modelAttr.isPtrTime() {
				if attrValue == nil {
					continue
				}
				if modelAttr.isIso8601() {
					var tm string
					if v.Kind() == reflect.String {
						tm = v.Interface().(string)
					} else {
						err = IErrInvalidISO8601
						return
					}
					var t time.Time
					t, err = time.Parse(iso8601TimeFormat, tm)
					if err != nil {
						err = IErrInvalidISO8601
						return
					}

					if modelAttr.isPtrTime() {
						fieldValue.Set(reflect.ValueOf(&t))
					} else {
						fieldValue.Set(reflect.ValueOf(t))
					}
				} else {
					var at int64

					if v.Kind() == reflect.Float64 {
						at = int64(v.Interface().(float64))
					} else if v.Kind() == reflect.Int {
						at = v.Int()
					} else {
						err = IErrInvalidTime
						return
					}

					t := time.Unix(at, 0)

					if modelAttr.isPtrTime() {
						fieldValue.Set(reflect.ValueOf(&t))
					} else {
						fieldValue.Set(reflect.ValueOf(t))
					}
				}
				continue
			}

			if fieldValue.Type() == reflect.TypeOf([]string{}) {
				if attrValue == nil {
					continue
				}
				values := make([]string, v.Len())
				for i := 0; i < v.Len(); i++ {
					elem := v.Index(i)
					if elem.IsNil() {
						errObj := ErrInvalidJSONFieldValue.Copy()
						errObj.Detail = fmt.Sprintf("Invalid field value for the '%s' field. The field doesn't allow 'null' value", attrName)
						err = errObj
						return
					}
					strVal, ok := elem.Interface().(string)
					if !ok {
						errObj := ErrInvalidJSONFieldValue.Copy()
						errObj.Detail = fmt.Sprintf("Invalid field value for the '%s' field. The field allow only string values. Field type: %v. Value: %v", attrName, elem.Kind(), elem.Interface())
						err = errObj
						return
					}
					values[i] = strVal
				}

				fieldValue.Set(reflect.ValueOf(values))

				continue
			}

			if v.Kind() == reflect.Float64 {
				floatValue := v.Interface().(float64)

				// The field may or may not be a pointer to a numeric; the kind var
				// will not contain a pointer type
				var kind reflect.Kind
				if fieldValue.Kind() == reflect.Ptr {
					kind = fieldType.Type.Elem().Kind()
				} else {
					kind = fieldType.Type.Kind()
				}

				var numericValue reflect.Value

				switch kind {
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
					return IErrUnknownFieldNumberType
				}

				assign(fieldValue, numericValue)
				continue
			}

			// Field was a Pointer type
			if fieldValue.Kind() == reflect.Ptr {
				if attrValue == nil {
					continue
				}
				var concreteVal reflect.Value

				switch cVal := attrValue.(type) {
				case string:
					concreteVal = reflect.ValueOf(&cVal)
				case bool:
					concreteVal = reflect.ValueOf(&cVal)
				case complex64:
					concreteVal = reflect.ValueOf(&cVal)
				case complex128:
					concreteVal = reflect.ValueOf(&cVal)
				case uintptr:
					concreteVal = reflect.ValueOf(&cVal)
				default:
					err = IErrUnsupportedPtrType
					return
				}

				if fieldValue.Type() != concreteVal.Type() {
					err = IErrUnsupportedPtrType
					return
				}

				fieldValue.Set(concreteVal)
				continue
			}

			// Check if the value is not a map
			if fieldValue.Kind() == reflect.Map {
				if v.Kind() != reflect.Map {
					if attrValue == nil {
						c.log().Debug("Nil attr value")
						continue
					}

					errObj := ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("Invalid field value for the '%s' field. The field doesn't allow 'null' value", attrName)
					err = errObj
					return
				}

				// allow only k,v with k as string

				if v.Type().Key().Kind() != reflect.String {
					c.log().Debug("Unmarshal map field:'%s'. Incoming field key type: '%s'", v.Type().String())
					errObj := ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("Attribute Field: '%s' in the collection: '%s' is a hashmap type with keys of 'string' type.", attrName, mStruct.collectionType)
					err = errObj
					return
				}

				mapValueType := fieldValue.Type().Elem()
				var isPtrValue bool

				if mapValueType.Kind() == reflect.Ptr {
					isPtrValue = true
					mapValueType = mapValueType.Elem()
				}

				keys := v.MapKeys()
				if len(keys) > 0 && fieldValue.IsNil() {
					fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
				}

				setMapValues := func(key, mpIndex reflect.Value) error {
					// Check if the value is of pointer type

					mpIndexValue := mpIndex.Interface()
					if isPtrValue {
						// check if mpIndex is not nil
						if mpIndexValue == nil {
							fieldValue.SetMapIndex(key, reflect.Zero(fieldValue.Type().Elem()))
							return nil
						}
					}

					if mapValueType.Kind() == reflect.String {
						// check if the value is string
						mpString, ok := mpIndexValue.(string)
						if !ok {
							err := ErrInvalidJSONFieldValue.Copy()
							err.Detail = fmt.Sprintf("Attribute field: '%s' in the collection: '%s' is a hashmap with values of type: 'string'. Provided value is not valid: '%v'", attrName, mpIndex, mpIndex.Interface())
							return err
						}

						// if the map value type is of pointer type get the address value
						if isPtrValue {
							fieldValue.SetMapIndex(key, reflect.ValueOf(&mpString))
						} else {
							fieldValue.SetMapIndex(key, reflect.ValueOf(mpString))
						}
						return nil
					}

					c.log().Debugf("mpIndex.Kind(): '%v' mpIndex.Type().Kind(): '%v'", mpIndex.Type(), mpIndex.Type().Kind())

					// the map attribute allow only strings and numeric values as a value
					// json.Decoder sets the json.Number as float64
					// the strings are already checked

					floatValue, ok := mpIndexValue.(float64)

					if !ok {
						c.log().Debugf("mpIndex Kind: %v", mpIndex.Kind())
						errObj := ErrInvalidJSONFieldValue.Copy()
						errObj.Detail = fmt.Sprintf("Attribute field: '%s' in the collection: '%s' is a hashmap with values of type: %s. Provided value is not valid: '%v'", attrName, mStruct.collectionType, mapValueType.Kind().String(), mpIndex.Interface())
						return errObj
					}

					var numericValue reflect.Value
					switch mapValueType.Kind() {
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
						return IErrUnknownFieldNumberType
					}

					if isPtrValue {
						fieldValue.SetMapIndex(key, numericValue)
					} else {
						fieldValue.SetMapIndex(key, numericValue.Elem())
					}

					return nil
				}

				for _, k := range keys {
					err = setMapValues(k, v.MapIndex(k))
					if err != nil {
						return
					}
				}
				continue
			}

			// As a final catch-all, ensure types line up to avoid a runtime panic.
			if fieldValue.Kind() != v.Kind() {
				unmErr := new(json.UnmarshalFieldError)
				unmErr.Field = fieldType
				unmErr.Type = fieldValue.Type()
				unmErr.Key = attrName
				if c != nil {
					c.log().Debugf("Error in unmarshal field: %v\n", unmErr)
				}
				err = unmErr
				return
			}
			fieldValue.Set(reflect.ValueOf(attrValue))
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
