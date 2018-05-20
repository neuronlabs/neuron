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

type scopeKind int

const (
	rootKind scopeKind = iota
	includedKind
	relationshipKind
	relatedKind
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

	// IncludeIDValues contain unique values for given include fields
	// the key is the - primary key value
	IncludeIDValues *HashSet

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
	totalIncludeCount int
	kind              scopeKind

	// CollectionScope is a pointer to the scope containing
	collectionScope *Scope
	rootScope       *Scope

	currentIncludedFieldIndex int
	isRelationship            bool
}

// Returns the collection name for given scope
func (s *Scope) GetCollection() string {
	return s.Struct.collectionType
}

// GetCollectionScope gets the collection root scope for given scope.
// Used for included Field scopes for getting their model root scope, that contains all
func (s *Scope) GetCollectionScope() *Scope {
	return s.collectionScope
}

// GetValueAddress gets the address of the value for given scope
// in order to set it use the SetValueFromAddressable
func (s *Scope) GetValueAddress() interface{} {
	return s.valueAddress
}

// GetTotalIncludeFieldCount gets the count for all included Fields. May be used
// as a wait group counter.
func (s *Scope) GetTotalIncludeFieldCount() int {
	return s.totalIncludeCount
}

// GetPresetRelationshipScope - for given root Scope 's' gets the value of the relationship
// used for given request and set it's value into relationshipScope.
// returns an error if the value is not set or there is no relationship includedField
// for given scope.
func (s *Scope) GetPresetRelationshipScope() (relScope *Scope, err error) {
	if len(s.IncludedFields) != 1 {
		return nil, errors.New("Provided invalid IncludedFields for given scope.")
	}
	if err = s.setIncludedFieldValue(s.IncludedFields[0]); err != nil {
		return
	}
	relScope = s.IncludedFields[0].Scope
	return
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

// NextIncludedField allows iteration over the IncludedFields.
// If there is any included field it changes the current field index to the next available.
func (s *Scope) NextIncludedField() bool {
	if s.currentIncludedFieldIndex >= len(s.IncludedFields)-1 {
		return false
	}

	s.currentIncludedFieldIndex++
	return true
}

// CurrentIncludedField gets current included field, based on the index
func (s *Scope) CurrentIncludedField() (*IncludeField, error) {
	if s.currentIncludedFieldIndex == -1 || s.currentIncludedFieldIndex > len(s.IncludedFields)-1 {
		return nil, errors.New("Getting non-existing included field.")
	}

	return s.IncludedFields[s.currentIncludedFieldIndex], nil
}

// ResetIncludedField resets the current included field pointer
func (s *Scope) ResetIncludedField() {
	s.currentIncludedFieldIndex = -1
}

// SetIDFilters sets the ID Filter for given values.
func (s *Scope) SetIDFilters(idValues ...interface{}) {
	s.setIDFilterValues(idValues...)
}

// SetPrimaryFilters sets the primary filter for given values.
func (s *Scope) SetPrimaryFilters(values ...interface{}) {
	s.setIDFilterValues(values...)
}

// Sets the LanguageFilter for given scope.
// If the scope's model does not support i18n it does nothing.
func (s *Scope) SetLanguageFilter(languages ...interface{}) {
	s.setLanguageFilterValues(languages...)
}

// SetValueFromAddressable - lack of generic makes it hard for preparing addressable value.
// While getting the addressable value with GetValueAddress, this function makes use of it
// by setting the Value from addressable.
func (s *Scope) SetValueFromAddressable() {
	s.setValueFromAddresable()
}

// initialize new scope with added primary field to fieldset
func newScope(modelStruct *ModelStruct) *Scope {
	scope := &Scope{
		Struct:                    modelStruct,
		Fieldset:                  make(map[string]*StructField),
		currentIncludedFieldIndex: -1,
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

func (s *Scope) setIDFilterValues(values ...interface{}) {
	s.setPrimaryFilterValues(s.Struct.primary, values...)
	return
}

func (s *Scope) setLanguageFilterValues(values ...interface{}) {
	if s.Struct.language == nil {
		return
	}
	s.setPrimaryFilterValues(s.Struct.language, values...)
	return
}

func (s *Scope) setPrimaryFilterValues(primField *StructField, values ...interface{}) {
	filter := s.getOrCreatePrimaryFilter(primField)

	if filter.Values == nil {
		filter.Values = make([]*FilterValues, 0)
	}

	fv := &FilterValues{}
	fv.Values = append(fv.Values, values...)
	fv.Operator = OpIn
	filter.Values = append(filter.Values, fv)
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
			fField = s.getOrCreateIDFilter()
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
			fField = s.getOrCreateIDFilter()
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

func (s *Scope) getOrCreatePrimaryFilter(primField *StructField) (filter *FilterField) {
	if s.PrimaryFilters == nil {
		s.PrimaryFilters = []*FilterField{}
	}

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

func (s *Scope) getOrCreateIDFilter() (filter *FilterField) {
	return s.getOrCreatePrimaryFilter(s.Struct.primary)
}

func (s *Scope) getOrCreateLangaugeFilter() (filter *FilterField) {
	if s.Struct.language == nil {
		return nil
	}
	return s.getOrCreatePrimaryFilter(s.Struct.language)
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

	// IncludedScopes for root are always set
	s.IncludedScopes = make(map[*ModelStruct]*Scope)

	var includedMap map[string]int

	// many includes flag if there is more than one include
	var manyIncludes bool = len(includedList) > 1

	if manyIncludes {
		includedMap = make(map[string]int)
	}

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
	var includedField *IncludeField
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

		// create new included field
		includedField = s.getOrCreateIncludeField(relationField)

		errs = includedField.Scope.buildInclude(included[index+1:])
		if errs != nil {
			return
		}

	} else {
		// create new includedField if the field was not already created during nested process.
		includedField = s.getOrCreateIncludeField(relationField)
	}

	includedField.Scope.kind = includedKind
	return
}

func (s *Scope) copyIncludedFilters() {
	for _, includedField := range s.IncludedFields {
		includedField.copyFilters()
	}
}

// createModelsRootScope creates scope for given model (mStruct) and
// stores it within the rootScope.IncludedScopes.
// Used for collection unique root scopes
// (filters, fieldsets etc. for given collection scope)
func (s *Scope) createModelsRootScope(mStruct *ModelStruct) *Scope {
	scope := s.createModelsScope(mStruct)
	scope.rootScope.IncludedScopes[mStruct] = scope
	scope.IncludeIDValues = NewHashSet()
	return scope
}

// getModelsRootScope returns the scope for given model that is stored within
// the rootScope
func (s *Scope) getModelsRootScope(mStruct *ModelStruct) (collRootScope *Scope) {
	if s.rootScope == nil {
		// if 's' is root scope and is related to model that is looking for
		if s.Struct == mStruct {
			return s
		}

		return s.IncludedScopes[mStruct]
	}

	return s.rootScope.IncludedScopes[mStruct]
}

// getOrCreateModelsRootScope gets ModelsRootScope and if it is null it creates new.
func (s *Scope) getOrCreateModelsRootScope(mStruct *ModelStruct) *Scope {
	scope := s.getModelsRootScope(mStruct)
	if scope == nil {
		scope = s.createModelsRootScope(mStruct)
	}
	return scope
}

// createsModelsScope
func (s *Scope) createModelsScope(mStruct *ModelStruct) *Scope {
	scope := newScope(mStruct)
	if s.rootScope == nil {
		scope.rootScope = s
	} else {
		scope.rootScope = s.rootScope
	}
	return scope
}

// createOrGetIncludeField checks if given include field exists within given scope.
// if not found create new include field.
// returns the include field
func (s *Scope) getOrCreateIncludeField(
	field *StructField,
) (includeField *IncludeField) {
	for _, included := range s.IncludedFields {
		if included.getFieldIndex() == field.getFieldIndex() {
			return included
		}
	}

	return s.createIncludedField(field)
}

func (s *Scope) createIncludedField(
	field *StructField,
) (includeField *IncludeField) {
	includeField = newIncludeField(field, s)
	if s.IncludedFields == nil {
		s.IncludedFields = make([]*IncludeField, 0)
	}

	s.IncludedFields = append(s.IncludedFields, includeField)
	return
}

// setIncludedFieldValue - used while getting the Relationship Scope,
// and the 's' has the 'includedField' value in it's value.
func (s *Scope) setIncludedFieldValue(includeField *IncludeField) error {
	v := reflect.ValueOf(s.Value)
	switch v.Kind() {
	case reflect.Ptr:
		includeField.setRelatedValue(v.Elem())
	default:
		return fmt.Errorf("Scope has invalid value type: %s", v.Type())
	}
	return nil
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

func getURLVariables(req *http.Request, mStruct *ModelStruct, indexFirst, indexSecond int,
) (valueFirst, valueSecond string, err error) {

	path := req.URL.Path
	var invalidURL = func() error {
		return fmt.Errorf("Provided url is invalid for getting url variables: '%s' with indexes: '%d'/ '%d'", path, indexFirst, indexSecond)
	}
	pathSplitted := strings.Split(path, "/")
	if indexFirst > len(pathSplitted)-1 {
		err = invalidURL()
		return
	}
	var collectionIndex int = -1
	if mStruct.collectionURLIndex != -1 {
		collectionIndex = mStruct.collectionURLIndex
	} else {
		for i, splitted := range pathSplitted {
			if splitted == mStruct.collectionType {
				collectionIndex = i
				break
			}
		}
		if collectionIndex == -1 {
			err = fmt.Errorf("The url for given request does not contain collection name: %s", mStruct.collectionType)
			return
		}
	}

	if collectionIndex+indexFirst > len(pathSplitted)-1 {
		err = invalidURL()
		return
	}
	valueFirst = pathSplitted[collectionIndex+indexFirst]

	if indexSecond > 0 {
		if collectionIndex+indexSecond > len(pathSplitted)-1 {
			err = invalidURL()
			return
		}
		valueSecond = pathSplitted[collectionIndex+indexSecond]
	}
	return
}

func getID(req *http.Request, mStruct *ModelStruct) (id string, err error) {
	id, _, err = getURLVariables(req, mStruct, 1, -1)
	return
}

func getRelationship(req *http.Request, mStruct *ModelStruct) (relationship string, err error) {
	relationship, _, err = getURLVariables(req, mStruct, 2, -1)
	return
}

func getIDAndRelationship(req *http.Request, mStruct *ModelStruct,
) (id, relationship string, err error) {
	return getURLVariables(req, mStruct, 1, 3)

}

func getIDAndRelated(req *http.Request, mStruct *ModelStruct,
) (id, related string, err error) {
	return getURLVariables(req, mStruct, 1, 2)
}
