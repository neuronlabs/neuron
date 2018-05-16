package jsonapi

import (
	"reflect"
	"strings"
)

// IncludeScope is the includes information scope
// it contains the field to include from the root scope
// related subscope, and subfields to include.
type IncludeField struct {
	*StructField

	// RelatedScope is the scope where given include is described
	Scope *Scope

	// if subField is included
	IncludedSubfields []*IncludeField

	RootScope *Scope
}

func newIncludeField(field *StructField, scope *Scope) *IncludeField {
	includeField := new(IncludeField)
	includeField.StructField = field
	includeField.Scope = scope.createModelsScope(field.relatedStruct)
	return includeField
}

func (i *IncludeField) getNonUsedFromSingle(value reflect.Value) (interface{}, error) {

	fieldValue := value.Elem().Field(i.getFieldIndex())

	switch fieldValue.Kind() {
	case reflect.Slice:
		for j := 0; j < fieldValue.Len(); ji++ {
			// set primary field within scope for given model struct
			elem := fieldValue.Index(j)
			if !elem.IsNil() {
				primValue := elem.Elem().Field(primIndex)

				if primValue.IsValid() {

					includedScope.IncludeIDValues[primValue.Interface()] = struct{}{}
				}
			}
		}
	case reflect.Ptr:
		if !fieldValue.IsNil() {
			primValue := fieldValue.Elem().Field(primIndex)
			if primValue.IsValid() {
				includedScope.IncludeIDValues[primValue.Interface()] = struct{}{}
			}
		} else {
			// error
		}
	default:
		err = IErrUnexpectedType
		return
	}

	return nil, nil
}

func (i *IncludeField) GetNonUsedIDS() ([]interface{}, error) {
	nonUsed := map[interface{}]struct{}{}
	i.Scope.IncludeIDValues.Lock()
	defer i.Scope.IncludeIDValues.Unlock()

	v := reflect.ValueOf(i.RootScope.Value)
	switch v.Kind() {
	case reflect.Slice:
		for j := 0; j < v.Len(); j++ {
			elem := v.Index(j)
			if elem.IsNil() {
				continue
			}

			value, err := i.getNonUsedFromSingle(elem)
			if err != nil {
				return nil, err
			}
			if _, ok := nonUsed[value]; !ok {
				nonUsed[value] = struct{}{}
			}
		}
	case reflect.Ptr:
		value, err := i.getNonUsedFromSingle(v)
		if err != nil {
			return nil, err
		}
		if _, ok := nonUsed[value]; !ok {
			nonUsed[value] = struct{}{}
		}
	default:
		err := IErrUnexpectedType
		return nil, err
	}

	notUsedIDS := make([]interface{}, len(nonUsed))

	j := 0
	for uniqueID := range nonUsed {
		notUsedIDS[j] = uniqueID
		j++
	}
	return notUsedIDS, nil
}

func (i *IncludeField) buildNestedInclude(nested string, scope *Scope,
) (errs []*ErrorObject) {

	var (
		nestedInclude *IncludeField
		relationField *StructField
		ok            bool
	)

	// fmt.Printf("Nested include: %v, mStruct: %v, relStruct: %v.\n", i.fieldName, i.mStruct, i.relatedStruct)
	// check in the field's model if it contains this nested include
	relationField, ok = i.relatedStruct.relationships[nested]
	if !ok {
		// if no relationship found, then check if it is possible to separate dots from 'nested'
		// no relationship found check nesteds
		index := strings.Index(nested, annotationNestedSeperator)
		if index == -1 {
			errs = append(errs, errNoRelationship(i.relatedStruct.collectionType, nested))
			return
		}

		// field part of included (field.subfield)
		field := nested[:index]
		relationField, ok = i.relatedStruct.relationships[field]
		if !ok {
			// still not found - then add error and return
			errs = append(errs, errNoRelationship(i.relatedStruct.collectionType, field))
			return
		}

		nestedInclude = i.getOrCreateNestedInclude(relationField)
		// build recursively nested fields
		errs = append(errs, nestedInclude.buildNestedInclude(nested[index+1:], scope)...)
	} else {
		nestedInclude = i.getOrCreateNestedInclude(relationField)
	}

	_ = scope.getOrCreateIncludedScope(nestedInclude.relatedStruct)
	return
}

// getOrCreateNestedInclude - get from includedSubfiedls or if no such field
// create new included.
func (i *IncludeField) getOrCreateNestedInclude(field *StructField) *IncludeField {
	if i.IncludedSubfields == nil {
		i.IncludedSubfields = make([]*IncludeField, 0)
	}
	for _, subfield := range i.IncludedSubfields {
		if subfield.getFieldIndex() == field.getFieldIndex() {
			return subfield
		}
	}
	includeField := new(IncludeField)
	includeField.StructField = field
	i.IncludedSubfields = append(i.IncludedSubfields, includeField)
	return includeField
}

func (i *IncludeField) includeSubfield(includeField *IncludeField) {
	i.IncludedSubfields = append(i.IncludedSubfields, includeField)
}
