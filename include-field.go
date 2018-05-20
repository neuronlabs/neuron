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
}

// GetNonUsedPrimaries gets the id values from the RelatedScope, checks which id values were
// already stored within the colleciton root scope and return new ones.
func (i *IncludeField) GetNonUsedPrimaries() (notUsedIDS []interface{}, err error) {

	// nonUsed makes it possible to get unique ids that are not already used
	nonUsed := map[interface{}]struct{}{}

	// Lock the HashSet for given collection
	i.Scope.collectionScope.IncludeIDValues.Lock()
	defer i.Scope.collectionScope.IncludeIDValues.Unlock()

	// Get the value from the RelatedScope
	v := reflect.ValueOf(i.RelatedScope.Value)

	switch v.Kind() {
	case reflect.Slice:
		for j := 0; j < v.Len(); j++ {
			elem := v.Index(j)
			if elem.IsNil() {
				continue
			}

			if err := i.getNonUsedFromSingle(elem, nonUsed); err != nil {
				return nil, err
			}

		}
	case reflect.Ptr:

		if err := i.getNonUsedFromSingle(v, nonUsed); err != nil {
			return nil, err
		}
	default:
		err := IErrUnexpectedType
		fmt.Println(v)
		return nil, err
	}

	// Copy the notUsed map into array
	notUsedIDS = make([]interface{}, len(nonUsed))

	j := 0
	for uniqueID := range nonUsed {
		notUsedIDS[j] = uniqueID
		j++
	}

	return notUsedIDS, nil
}

func newIncludeField(field *StructField, scope *Scope) *IncludeField {
	includeField := new(IncludeField)
	includeField.StructField = field

	// Set NewScope for given field
	includeField.Scope = scope.createModelsScope(field.relatedStruct)
	// Set the root collection scope for given scope
	includeField.Scope.collectionScope = scope.getOrCreateModelsRootScope(field.relatedStruct)

	// Set relatedScope for given incldudedField
	includeField.RelatedScope = scope

	includeField.Scope.rootScope.totalIncludeCount++

	return includeField
}

func (i *IncludeField) getNonUsedFromSingle(
	value reflect.Value,
	uniques map[interface{}]struct{},
) error {

	fieldValue := value.Elem().Field(i.getFieldIndex())
	primIndex := i.relatedStruct.primary.getFieldIndex()

	setCollectionIDs := func(ptr reflect.Value) {
		primValue := ptr.Elem().Field(primIndex)

		if primValue.IsValid() {
			v := primValue.Interface()
			if _, ok := i.Scope.collectionScope.IncludeIDValues.values[v]; !ok {
				// add to collection IDs
				i.Scope.collectionScope.IncludeIDValues.values[v] = struct{}{}
				if _, ok = uniques[v]; !ok {
					uniques[v] = struct{}{}
				}
			}
		}
	}

	switch fieldValue.Kind() {
	case reflect.Slice:
		for j := 0; j < fieldValue.Len(); j++ {
			// set primary field within scope for given model struct
			elem := fieldValue.Index(j)
			if !elem.IsNil() {
				setCollectionIDs(elem)
			}
		}
	case reflect.Ptr:
		if !fieldValue.IsNil() {
			primValue := fieldValue.Elem().Field(primIndex)
			if primValue.IsValid() {
				setCollectionIDs(fieldValue)
			}
		}
	default:
		err := IErrUnexpectedType
		return err
	}

	return nil
}

func (i *IncludeField) prepareScopeValue() {
	switch i.StructField.refStruct.Type.Kind() {
	case reflect.Ptr:
	case reflect.Slice:
	}

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

// func (i *IncludeField) copySingleValue(scopeValue, value reflect.Value) {
// 	fieldValue := value.Elem().Field(i.getFieldIndex())
// 	primIndex := i.relatedStruct.primary.getFieldIndex()

// 	switch fieldValue.Kind() {
// 	case reflect.Slice:
// 		for j := 0; j < fieldValue.Len(); j++ {
// 			// set primary field within scope for given model struct
// 			elem := fieldValue.Index(j)
// 			if !elem.IsNil() {
// 				reflect.Append(scopeValue, elem)
// 			}
// 		}
// 	case reflect.Ptr:

// 		if !fieldValue.IsNil() {
// 			primValue := fieldValue.Elem().Field(primIndex)
// 			if primValue.IsValid() {
// 				scopeValue.Set(primVa)
// 			}
// 		}
// 	default:
// 		err := IErrUnexpectedType
// 		return err
// 	}
// }

// copyFilters copies the filters from the collection root scope
// used to copy the filters created by the query
func (i *IncludeField) copyFilters() {
	// copy primaries
	i.Scope.PrimaryFilters = i.Scope.collectionScope.PrimaryFilters

	// copy attribute filters
	i.Scope.AttributeFilters = i.Scope.collectionScope.AttributeFilters

	// relationships
	i.Scope.RelationshipFilters = i.Scope.collectionScope.RelationshipFilters

	for _, nested := range i.Scope.IncludedFields {
		nested.copyFilters()
	}
}

// getOrCreateNestedInclude - get from includedSubfiedls or if no such field
// // create new included.
// func (i *IncludeField) getOrCreateNestedInclude(field *StructField) *IncludeField {
// 	if i.IncludedSubfields == nil {
// 		i.IncludedSubfields = make([]*IncludeField, 0)
// 	}
// 	for _, subfield := range i.IncludedSubfields {
// 		if subfield.getFieldIndex() == field.getFieldIndex() {
// 			return subfield
// 		}
// 	}

// 	includeField := newIncludeField(field, i.Scope)

// 	i.IncludedSubfields = append(i.IncludedSubfields, includeField)
// 	return includeField
// }

// func (i *IncludeField) includeSubfield(includeField *IncludeField) {
// 	i.IncludedSubfields = append(i.IncludedSubfields, includeField)
// }
