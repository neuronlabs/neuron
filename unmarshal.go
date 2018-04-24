package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	unsuportedStructTagMsg = "Unsupported jsonapi tag annotation, %s"
)

var (
	// ErrInvalidTime is returned when a struct has a time.Time type field, but
	// the JSON value was not a unix timestamp integer.
	ErrInvalidTime = errors.New("Only numbers can be parsed as dates, unix timestamps")
	// ErrInvalidISO8601 is returned when a struct has a time.Time type field and includes
	// "iso8601" in the tag spec, but the JSON value was not an ISO8601 timestamp string.
	ErrInvalidISO8601 = errors.New("Only strings can be parsed as dates, ISO8601 timestamps")
	// ErrUnknownFieldNumberType is returned when the JSON value was a float
	// (numeric) but the Struct field was a non numeric type (i.e. not int, uint,
	// float, etc)
	ErrUnknownFieldNumberType = errors.New("The struct field was not of a known number type")
	// ErrUnsupportedPtrType is returned when the Struct field was a pointer but
	// the JSON value was of a different type
	ErrUnsupportedPtrType = errors.New("Pointer type in struct is not supported")
	// ErrInvalidType is returned when the given type is incompatible with the expected type.
	ErrInvalidType = errors.New("Invalid type provided") // I wish we used punctuation.
	// ErrInvalidQuery

)

var (
	// ErrBadJSONAPIStructTag is returned when the Struct field's JSON API
	// annotation is invalid.
	ErrBadJSONAPIStructTag = errors.New("Bad jsonapi struct tag format")
	// ErrBadJSONAPIID is returned when the Struct JSON API annotated "id" field
	// was not a valid numeric type.
	ErrBadJSONAPIID = errors.New(
		"id should be either string, int(8,16,32,64) or uint(8,16,32,64)")
	// ErrExpectedSlice is returned when a variable or argument was expected to
	// be a slice of *Structs; MarshalMany will return this error when its
	// interface{} argument is invalid.
	ErrExpectedSlice = errors.New("models should be a slice of struct pointers")
	// ErrUnexpectedType is returned when marshalling an interface; the interface
	// had to be a pointer or a slice; otherwise this error is returned.
	ErrUnexpectedType = errors.New("models should be a struct pointer or slice of struct pointers")
)

func UnmarshalScopeOne(in io.Reader, c *Controller) (*Scope, error) {
	return unmarshalScopeOne(in, c)
}

func unmarshalScopeOne(in io.Reader, c *Controller) (*Scope, error) {
	payload := new(OnePayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, err
	}

	collection := payload.Data.Type
	mStruct := c.Models.GetByCollection(collection)
	if mStruct == nil {
		errObj := ErrInvalidResourceName.Copy()
		errObj.Detail = fmt.Sprintf("The specified collection: '%s' is not recognized by the server.", collection)

		if similar := c.Models.getSimilarCollections(collection); len(similar) != 0 {
			errObj.Detail += "Suggested collections: "
			for _, sim := range similar[:len(similar)-1] {
				errObj.Detail += fmt.Sprintf("%s, ", sim)
			}
			errObj.Detail += fmt.Sprintf("%s.", similar[len(similar)-1])
		}
		return nil, errObj
	}

	scope := newRootScope(mStruct, false)

	var err error
	if payload.Included != nil {
		includedMap := make(map[string]*Node)
		for _, included := range payload.Included {
			key := fmt.Sprintf("%s,%s", included.Type, included.ID)
			includedMap[key] = included
		}
		err = unmarshalNode(payload.Data, reflect.ValueOf(scope.Value), &includedMap)
	} else {
		err = unmarshalNode(payload.Data, reflect.ValueOf(scope.Value), nil)
	}
	if err != nil {
		return nil, err
	}

	return scope, nil
}

// func unmarshalScopeOne(in io.Reader, c *Controller) (*Scope, error) {
// 	payload := new(OnePayload)

// 	if err := json.NewDecoder(in).Decode(payload); err != nil {
// 		return nil, err
// 	}

// 	collection := payload.Data.Type
// 	mStruct := c.Models.GetByCollection(collection)
// 	if mStruct == nil {
// 		errObj := ErrInvalidResourceName.Copy()
// 		errObj.Detail = fmt.Sprintf("The specified collection: '%s' is not recognized by the server.", collection)

// 		if similar := c.Models.getSimilarCollections(collection); len(similar) != 0 {
// 			errObj.Detail += "Suggested collections: "
// 			for _, sim := range similar[:len(similar)-1] {
// 				errObj.Detail += fmt.Sprintf("%s, ", sim)
// 			}
// 			errObj.Detail += fmt.Sprintf("%s.", similar[len(similar)-1])
// 		}
// 		return nil, errObj
// 	}
// 	scope := newRootScope(mStruct, false)

// 	return scope, nil
// }

// func unmarshalNode(node *Node, scope *Scope) error {
// 	t := reflect.TypeOf(scope.Value)
// 	var isSlice bool
// 	if t.Kind() == reflect.Slice {
// 		isSlice = true
// 	}

