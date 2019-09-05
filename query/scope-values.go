package query

import (
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal/safemap"
)

func (s *Scope) getPrimaryFieldValues() ([]interface{}, error) {
	if s.Value == nil {
		return nil, errors.NewDet(class.QueryNoValue, "no scope value provided")
	}

	primaryIndex := s.mStruct.Primary().FieldIndex()
	values := []interface{}{}

	addPrimaryValue := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(primaryIndex)
		values = append(values, primaryValue.Interface())
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.QueryValueType, "invalid query value type")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)
			if single.Kind() != reflect.Ptr {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.NewDet(class.QueryValueType, "one fo the slice values is of invalid type")
			}

			if single.IsNil() {
				continue
			}

			single = single.Elem()
			if single.Kind() != reflect.Struct {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.NewDet(class.QueryValueType, "one fo the slice values is of invalid type")
			}
			addPrimaryValue(single)
		}
	case reflect.Struct:
		addPrimaryValue(v)
	default:
		return nil, errors.NewDet(class.QueryValueType, "one query value is of invalid type")
	}
	return values, nil
}

// GetForeignKeyValues gets the values of the foreign key struct field
func (s *Scope) getForeignKeyValues(foreign *mapping.StructField) ([]interface{}, error) {
	if s.mStruct != foreign.Struct() {
		log.Debugf("Scope's ModelStruct: %s, doesn't match foreign key ModelStruct: '%s' ", s.mStruct.Collection(), foreign.Struct().Collection())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key mismatched ModelStruct")
	} else if foreign.FieldKind() != mapping.KindForeignKey {
		log.Debugf("'foreign' field is not a ForeignKey: %s", foreign.FieldKind())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key is not a valid ForeignKey")
	} else if s.Value == nil {
		return nil, errors.NewDet(class.QueryNoValue, "provided nil scope value")
	}

	// initialize the array
	values := []interface{}{}

	// set the adding functino
	addForeignKey := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(foreign.FieldIndex())
		values = append(values, primaryValue.Interface())
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.QueryNoValue, "provided no scope value")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)

			if single.Kind() != reflect.Ptr {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.NewDet(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			if single.IsNil() {
				continue
			}
			single = single.Elem()
			if single.Kind() != reflect.Struct {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.NewDet(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			addForeignKey(single)
		}
	case reflect.Struct:
		addForeignKey(v)
	default:
		return nil, errors.NewDet(class.QueryValueType, "invalid query value type")
	}

	return values, nil
}

// GetUniqueForeignKeyValues gets the unique values of the foreign key struct field
func (s *Scope) getUniqueForeignKeyValues(foreign *mapping.StructField) ([]interface{}, error) {
	if s.mStruct != foreign.Struct() {
		log.Debugf("Scope's ModelStruct: %s, doesn't match foreign key ModelStruct: '%s' ", s.mStruct.Collection(), foreign.Struct().Collection())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key mismatched ModelStruct")
	} else if foreign.FieldKind() != mapping.KindForeignKey {
		log.Debugf("'foreign' field is not a ForeignKey: %s", foreign.FieldKind())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key is not a valid ForeignKey")
	} else if s.Value == nil {
		return nil, errors.NewDet(class.QueryNoValue, "provided nil scope value")
	}

	// initialize the array
	foreigns := make(map[interface{}]struct{})
	values := []interface{}{}

	// set the adding functino
	addForeignKey := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(foreign.FieldIndex())
		foreigns[primaryValue.Interface()] = struct{}{}
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.QueryNoValue, "provided no scope value")
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)

			if single.Kind() != reflect.Ptr {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.NewDet(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			if single.IsNil() {
				continue
			}
			single = single.Elem()
			if single.Kind() != reflect.Struct {
				log.Debugf("Getting Primary values from the toMany scope. One of the scope values is of invalid type: %T", single.Interface())
				return nil, errors.NewDet(class.QueryValueType, "one of the scope's slice value is of invalid type")
			}

			addForeignKey(single)
		}
	case reflect.Struct:
		addForeignKey(v)
	default:
		return nil, errors.NewDet(class.QueryValueType, "invalid query value type")
	}

	for foreign := range foreigns {
		values = append(values, foreign)
	}

	return values, nil
}

// setCollectionValues iterate over the scope's Value field and add it to the collection root
// scope. If the collection root scope contains value with given primary field it checks if given
// scope contains included fields that are not within fieldset. If so it adds the included field
// value to the value that were inside the collection root scope.
func (s *Scope) setCollectionValues() error {
	if s.collectionScope.includedValues == nil {
		s.collectionScope.includedValues = safemap.New()
	}
	s.collectionScope.includedValues.Lock()
	defer s.collectionScope.includedValues.Unlock()

	primIndex := s.mStruct.Primary().FieldIndex()

	setValueToCollection := func(value reflect.Value) {
		primaryValue := value.Elem().FieldByIndex(primIndex)
		if !primaryValue.IsValid() {
			return
		}

		primary := primaryValue.Interface()
		insider, ok := s.collectionScope.includedValues.UnsafeGet(primary)
		if !ok {
			s.collectionScope.includedValues.UnsafeSet(primary, value.Interface())
			return
		}

		if insider == nil {
			// in order to prevent the nil values set within given key
			s.collectionScope.includedValues.UnsafeSet(primary, value.Interface())
		} else if s.hasFieldNotInFieldset {
			// this scopes value should have more fields
			insideValue := reflect.ValueOf(insider)

			for _, included := range s.includedFields {
				// only the fields that are not in the fieldset should be added
				if included.NotInFieldset {
					// get the included field index
					index := included.FieldIndex()

					// check if included field in the collection Values has this field
					if insideField := insideValue.Elem().FieldByIndex(index); !insideField.IsNil() {
						thisField := value.Elem().FieldByIndex(index)
						if thisField.IsNil() {
							thisField.Set(insideField)
						}
					}
				}
			}
		}
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return errors.NewDet(class.QueryValueType, "invalid scope value type")
	}
	if v.IsNil() {
		return nil
	}
	tempV := v.Elem()

	switch tempV.Kind() {
	case reflect.Slice:
		for i := 0; i < tempV.Len(); i++ {
			elem := tempV.Index(i)
			if elem.Type().Kind() != reflect.Ptr {
				return errors.NewDet(class.QueryValueType, "invalid scope value type")
			}
			if !elem.IsNil() {
				setValueToCollection(elem)
			}
		}
	case reflect.Struct:
		setValueToCollection(v)
	default:
		err := errors.NewDet(class.QueryValueType, "invalid scope value type")
		return err
	}

	return nil
}
