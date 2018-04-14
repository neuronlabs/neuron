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

// URLParams defines url parameters as a map[string]string
type URLParams map[string]string

type paramsKey struct{}

// ParamsKey is the variable for setting URLParams in the request context.
var ParamsKey = paramsKey{}

// BuildScopeMulti - build scope for the list request
func BuildScopeMulti(req *http.Request, model interface{},
) (scope *Scope, errs []*ErrorObject, err error) {
	// Get ModelStruct
	var mStruct *ModelStruct
	mStruct, err = GetModelStruct(model)
	if err != nil {
		return
	}

	// // Check if parameters are in the request context
	// params, ok := req.Context().Value(ParamsKey).(URLParams)
	// if !ok {
	// 	err = ErrNoParamsInContext
	// 	return
	// }

	scope = newRootScope(mStruct)

	// overloadPreventer - is a warden upon invalid query parameters
	var overloadPreventer int = 5

	// Get URLQuery
	q := req.URL.Query()

	// Check first included in order to create subscopes
	included, ok := q[QueryParamInclude]
	if ok {
		// build included scopes
		errs, err = scope.buildIncludedScopes(included...)
		if err != nil {
			return
		}
		overloadPreventer -= len(errs)
	}

	var errObj *ErrorObject
	var errorObjects []*ErrorObject

	for key, value := range q {
		switch {
		case key == QueryParamInclude:
			continue
		case key == QueryParamSort:
			errorObjects = scope.setSortFields(value...)
			errs = append(errs, errorObjects...)
			overloadPreventer -= len(errorObjects)
		case key == QueryParamPageLimit:
			scope.Pagination.Limit, errObj = scope.preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overloadPreventer--
				break
			}
		case key == QueryParamPageOffset:
			scope.Pagination.Offset, errObj = scope.preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overloadPreventer--
			}
		case key == QueryParamPageNumber:
			scope.Pagination.PageNumber, errObj = scope.preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overloadPreventer--
			}
		case key == QueryParamPageSize:
			scope.Pagination.PageSize, errObj = scope.preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overloadPreventer--
			}
		case strings.HasPrefix(key, QueryParamFilter):
			// filter[field]

		case strings.HasPrefix(key, QueryParamFields):
			// fields[collection]
			// check if is in attrs or

		default:
			// Check if it is an attribute to query for
			var errorObjects []*ErrorObject
			errorObjects = scope.checkFields(key)
			overloadPreventer -= len(errorObjects)
			errs = append(errs, errorObjects...)
		}
		if overloadPreventer <= 0 {
			return
		}
	}
	return
}

func BuildScopeSingle(req *http.Request, model interface{},
) (scope *Scope, errs []*ErrorObject, err error) {
	// get model type
	return

}

type Scope struct {
	// Struct is a modelStruct this scope is based on
	Struct *ModelStruct

	// RelatedField is a structField to which Subscope this one belongs to
	RelatedField *StructField

	// Root defines the root of the provided scope
	Root *Scope

	// Value is a value for given subscope
	Value interface{}

	// Filters contain fields by which the main value is being filtered.
	Filters []*FilterField

	// Fields represents fields used for this subscope - jsonapi 'fields[collection]'
	Fields []*StructField

	// Included subscopes
	SubScopes []*Scope

	// SortFields
	Sorts []*SortField

	// Pagination
	Pagination *Pagination

	isRoot           bool
	collectionScopes map[string]*Scope
}

func newRootScope(mStruct *ModelStruct) *Scope {
	scope := newSubScope(mStruct, nil)
	scope.collectionScopes = make(map[string]*Scope)
	scope.collectionScopes[mStruct.collectionType] = scope
	scope.isRoot = true
	return scope
}

func newSubScope(modelStruct *ModelStruct, relatedField *StructField) *Scope {
	scope := &Scope{Struct: modelStruct, RelatedField: relatedField}
	scope.Value = reflect.New(modelStruct.modelType)
	return scope
}

func (s *Scope) GetFields() []*StructField {
	return s.Fields
}

func (s *Scope) GetSortFields() (fields []*SortField) {
	return s.Sorts
}

// SetSortFields sets the sort fields for given string array.
func (s *Scope) setSortFields(sortFields ...string) (errs []*ErrorObject) {
	var err *ErrorObject
	var order Order
	var fields map[string]int = make(map[string]int)

	// If the number of sort fields is too long then do not allow
	if len(sortFields) > s.Struct.getSortFieldCount() {
		err = ErrOutOfRangeQueryParameterValue.Copy()
		err.Detail = fmt.Sprintf("The sort parameter for the '%v' collection is too long. The permissible length is up to the number of collection attributes with has-one relationships.", s.Struct.collectionType)
		errs = append(errs, err)
		return
	}

	for i, sortField := range sortFields {
		if sortField[0] == '-' {
			order = DescendingOrder
			sortField = sortField[1:]

		} else {
			order = AscendingOrder
		}

		// check if no dups provided
		count := fields[sortField]
		count++

		fields[sortField] = count
		if count > 1 {
			if count == 2 {
				err = ErrInvalidQueryParameter.Copy()
				err.Detail = fmt.Sprintf("Sort parameter: %v used more than once.", sortField)
				errs = append(errs, err)
				continue
			} else if count > 2 {
				break
			}
		}

		sField, ok := s.Struct.attributes[sortField]
		if !ok {
			sField, ok = s.Struct.relationships[sortField]
			if !ok {
				err = ErrInvalidQueryParameter.Copy()
				err.Detail = fmt.Sprintf("Provided sort parameter: '%v' is not valid for '%v' collection.", sortFields[i], s.Struct.collectionType)
				errs = append(errs, err)
				continue
			}
		}
		sort := &SortField{StructField: sField, Order: order}
		s.Sorts = append(s.Sorts, sort)
	}
	return
}