// 	modelType := scope.Struct.modelType
// 	modelValue := reflect.New(modelType)

// 	primary := modelValue.Field(scope.Struct.primary.getFieldIndex())
// 	/** if client id set to cliend it */
// 	err := setPrimaryField(node.ID, primary)
// 	if err != nil {
// 		// maybe ErrObj
// 		return err
// 	}

// 	for attr, attrVal := range node.Attributes {
// 		sField, ok := scope.Struct.attributes[attr]
// 		if !ok {
// 			// invalid attribute field
// 			errObj := ErrInvalidJSONFieldValue.Copy()
// 			errObj.Detail = fmt.Sprintf("An object of '%s' collection does not have attribute named: '%s'", node.Type, attr)
// 			return errObj
// 		}
// 		attrType := reflect.TypeOf(attrVal)
// 		switch attrType.Kind() {
// 		case reflect.Bool:
// 			if sField.refStruct.Type.Kind() == reflect.Bool {
// 				modelValue.Field(sField.getFieldIndex()).SetBool(attrVal.(bool))
// 			} else {
// 				//error
// 			}
// 		case reflect.Slice:
// 			if sField.refStruct.Type.Kind() == reflect.Slice {
// 				// check type
// 			}
// 		case reflect.Float64:
// 		case reflect.Struct:
// 		case reflect.Int:
// 		case reflect.Float64:

// 		}

// 	}

// 	val := reflect.ValueOf(scope.Value)

// 	if isSlice {
// 		scope.Value = reflect.Append(val, modelValue)
// 	} else {
// 		val = modelValue
// 	}

// 	return nil
// }

func UnmarshalPayload(in io.Reader, model interface{}) error {
	payload := new(OnePayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return err
	}

	if payload.Included != nil {
		includedMap := make(map[string]*Node)
		for _, included := range payload.Included {
			key := fmt.Sprintf("%s,%s", included.Type, included.ID)
			includedMap[key] = included
		}

		return unmarshalNode(payload.Data, reflect.ValueOf(model), &includedMap)
	}
	return unmarshalNode(payload.Data, reflect.ValueOf(model), nil)
}

// UnmarshalManyPayload converts an io into a set of struct instances using
// jsonapi tags on the type's struct fields.
func UnmarshalManyPayload(in io.Reader, t reflect.Type) ([]interface{}, error) {
	payload := new(ManyPayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, err
	}

	models := []interface{}{}         // will be populated from the "data"
	includedMap := map[string]*Node{} // will be populate from the "included"

	if payload.Included != nil {
		for _, included := range payload.Included {
			key := fmt.Sprintf("%s,%s", included.Type, included.ID)
			includedMap[key] = included
		}
	}

	for _, data := range payload.Data {
		model := reflect.New(t.Elem())
		err := unmarshalNode(data, model, &includedMap)
		if err != nil {
			return nil, err
		}
		models = append(models, model.Interface())
	}

	return models, nil
}

