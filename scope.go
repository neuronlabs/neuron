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
	maxPermissibleDuplicates = 3
)

type Scope struct {
	// Struct is a modelStruct this scope is based on
	Struct *ModelStruct

	// RelatedField is a structField to which Subscope this one belongs to
	RelatedField *StructField

	// Root defines the root of the provided scope
	Root *Scope

	// Included subscopes
	SubScopes []*Scope

	// Value is a value for given subscope
	Value interface{}

	// Filters contain fields by which the main value is being filtered.
	Filters map[int]*FilterScope

	// Fields represents fields used for this subscope - jsonapi 'fields[collection]'
	Fields []*StructField

	// SortScopes
	Sorts []*SortScope

	// PaginationScope
	PaginationScope *PaginationScope

	collectionScopes map[string]*Scope

	currentErrorCount int
}

func newRootScope(mStruct *ModelStruct, multiple bool) *Scope {
	scope := newSubScope(mStruct, nil, multiple)
	scope.collectionScopes = make(map[string]*Scope)
	scope.collectionScopes[mStruct.collectionType] = scope
	return scope
}

func newSubScope(modelStruct *ModelStruct, relatedField *StructField, multiple bool) *Scope {
	scope := &Scope{
		Struct:       modelStruct,
		RelatedField: relatedField,
		Fields:       []*StructField{modelStruct.primary},
	}
	var (
		makeSlice = func(tp reflect.Type) {
			scope.Value = reflect.New(reflect.SliceOf(reflect.TypeOf(reflect.New(tp).Interface()))).Interface()
		}
		makePtr = func(tp reflect.Type) {
			scope.Value = reflect.New(tp).Interface()
		}
	)

	var tp reflect.Type

	if relatedField == nil {
		tp = modelStruct.modelType
	} else {
		tp = relatedField.relatedModelType
	}
	if multiple {
		makeSlice(tp)
	} else {
		makePtr(tp)
	}

	// fmt.Printf("Scope val type: %v\n", reflect.TypeOf(scope.Value))

	return scope
}

// func (s *Scope) GetFields() []*StructField {
// 	return s.Fields
// }

// func (s *Scope) GetSortScopes() (fields []*SortScope) {
// 	return s.Sorts
// }

func (s *Scope) GetFilterScopes() (filters []*FilterScope) {
	for _, fScope := range s.Filters {
		filters = append(filters, fScope)
	}
	return
}

func (s *Scope) GetManyValues() ([]interface{}, error) {
	var v interface{}
	v = reflect.ValueOf(s.Value).Elem().Interface()
	return convertToSliceInterface(&v)
}

// func (s *Scope) SetFilteredField(fieldApiName string, values []string, operator FilterOperator,
// ) (errs []*ErrorObject, err error) {
// 	if operator > OpEndsWith {
// 		err = fmt.Errorf("Invalid operator provided: '%s'", operator)
// 		return
// 	}
// 	_, errs, err = s.newFilterScope(s.Struct.collectionType, values, s.Struct, fieldApiName, operator.String())
// 	return
// }

func (s *Scope) setPrimaryFilterScope(value string) (errs []*ErrorObject) {
	_, errs = s.newFilterScope(s.Struct.collectionType, []string{value}, s.Struct, annotationID, annotationEqual)
	return
}

// SetSortScopes sets the sort fields for given string array.
func (s *Scope) setSortScopes(sorts ...string) (errs []*ErrorObject) {
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

		invalidField = s.newSortScope(sort, order)
		if invalidField {
			badField(sort)
			continue
		}

	}
	return
}

