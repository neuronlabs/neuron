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

	// overloadPreventer - is a warden upon invalid query parameters
	var (
		overloadPreventer int = 5
		errObj            *ErrorObject
		errorObjects      []*ErrorObject
		addErrors         = func(errObjects ...*ErrorObject) {
			errs = append(errs, errObjects...)
			overloadPreventer -= len(errObjects)
		}
	)

	scope = newRootScope(mStruct)

	// Get URLQuery
	q := req.URL.Query()

	// Check first included in order to create subscopes
	included, ok := q[QueryParamInclude]
	if ok {
		// build included scopes
		errorObjects, err = scope.buildIncludedScopes(included...)
		addErrors(errorObjects...)
		if err != nil || len(errs) > 0 {
			return
		}
	}

	for key, value := range q {
		if len(value) > 1 {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' set more than once.", key)
			addErrors(errObj)
			continue
		}
		switch {
		case key == QueryParamInclude:
			continue
		case key == QueryParamPageLimit:
			errObj = scope.preparePaginatedValue(key, value[0], 0)
			if errObj != nil {
				addErrors(errObj)
				break
			}
		case key == QueryParamPageOffset:
			errObj = scope.preparePaginatedValue(key, value[0], 1)
			if errObj != nil {
				addErrors(errObj)
			}
		case key == QueryParamPageNumber:
			errObj = scope.preparePaginatedValue(key, value[0], 2)
			if errObj != nil {
				addErrors(errObj)
			}
		case key == QueryParamPageSize:
			errObj = scope.preparePaginatedValue(key, value[0], 3)
			if errObj != nil {
				addErrors(errObj)
			}
		case strings.HasPrefix(key, QueryParamFilter):
			// filter[field]
			var splitted []string
			// get other operators
			var er error
			splitted, er = splitBracketParameter(key[len(QueryParamFilter):])
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The filter paramater is of invalid form. %s", er)
				addErrors(errObj)
				continue
			}

			collection := splitted[0]

			filterScope := scope.collectionScopes[collection]
			if filterScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The collection: '%s' is invalid or not included in query.", collection)
				addErrors(errObj)
				continue
			}

			splitValues := strings.Split(value[0], annotationSeperator)

			_, errorObjects, err = filterScope.
				newFilterField(splitted[0], splitValues, filterScope.Struct, splitted[1:]...)
			addErrors(errorObjects...)
			if err != nil {
				return
			}

		case key == QueryParamSort:
			splitted := strings.Split(value[0], annotationSeperator)

			errorObjects = scope.setSortFields(splitted...)
			addErrors(errorObjects...)
		case strings.HasPrefix(key, QueryParamFields):
			// fields[collection]
			var splitted []string
			var er error
			splitted, er = splitBracketParameter(key[len(QueryParamFields):])
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter is of invalid form. %s", er)
				addErrors(errObj)
				continue
			}
			if len(splitted) != 1 {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter: '%s' is of invalid form. Nested 'fields' is not supported.", key)
				addErrors(errObj)
				continue
			}
			collection := splitted[0]
			fieldsScope := scope.collectionScopes[collection]
			if fieldsScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The collection: '%s' in fields parameter is invalid or not included to the query.", collection)
				addErrors(errObj)
				continue
			}
			splitValues := strings.Split(value[0], annotationSeperator)
			errorObjects = fieldsScope.setWorkingFields(splitValues...)
			addErrors(errorObjects...)

		default:
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' is unsupported.", key)
			addErrors(errObj)
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
	// Get ModelStruct
	var mStruct *ModelStruct
	mStruct, err = GetModelStruct(model)
	if err != nil {
		return
	}
	var (
		overloadPreventer = 2
		addErrors         = func(errObjects ...*ErrorObject) {
			errs = append(errs, errObjects...)
			overloadPreventer -= len(errObjects)
		}
		errObj       *ErrorObject
		errorObjects []*ErrorObject
		id           string
	)
	id, err = getID(req, mStruct)
	if err != nil {
		addErrors(ErrInternalError.Copy())
		return
	}

	q := req.URL.Query()

	scope = newRootScope(mStruct)
	errorObjects, err = scope.setPrimaryFilterField(id)
	if err != nil || len(errorObjects) != 0 {
		errObj = ErrInternalError.Copy()
		errs = append(errs, errObj)
		return
	}

	// Check first included in order to create subscopes
	included, ok := q[QueryParamInclude]
	if ok {
		// build included scopes
		errorObjects, err = scope.buildIncludedScopes(included...)
		addErrors(errorObjects...)
		if err != nil || len(errs) > 0 {
			return
		}
	}
	for key, values := range q {
		if len(values) > 1 {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' set more than once.", key)
			addErrors(errObj)
			continue
		}

		switch {
		case key == QueryParamInclude:
			continue
		case strings.HasPrefix(key, QueryParamFields):
			// fields[collection]
			var splitted []string
			var er error
			splitted, er = splitBracketParameter(key[len(QueryParamFields):])
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter is of invalid form. %s", er)
				addErrors(errObj)
				continue
			}
			if len(splitted) != 1 {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter: '%s' is of invalid form. Nested 'fields' is not supported.", key)
				addErrors(errObj)
				continue
			}
			collection := splitted[0]
			fieldsScope := scope.collectionScopes[collection]
			if fieldsScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The collection: '%s' in fields parameter is invalid or not included to the query.", collection)
				addErrors(errObj)
				continue
			}
			splitValues := strings.Split(values[0], annotationSeperator)
			errorObjects = fieldsScope.setWorkingFields(splitValues...)
			addErrors(errorObjects...)
		default:
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' is unsupported.", key)
			addErrors(errObj)
		}

		if overloadPreventer <= 0 {
			return
		}
	}

	return
}

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
	Filters map[int]*FilterField

	// Fields represents fields used for this subscope - jsonapi 'fields[collection]'
	Fields []*StructField

	// SortFields
	Sorts []*SortField

	// Pagination
	Pagination *Pagination

	collectionScopes map[string]*Scope
}

