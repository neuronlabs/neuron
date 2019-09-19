package query

import (
	"reflect"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron-core/class"
	"github.com/neuronlabs/neuron-core/log"
	"github.com/neuronlabs/neuron-core/mapping"

	"github.com/neuronlabs/neuron-core/internal"
)

// IncludeField is the includes information scope
// it contains the field to include from the root scope
// related subscope, and subfields to include.
type IncludeField struct {
	*mapping.StructField
	// Scope is the query scope that contains the values and filters for given
	// include field
	Scope *Scope
	// RelatedScope is a pointer to the main scope where the IncludedField is stored.
	RelatedScope  *Scope
	NotInFieldset bool
}

// GetMissingPrimaries gets the id values from the RelatedScope, checks which id values were
// already stored within the collection root scope and return new ones.
func (i *IncludeField) GetMissingPrimaries() ([]interface{}, error) {
	return i.getMissingPrimaries()
}

func (i *IncludeField) copy(relatedScope *Scope, root *Scope) *IncludeField {
	included := &IncludeField{StructField: i.StructField, NotInFieldset: i.NotInFieldset}
	included.Scope = i.Scope.copy(false, root)
	included.RelatedScope = relatedScope
	return included
}

// newIncludedField creates a included field within 'scope' for provided 'field'.
func newIncludeField(field *mapping.StructField, scope *Scope) *IncludeField {
	includeField := new(IncludeField)
	includeField.StructField = field

	// Set NewScope for given field
	includeField.Scope = scope.createModelsScope(field.Relationship().Struct())

	// Set the root collection scope for given scope
	includeField.Scope.collectionScope = scope.getOrCreateModelsRootScope(field.Relationship().Struct())
	if _, ok := includeField.Scope.collectionScope.Fieldset[includeField.NeuronName()]; !ok {
		includeField.NotInFieldset = true
		scope.hasFieldNotInFieldset = true
	}

	// Set relatedScope for given incldudedField
	includeField.RelatedScope = scope
	includeField.Scope.rootScope.totalIncludeCount++

	return includeField
}

func (i *IncludeField) copyScopeBoundaries() {
	// copy primaries
	i.Scope.PrimaryFilters = make([]*FilterField, len(i.Scope.collectionScope.PrimaryFilters))
	copy(i.Scope.PrimaryFilters, i.Scope.collectionScope.PrimaryFilters)

	// copy attribute filters
	i.Scope.AttributeFilters = make([]*FilterField, len(i.Scope.collectionScope.AttributeFilters))
	copy(i.Scope.AttributeFilters, i.Scope.collectionScope.AttributeFilters)

	// copy filterKeyFilters
	i.Scope.FilterKeyFilters = make([]*FilterField, len(i.Scope.collectionScope.FilterKeyFilters))
	copy(i.Scope.FilterKeyFilters, i.Scope.collectionScope.FilterKeyFilters)

	// relationships
	i.Scope.RelationFilters = make([]*FilterField, len(i.Scope.collectionScope.RelationFilters))
	copy(i.Scope.RelationFilters, i.Scope.collectionScope.RelationFilters)

	//copy foreignKeyFilters
	i.Scope.ForeignFilters = make([]*FilterField, len(i.Scope.collectionScope.ForeignFilters))
	copy(i.Scope.ForeignFilters, i.Scope.collectionScope.ForeignFilters)

	// copy language filters
	if i.Scope.collectionScope.LanguageFilters != nil {
		i.Scope.LanguageFilters = i.Scope.collectionScope.LanguageFilters
	}

	// fieldset is taken by reference - copied if there is nested
	i.Scope.Fieldset = i.Scope.collectionScope.Fieldset

	i.Scope.store[internal.ControllerStoreKey] = i.Scope.collectionScope.store[internal.ControllerStoreKey]

	for _, nested := range i.Scope.includedFields {
		// if the nested include is not found within the collection fieldset
		// the 'i'.Scope should have a new (not reference) Fieldset
		// with the nested field added to it
		if nested.NotInFieldset {
			// make a new fieldset if it is the same reference
			if len(i.Scope.Fieldset) == len(i.Scope.collectionScope.Fieldset) {
				// if there is more than one nested this would not happen
				i.Scope.Fieldset = make(map[string]*mapping.StructField)
				// copy fieldset
				for key, field := range i.Scope.collectionScope.Fieldset {
					i.Scope.Fieldset[key] = field
				}
			}

			//add nested
			i.Scope.Fieldset[nested.NeuronName()] = nested.StructField
		}
		nested.copyScopeBoundaries()
	}
}

