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

	for key, value := range q {
		switch {
		case key == QueryParamInclude:
			continue
		case key == QueryParamSort:
			errorObjects := scope.checkSortFields(value...)
			overloadPreventer -= len(errorObjects)
			errs = append(errs, errorObjects...)
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
			// filter[collection]
		case strings.HasPrefix(key, QueryParamFields):
			// fields[collection]
			// check if is in attrs or
		case key == QueryParamSort:
			var errorObjects []*ErrorObject
			errorObjects = scope.setSortFields(value...)
			errs = append(errs, errorObjects...)
			overloadPreventer -= len(errorObjects)
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
	Filters [][]interface{}

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
	scope := newSubScope(mStruct)
	scope.collectionScopes = make(map[string]*Scope)
	scope.collectionScopes[mStruct.collectionType] = scope
	scope.isRoot = true
	return scope
}

func newSubScope(modelStruct *ModelStruct) *Scope {
	scope := &Scope{Struct: modelStruct}
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
	var errored bool
	var order Order
	for _, sortField := range sortFields {
		if sortField[0] == '-' {
			order = DescendingOrder
			sortField = sortField[1:]
		} else {
			order = AscendingOrder
		}
		err = s.Struct.checkAttribute(sortField)
		if err != nil {
			errs = append(errs, err)
			errored = true
			continue
		}
		if errored {
			continue
		}
		sort := &SortField{StructField: s.Struct.attributes[sortField], Order: order}
		s.Sorts = append(s.Sorts, sort)
	}
	return
}

func (s *Scope) buildIncludedScopes(includedList ...string,
) (errs []*ErrorObject, err error) {
	var errorObjects []*ErrorObject
	for _, included := range includedList {
		if strings.Count(included, annotationNestedSeperator) > maxNestedRelLevel+1 {
			errs = append(errs, ErrTooManyNestedRelationships(included))
			continue
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
		sub = s.getCollectionScope(seperated)
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
			sub = newSubScope(relatedMStruct)
		}
		var errorObjects []*ErrorObject
		errorObjects, err = sub.buildSubScopes(included[index+1:], collectionScopes)
		errs = append(errs, errorObjects...)
	} else {
		// check for duplicates
		sub = s.getCollectionScope(included)
		if sub == nil {
			// if sField was found
			relatedMStruct := cacheModelMap.Get(sField.refStruct.Type)
			if relatedMStruct == nil {
				err = errNoModelMappedForRel(
					sField.refStruct.Type,
					s.Struct.modelType,
					sField.fieldName,
				)
				errs = append(errs, ErrInternalError.Copy())
				return
			}
			sub = newSubScope(relatedMStruct)
		}
	}
	// map collection type to subscope
	collectionScopes[sub.Struct.collectionType] = sub
	s.SubScopes = append(s.SubScopes, sub)
	return
}

func (s *Scope) checkSortFields(fields ...string,
) []*ErrorObject {
	return s.Struct.checkAttributes(fields...)
}

func (s *Scope) checkFilterFields(fields ...string,
) []*ErrorObject {
	return s.Struct.checkAttributes(fields...)
}

func (s *Scope) checkFields(fields ...string) []*ErrorObject {
	return s.Struct.checkFields(fields...)
}

func (s *Scope) getCollectionScope(collection string) *Scope {
	for _, sub := range s.SubScopes {
		if sub.Struct.collectionType == collection {
			return sub
		}
	}
	return nil
}

func (s *Scope) setSubScopes() {
	for _, subscope := range s.SubScopes {
		subscope.Value = reflect.New(subscope.Struct.modelType)
		// val := reflect.ValueOf(sub.Value)

	}
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
	StructField *StructField
	Order       Order
}

// get collection name from query parameters containing '[...]' i.e. fields[collection]
func retrieveCollectionName(queryParam string) (string, *ErrorObject) {
	opened := strings.Index(queryParam, annotationOpenedBracket)

	if opened == -1 {
		// error leci
		err := ErrUnsupportedQueryParameter.Copy()
		err.Detail = fmt.Sprintf("The %v query parameter is not supported.", queryParam)
		return "", err
	}

	closed := strings.Index(queryParam, annotationClosedBracket)
	if closed == -1 || closed != len(queryParam)-1 {
		// znów error
		err := ErrUnsupportedQueryParameter.Copy()
		err.Detail = fmt.Sprintf("The %v query parameter is not supported.", queryParam)
		return "", err
	}

	collection := queryParam[opened+1 : closed]
	return collection, nil
}