func (s *Scope) newSortScope(sort string, order Order) (invalidField bool) {
	var (
		sField    *StructField
		ok        bool
		sortScope *SortScope
	)

	splitted := strings.Split(sort, annotationNestedSeperator)
	l := len(splitted)
	switch {
	case l == 1:
		if sort == annotationID {
			sField = s.Struct.primary
		} else {
			sField, ok = s.Struct.attributes[sort]
			if !ok {
				invalidField = true
				return
			}
		}
		sortScope = &SortScope{Field: sField, Order: order}
		s.Sorts = append(s.Sorts, sortScope)
	case l <= (maxNestedRelLevel + 1):
		sField, ok = s.Struct.relationships[splitted[0]]

		if !ok {

			invalidField = true
			return
		}
		// if true then the nested should be an attribute for given
		var found bool
		for i := range s.Sorts {
			if s.Sorts[i].Field.getFieldIndex() == sField.getFieldIndex() {
				sortScope = s.Sorts[i]
				found = true
				break
			}
		}
		if !found {
			sortScope = &SortScope{Field: sField}
		}
		invalidField = sortScope.setRelationScope(splitted[1:], order)
		if !found && !invalidField {
			s.Sorts = append(s.Sorts, sortScope)
		}
		return
	default:
		invalidField = true
	}
	return

}

func (s *Scope) buildIncludedScopes(includedList ...string,
) (errs []*ErrorObject) {
	var errorObjects []*ErrorObject
	var errObj *ErrorObject

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

	for _, included := range includedList {
		annotCount := strings.Count(included, annotationNestedSeperator)
		if annotCount > maxNestedRelLevel {
			errs = append(errs, ErrTooManyNestedRelationships(included))
			continue
		}

		if manyIncludes {
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

		errorObjects = s.buildSubScopes(included, s.collectionScopes)
		errs = append(errs, errorObjects...)
	}
	return
}

// build sub scopes for provided included argument.
func (s *Scope) buildSubScopes(included string, collectionScopes map[string]*Scope,
) (errs []*ErrorObject) {

	var sub *Scope

	// Check if the included is in relationships
	sField, ok := s.Struct.relationships[included]
	if !ok {
		index := strings.Index(included, annotationNestedSeperator)
		if index == -1 {
			errs = append(errs, errNoRelationship(s.Struct.collectionType, included))
			return
		}

		seperated := included[:index]
		sField, ok := s.Struct.relationships[seperated]
		if !ok {
			errs = append(errs, errNoRelationship(s.Struct.collectionType, seperated))
			return
		}
		// Check if no other scope for this collection within given 'subscope' exists
		sub = s.getScopeForField(sField)
		if sub == nil {
			var isSlice bool
			if sField.jsonAPIType == RelationshipMultiple {
				isSlice = true
			}
			sub = newSubScope(sField.relatedStruct, sField, isSlice)

		}
		errs = sub.buildSubScopes(included[index+1:], collectionScopes)
	} else {
		// check for duplicates
		sub = s.getScopeForField(sField)
		if sub == nil {
			var isSlice bool
			if sField.jsonAPIType == RelationshipMultiple {
				isSlice = true
			}
			sub = newSubScope(sField.relatedStruct, sField, isSlice)
		}
	}
	// map collection type to subscope
	collectionScopes[sub.Struct.collectionType] = sub
	sub.Fields = sub.Struct.fields
	s.SubScopes = append(s.SubScopes, sub)
	sub.Root = s
	return
}

// splitted contains filter fields after filter[collection]
// i.e. /blogs?filter[blogs][posts][id][ne]=10
// splitted should be then [posts, id, ne]
func (s *Scope) newFilterScope(
	collection string,
	values []string,
	m *ModelStruct,
	splitted ...string,
) (fField *FilterScope, errs []*ErrorObject) {
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

		// creates new filter field if is nil and adds it to the scope
		setFilterScope = func() {
			// get the filter field from the scope if exists.

			var sameStruct bool = m == s.Struct
			if sameStruct {
				fField = s.Filters[sField.getFieldIndex()]
			}

			if fField == nil {
				fField = new(FilterScope)
				fField.Field = sField
				if sameStruct {
					s.Filters[sField.getFieldIndex()] = fField
				}
			}

		}
	)

	if s.Filters == nil {
		s.Filters = make(map[int]*FilterScope)
	}

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
			sField = m.primary
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
		}
		setFilterScope()
		op = OpEqual
		errObjects = fField.setValues(m.collectionType, values, op)
		errs = append(errs, errObjects...)

	case 2:
		if fieldName == "id" {
			sField = m.primary
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
				setFilterScope()

				// relFilter is a FilterScope for specific field in relationship
				var relFilter *FilterScope

				relFilter, errObjects = s.newFilterScope(
					fieldName,
					values,
					sField.relatedStruct,
					splitted[1:]...,
				)
				errs = append(errs, errObjects...)
				fField.appendRelFilter(relFilter)
				return
			}
		}
		// jeżeli attribute ->
		op, ok = operatorsValue[splitted[1]]
		if !ok {
			invalidOperator(splitted[1])
			return
		}
		setFilterScope()
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
		setFilterScope()

		var relFilter *FilterScope

		// get relationship's filter for specific field (filterfield)
		relFilter, errObjects = s.newFilterScope(
			fieldName,
			values,
			sField.relatedStruct,
			splitted[1:]...,
		)
		errs = append(errs, errObjects...)
		fField.appendRelFilter(relFilter)

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
		return
	}

	return
}

