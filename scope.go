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

	scope = newScope(mStruct)

	// overLoadPreventer - is a warden upon invalid query parameters
	var overLoadPreventer int = 5

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
		overLoadPreventer -= len(errs)
	}

	var errObj *ErrorObject

	for key, value := range q {
		switch {
		case key == QueryParamInclude:
			continue
		case key == QueryParamSort:
			errorObjects := scope.RootScope.checkSortFields(value...)
			overLoadPreventer -= len(errorObjects)
			errs = append(errs, errorObjects...)
		case key == QueryParamPageLimit:
			scope.RootScope.Pagination.Limit, errObj = scope.RootScope.
				preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overLoadPreventer--
				break
			}
		case key == QueryParamPageOffset:
			scope.RootScope.Pagination.Offset, errObj = scope.RootScope.
				preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overLoadPreventer--
			}
		case key == QueryParamPageNumber:
			scope.RootScope.Pagination.PageNumber, errObj = scope.RootScope.
				preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overLoadPreventer--
			}
		case key == QueryParamPageSize:
			scope.RootScope.Pagination.PageSize, errObj = scope.RootScope.
				preparePaginatedValue(key, value[0])
			if errObj != nil {
				errs = append(errs, errObj)
				overLoadPreventer--
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
			overLoadPreventer -= len(errorObjects)
		default:
			// Check if it is an attribute to query for
			var errorObjects []*ErrorObject
			errorObjects = scope.RootScope.checkFields(key)
			overLoadPreventer -= len(errorObjects)
			errs = append(errs, errorObjects...)
		}
		if overLoadPreventer <= 0 {
			return
		}
	}
	return
}

func BuildScopeSingle(req *http.Request, model interface{},
) (scope *SubScope, errs []*ErrorObject, err error) {
	// get model type
	return

}

type Scope struct {
	RootScope *SubScope
	Error     error

	included         []string
	collectionScopes map[string]*SubScope
}

func newScope(mStruct *ModelStruct) *Scope {
	scope := &Scope{
		RootScope:        newSubScope(mStruct),
		collectionScopes: make(map[string]*SubScope),
	}
	scope.collectionScopes[mStruct.collectionType] = scope.RootScope
	return scope
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
		err = s.RootScope.Struct.checkAttribute(sortField)
		if err != nil {
			errs = append(errs, err)
			errored = true
			continue
		}
		if errored {
			continue
		}
		sort := &SortField{StructField: s.RootScope.Struct.attributes[sortField], Order: order}
		s.RootScope.Sorts = append(s.RootScope.Sorts, sort)
	}
	return
}

func (s *Scope) buildIncludedScopes(includedList ...string,
) (errs []*ErrorObject, err error) {
	var errorObjects []*ErrorObject
	for _, included := range includedList {
		errorObjects, err = s.RootScope.buildSubScopes(included, s.collectionScopes)
		errs = append(errs, errorObjects...)
		if err != nil {
			return
		}
	}
	return
}

type SubScope struct {
	Struct *ModelStruct
	Value  interface{}

	// Filters contain fields by which the main value is being filtered.
	Filters interface{}

	// RelatedField is a structField to which Subscope this one belongs to
	RelatedField *StructField
	Fields       []*StructField
	SubScopes    []*SubScope
	Sorts        []*SortField
	Pagination   *Pagination
}

func newSubScope(modelStruct *ModelStruct) *SubScope {
	scope := &SubScope{Struct: modelStruct}
	scope.Value = reflect.New(modelStruct.modelType)
	return scope
}

func (s *SubScope) GetFields() []*StructField {
	return s.Fields
}

func (s *SubScope) GetSortFields() (fields []*SortField) {
	return s.Sorts
}

// build sub scopes for provided included argument.
func (s *SubScope) buildSubScopes(included string, collectionScopes map[string]*SubScope,
) (errs []*ErrorObject, err error) {

	var sub *SubScope

	// Check if the included is in relationships
	sField, ok := s.Struct.relationships[included]
	if !ok {
		// get index of 'annotationNestedSeperator'
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
			// if sField.relationship == nil {
			// 	err = errNoRelationshipInModel(sField.refStruct.Type,
			// 		s.Struct.modelType, included)
			// 	errs = append(errs, ErrInternalError.Copy())
			// 	return
			// }

			relatedMStruct := cacheModelMap.Get(sField.relatedType)
			if relatedMStruct == nil {
				err = errNoModelMappedForRel(
					sField.relatedType,
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

func (s *SubScope) checkSortFields(fields ...string,
) []*ErrorObject {
	return s.Struct.checkAttributes(fields...)
}

func (s *SubScope) checkFilterFields(fields ...string,
) []*ErrorObject {
	return s.Struct.checkAttributes(fields...)
}

func (s *SubScope) checkFields(fields ...string) []*ErrorObject {
	return s.Struct.checkFields(fields...)
}

func (s *SubScope) getCollectionScope(collection string) *SubScope {
	for _, sub := range s.SubScopes {
		if sub.Struct.collectionType == collection {
			return sub
		}
	}
	return nil
}

func (s *SubScope) setSubScopes() {
	for _, subscope := range s.SubScopes {
		subscope.Value = reflect.New(subscope.Struct.modelType)
		// val := reflect.ValueOf(sub.Value)

	}
}

func (s *SubScope) preparePaginatedValue(key, value string) (int, *ErrorObject) {
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
