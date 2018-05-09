package jsonapi

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNoParamsInContext = errors.New("No parameters in the request Context.")
)

var (
	// used for errors
	maxPermissibleDuplicates = 3
)

// Scope contains information about given query for specific collection
// if the query defines the different collection than the main scope, then
// every detail about querying (fieldset, filters, sorts) are within new scopes
// kept in the Subscopes
type Scope struct {
	// Struct is a modelStruct this scope is based on
	Struct *ModelStruct

	// Value is the values or / value of the queried object / objects
	Value        interface{}
	valueAddress interface{}

	// CollectionScopes contains filters, fieldsets and values for included collections
	// every collection that is inclued would contain it's subscope
	// if filters, fieldsets are set for non-included scope error should occur
	IncludedScopes map[*ModelStruct]*Scope

	// Included contain fields to include. If the included field is a relationship type, then
	// specific includefield contains information about it
	IncludedFields []*IncludeField

	// IncludeValues contain unique values for given include fields
	// the key is the - primary key value
	IncludeValues map[interface{}]struct{}

	// PrimaryFilters contain filter for the primary field
	PrimaryFilters []*FilterField

	// RelationshipFilters contain relationship field filters
	RelationshipFilters []*FilterField

	// AttributeFilters contain fitler for the attribute fields
	AttributeFilters []*FilterField

	// Fields represents fieldset used for this scope - jsonapi 'fields[collection]'
	Fieldset map[string]*StructField

	// SortFields
	Sorts []*SortField

	// Pagination
	Pagination *Pagination

	IsMany            bool
	errorLimit        int
	maxNestedLevel    int
	currentErrorCount int
}

// Returns the collection name for given scope
func (s *Scope) GetCollection() string {
	return s.Struct.collectionType
}

func (s *Scope) GetValueAddress() interface{} {
	return s.valueAddress
}

// NewValueMany creates empty slice of ptr value for given scope
// value is of type []*ModelStruct.Type
func (s *Scope) NewValueMany() {
	s.newValueMany()
}

// NewValueSingle creates new value for given scope of a type *ModelStruct.Type
func (s *Scope) NewValueSingle() {
	s.newValueSingle()
}

// SetIncludedPrimaries sets the primary fields for given model
func (s *Scope) SetIncludedPrimaries() (err error) {
	if s.Value == nil {
		err = errors.New("Nil value provided for scope.")
		return
	}

	v := reflect.ValueOf(s.Value)
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.IsNil() {
				continue
			}
			if err = s.setIncludedPrimary(elem.Interface()); err != nil {
				return
			}
		}
	case reflect.Ptr:
		if err = s.setIncludedPrimary(s.Value); err != nil {
			return
		}
	default:
		err = IErrUnexpectedType
		return
	}

	for _, includedScope := range s.IncludedScopes {
		includedScope.setIncludedFilterValues()
	}
	return
}

func (s *Scope) SetValueFromAddressable() {
	s.setValueFromAddresable()
}

// initialize new scope with added primary field to fieldset
func newScope(modelStruct *ModelStruct) *Scope {
	scope := &Scope{
		Struct:   modelStruct,
		Fieldset: make(map[string]*StructField),
	}

	// set all fields
	for _, field := range modelStruct.fields {
		scope.Fieldset[field.jsonAPIName] = field
	}

	return scope
}

/**

FIELDSET

*/

// fields[collection] = field1, field2
func (s *Scope) buildFieldset(fields ...string) (errs []*ErrorObject) {
	var (
		errObj *ErrorObject
	)

	if len(fields) > s.Struct.getWorkingFieldCount() {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Too many fields to set.")
		errs = append(errs, errObj)
		return
	}

	s.Fieldset = make(map[string]*StructField)

	for _, field := range fields {

		sField, err := s.Struct.checkField(field)
		if err != nil {
			if field == "id" {
				err = ErrInvalidQueryParameter.Copy()
				err.Detail = "Invalid fields parameter. 'id' is not a field - it is primary key."
			}
			errs = append(errs, err)
			continue
		}
		_, ok := s.Fieldset[sField.jsonAPIName]
		if ok {
			// duplicate
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Duplicated fieldset parameter: '%s' for: '%s' collection.", field, s.Struct.collectionType)
			errs = append(errs, errObj)
			if len(errs) > maxPermissibleDuplicates {
				return
			}
			continue
		}
		s.Fieldset[sField.jsonAPIName] = sField
	}

	return

}

