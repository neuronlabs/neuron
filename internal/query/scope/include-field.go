package scope

import (
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query/filters"
	"github.com/neuronlabs/neuron/internal/query/sorts"
	"github.com/neuronlabs/neuron/log"
	"reflect"
)

// IncludeField is the includes information scope
// it contains the field to include from the root scope
// related subscope, and subfields to include.
type IncludeField struct {
	*models.StructField

	// Scope is the scope that contains the values and filters for given
	// include field
	Scope *Scope

	// RelatedScope is a pointer to the scope where the IncludedField is stored.
	RelatedScope *Scope

	NotInFieldset bool
}

func (i *IncludeField) copy(relatedScope *Scope, root *Scope) *IncludeField {
	included := &IncludeField{StructField: i.StructField, NotInFieldset: i.NotInFieldset}
	included.Scope = i.Scope.copy(false, root)
	included.RelatedScope = relatedScope
	return included
}

// GetMissingPrimaries gets the id values from the RelatedScope, checks which id values were
// already stored within the colleciton root scope and return new ones.
func (i *IncludeField) GetMissingPrimaries() ([]interface{}, error) {
	return i.getMissingPrimaries()
}

func newIncludeField(field *models.StructField, scope *Scope) *IncludeField {
	includeField := new(IncludeField)
	includeField.StructField = field

	// Set NewScope for given field

	includeField.Scope = scope.createModelsScope(field.Relationship().Struct())

	// Set the root collection scope for given scope
	includeField.Scope.collectionScope = scope.getOrCreateModelsRootScope(field.Relationship().Struct())
	if _, ok := includeField.Scope.collectionScope.fieldset[includeField.ApiName()]; !ok {
		includeField.NotInFieldset = true
		scope.hasFieldNotInFieldset = true
	}

	// Set relatedScope for given incldudedField
	includeField.RelatedScope = scope

	includeField.Scope.rootScope.totalIncludeCount++

	return includeField
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
		return nil, internal.IErrUnexpectedType
	}

	// Check if is nil
	if !v.IsNil() {
		v = v.Elem()
		switch v.Kind() {
		case reflect.Slice:
			log.Debugf("Getting from slice")
			for j := 0; j < v.Len(); j++ {
				elem := v.Index(j)
				if elem.Kind() == reflect.Ptr {
					if elem.IsNil() {
						continue
					}
					elem = elem.Elem()
				}
				log.Debugf("i'th: %d element", j)
				if err := i.getMissingFromSingle(elem, uniqueMissing); err != nil {
					return nil, err
				}

			}
		case reflect.Struct:
			log.Debugf("Getting from single")
			if err := i.getMissingFromSingle(v, uniqueMissing); err != nil {
				return nil, err
			}

		default:
			err := internal.IErrUnexpectedType
			log.Errorf("Unexpected Included Scope Value type: %s", v.Type())
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

func (i *IncludeField) getMissingFromSingle(
	value reflect.Value,
	uniqueMissing map[interface{}]struct{},
) error {

	var (
		fieldValue = value.FieldByIndex(i.FieldIndex())

		// get related model's primary index
		primIndex = models.FieldsRelatedModelStruct(i.StructField).PrimaryField().FieldIndex()

		// setCollectionValues sets the relationship field primary index into the uniqueMissing map
		setCollectionValues = func(model reflect.Value) {
			primValue := model.FieldByIndex(primIndex)

			if primValue.IsValid() {
				primary := primValue.Interface()

				if _, ok := i.Scope.collectionScope.includedValues.Values()[primary]; !ok {
					// add to collection IDs
					i.Scope.collectionScope.includedValues.UnsafeSet(primary, nil)
					if _, ok = uniqueMissing[primary]; !ok {
						uniqueMissing[primary] = struct{}{}
					} else {
						log.Debugf("Primary: '%v' already exists", primary)
					}
				}
			} else {
				log.Debugf("Primary value is not valid")
			}
		}
	)

	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return nil
		}
		fieldValue = fieldValue.Elem()
	}

	// Get the type of the value
	switch fieldValue.Kind() {
	case reflect.Slice:
		log.Debugf("Field is Slice")
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
		log.Debugf("Field is Struct")
		setCollectionValues(fieldValue)
	default:
		log.Debugf("Unexpect type: '%s' in the relationship field's value.", fieldValue.Type())
		err := internal.IErrUnexpectedType
		return err
	}

	return nil
}

func (i *IncludeField) setRelationshipValue(relatedValue reflect.Value) {
	var includedScopeValue reflect.Value
	fieldValue := relatedValue.FieldByIndex(i.FieldIndex())

	switch i.FieldType().Kind() {
	case reflect.Slice:
		if fieldValue.Len() == 0 {
			i.Scope.Value = reflect.New(i.FieldType()).Elem().Interface()
			return
		}
		includedScopeValue = reflect.New(i.FieldType()).Elem()
		includedScopeValue.Set(fieldValue)
	case reflect.Ptr:
		if fieldValue.IsNil() {
			i.Scope.Value = nil
			return
		}
		includedScopeValue = reflect.New(i.FieldType().Elem())
		includedScopeValue.Elem().Set(fieldValue.Elem())
	}

	i.Scope.Value = includedScopeValue.Interface()
	return
}