func (i *IncludeField) getMissingPrimaries() ([]interface{}, error) {
	// uniqueMissing makes it possible to get unique ids that are not already used
	uniqueMissing := map[interface{}]struct{}{}

	// Lock the SafeHashMap for given collection
	i.Scope.collectionScope.includedValues.Lock()
	defer i.Scope.collectionScope.includedValues.Unlock()

	// Get the value from the RelatedScope
	v := reflect.ValueOf(i.RelatedScope.Value)

	// RelatedScope Value must be a pointer type
	if v.Kind() != reflect.Ptr {
		return nil, errors.NewDet(class.QueryValueType, "included scope with invalid value")
	}

	// Check if is nil
	if !v.IsNil() {
		v = v.Elem()
		switch v.Kind() {
		case reflect.Slice:
			if log.Level() == log.LDEBUG3 {
				log.Debug3f("Getting from slice")
			}

			for j := 0; j < v.Len(); j++ {
				elem := v.Index(j)
				if elem.Kind() == reflect.Ptr {
					if elem.IsNil() {
						continue
					}
					elem = elem.Elem()
				}
				if log.Level() == log.LDEBUG3 {
					log.Debug3f("i'th: %d element: %v", j, elem.Interface())
				}
				if err := i.getMissingFromSingle(elem, uniqueMissing); err != nil {
					return nil, err
				}
			}
		case reflect.Struct:
			if err := i.getMissingFromSingle(v, uniqueMissing); err != nil {
				return nil, err
			}
		default:
			log.Errorf("Unexpected Included Scope Value type: %s", v.Type())
			err := errors.NewDet(class.QueryValueType, "unexpected included scope value type")
			return nil, err
		}
	}

	// Copy the notUsed map into array
	missingIDs := make([]interface{}, len(uniqueMissing))

	j := 0
	for uniqueID := range uniqueMissing {
		missingIDs[j] = uniqueID
		j++
	}

	return missingIDs, nil
}

func (i *IncludeField) getMissingFromSingle(value reflect.Value, uniqueMissing map[interface{}]struct{}) error {
	fieldValue := value.FieldByIndex(i.FieldIndex())
	// get related model's primary index
	primIndex := i.StructField.Relationship().Struct().Primary().FieldIndex()

	// setCollectionValues sets the relationship field primary index into the uniqueMissing map
	setCollectionValues := func(model reflect.Value) {
		primValue := model.FieldByIndex(primIndex)
		primary := primValue.Interface()

		if _, ok := i.Scope.collectionScope.includedValues.Values()[primary]; !ok {
			// add to collection IDs
			i.Scope.collectionScope.includedValues.UnsafeSet(primary, nil)
			if _, ok = uniqueMissing[primary]; !ok {
				uniqueMissing[primary] = struct{}{}
			} else {
				if log.Level() == log.LDEBUG3 {
					log.Debug3f("Primary: '%v' already exists - duplicated value", primary)
				}
			}
		}
	}

	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return nil
		}
		fieldValue = fieldValue.Elem()
	}

	// Get the type of the value
	switch fieldValue.Kind() {
	case reflect.Slice:
		for j := 0; j < fieldValue.Len(); j++ {
			// set primary field within scope for given model struct
			// elem is the model at j'th index in the slice
			elem := fieldValue.Index(j)

			// it may be a pointer to struct
			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}
			setCollectionValues(elem)
		}
	case reflect.Struct:
		log.Debug3f("Field is Struct")
		setCollectionValues(fieldValue)
	default:
		log.Debug3f("Unexpect type: '%s' in the relationship field's value.", fieldValue.Type())
		err := errors.NewDet(class.QueryValueType, "unexpected included scope value type")
		return err
	}
	return nil
}