/**

FILTERS

*/

func (s *Scope) setIncludedFilterValues() {
	var (
		values []interface{}
		i      int
	)
	if len(s.IncludeValues) == 0 {
		return
	}

	values = make([]interface{}, len(s.IncludeValues))
	for v := range s.IncludeValues {
		values[i] = v
		i++
	}
	s.setPrimaryFilterValues(values...)

	return
}

func (s *Scope) setPrimaryFilterValues(values ...interface{}) {
	filter := s.getOrCreatePrimaryFilter()

	if filter.Values == nil {
		filter.Values = make([]*FilterValues, 0)
	}

	fv := &FilterValues{}
	fv.Values = append(fv.Values, values...)
	fv.Operator = OpIn
	filter.Values = append(filter.Values, fv)

	return
}

func (s *Scope) setPrimaryFilterfield(value string) (errs []*ErrorObject) {
	_, errs = s.buildFilterfield(s.Struct.collectionType, []string{value}, s.Struct, annotationID, annotationEqual)
	return
}

// splitted contains filter fields after filter[collection]
// i.e. /blogs?filter[blogs][posts][id][ne]=10
// splitted should be then [posts, id, ne]
func (s *Scope) buildFilterfield(
	collection string,
	values []string,
	m *ModelStruct,
	splitted ...string,
) (fField *FilterField, errs []*ErrorObject) {
	var (
		sField    *StructField
		op        FilterOperator
		ok        bool
		fieldName string

		errObj     *ErrorObject
		errObjects []*ErrorObject
		// private function for returning ErrObject
		invalidName = func(fieldName, collection string) {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Provided invalid filter field name: '%s' for the '%s' collection.", fieldName, collection)
			errs = append(errs, errObj)
			return
		}
		invalidOperator = func(operator string) {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Provided invalid filter operator: '%s' for the '%s' field.", operator, fieldName)
			errs = append(errs, errObj)
			return
		}
	)

	// check if any parameters are set for filtering
	if len(splitted) == 0 {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprint("Too few filter parameters. Valid format is: filter[collection][field][subfield|operator]([operator])*.")
		errs = append(errs, errObj)
		return
	}

	// for all cases first value should be a fieldName
	fieldName = splitted[0]

	switch len(splitted) {
	case 1:
		// if there is only one argument it must be an attribute.
		if fieldName == "id" {
			fField = s.getOrCreatePrimaryFilter()
		} else {
			sField, ok = m.attributes[fieldName]
			if !ok {
				_, ok = m.relationships[fieldName]
				if ok {
					errObj = ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprintf("Provided filter field: '%s' is a relationship. In order to filter a relationship specify the relationship's field. i.e. '/%s?filter[%s][id]=123'", fieldName, m.collectionType, fieldName)
					errs = append(errs, errObj)
					return
				}
				invalidName(fieldName, m.collectionType)
				return
			}
			fField = s.getOrCreateAttributeFilter(sField)
		}
		errObjects = fField.setValues(m.collectionType, values, OpEqual)
		errs = append(errs, errObjects...)

	case 2:
		if fieldName == "id" {
			fField = s.getOrCreatePrimaryFilter()
		} else {
			sField, ok = m.attributes[fieldName]
			if !ok {
				// jeżeli relacja ->
				sField, ok = m.relationships[fieldName]
				if !ok {
					invalidName(fieldName, m.collectionType)
					return
				}

				// if field were already used
				fField = s.getOrCreateRelationshipFilter(sField)

				// relFilter is a FilterField for specific field in relationship
				var relFilter *FilterField

				relFilter, errObjects = s.buildFilterfield(
					fieldName,
					values,
					sField.relatedStruct,
					splitted[1:]...,
				)
				errs = append(errs, errObjects...)
				fField.addSubfieldFilter(relFilter)
				return
			}
			fField = s.getOrCreateAttributeFilter(sField)
		}
		// it is an attribute filter
		op, ok = operatorsValue[splitted[1]]
		if !ok {
			invalidOperator(splitted[1])
			return
		}
		errObjects = fField.setValues(m.collectionType, values, op)
		errs = append(errs, errObjects...)
	case 3:
		// musi być relacja
		sField, ok = m.relationships[fieldName]
		if !ok {
			// moze ktos podal attribute
			_, ok = m.attributes[fieldName]
			if !ok {
				invalidName(fieldName, m.collectionType)
				return
			}
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Too many parameters for the attribute field: '%s'.", fieldName)
			errs = append(errs, errObj)
			return
		}
		fField = s.getOrCreateRelationshipFilter(sField)

		var relFilter *FilterField

		// get relationship's filter for specific field (filterfield)
		relFilter, errObjects = s.buildFilterfield(
			fieldName,
			values,
			sField.relatedStruct,
			splitted[1:]...,
		)
		errs = append(errs, errObjects...)
		fField.addSubfieldFilter(relFilter)

	default:
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.
			Sprintf("Too many filter parameters for '%s' collection. ", collection)
		errs = append(errs, errObj)
		_, ok = m.attributes[fieldName]
		if !ok {
			_, ok = m.relationships[fieldName]
			if !ok {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.
					Sprintf("Invalid field name: '%s' for '%s' collection.", fieldName, collection)
				errs = append(errs, errObj)
			}
		}
	}
	return
}