func unmarshalNode(data *Node, model reflect.Value, included *map[string]*Node) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("data is not a jsonapi representation of '%v'", model.Type())
		}
	}()

	modelValue := model.Elem()
	modelType := model.Type().Elem()

	var er error

	for i := 0; i < modelValue.NumField(); i++ {
		fieldType := modelType.Field(i)
		tag := fieldType.Tag.Get("jsonapi")
		if tag == "" {
			continue
		}

		fieldValue := modelValue.Field(i)

		args := strings.Split(tag, ",")

		if len(args) < 1 {
			er = ErrBadJSONAPIStructTag
			break
		}

		annotation := args[0]

		if (annotation == annotationClientID && len(args) != 1) ||
			(annotation != annotationClientID && len(args) < 2) {
			er = ErrBadJSONAPIStructTag
			break
		}

		if annotation == annotationPrimary {
			if data.ID == "" {
				continue
			}

			// Check the JSON API Type
			if data.Type != args[1] {
				er = fmt.Errorf(
					"Trying to Unmarshal an object of type %#v, but %#v does not match",
					data.Type,
					args[1],
				)
				break
			}

			// ID will have to be transmitted as astring per the JSON API spec
			v := reflect.ValueOf(data.ID)

			// Deal with PTRS
			var kind reflect.Kind
			if fieldValue.Kind() == reflect.Ptr {
				kind = fieldType.Type.Elem().Kind()
			} else {
				kind = fieldType.Type.Kind()
			}

			// Handle String case
			if kind == reflect.String {
				assign(fieldValue, v)
				continue
			}

			// Value was not a string... only other supported type was a numeric,
			// which would have been sent as a float value.
			floatValue, err := strconv.ParseFloat(data.ID, 64)
			if err != nil {
				// Could not convert the value in the "id" attr to a float
				er = ErrBadJSONAPIID
				break
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
				er = ErrBadJSONAPIID
				break
			}

			assign(fieldValue, idValue)
		} else if annotation == annotationClientID {
			if data.ClientID == "" {
				continue
			}

			fieldValue.Set(reflect.ValueOf(data.ClientID))
		} else if annotation == annotationAttribute {
			attributes := data.Attributes
			if attributes == nil || len(data.Attributes) == 0 {
				continue
			}

			var iso8601 bool

			if len(args) > 2 {
				for _, arg := range args[2:] {
					if arg == annotationISO8601 {
						iso8601 = true
					}
				}
			}

			val := attributes[args[1]]

			// continue if the attribute was not included in the request
			if val == nil {
				continue
			}

			v := reflect.ValueOf(val)

			// Handle field of type time.Time
			if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				if iso8601 {
					var tm string
					if v.Kind() == reflect.String {
						tm = v.Interface().(string)
					} else {
						er = ErrInvalidISO8601
						break
					}

					t, err := time.Parse(iso8601TimeFormat, tm)
					if err != nil {
						er = ErrInvalidISO8601
						break
					}

					fieldValue.Set(reflect.ValueOf(t))

					continue
				}

				var at int64

				if v.Kind() == reflect.Float64 {
					at = int64(v.Interface().(float64))
				} else if v.Kind() == reflect.Int {
					at = v.Int()
				} else {
					return ErrInvalidTime
				}

				t := time.Unix(at, 0)

				fieldValue.Set(reflect.ValueOf(t))

				continue
			}

			if fieldValue.Type() == reflect.TypeOf([]string{}) {
				values := make([]string, v.Len())
				for i := 0; i < v.Len(); i++ {
					values[i] = v.Index(i).Interface().(string)
				}

				fieldValue.Set(reflect.ValueOf(values))

				continue
			}

			if fieldValue.Type() == reflect.TypeOf(new(time.Time)) {
				if iso8601 {
					var tm string
					if v.Kind() == reflect.String {
						tm = v.Interface().(string)
					} else {
						er = ErrInvalidISO8601
						break
					}

					v, err := time.Parse(iso8601TimeFormat, tm)
					if err != nil {
						er = ErrInvalidISO8601
						break
					}

					t := &v

					fieldValue.Set(reflect.ValueOf(t))

					continue
				}

				var at int64

				if v.Kind() == reflect.Float64 {
					at = int64(v.Interface().(float64))
				} else if v.Kind() == reflect.Int {
					at = v.Int()
				} else {
					return ErrInvalidTime
				}

				v := time.Unix(at, 0)
				t := &v

				fieldValue.Set(reflect.ValueOf(t))

				continue
			}

			// JSON value was a float (numeric)
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
					return ErrUnknownFieldNumberType
				}

				assign(fieldValue, numericValue)
				continue
			}

			// Field was a Pointer type
			if fieldValue.Kind() == reflect.Ptr {
				var concreteVal reflect.Value

				switch cVal := val.(type) {
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
					return ErrUnsupportedPtrType
				}

				if fieldValue.Type() != concreteVal.Type() {
					return ErrUnsupportedPtrType
				}

				fieldValue.Set(concreteVal)
				continue
			}

			// As a final catch-all, ensure types line up to avoid a runtime panic.
			if fieldValue.Kind() != v.Kind() {
				return ErrInvalidType
			}
			fieldValue.Set(reflect.ValueOf(val))

		} else if annotation == annotationRelation {
			isSlice := fieldValue.Type().Kind() == reflect.Slice

			if data.Relationships == nil || data.Relationships[args[1]] == nil {
				continue
			}

			if isSlice {
				// to-many relationship
				relationship := new(RelationshipManyNode)

				buf := bytes.NewBuffer(nil)

				json.NewEncoder(buf).Encode(data.Relationships[args[1]])
				json.NewDecoder(buf).Decode(relationship)

				data := relationship.Data
				models := reflect.New(fieldValue.Type()).Elem()

				for _, n := range data {
					m := reflect.New(fieldValue.Type().Elem().Elem())

					if err := unmarshalNode(
						fullNode(n, included),
						m,
						included,
					); err != nil {
						er = err
						break
					}

					models = reflect.Append(models, m)
				}

				fieldValue.Set(models)
			} else {
				// to-one relationships
				relationship := new(RelationshipOneNode)

				buf := bytes.NewBuffer(nil)

				json.NewEncoder(buf).Encode(
					data.Relationships[args[1]],
				)
				json.NewDecoder(buf).Decode(relationship)

				/*
					http://jsonapi.org/format/#document-resource-object-relationships
					http://jsonapi.org/format/#document-resource-object-linkage
					relationship can have a data node set to null (e.g. to disassociate the relationship)
					so unmarshal and set fieldValue only if data obj is not null
				*/
				if relationship.Data == nil {
					continue
				}

				m := reflect.New(fieldValue.Type().Elem())
				if err := unmarshalNode(
					fullNode(relationship.Data, included),
					m,
					included,
				); err != nil {
					er = err
					break
				}

				fieldValue.Set(m)

			}

		} else {
			er = fmt.Errorf(unsuportedStructTagMsg, annotation)
		}
	}

	return er
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