func newRootScope(mStruct *ModelStruct) *Scope {
	scope := newSubScope(mStruct, nil)
	scope.collectionScopes = make(map[string]*Scope)
	scope.collectionScopes[mStruct.collectionType] = scope
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

func (s *Scope) SetFilteredField(fieldApiName string, values []string, operator FilterOperator,
) (errs []*ErrorObject, err error) {
	if operator > OpEndsWith {
		err = fmt.Errorf("Invalid operator provided: '%s'", operator)
		return
	}
	_, errs, err = s.newFilterField(s.Struct.collectionType, values, s.Struct, fieldApiName, operator.String())
	return
}

func (s *Scope) setPrimaryFilterField(value string) (errs []*ErrorObject, err error) {
	_, errs, err = s.newFilterField(s.Struct.collectionType, []string{value}, s.Struct, annotationID, annotationEqual)
	return
}

// SetSortFields sets the sort fields for given string array.
func (s *Scope) setSortFields(sortFields ...string) (errs []*ErrorObject) {
	var err *ErrorObject
	var order Order
	var fields map[string]int = make(map[string]int)
	var sField *StructField
	var ok bool

	// If the number of sort fields is too long then do not allow
	if len(sortFields) > s.Struct.getSortFieldCount() {
		err = ErrOutOfRangeQueryParameterValue.Copy()
		err.Detail = fmt.Sprintf("There are too many sort parameters for the '%v' collection.", s.Struct.collectionType)
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

		if sortField == "id" {
			sField = s.Struct.primary
		} else {
			sField, ok = s.Struct.attributes[sortField]
			if !ok {
				sField, ok = s.Struct.relationships[sortField]
				if !ok {
					err = ErrInvalidQueryParameter.Copy()
					err.Detail = fmt.Sprintf("Provided sort parameter: '%v' is not valid for '%v' collection.", sortFields[i], s.Struct.collectionType)
					errs = append(errs, err)
					continue
				}
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
			relatedMStruct := cacheModelMap.Get(sField.getRelatedModelType())
			if relatedMStruct == nil {
				err = errNoModelMappedForRel(
					sField.getRelatedModelType(),
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
			relatedMStruct := cacheModelMap.Get(sField.getRelatedModelType())
			if relatedMStruct == nil {
				err = errNoModelMappedForRel(
					sField.refStruct.Type,
					sField.getRelatedModelType(),
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
	sub.Root = s
	return
}

// splitted contains filter fields after filter[collection]
// i.e. /blogs?filter[blogs][posts][id][ne]=10
// splitted should be then [posts, id, ne]
func (s *Scope) newFilterField(
	collection string,
	values []string,
	m *ModelStruct,
	splitted ...string,
) (fField *FilterField, errs []*ErrorObject, err error) {
	var (
		sField     *StructField
		relMStruct *ModelStruct
		op         FilterOperator
		ok         bool
		fieldName  string

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
		setFilterField = func() {
			// get the filter field from the scope if exists.

			var sameStruct bool = m == s.Struct
			if sameStruct {
				fField = s.Filters[sField.getFieldIndex()]
			}

			if fField == nil {
				fField = new(FilterField)
				fField.StructField = sField
				if sameStruct {
					s.Filters[sField.getFieldIndex()] = fField
				}
			}

		}
	)

	if s.Filters == nil {
		s.Filters = make(map[int]*FilterField)
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
		setFilterField()
		op = OpEqual
		errObjects, err = fField.setValues(m.collectionType, values, op)
		if err != nil {
			return
		}
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
				setFilterField()

				// get model type for given field

				relMStruct = cacheModelMap.Get(sField.getRelatedModelType())
				if relMStruct == nil {
					// err is internal
					err = fmt.Errorf("Unmapped model: %s.", sField.getRelatedModelType())
					errs = append(errs, ErrInternalError.Copy())
					return
				}

				// relFilter is a FilterField for specific field in relationship
				var relFilter *FilterField

				relFilter, errObjects, err = s.newFilterField(
					fieldName,
					values,
					relMStruct,
					splitted[1:]...,
				)
				errs = append(errs, errObjects...)
				if err != nil || len(errs) > 0 {
					return
				}
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
		setFilterField()
		errObjects, err = fField.setValues(m.collectionType, values, op)
		errs = append(errs, errObjects...)
		if err != nil {
			return
		}
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
		setFilterField()
		relMStruct = cacheModelMap.Get(sField.getRelatedModelType())
		if relMStruct == nil {
			err = fmt.Errorf("Unmapped model: %s.", sField.getRelatedModelType())
			errs = append(errs, ErrInternalError.Copy())
			return
		}

		var relFilter *FilterField

		// get relationship's filter for specific field (filterfield)
		relFilter, errObjects, err = s.newFilterField(
			fieldName,
			values,
			relMStruct,
			splitted[1:]...,
		)
		errs = append(errs, errObjects...)
		if err != nil {
			return
		}

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

// fields[collection] = field1, field2
func (s *Scope) setWorkingFields(fields ...string) (errs []*ErrorObject) {
	var (
		errObj *ErrorObject
	)

	if len(fields) > s.Struct.getWorkingFieldCount() {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Too many working fields to include ")
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

	if s.Pagination == nil {
		s.Pagination = &Pagination{}
	}
	switch index {
	case 0:
		s.Pagination.Limit = val
	case 1:
		s.Pagination.Offset = val
	case 2:
		s.Pagination.PageNumber = val
	case 3:
		s.Pagination.PageSize = val
	}
	return nil
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
	opened := strings.Index(queryParam, string(annotationOpenedBracket))

	if opened == -1 {
		// no opening
		err := ErrUnsupportedQueryParameter.Copy()
		err.Detail = fmt.Sprintf("The %v query parameter is not supported.", queryParam)
		return "", err
	}

	closed := strings.Index(queryParam, string(annotationClosedBracket))
	if closed == -1 || closed != len(queryParam)-1 {
		// no closing
		err := ErrUnsupportedQueryParameter.Copy()
		err.Detail = fmt.Sprintf("The %v query parameter is not supported.", queryParam)
		return "", err
	}

	collection := queryParam[opened+1 : closed]
	return collection, nil
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