func (s *Scope) getOrCreatePrimaryFilter() (filter *FilterField) {
	if s.PrimaryFilters == nil {
		s.PrimaryFilters = []*FilterField{}
	}

	primField := s.Struct.primary

	for _, pf := range s.PrimaryFilters {
		if pf.getFieldIndex() == primField.getFieldIndex() {
			filter = pf
			return
		}
	}

	// if not found within primary filters
	filter = &FilterField{StructField: primField}
	s.PrimaryFilters = append(s.PrimaryFilters, filter)

	return filter
}

func (s *Scope) getOrCreateAttributeFilter(
	sField *StructField,
) (filter *FilterField) {
	if s.AttributeFilters == nil {
		s.AttributeFilters = []*FilterField{}
	}

	for _, attrFilter := range s.AttributeFilters {
		if attrFilter.getFieldIndex() == sField.getFieldIndex() {
			filter = attrFilter
			return
		}
	}

	filter = &FilterField{StructField: sField}
	s.AttributeFilters = append(s.AttributeFilters, filter)
	return filter
}

func (s *Scope) getOrCreateRelationshipFilter(sField *StructField) (filter *FilterField) {
	if s.RelationshipFilters == nil {
		s.RelationshipFilters = []*FilterField{}
	}

	for _, relFilter := range s.RelationshipFilters {
		if relFilter.getFieldIndex() == sField.getFieldIndex() {
			filter = relFilter
			return
		}
	}

	filter = &FilterField{StructField: sField}
	s.RelationshipFilters = append(s.RelationshipFilters, filter)
	return filter
}

/**

INCLUDES

*/

// buildIncludeList provide fast checks for the includedList
// if given include passes use buildInclude method on it.
func (s *Scope) buildIncludeList(includedList ...string,
) (errs []*ErrorObject) {
	var errorObjects []*ErrorObject
	var errObj *ErrorObject

	// check if the number of included fields is possible
	if len(includedList) > s.Struct.getMaxIncludedCount() {
		errObj = ErrOutOfRangeQueryParameterValue.Copy()
		errObj.Detail = fmt.Sprintf("Too many included parameter values for: '%s' collection.",
			s.Struct.collectionType)
		errs = append(errs, errObj)
		return
	}

	var includedMap map[string]int

	// many includes flag if there is more than one include
	var manyIncludes bool = len(includedList) > 1

	if manyIncludes {
		includedMap = make(map[string]int)
	}
	s.IncludedScopes = make(map[*ModelStruct]*Scope)

	// having multiple included in the query
	for _, included := range includedList {

		// check the nested level of every included
		annotCount := strings.Count(included, annotationNestedSeperator)
		if annotCount > s.maxNestedLevel {
			errs = append(errs, ErrTooManyNestedRelationships(included))
			continue
		}

		// if there are more than one include
		if manyIncludes {

			// assert no duplicates are provided in the include list
			includedCount := includedMap[included]
			includedCount++
			includedMap[included] = includedCount
			if annotCount == 0 && includedCount > 1 {
				if includedCount == 2 {
					errObj = ErrInvalidQueryParameter.Copy()
					errObj.Detail = fmt.Sprintf("Included parameter '%s' used more than once.", included)
					errs = append(errs, errObj)
					continue
				} else if includedCount >= maxPermissibleDuplicates {
					break
				}
			}
		}
		errorObjects = s.buildInclude(included)
		errs = append(errs, errorObjects...)
	}
	return
}

