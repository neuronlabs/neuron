package query

import (
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"
)

func (s *Scope) getPrimaryFieldValues() ([]interface{}, error) {
	if s.Value == nil {
		return nil, errors.NewDet(class.QueryNilValue, "no scope value provided")
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
	switch {
	case s.mStruct != foreign.Struct():
		log.Debugf("Scope's ModelStruct: %s, doesn't match foreign key ModelStruct: '%s' ", s.mStruct.Collection(), foreign.Struct().Collection())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key mismatched ModelStruct")
	case foreign.Kind() != mapping.KindForeignKey:
		log.Debugf("'foreign' field is not a ForeignKey: %s", foreign.Kind())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key is not a valid ForeignKey")
	case s.Value == nil:
		return nil, errors.NewDet(class.QueryNilValue, "provided nil scope value")
	}

	// initialize the array
	values := []interface{}{}

	// set the adding function
	addForeignKey := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(foreign.FieldIndex())
		values = append(values, primaryValue.Interface())
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.QueryNilValue, "provided no scope value")
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
	switch {
	case s.mStruct != foreign.Struct():
		log.Debugf("Scope's ModelStruct: %s, doesn't match foreign key ModelStruct: '%s' ", s.mStruct.Collection(), foreign.Struct().Collection())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key mismatched ModelStruct")
	case foreign.Kind() != mapping.KindForeignKey:
		log.Debugf("'foreign' field is not a ForeignKey: %s", foreign.Kind())
		return nil, errors.NewDet(class.InternalQueryInvalidField, "foreign key is not a valid ForeignKey")
	case s.Value == nil:
		return nil, errors.NewDet(class.QueryNilValue, "provided nil scope value")
	}
	// initialize the array
	foreignKeys := make(map[interface{}]struct{})
	values := []interface{}{}

	// set the adding function
	addForeignKey := func(single reflect.Value) {
		primaryValue := single.FieldByIndex(foreign.FieldIndex())
		foreignKeys[primaryValue.Interface()] = struct{}{}
	}

	v := reflect.ValueOf(s.Value)
	if v.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.QueryNilValue, "provided no scope value")
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

	for foreignKey := range foreignKeys {
		values = append(values, foreignKey)
	}
	return values, nil
}
