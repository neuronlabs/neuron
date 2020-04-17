package query

import (
	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/mapping"
)

// IncludedField is the includes information scope
// it contains the field to include from the root scope
// related subScope, and subfields to include.
type IncludedField struct {
	StructField *mapping.StructField
	Fieldset    map[string]*mapping.StructField
}

func (i *IncludedField) copy() *IncludedField {
	copiedIncludedField := &IncludedField{StructField: i.StructField}
	if i.Fieldset != nil {
		copiedIncludedField.Fieldset = make(map[string]*mapping.StructField, len(i.Fieldset))
		for k, v := range i.Fieldset {
			copiedIncludedField.Fieldset[k] = v
		}
	}
	return copiedIncludedField
}

// SetFieldset sets the fieldset for given included
func (i *IncludedField) setFieldset(fields ...string) error {
	model := i.StructField.Relationship().Struct()
	for _, field := range fields {
		sField, ok := model.Field(field)
		if !ok {
			return errors.NewDetf(class.QueryFieldsetUnknownField,
				"provided invalid field: '%s' in the fieldset of included field: '%s'", field, i.StructField.Name())
		}
		// check if the field is not already set
		_, ok = i.Fieldset[sField.NeuronName()]
		if ok {
			return errors.NewDetf(class.QueryFieldsetDuplicate,
				"provided field: '%s' in the fieldset of included field: '%s' is already in the included fieldset",
				field, i.StructField.Name())
		}
		i.Fieldset[sField.NeuronName()] = sField
	}
	return nil
}

// // GetMissingPrimaries gets the id values from the RelatedScope, checks which id values were
// // already stored within the collection root scope and return new ones.
// func (i *IncludedField) GetMissingPrimaries() ([]interface{}, error) {
// 	return i.getMissingPrimaries()
// }
//
// func (i *IncludedField) copy(relatedScope *Scope, root *Scope) *IncludedField {
// 	included := &IncludedField{StructField: i.StructField, NotInFieldset: i.NotInFieldset}
// 	included.Scope = i.Scope.copy(false, root)
// 	included.RelatedScope = relatedScope
// 	return included
// }
//
// // newIncludedField creates a included field within 'scope' for provided 'field'.
// func newIncludeField(field *mapping.StructField, scope *Scope) *IncludedField {
// 	includeField := new(IncludedField)
// 	includeField.StructField = field
//
// 	// Set NewScope for given field
// 	includeField.Scope = scope.createModelsScope(scope.Controller(), field.Relationship().Struct())
//
// 	// Set the root collection scope for given scope
// 	includeField.Scope.collectionScope = scope.getOrCreateModelsRootScope(scope.Controller(), field.Relationship().Struct())
// 	if _, ok := includeField.Scope.collectionScope.Fieldset[includeField.NeuronName()]; !ok {
// 		includeField.NotInFieldset = true
// 		scope.hasFieldNotInFieldset = true
// 	}
//
// 	// Set relatedScope for given incldudedField
// 	includeField.RelatedScope = scope
// 	includeField.Scope.rootScope.totalIncludeCount++
//
// 	return includeField
// }
//
// func (i *IncludedField) getMissingPrimaries() ([]interface{}, error) {
// 	// uniqueMissing makes it possible to get unique ids that are not already used
// 	uniqueMissing := map[interface{}]struct{}{}
//
// 	// Lock the SafeHashMap for given collection
// 	i.Scope.collectionScope.includedValues.Lock()
// 	defer i.Scope.collectionScope.includedValues.Unlock()
//
// 	// Get the value from the RelatedScope
// 	v := reflect.ValueOf(i.RelatedScope.Value)
//
// 	// RelatedScope Value must be a pointer type
// 	if v.Kind() != reflect.Ptr {
// 		return nil, errors.NewDet(class.QueryValueType, "included scope with invalid value")
// 	}
//
// 	// Check if is nil
// 	if !v.IsNil() {
// 		v = v.Elem()
// 		switch v.Kind() {
// 		case reflect.Slice:
// 			if log.Level() == log.LevelDebug3 {
// 				log.Debug3f("Getting from slice")
// 			}
//
// 			for j := 0; j < v.Len(); j++ {
// 				elem := v.Index(j)
// 				if elem.Kind() == reflect.Ptr {
// 					if elem.IsNil() {
// 						continue
// 					}
// 					elem = elem.Elem()
// 				}
// 				if log.Level() == log.LevelDebug3 {
// 					log.Debug3f("i'th: %d element: %v", j, elem.Interface())
// 				}
// 				if err := i.getMissingFromSingle(elem, uniqueMissing); err != nil {
// 					return nil, err
// 				}
// 			}
// 		case reflect.Struct:
// 			if err := i.getMissingFromSingle(v, uniqueMissing); err != nil {
// 				return nil, err
// 			}
// 		default:
// 			log.Errorf("Unexpected Included Scope Value type: %s", v.Type())
// 			err := errors.NewDet(class.QueryValueType, "unexpected included scope value type")
// 			return nil, err
// 		}
// 	}
//
// 	// Copy the notUsed map into array
// 	missingIDs := make([]interface{}, len(uniqueMissing))
//
// 	j := 0
// 	for uniqueID := range uniqueMissing {
// 		missingIDs[j] = uniqueID
// 		j++
// 	}
//
// 	return missingIDs, nil
// }
//
// func (i *IncludedField) getMissingFromSingle(value reflect.Value, uniqueMissing map[interface{}]struct{}) error {
// 	fieldValue := value.FieldByIndex(i.FieldIndex())
// 	// get related model's primary index
// 	primIndex := i.StructField.Relationship().Struct().Primary().FieldIndex()
//
// 	// setCollectionValues sets the relationship field primary index into the uniqueMissing map
// 	setCollectionValues := func(model reflect.Value) {
// 		primValue := model.FieldByIndex(primIndex)
// 		primary := primValue.Interface()
//
// 		if _, ok := i.Scope.collectionScope.includedValues.Values()[primary]; !ok {
// 			// add to collection IDs
// 			i.Scope.collectionScope.includedValues.UnsafeSet(primary, nil)
// 			if _, ok = uniqueMissing[primary]; !ok {
// 				uniqueMissing[primary] = struct{}{}
// 			} else if log.Level() == log.LevelDebug3 {
// 				log.Debug3f("Primary: '%v' already exists - duplicated value", primary)
// 			}
// 		}
// 	}
//
// 	if fieldValue.Kind() == reflect.Ptr {
// 		if fieldValue.IsNil() {
// 			return nil
// 		}
// 		fieldValue = fieldValue.Elem()
// 	}
//
// 	// Get the type of the value
// 	switch fieldValue.Kind() {
// 	case reflect.Slice:
// 		for j := 0; j < fieldValue.Len(); j++ {
// 			// set primary field within scope for given model struct
// 			// elem is the model at j'th index in the slice
// 			elem := fieldValue.Index(j)
//
// 			// it may be a pointer to struct
// 			if elem.Kind() == reflect.Ptr {
// 				if elem.IsNil() {
// 					continue
// 				}
// 				elem = elem.Elem()
// 			}
// 			setCollectionValues(elem)
// 		}
// 	case reflect.Struct:
// 		log.Debug3f("Field is Struct")
// 		setCollectionValues(fieldValue)
// 	default:
// 		log.Debug3f("Unexpect type: '%s' in the relationship field's value.", fieldValue.Type())
// 		err := errors.NewDet(class.QueryValueType, "unexpected included scope value type")
// 		return err
// 	}
// 	return nil
// }