// buildInclude searches for the relationship field within given scope
// if not found, then tries to seperate the 'included' argument
// by the 'annotationNestedSeperator'. If seperated correctly
// it tries to create nested fields.
// adds IncludeScope for given field.
func (s *Scope) buildInclude(included string) (errs []*ErrorObject) {
	// relationScope is the scope for the included field
	var (
		includeField  *IncludeField
		isNew         bool
		relationScope *Scope
	)

	// search for the 'included' in the model's
	relationField, ok := s.Struct.relationships[included]
	if !ok {
		// no relationship found check nesteds
		index := strings.Index(included, annotationNestedSeperator)
		if index == -1 {
			errs = append(errs, errNoRelationship(s.Struct.collectionType, included))
			return
		}

		// root part of included (root.subfield)
		field := included[:index]
		relationField, ok := s.Struct.relationships[field]
		if !ok {
			errs = append(errs, errNoRelationship(s.Struct.collectionType, field))
			return
		}

		// get or create new scope for the included field's collection
		relationScope = s.getOrCreateIncludedScope(relationField.relatedStruct)

		// create new included field
		includeField, isNew = s.getOrCreateIncludeField(relationField, relationScope)

		// set nested fields for includeField
		errs = append(errs, includeField.buildNestedInclude(included[index+1:], s)...)
		if errs != nil {
			return
		}
	} else {
		// get or create new scope for the included field's collection
		relationScope = s.getOrCreateIncludedScope(relationField.relatedStruct)

		// create new includedField
		includeField, isNew = s.getOrCreateIncludeField(relationField, relationScope)
	}

	if isNew {
		s.IncludedFields = append(s.IncludedFields, includeField)
	}

	return
}

func (s *Scope) getIncludedScope(mStruct *ModelStruct) *Scope {
	if s.Struct == mStruct {
		return s
	}
	return s.IncludedScopes[mStruct]
}

// createOrGetIncludedScope checks if includeScope exists within given scope.
// if exists returns it, otherwise create new and returns it.
func (s *Scope) getOrCreateIncludedScope(mStruct *ModelStruct) *Scope {
	if s.Struct == mStruct {
		return s
	}

	if s.IncludedScopes == nil {
		s.IncludedScopes = make(map[*ModelStruct]*Scope)
		return s.createIncludedScope(mStruct)
	}

	scope, ok := s.IncludedScopes[mStruct]
	if !ok {
		scope = s.createIncludedScope(mStruct)
	}
	return scope
}

func (s *Scope) createIncludedScope(mStruct *ModelStruct) *Scope {
	scope := newScope(mStruct)
	scope.IsMany = true
	s.IncludedScopes[mStruct] = scope
	// included scope is always treated as many

	return scope
}

// createOrGetIncludeField checks if given include field exists within given scope.
// if not found create new include field.
// returns the include field and the flag 'isNew' if the field were newly created.
func (s *Scope) getOrCreateIncludeField(field *StructField, scope *Scope,
) (includeField *IncludeField, isNew bool) {
	for _, included := range s.IncludedFields {
		if included.getFieldIndex() == field.getFieldIndex() {
			return included, false
		}
	}
	return newIncludeField(field, scope), true
}

func (s *Scope) setIncludedPrimary(value interface{}) (err error) {

	v := reflect.ValueOf(value).Elem()

	for _, field := range s.IncludedFields {
		// included field is a relationship
		fieldValue := v.Field(field.getFieldIndex())
		// if fieldValue.IsNil() {
		// 	continue
		// }

		includedScope, ok := s.IncludedScopes[field.relatedStruct]
		if !ok {
			err = fmt.Errorf("Given includeField: '%s' not found within included scopes. ", field.fieldName)
			return
		}

		primIndex := field.relatedStruct.primary.getFieldIndex()

		if includedScope.IncludeValues == nil {
			includedScope.IncludeValues = make(map[interface{}]struct{})
		}
		switch fieldValue.Kind() {
		case reflect.Slice:
			for i := 0; i < fieldValue.Len(); i++ {
				// set primary field within scope for given model struct
				elem := fieldValue.Index(i)
				if !elem.IsNil() {
					primValue := elem.Elem().Field(primIndex)

					if primValue.IsValid() {
						includedScope.IncludeValues[primValue.Interface()] = struct{}{}
					}
				}
			}
		case reflect.Ptr:
			if !fieldValue.IsNil() {
				primValue := fieldValue.Elem().Field(primIndex)
				if primValue.IsValid() {
					includedScope.IncludeValues[primValue.Interface()] = struct{}{}
				}
			} else {
				// error
			}
		default:
			err = IErrUnexpectedType
			return
		}

	}
	return
}