func (i *IncludeField) copyScopeBoundaries() {
	// copy primaries
	i.Scope.primaryFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.primaryFilters))
	copy(i.Scope.primaryFilters, i.Scope.collectionScope.primaryFilters)

	// copy attribute filters
	i.Scope.attributeFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.attributeFilters))
	copy(i.Scope.attributeFilters, i.Scope.collectionScope.attributeFilters)

	// copy filterKeyFilters
	i.Scope.keyFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.keyFilters))
	copy(i.Scope.keyFilters, i.Scope.collectionScope.keyFilters)

	// relationships
	i.Scope.relationshipFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.relationshipFilters))
	copy(i.Scope.relationshipFilters, i.Scope.collectionScope.relationshipFilters)

	//copy foreignKeyFilters
	i.Scope.foreignFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.foreignFilters))
	copy(i.Scope.foreignFilters, i.Scope.collectionScope.foreignFilters)

	// copy language filters

	if i.Scope.collectionScope.languageFilters != nil {
		i.Scope.languageFilters = i.Scope.collectionScope.languageFilters
	}

	// fieldset is taken by reference - copied if there is nested
	i.Scope.fieldset = i.Scope.collectionScope.fieldset
	i.Scope.ctx = i.Scope.collectionScope.ctx

	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.UseLinks); ok {
		i.Scope.fContainer.Set(flags.UseLinks, f)
	}

	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.ReturnPatchContent); ok {
		i.Scope.fContainer.Set(flags.ReturnPatchContent, f)
	}

	for _, nested := range i.Scope.includedFields {
		// if the nested include is not found within the collection fieldset
		// the 'i'.Scope should have a new (not reference) Fieldset
		// with the nested field added to it
		if nested.NotInFieldset {
			// make a new fieldset if it is the same reference
			if len(i.Scope.fieldset) == len(i.Scope.collectionScope.fieldset) {
				// if there is more than one nested this would not happen
				i.Scope.fieldset = make(map[string]*models.StructField)
				// copy fieldset
				for key, field := range i.Scope.collectionScope.fieldset {
					i.Scope.fieldset[key] = field
				}
			}

			//add nested
			i.Scope.fieldset[nested.ApiName()] = nested.StructField
		}

		nested.copyScopeBoundaries()
	}

}

func (i *IncludeField) copyPresetFullParameters() {
	// copy primaries
	i.Scope.primaryFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.primaryFilters))
	copy(i.Scope.primaryFilters, i.Scope.collectionScope.primaryFilters)

	// copy attribute filters
	i.Scope.attributeFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.attributeFilters))
	copy(i.Scope.attributeFilters, i.Scope.collectionScope.attributeFilters)

	// copy filterKeyFilters
	i.Scope.keyFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.keyFilters))
	copy(i.Scope.keyFilters, i.Scope.collectionScope.keyFilters)

	// relationships
	i.Scope.relationshipFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.relationshipFilters))
	copy(i.Scope.relationshipFilters, i.Scope.collectionScope.relationshipFilters)

	//copy foreignKeyFilters
	i.Scope.foreignFilters = make([]*filters.FilterField, len(i.Scope.collectionScope.foreignFilters))
	copy(i.Scope.foreignFilters, i.Scope.collectionScope.foreignFilters)

	// copy language filters
	if i.Scope.collectionScope.languageFilters != nil {
		i.Scope.languageFilters = i.Scope.collectionScope.languageFilters
	}

	// fieldset is taken by reference - copied if there is nested
	// i.Scope.fieldset = i.Scope.collectionScope.fieldset

	i.Scope.sortFields = make([]*sorts.SortField, len(i.Scope.collectionScope.sortFields))
	copy(i.Scope.sortFields, i.Scope.collectionScope.sortFields)

	i.Scope.pagination = i.Scope.collectionScope.pagination
	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.UseLinks); ok {
		i.Scope.fContainer.Set(flags.UseLinks, f)
	}

	if f, ok := i.Scope.collectionScope.fContainer.Get(flags.ReturnPatchContent); ok {
		i.Scope.fContainer.Set(flags.ReturnPatchContent, f)
	}

	for _, nested := range i.Scope.includedFields {
		// if the nested include is not found within the collection fieldset
		// the 'i'.Scope should have a new (not reference) Fieldset
		// with the nested field added to it
		if nested.NotInFieldset {
			// make a new fieldset if it is the same reference
			if len(i.Scope.fieldset) == len(i.Scope.collectionScope.fieldset) {
				// if there is more than one nested this would not happen
				i.Scope.fieldset = make(map[string]*models.StructField)
				// copy fieldset
				for key, field := range i.Scope.collectionScope.fieldset {
					i.Scope.fieldset[key] = field
				}
			}

			//add nested
			i.Scope.fieldset[nested.ApiName()] = nested.StructField
		}

		nested.copyPresetFullParameters()
	}
}