func (s *Scope) getScopeForField(sField *StructField) *Scope {
	for _, sub := range s.SubScopes {
		if sub.RelatedField == sField {
			return sub
		}
	}
	return nil
}

// fields[collection] = field1, field2
func (s *Scope) setWorkingFields(fields ...string) (errs []*ErrorObject) {
	var (
		errObj *ErrorObject
	)
	fmt.Println(fields)
	if len(fields) > s.Struct.getWorkingFieldCount() {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Too many fields to set.")
		errs = append(errs, errObj)
		return
	}

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
		s.Fields = append(s.Fields, sField)
	}

	return

}

func (s *Scope) preparePaginatedValue(key, value string, index int) *ErrorObject {
	val, err := strconv.Atoi(value)
	if err != nil {
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided query parameter: %v, contains invalid value: %v. Positive integer value is required.", key, value)
		return errObj
	}

	if s.PaginationScope == nil {
		s.PaginationScope = &PaginationScope{}
	}
	switch index {
	case 0:
		s.PaginationScope.Limit = val
		s.PaginationScope.Type = OffsetPaginate
	case 1:
		s.PaginationScope.Offset = val
		s.PaginationScope.Type = OffsetPaginate
	case 2:
		s.PaginationScope.PageNumber = val
		s.PaginationScope.Type = PagePaginate
	case 3:
		s.PaginationScope.PageSize = val
		s.PaginationScope.Type = PagePaginate
	}
	return nil
}

// PaginationType is the enum that describes the type of pagination
type PaginationType int

const (
	OffsetPaginate PaginationType = iota
	PagePaginate
	CursorPaginate
)

type PaginationScope struct {
	Limit      int
	Offset     int
	PageNumber int
	PageSize   int

	UseTotal bool
	Total    int

	// Describes which pagination type to use.
	Type PaginationType
}

func (p *PaginationScope) GetLimitOffset() (limit, offset int) {
	switch p.Type {
	case OffsetPaginate:
		limit = p.Limit
		offset = p.Offset
	case PagePaginate:
		limit = p.PageSize
		offset = p.PageNumber * p.PageSize
	case CursorPaginate:
		// not implemented yet
	}
	return
}

func (p *PaginationScope) check() error {
	var offsetBased, pageBased bool
	if p.Limit != 0 || p.Offset != 0 {
		offsetBased = true
	}
	if p.PageNumber != 0 || p.PageSize != 0 {
		pageBased = true
	}

	if offsetBased && pageBased {
		err := errors.New("Both offset-based and page-based pagination are set.")
		return err
	}
	return nil
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