/**

PAGINATION

*/
func (s *Scope) preparePaginatedValue(key, value string, index int) *ErrorObject {
	val, err := strconv.Atoi(value)
	if err != nil {
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided query parameter: %v, contains invalid value: %v. Positive integer value is required.", key, value)
		return errObj
	}

	if s.Pagination == nil {
		s.Pagination = &Pagination{}
	}
	switch index {
	case 0:
		s.Pagination.Limit = val
		s.Pagination.Type = OffsetPaginate
	case 1:
		s.Pagination.Offset = val
		s.Pagination.Type = OffsetPaginate
	case 2:
		s.Pagination.PageNumber = val
		s.Pagination.Type = PagePaginate
	case 3:
		s.Pagination.PageSize = val
		s.Pagination.Type = PagePaginate
	}
	return nil
}

/**

SORTS

*/
// setSortFields sets the sort fields for given string array.
func (s *Scope) buildSortFields(sorts ...string) (errs []*ErrorObject) {
	var (
		err      *ErrorObject
		order    Order
		fields   map[string]int = make(map[string]int)
		badField                = func(fieldName string) {
			err = ErrInvalidQueryParameter.Copy()
			err.Detail = fmt.Sprintf("Provided sort parameter: '%v' is not valid for '%v' collection.", fieldName, s.Struct.collectionType)
			errs = append(errs, err)
		}
		invalidField bool
	)

	// If the number of sort fields is too long then do not allow
	if len(sorts) > s.Struct.getSortScopeCount() {
		err = ErrOutOfRangeQueryParameterValue.Copy()
		err.Detail = fmt.Sprintf("There are too many sort parameters for the '%v' collection.", s.Struct.collectionType)
		errs = append(errs, err)
		return
	}

	for _, sort := range sorts {
		if sort[0] == '-' {
			order = DescendingOrder
			sort = sort[1:]

		} else {
			order = AscendingOrder
		}

		// check if no dups provided
		count := fields[sort]
		count++

		fields[sort] = count
		if count > 1 {
			if count == 2 {
				err = ErrInvalidQueryParameter.Copy()
				err.Detail = fmt.Sprintf("Sort parameter: %v used more than once.", sort)
				errs = append(errs, err)
				continue
			} else if count > 2 {
				break
			}
		}

		invalidField = newSortField(sort, order, s)
		if invalidField {
			badField(sort)
			continue
		}

	}
	return
}

// VALUES

func (s *Scope) newValueSingle() {
	s.Value = reflect.New(s.Struct.modelType).Interface()
	s.valueAddress = s.Value
}

func (s *Scope) newValueMany() {
	val := reflect.New(reflect.SliceOf(reflect.New(s.Struct.modelType).Type()))

	s.Value = val.Elem().Interface()
	s.valueAddress = val.Interface()
}

func (s *Scope) setValueFromAddresable() {
	s.Value = reflect.ValueOf(s.valueAddress).Elem().Interface()
}

func getID(req *http.Request, mStruct *ModelStruct) (id string, err error) {
	path := req.URL.Path
	pathSplitted := strings.Split(path, "/")
	var idIndex int = -1
	if mStruct.collectionURLIndex != -1 {
		idIndex = mStruct.collectionURLIndex + 1
	} else {
		for i, spl := range pathSplitted {
			if spl == mStruct.collectionType {
				idIndex = i + 1
				break
			}
		}
		if idIndex == -1 {
			err = fmt.Errorf("The url for given request does not contain collection name: %s", mStruct.collectionType)
			return
		}
	}

	if idIndex > len(pathSplitted)-1 {
		err = errors.New("Given request does not use id for the model.")
		return
	}
	id = pathSplitted[idIndex]
	return
}
