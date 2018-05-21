package jsonapi

import (
	"fmt"
	"reflect"
)

// IncludeScope is the includes information scope
// it contains the field to include from the root scope
// related subscope, and subfields to include.
type IncludeField struct {
	*StructField

	// Scope is the scope that contains the values and filters for given
	// include field
	Scope *Scope

	// RelatedScope defines the scope where the IncludedField is stored.
	RelatedScope *Scope

	NotInFieldset bool
}

// GetNonUsedPrimaries gets the id values from the RelatedScope, checks which id values were
// already stored within the colleciton root scope and return new ones.
func (i *IncludeField) GetMissingObjects() ([]interface{}, error) {

	// uniqueMissing makes it possible to get unique ids that are not already used
	uniqueMissing := map[interface{}]struct{}{}

	// Lock the SafeHashMap for given collection
	i.Scope.collectionScope.IncludedValues.Lock()
	defer i.Scope.collectionScope.IncludedValues.Unlock()

	// Get the value from the RelatedScope
	v := reflect.ValueOf(i.RelatedScope.Value)

	switch v.Kind() {
	case reflect.Slice:
		for j := 0; j < v.Len(); j++ {
			elem := v.Index(j)
			if elem.IsNil() {
				continue
			}

			if err := i.getMissingFromSingle(elem, uniqueMissing); err != nil {
				return nil, err
			}

		}
	case reflect.Ptr:
		if err := i.getMissingFromSingle(v, uniqueMissing); err != nil {
			return nil, err
		}
	default:
		err := IErrUnexpectedType
		fmt.Println(v)
		return nil, err
	}

	// Copy the notUsed map into array
	missingIDs := make([]interface{}, len(uniqueMissing))

	j := 0
	for uniqueID := range uniqueMissing {
		missingIDs[j] = uniqueID
		j++
	}

	fmt.Println(missingIDs)

	return missingIDs, nil
}

func newIncludeField(field *StructField, scope *Scope) *IncludeField {
	includeField := new(IncludeField)
	includeField.StructField = field

	// Set NewScope for given field
	includeField.Scope = scope.createModelsScope(field.relatedStruct)

	// Set the root collection scope for given scope
	includeField.Scope.collectionScope = scope.getOrCreateModelsRootScope(field.relatedStruct)
	if _, ok := includeField.Scope.collectionScope.Fieldset[includeField.jsonAPIName]; !ok {
		includeField.NotInFieldset = true
		scope.hasFieldNotInFieldset = true
	}

	// Set relatedScope for given incldudedField
	includeField.RelatedScope = scope

	includeField.Scope.rootScope.totalIncludeCount++

	return includeField
}

func (i *IncludeField) getMissingFromSingle(
	value reflect.Value,
	uniqueMissing map[interface{}]struct{},
) error {
	var (
		fieldValue = value.Elem().Field(i.getFieldIndex())
		primIndex  = i.relatedStruct.primary.getFieldIndex()

		setCollectionValues = func(ptr reflect.Value) {
			primValue := ptr.Elem().Field(primIndex)

			if primValue.IsValid() {
				primary := primValue.Interface()

				if _, ok := i.Scope.collectionScope.IncludedValues.values[primary]; !ok {
					// add to collection IDs
					i.Scope.collectionScope.IncludedValues.values[primary] = nil
					if _, ok = uniqueMissing[primary]; !ok {
						uniqueMissing[primary] = struct{}{}
						fmt.Printf("Setting id: %v\n", primary)
					}
				}
			}
		}
	)

	// Get the type of the value
	switch fieldValue.Kind() {
	case reflect.Slice:
		for j := 0; j < fieldValue.Len(); j++ {
			// set primary field within scope for given model struct
			elem := fieldValue.Index(j)
			if !elem.IsNil() {
				setCollectionValues(elem)
			}
		}
	case reflect.Ptr:
		if !fieldValue.IsNil() {
			primValue := fieldValue.Elem().Field(primIndex)
			if primValue.IsValid() {
				setCollectionValues(fieldValue)
			}
		}
	default:
		err := IErrUnexpectedType
		return err
	}

	return nil
}

func (i *IncludeField) setRelatedValue(relatedValue reflect.Value) {
	var includedScopeValue reflect.Value

	fieldValue := relatedValue.Field(i.getFieldIndex())
	switch i.refStruct.Type.Kind() {
	case reflect.Slice:
		includedScopeValue = reflect.New(i.refStruct.Type).Elem()
		includedScopeValue.Set(fieldValue)
	case reflect.Ptr:
		includedScopeValue = reflect.New(i.refStruct.Type.Elem())
		includedScopeValue.Elem().Set(fieldValue.Elem())
	}

	i.Scope.Value = includedScopeValue.Interface()
}

func (i *IncludeField) copyScopeBoundaries() {
	// copy primaries
	copy(i.Scope.PrimaryFilters, i.Scope.collectionScope.PrimaryFilters)

	// copy attribute filters
	copy(i.Scope.AttributeFilters, i.Scope.collectionScope.AttributeFilters)

	// relationships
	copy(i.Scope.RelationshipFilters, i.Scope.collectionScope.RelationshipFilters)

	// fieldset is taken by reference - copied if there is nested
	i.Scope.Fieldset = i.Scope.collectionScope.Fieldset

	for _, nested := range i.Scope.IncludedFields {
		// if the nested include is not found within the collection fieldset
		// the 'i'.Scope should have a new (not reference) Fieldset
		// with the nested field added to it
		if nested.NotInFieldset {
			// make a new fieldset if it is the same reference
			if len(i.Scope.Fieldset) == len(i.Scope.collectionScope.Fieldset) {
				// if there is more than one nested this would not happen
				i.Scope.Fieldset = make(map[string]*StructField)
				// copy fieldset
				for key, field := range i.Scope.collectionScope.Fieldset {
					i.Scope.Fieldset[key] = field
				}
			}

			//add nested
			i.Scope.Fieldset[nested.jsonAPIName] = nested.StructField
		}

		nested.copyScopeBoundaries()
	}

}