func (s *Scope) buildIncludedScopes(includedList ...string,
) (errs []*ErrorObject, err error) {
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

		errorObjects, err = s.buildSubScopes(included, s.collectionScopes)
		errs = append(errs, errorObjects...)
		if err != nil {
			return
		}
	}
	return
}

// build sub scopes for provided included argument.
func (s *Scope) buildSubScopes(included string, collectionScopes map[string]*Scope,
) (errs []*ErrorObject, err error) {

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
			relatedMStruct := cacheModelMap.Get(sField.relatedModelType)
			if relatedMStruct == nil {
				err = errNoModelMappedForRel(
					sField.relatedModelType,
					sField.refStruct.Type,
					sField.fieldName,
				)
				errs = append(errs, ErrInternalError.Copy())
				return
			}
			sub = newSubScope(relatedMStruct, sField)

		}
		errs, err = sub.buildSubScopes(included[index+1:], collectionScopes)
	} else {
		// check for duplicates
		sub = s.getScopeForField(sField)
		if sub == nil {
			// if sField was found
			relatedMStruct := cacheModelMap.Get(sField.relatedModelType)
			if relatedMStruct == nil {
				err = errNoModelMappedForRel(
					sField.refStruct.Type,
					sField.relatedModelType,
					sField.fieldName,
				)
				errs = append(errs, ErrInternalError.Copy())
				return
			}
			sub = newSubScope(relatedMStruct, sField)
		}
	}
	// map collection type to subscope
	collectionScopes[sub.Struct.collectionType] = sub
	s.SubScopes = append(s.SubScopes, sub)
	return
}

func (s *Scope) checkFilterField(field string, values ...string,
) (errorObjects []*ErrorObject, err error) {
	// checks if given fields exists in the attributes
	sField := s.Struct.attributes[field]

	if sField == nil {
		// if field is not in attributes if given field exists in relationships
		sField = s.Struct.relationships[field]

		if sField == nil {
			// if given scope's model does not contain given field
			errObj := ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Invalid filter parameter. The collection '%s' does not contain field: '%s' ", s.Struct.collectionType, field)
			errorObjects = append(errorObjects, errObj)
			return
		}
	}
	filterField := &FilterField{StructField: sField, Values: []interface{}{values}}

	errorObjects = filterField.checkValues()
	if len(errorObjects) != 0 {
		return
	}

	errorObjects = filterField.setValues()
	if len(errorObjects) != 0 {
		return
	}

	s.Filters = append(s.Filters, filterField)
	return
}

func (s *Scope) checkFields(fields ...string) []*ErrorObject {
	return s.Struct.checkFields(fields...)
}

func (s *Scope) getScopeForField(sField *StructField) *Scope {
	for _, sub := range s.SubScopes {
		if sub.RelatedField == sField {
			return sub
		}
	}
	return nil
}

func (s *Scope) preparePaginatedValue(key, value string) (int, *ErrorObject) {
	val, err := strconv.Atoi(value)
	if err != nil {
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided query parameter: %v, contains invalid value: %v. Positive integer value is required.", key, value)
		return val, errObj
	}

	if s.Pagination == nil {
		s.Pagination = &Pagination{}
	}
	return val, nil
}

type Pagination struct {
	Limit      int
	Offset     int
	PageNumber int
	PageSize   int
}

func (p *Pagination) CheckParameters() error {
	return nil
}

type Order int

const (
	AscendingOrder Order = iota
	DescendingOrder
)

type SortField struct {
	*StructField
	Order Order
}

// get collection name from query parameters containing '[...]' i.e. fields[collection]
func retrieveCollectionName(queryParam string) (string, *ErrorObject) {
	opened := strings.Index(queryParam, annotationOpenedBracket)

	if opened == -1 {
		// no opening
		err := ErrUnsupportedQueryParameter.Copy()
		err.Detail = fmt.Sprintf("The %v query parameter is not supported.", queryParam)
		return "", err
	}

	closed := strings.Index(queryParam, annotationClosedBracket)
	if closed == -1 || closed != len(queryParam)-1 {
		// no closing
		err := ErrUnsupportedQueryParameter.Copy()
		err.Detail = fmt.Sprintf("The %v query parameter is not supported.", queryParam)
		return "", err
	}

	collection := queryParam[opened+1 : closed]
	return collection, nil
}
