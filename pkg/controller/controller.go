package controller

import (
	"errors"
	"fmt"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/gateway/endpoint"
	"github.com/kucjac/jsonapi/pkg/gateway/modelhandler"
	"github.com/kucjac/jsonapi/pkg/namer"
	"github.com/kucjac/jsonapi/pkg/query"
	"github.com/kucjac/uni-logger"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Controller is the data structure that is responsible for controlling all the models
// within single API
type Controller struct {
	// APIURLBase is a url prefix for the resources. (I.e. having APIURLBase = "/api/v1" and
	// resource with collection posts this would lead to links -> /api/v1/posts)
	APIURLBase string

	// Models is a *ModelStruct map that
	Models *ModelMap

	// Locale defines the coverage of the i18n and l10n for provided system
	Locale language.Coverage

	// Matcher is the language matcher for the provided controler
	Matcher language.Matcher

	// ErrorLimitMany defiens how fast should the many scope building function finish when error
	// occurs.
	ErrorLimitMany int

	// ErrorLimitSingle defines how fast should the one scope building function finish when error
	// occurs
	ErrorLimitSingle int

	// ErrorLimitRelated defines the upper boundaries for the error count while building
	// related  or relationship scope.
	ErrorLimitRelated int

	// IncludeNestedLimit is a maximum value for nested includes (i.e. IncludeNestedLimit = 1
	// allows ?include=posts.comments but does not allow ?include=posts.comments.author)
	IncludeNestedLimit int

	// FilterValueLimit is a maximum length of the filter values
	FilterValueLimit int

	// Flags is the container for all default flags
	Flags *flags.Container

	logger unilogger.LeveledLogger

	// Namer defines the function strategy how the model's and it's fields are being named
	NamerFunc namer.Namer

	flags map[int]bool

	// StrictUnmarshalMode if set to true, the incoming data cannot contain
	// any unknown fields
	StrictUnmarshalMode bool

	Operators query.OperatorContainer
}

// New Creates raw *jsonapi.Controller with no limits and links.
func New(coverage ...interface{}) *Controller {
	c := &Controller{
		Models:             newModelMap(),
		ErrorLimitMany:     1,
		ErrorLimitSingle:   1,
		ErrorLimitRelated:  1,
		IncludeNestedLimit: 1,
		FilterValueLimit:   5,
		Flags:              flags.New(),
		NamerFunc:          NamingSnake,
		Operators:          NewOpContainer(),
	}

	c.Flags.Set(flags.UseFilterValueLimit, true)
	c.Locale = language.NewCoverage(coverage...)
	c.Matcher = language.NewMatcher(c.Locale.Tags())
	return c
}

// Default creates new *jsonapi.Controller with preset limits:
// 	ErrorLimitMany:		5
// 	ErrorLimitSingle:	2
//	IncludeNestedLimit:	1
// Controller has also set the FlagUseLinks flag to true.
func Default(coverage ...interface{}) *Controller {
	c := &Controller{
		// APIURLBase:         "/",
		Models:             newModelMap(),
		ErrorLimitMany:     5,
		ErrorLimitSingle:   2,
		ErrorLimitRelated:  2,
		IncludeNestedLimit: 1,
		FilterValueLimit:   30,
		Flags:              flags.New(),
		NamerFunc:          NamingSnake,
		Operators:          NewOpContainer(),
	}

	c.Flags.Set(flags.UseLinks, true)
	c.Flags.Set(flags.UseFilterValueLimit, true)
	c.Flags.Set(flags.AddMetaCountList, false)
	c.Flags.Set(flags.ReturnPatchContent, true)

	c.Locale = language.NewCoverage(coverage...)
	c.Matcher = language.NewMatcher(c.Locale.Tags())
	return c
}

// SetLogger sets the logger for the controller operations
func (c *Controller) SetLogger(logger unilogger.LeveledLogger) {
	c.logger = logger
}

/**

PRESETS

*/

func (c *Controller) BuildScopeList(
	req *http.Request, endpoint *endpoint.Endpoint, model *modelhandler.ModelHandler,
) (scope *query.Scope, errs []*ErrorObject, err error) {
	// Get ModelStruct

	mStruct := c.Models.Get(model.ModelType)
	if mStruct == nil {
		err = IErrModelNotMapped
		return
	}

	// overloadPreventer - is a warden upon invalid query parameters
	var (
		errObj       *ErrorObject
		errorObjects []*ErrorObject
		addErrors    = func(errObjects ...*ErrorObject) {
			errs = append(errs, errObjects...)
			scope.currentErrorCount += len(errObjects)
		}
	)
	scope = newRootScope(mStruct)
	scope.ctx = req.Context()
	scope.logger = c.log()

	scope.IncludedScopes = make(map[*ModelStruct]*Scope)

	scope.maxNestedLevel = c.IncludeNestedLimit
	scope.collectionScope = scope
	scope.IsMany = true

	scope.setFlags(endpoint, model, c)

	// Get URLQuery
	q := req.URL.Query()

	// Check first included in order to create subscopes
	included, ok := q[QueryParamInclude]
	if ok {
		// build included scopes
		includedFields := strings.Split(included[0], annotationSeperator)
		errorObjects = scope.buildIncludeList(includedFields...)
		addErrors(errorObjects...)
		if len(errs) > 0 {
			return
		}
	}

	// set included language query parameter

	languages, ok := q[QueryParamLanguage]
	if ok {
		errorObjects, err = c.setIncludedLangaugeFilters(scope, languages[0])
		if err != nil {
			return
		}
		if len(errorObjects) > 0 {
			addErrors(errorObjects...)
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
		case key == QueryParamInclude, key == QueryParamLanguage:
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

			colModel := c.Models.GetByCollection(collection)
			if colModel == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid collection: '%s' in the filter query.", collection)
				addErrors(errObj)
				continue
			}

			filterScope := scope.getModelsRootScope(colModel)
			if filterScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The collection: '%s' is not included in query.", collection)
				addErrors(errObj)
				continue
			}

			splitValues := strings.Split(value[0], annotationSeperator)
			_, errorObjects = buildFilterField(filterScope, collection, splitValues, c, colModel, scope.Flags(), splitted[1:]...)
			addErrors(errorObjects...)
		case key == QueryParamSort:
			splitted := strings.Split(value[0], annotationSeperator)
			errorObjects = scope.buildSortFields(splitted...)
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
			fieldModel := c.Models.GetByCollection(collection)
			if fieldModel == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid collection: '%s' for the fields query.", collection)
				continue
			}

			fieldsScope := scope.getModelsRootScope(fieldModel)
			if fieldsScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter collection: '%s' is not included in the query.", collection)
				addErrors(errObj)
				continue
			}
			splitValues := strings.Split(value[0], annotationSeperator)
			errorObjects = fieldsScope.buildFieldset(splitValues...)
			addErrors(errorObjects...)
		case key == QueryParamPageTotal:
			scope.Flags().Set(flags.AddMetaCountList, true)
		case key == QueryParamLinks:
			var er error
			var links bool
			links, er = strconv.ParseBool(value[0])
			if er != nil {
				addErrors(ErrInvalidQueryParameter.Copy().WithDetail("Provided value for the links parameter is not a valid bool"))
			}
			scope.Flags().Set(flags.UseLinks, links)
		default:
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' is unsupported.", key)
			addErrors(errObj)
		}

		if scope.currentErrorCount >= c.ErrorLimitMany {
			return
		}
	}
	if scope.Pagination != nil {
		er := scope.Pagination.check()
		if er != nil {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Pagination parameters are not valid. %s", er)
			addErrors(errObj)
		}
	}

	// Copy the query.for the included fields
	scope.copyIncludedBoundaries()

	return
}

func (c *Controller) BuildScopeSingle(
	req *http.Request, endpoint *Endpoint, model *ModelHandler,
) (scope *Scope, errs []*ErrorObject, err error) {
	return c.buildScopeSingle(req, endpoint, model, nil)
}

// BuildScopeSingle builds the scope for given request and model.
// It gets and sets the ID from the 'http' request.
func (c *Controller) buildScopeSingle(
	req *http.Request, endpoint *Endpoint, model *ModelHandler, id interface{},
) (scope *Scope, errs []*ErrorObject, err error) {
	// get model type
	// Get ModelStruct
	mStruct := c.Models.Get(model.ModelType)
	if mStruct == nil {
		err = IErrModelNotMapped
		return
	}

	q := req.URL.Query()

	scope = newRootScope(mStruct)
	scope.ctx = scope.Context()
	scope.logger = c.log()

	if id == nil {
		errs, err = c.setIDFilter(req, scope)
		if err != nil {
			return
		}
		if len(errs) > 0 {
			return
		}
	} else {
		scope.SetPrimaryFilters(id)
	}

	scope.setFlags(endpoint, model, c)

	scope.maxNestedLevel = c.IncludeNestedLimit

	errs, err = c.buildQueryParametersSingle(scope, q)

	return
}

// BuildScopeRelated builds the scope for the related
func (c *Controller) BuildScopeRelated(
	req *http.Request, endpoint *Endpoint, model *ModelHandler,
) (scope *Scope, errs []*ErrorObject, err error) {
	mStruct := c.Models.Get(model.ModelType)
	if mStruct == nil {
		err = IErrModelNotMapped
		return
	}

	id, related, err := getIDAndRelated(req, mStruct)
	if err != nil {
		return
	}

	scope = &Scope{
		Struct:                    mStruct,
		Fieldset:                  make(map[string]*StructField),
		currentIncludedFieldIndex: -1,
		kind:   rootKind,
		logger: c.log(),
		ctx:    req.Context(),
	}

	scope.newValueSingle()

	scopeVal := reflect.ValueOf(scope.Value).Elem()
	prim := scopeVal.FieldByIndex(mStruct.primary.refStruct.Index)

	if er := setPrimaryField(id, prim); er != nil {
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided 'id' is of invalid type.")
		errs = append(errs, errObj)
		return
	}

	scope.collectionScope = scope

	relationField, ok := mStruct.relationships[related]
	if !ok {
		// invalid query parameter
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid related field name: '%s', for the collection: '%s'", related, mStruct.collectionType)
		errs = append(errs, errObj)
		return
	}

	errs = scope.setPrimaryFilterfield(c, id)
	if len(errs) > 0 {
		return
	}

	scope.Fieldset[related] = relationField

	if relationField.relationship.Kind == RelBelongsTo {
		fk := relationField.relationship.ForeignKey
		if fk != nil {
			scope.Fieldset[fk.jsonAPIName] = fk
		}
	}

	scope.IncludedScopes = make(map[*ModelStruct]*Scope)

	// preset related scope
	includedField := scope.getOrCreateIncludeField(relationField)
	includedField.Scope.kind = relatedKind

	q := req.URL.Query()
	languages, ok := q[QueryParamLanguage]
	if ok {
		errs, err = c.setIncludedLangaugeFilters(includedField.Scope, languages[0])
		if err != nil || len(errs) > 0 {
			return
		}
	}

	qLinks, ok := q[QueryParamLinks]
	if ok {
		var er error
		var links bool
		links, er = strconv.ParseBool(qLinks[0])
		if er != nil {
			errs = append(errs, ErrInvalidQueryParameter.Copy().WithDetail("Provided value for the links parameter is not a valid bool"))
			return
		}
		scope.Flags().Set(flags.UseLinks, links)
	}

	scope.copyIncludedBoundaries()

	return
}

func (c *Controller) BuildScopeRelationship(
	req *http.Request, endpoint *Endpoint, model *ModelHandler,
) (scope *Scope, errs []*ErrorObject, err error) {
	mStruct := c.Models.Get(model.ModelType)
	if mStruct == nil {
		err = IErrModelNotMapped
		return
	}

	id, relationship, err := getIDAndRelationship(req, mStruct)
	if err != nil {
		return
	}

	scope = &Scope{
		Struct:                    mStruct,
		Fieldset:                  make(map[string]*StructField),
		currentIncludedFieldIndex: -1,
		kind:   rootKind,
		logger: c.log(),
		ctx:    req.Context(),
	}
	scope.collectionScope = scope
	scope.newValueSingle()

	scopeVal := reflect.ValueOf(scope.Value).Elem()
	prim := scopeVal.FieldByIndex(mStruct.primary.refStruct.Index)

	if er := setPrimaryField(id, prim); er != nil {
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided 'id' is of invalid type.")
		errs = append(errs, errObj)
		return
	}

	// set primary field filter
	errs = scope.setPrimaryFilterfield(c, id)
	if len(errs) >= c.ErrorLimitRelated {
		return
	}

	relationField, ok := mStruct.relationships[relationship]
	if !ok {
		// invalid query parameter
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid relationship name: '%s', for the collection: '%s'", relationship, mStruct.collectionType)
		errs = append(errs, errObj)
		return
	}

	// preset root scope
	scope.Fieldset[relationship] = relationField

	if relationField.relationship != nil && relationField.relationship.Kind == RelBelongsTo {
		fk := relationField.relationship.ForeignKey
		if fk != nil {
			scope.Fieldset[fk.jsonAPIName] = fk
		}
	}
	scope.IncludedScopes = make(map[*ModelStruct]*Scope)
	scope.createModelsRootScope(relationField.relatedStruct)
	scope.IncludedScopes[relationField.relatedStruct].Fieldset = nil

	// preset relationship scope
	includedField := scope.getOrCreateIncludeField(relationField)
	includedField.Scope.kind = relationshipKind

	scope.copyIncludedBoundaries()

	return
}

// GetAndSetIDFilter is the method that gets the ID value from the request path, for given scope's
// model. Then it sets the id query.for given scope.
// If the url is constructed incorrectly it returns an internal error.
func (c *Controller) GetAndSetIDFilter(req *http.Request, scope *Scope) error {
	c.log().Debug("GetAndSetIDFIlter")

	id, err := getID(req, scope.Struct)
	if err != nil {
		return err
	}

	scope.setIDFilterValues(id)

	return nil
}

func (c *Controller) GetAndSetID(req *http.Request, scope *Scope) (prim interface{}, err error) {
	id, err := getID(req, scope.Struct)
	if err != nil {
		return nil, err
	}

	if scope.Value == nil {
		return nil, IErrNoValue
	}

	val := reflect.ValueOf(scope.Value)
	if val.IsNil() {
		return nil, IErrScopeNoValue
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	} else {
		return nil, IErrInvalidType
	}

	c.log().Debugf("Setting newPrim value")
	newPrim := reflect.New(scope.Struct.primary.refStruct.Type).Elem()

	err = setPrimaryField(id, newPrim)
	if err != nil {
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = "Provided invalid id value within the url."
		return nil, errObj
	}

	primVal := val.FieldByIndex(scope.Struct.primary.refStruct.Index)

	if reflect.DeepEqual(primVal.Interface(), reflect.Zero(scope.Struct.primary.refStruct.Type).Interface()) {
		primVal.Set(newPrim)
	} else {
		c.log().Debugf("Checking the values")
		if newPrim.Interface() != primVal.Interface() {
			errObj := ErrInvalidQueryParameter.Copy()
			errObj.Detail = "Provided invalid id value within the url. The id value doesn't match the primary field within the root object."
			return nil, errObj
		}
	}

	v := reflect.ValueOf(scope.Value)
	if v.Kind() != reflect.Ptr {
		return nil, IErrInvalidType
	}
	return primVal.Interface(), nil
}

// GetModelStruct returns the ModelStruct for provided model
// Returns error if provided model does not exists in the PrecomputedMap
func (c *Controller) GetModelStruct(model interface{}) (*ModelStruct, error) {
	return c.getModelStruct(model)
}

// GetCheckSetIDFilter the method that gets the ID value from the request path, for given scope
// model.then prepares the primary filter field for given id value.
// if an internal error occurs returns an 'error'.
// if user error occurs returns array of *ErrorObject's.
func (c *Controller) GetSetCheckIDFilter(req *http.Request, scope *Scope,
) (errs []*ErrorObject, err error) {
	return c.setIDFilter(req, scope)
}

func (c *Controller) Log() unilogger.LeveledLogger {
	return c.log()
}

// MarshalScope marshals given scope into jsonapi format.
func (c *Controller) MarshalScope(scope *Scope) (payloader Payloader, err error) {
	return marshalScope(scope, c)
}

// MustGetModelStruct gets (concurrently safe) the model struct from the cached model Map
// panics if the model does not exists in the map.
func (c *Controller) MustGetModelStruct(model interface{}) *ModelStruct {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		panic(err)
	}
	return mStruct
}

func (c *Controller) NewScope(model interface{}) (*Scope, error) {
	mStruct, err := c.GetModelStruct(model)
	if err != nil {
		return nil, err
	}

	scope := newRootScope(mStruct)
	scope.logger = c.log()

	return scope, nil
}

func (c *Controller) NewFilterFieldWithForeigns(fieldFilter string, values ...interface{},
) (filter *FilterField, err error) {

	if valid := strings.HasPrefix(fieldFilter, QueryParamFilter); !valid {
		err = fmt.Errorf("Invalid field filter argument provided: '%s'. The argument should be composed as: 'filter[collection][field]([subfield][operator])|([operator]).", fieldFilter)
		return
	}

	var splitted []string
	splitted, err = splitBracketParameter(fieldFilter[len(QueryParamFilter):])
	if err != nil {
		return
	}

	mStruct := c.Models.GetByCollection(splitted[0])
	if mStruct == nil {
		err = fmt.Errorf("The model for collection: '%s' is not precomputed within controller. Cannot clreate new filter field.", splitted[0])
		return
	}

	var (
		structField *StructField
		operator    *FilterOperator
	)
	handle3and4 := func() {
		structField = mStruct.relationships[splitted[1]]
		if structField == nil {
			err = fmt.Errorf("Invalid field name provided in fieldFilter: '%s'.", fieldFilter)
			return
		}
		filter = &FilterField{StructField: structField}

		if splitted[2] == "id" {
			structField = filter.relatedStruct.primary
		} else {
			if structField = filter.relatedStruct.attributes[splitted[2]]; structField == nil {
				if structField = filter.relatedStruct.foreignKeys[splitted[2]]; structField == nil {
					err = fmt.Errorf("Invalid subfield name provided: '%s' for the query: '%s'.", splitted[2], fieldFilter)
					return
				}
			}
		}
		subfieldFilter := &FilterField{StructField: structField}

		if len(splitted) == 4 {
			var ok bool
			operator, ok = c.Operators[splitted[3]]
			if !ok {
				err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[3], fieldFilter)
				return
			}
		} else {
			operator = OpIn
		}

		fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}
		for i, value := range values {
			t := reflect.TypeOf(value)
			if t == subfieldFilter.refStruct.Type {
				fv.Values[i] = value
			} else {
				err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, subfieldFilter.refStruct.Type)
				return
			}
		}

		subfieldFilter.Values = append(subfieldFilter.Values, fv)

		// Add subfield
		filter.Nested = append(filter.Nested, subfieldFilter)
	}

	handle2and3 := func() {
		if splitted[1] == "id" {
			structField = mStruct.primary
		} else {
			structField = mStruct.attributes[splitted[1]]
			if structField == nil {
				structField = mStruct.foreignKeys[splitted[1]]
				if structField == nil {
					if structField = mStruct.relationships[splitted[1]]; structField != nil {
						if len(splitted) == 3 {
							handle3and4()
						} else {
							err = fmt.Errorf("The relationship field: '%s' in the filter argument must specify the subfield. Filter: '%s'", splitted[1], fieldFilter)
						}
					} else {
						err = fmt.Errorf("Invalid field name: '%s' for the collection: '%s'.", splitted[1], splitted[0])
					}
					return
				}

			}
		}
		filter = &FilterField{StructField: structField}
		var ok bool
		if len(splitted) == 3 {
			operator, ok = operatorsValue[splitted[2]]
			if !ok {
				err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[2], fieldFilter)
				return
			}
		} else {
			operator = OpIn
		}
		fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}

		for i, value := range values {
			t := reflect.TypeOf(value)
			if t == filter.refStruct.Type {
				fv.Values[i] = value
			} else {
				err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, filter.refStruct.Type)
				return
			}
		}
		filter.Values = append(filter.Values, fv)
	}

	switch len(splitted) {
	case 2, 3:
		handle2and3()
	case 4:
		handle3and4()
	default:
		err = fmt.Errorf("The filter argument: '%s' is of invalid form.", fieldFilter)
		return
	}
	return
}

// NewFilterField creates new filter field based on the fieldFilter argument and provided values
//	Arguments:
//		@fieldFilter - jsonapi field name for provided model the field type must be of the same as
//				the last element of the preset. By default the filter operator is of 'in' type.
//				It must be of form: filter[collection][field]([operator]|([subfield][operator])).
//				The operator or subfield are not required.
//		@values - the values of type matching the filter field
func (c *Controller) NewFilterField(fieldFilter string, values ...interface{},
) (filter *FilterField, err error) {

	if valid := strings.HasPrefix(fieldFilter, QueryParamFilter); !valid {
		err = fmt.Errorf("Invalid field filter argument provided: '%s'. The argument should be composed as: 'filter[collection][field]([subfield][operator])|([operator]).", fieldFilter)
		return
	}

	var splitted []string
	splitted, err = splitBracketParameter(fieldFilter[len(QueryParamFilter):])
	if err != nil {
		return
	}

	mStruct := c.Models.GetByCollection(splitted[0])
	if mStruct == nil {
		err = fmt.Errorf("The model for collection: '%s' is not precomputed within controller. Cannot clreate new filter field.", splitted[0])
		return
	}

	var (
		structField *StructField
		operator    FilterOperator
	)
	handle3and4 := func() {
		structField = mStruct.relationships[splitted[1]]
		if structField == nil {
			err = fmt.Errorf("Invalid field name provided in fieldFilter: '%s'.", fieldFilter)
			return
		}
		filter = &FilterField{StructField: structField}

		if splitted[2] == "id" {
			structField = filter.relatedStruct.primary
		} else {
			if structField = filter.relatedStruct.attributes[splitted[2]]; structField == nil {
				err = fmt.Errorf("Invalid subfield name provided: '%s' for the query: '%s'.", splitted[2], fieldFilter)
				return
			}
		}
		subfieldFilter := &FilterField{StructField: structField}

		if len(splitted) == 4 {
			var ok bool
			operator, ok = operatorsValue[splitted[3]]
			if !ok {
				err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[3], fieldFilter)
				return
			}
		} else {
			operator = OpIn
		}

		fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}
		for i, value := range values {
			t := reflect.TypeOf(value)
			if t == subfieldFilter.refStruct.Type {
				fv.Values[i] = value
			} else {
				err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, subfieldFilter.refStruct.Type)
				return
			}
		}

		subfieldFilter.Values = append(subfieldFilter.Values, fv)

		// Add subfield
		filter.Nested = append(filter.Nested, subfieldFilter)
	}

	handle2and3 := func() {
		if splitted[1] == "id" {
			structField = mStruct.primary
		} else {
			structField = mStruct.attributes[splitted[1]]
			if structField == nil {
				if structField = mStruct.relationships[splitted[1]]; structField != nil {
					if len(splitted) == 3 {
						handle3and4()
					} else {
						err = fmt.Errorf("The relationship field: '%s' in the filter argument must specify the subfield. Filter: '%s'", splitted[1], fieldFilter)
					}
				} else {
					err = fmt.Errorf("Invalid field name: '%s' for the collection: '%s'.", splitted[1], splitted[0])
				}
				return
			}
		}
		filter = &FilterField{StructField: structField}
		var ok bool
		if len(splitted) == 3 {
			operator, ok = operatorsValue[splitted[2]]
			if !ok {
				err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[2], fieldFilter)
				return
			}
		} else {
			operator = OpIn
		}
		fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}

		for i, value := range values {
			t := reflect.TypeOf(value)
			if t == filter.refStruct.Type {
				fv.Values[i] = value
			} else {
				err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, filter.refStruct.Type)
				return
			}
		}
		filter.Values = append(filter.Values, fv)
	}

	switch len(splitted) {
	case 2, 3:
		handle2and3()
	case 4:
		handle3and4()
	default:
		err = fmt.Errorf("The filter argument: '%s' is of invalid form.", fieldFilter)
		return
	}
	return
}

// PrecomputeModels precomputes provided models, making it easy to check
// models relationships and  attributes.
func (c *Controller) PrecomputeModels(models ...interface{}) error {
	var err error
	if c.Models == nil {
		c.Models = newModelMap()
	}

	for _, model := range models {
		err = c.buildModelStruct(model, c.Models)
		if err != nil {
			return err
		}
	}
	for _, model := range c.Models.models {
		err = c.checkModelRelationships(model)
		if err != nil {
			return err
		}
		err = model.initCheckFieldTypes()
		if err != nil {
			return err
		}

		model.initComputeSortedFields()
		model.initComputeThisIncludedCount()
		c.Models.collections[model.collectionType] = model.modelType
	}

	for _, model := range c.Models.models {
		model.nestedIncludedCount = model.initComputeNestedIncludedCount(0, c.IncludeNestedLimit)

	}
	err = c.setRelationships()
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) SetAPIURL(url string) error {
	// manage the url
	c.APIURLBase = url
	return nil
}

// build filterfield
func (c *Controller) buildFilterField(
	s *Scope,
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

	if len(values) > c.FilterValueLimit {
		errObj = ErrOutOfRangeQueryParameterValue.Copy()
		errObj.Detail = fmt.Sprintf("The number of values for the filter: %s within collection: %s exceeds the permissible length: '%d'", strings.Join(splitted, annotationSeperator), collection, c.FilterValueLimit)
		errs = append(errs, errObj)
		return
	}

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

				sField, ok = m.filterKeys[fieldName]
				if !ok {
					invalidName(fieldName, m.collectionType)
					return
				}
				fField = s.getOrCreateFilterKeyFilter(sField)
			} else {
				if sField.isLanguage() {
					fField = s.getOrCreateLangaugeFilter()
				} else if sField.isMap() {
					// Map doesn't allow any default parameters
					// i.e. filter[collection][mapfield]=something
					errObj = ErrInvalidQueryParameter.Copy()
					errObj.Detail = "Cannot filter field of 'Hashmap' type with no operator"
					errs = append(errs, errObj)
					return
				} else {
					fField = s.getOrCreateAttributeFilter(sField)
				}
			}
		}

		errObjects = c.setFilterValues(fField, m.collectionType, values, OpEqual)
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

				errObj = buildNestedFilter(c, fField, values, splitted[1:]...)
				if errObj != nil {
					errs = append(errs, errObj)
				}

				return
			}
			if sField.isLanguage() {
				fField = s.getOrCreateLangaugeFilter()
			} else {
				fField = s.getOrCreateAttributeFilter(sField)
			}
		}
		// it is an attribute filter
		op, ok = operatorsValue[splitted[1]]
		if !ok {
			invalidOperator(splitted[1])
			return
		}

		errObjects = c.setFilterValues(fField, m.collectionType, values, op)
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

		errObj = buildNestedFilter(c, fField, values, splitted[1:]...)
		if errObj != nil {
			errs = append(errs, errObj)
		}

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

func (c *Controller) buildQueryParametersSingle(
	scope *Scope, q url.Values,
) (errs []*ErrorObject, err error) {

	var (
		addErrors = func(errObjects ...*ErrorObject) {
			errs = append(errs, errObjects...)
			scope.currentErrorCount += len(errObjects)
		}
		errObj       *ErrorObject
		errorObjects []*ErrorObject
	)
	// Check first included in order to create subscopes
	included, ok := q[QueryParamInclude]
	if ok {
		// build included scopes
		if len(included) != 1 {
			errObj := ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintln("Duplicated 'included' query parameter.")
			addErrors(errObj)
			return
		}
		includedFields := strings.Split(included[0], annotationSeperator)
		errorObjects = scope.buildIncludeList(includedFields...)
		addErrors(errorObjects...)
		if len(errs) > 0 {
			return
		}
	}

	languages, ok := q[QueryParamLanguage]
	if ok {
		errorObjects, err = c.setIncludedLangaugeFilters(scope, languages[0])
		if err != nil {
			return
		}
		if len(errorObjects) > 0 {
			addErrors(errorObjects...)
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
		case key == QueryParamInclude, key == QueryParamLanguage:
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

			fieldsetModel := c.Models.GetByCollection(collection)
			if fieldsetModel == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid collection: '%s' for the fields query.", collection)
				addErrors(errObj)
				continue
			}

			fieldsetScope := scope.getModelsRootScope(fieldsetModel)
			if fieldsetScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter collection: '%s' is not included in the query.", collection)
				addErrors(errObj)
				continue
			}
			splitValues := strings.Split(values[0], annotationSeperator)

			errorObjects = fieldsetScope.buildFieldset(splitValues...)
			addErrors(errorObjects...)
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

			colModel := c.Models.GetByCollection(collection)
			if colModel == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided invalid collection: '%s' in the filter query.", collection)
				addErrors(errObj)
				continue
			}

			filterScope := scope.getModelsRootScope(colModel)
			if filterScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The collection: '%s' is not included in query.", collection)
				addErrors(errObj)
				continue
			}

			splitValues := strings.Split(values[0], annotationSeperator)

			_, errorObjects = buildFilterField(filterScope, collection, splitValues, c, colModel, filterScope.Flags(), splitted[1:]...)

			addErrors(errorObjects...)
		case key == QueryParamLinks:
			var er error
			var links bool
			links, er = strconv.ParseBool(values[0])
			if er != nil {
				addErrors(ErrInvalidQueryParameter.Copy().WithDetail("Provided value for the links parameter is not a valid bool"))
			}
			scope.Flags().Set(flags.UseLinks, links)
		default:
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The query parameter: '%s' is unsupported.", key)
			addErrors(errObj)
		}

		if scope.currentErrorCount >= c.ErrorLimitSingle {
			return
		}
	}
	scope.copyIncludedBoundaries()
	return
}

func (c *Controller) buildPreparedPair(
	query, fieldFilter string,
	check bool,
) (presetScope *Scope, filter *FilterField) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	filter, err = c.NewFilterField(fieldFilter)
	if err != nil {
		err = fmt.Errorf("Invalid field filter provided for presetPair: '%s'. %v", fieldFilter, err)
		return
	}

	// mStruct := filter.mStruct

	// Parse the query
	var queryParsed url.Values
	queryParsed, err = url.ParseQuery(query)
	if err != nil {
		return
	}

	preset := queryParsed.Get(QueryParamPreset)
	if preset == "" {
		err = fmt.Errorf("Invalid query: '%s' . No preset parameter", query)
		return
	}

	presets := strings.Split(preset, annotationSeperator)
	if len(presets) != 1 {
		err = fmt.Errorf("Multiple presets are not allowed. Query: '%s'", query)
		return
	}

	presetValues := strings.Split(presets[0], annotationNestedSeperator)
	// if len(presetValues) <= 1 {
	// 	err = fmt.Errorf("Preset scope should contain at least two parameter values within its preset. Query: '%s'", query)
	// 	return
	// }

	rootModel := c.Models.GetByCollection(presetValues[0])
	if rootModel == nil {
		err = fmt.Errorf("Invalid query parameter: '%s'. No model found for the collection: '%s'. ", query, presetValues[0])
		return
	}

	presetScope = newRootScope(rootModel)
	presetScope.maxNestedLevel = 10000
	presetScope.Fieldset = nil

	if len(presetValues) > 1 {
		if errs := presetScope.buildIncludeList(strings.Join(presetValues[1:], annotationNestedSeperator)); len(errs) > 0 {
			err = fmt.Errorf("Invalid preset values. %s", errs)
			return
		}
	}

	var selectIncludedFieldset func(scope *Scope) error

	selectIncludedFieldset = func(scope *Scope) error {
		scope.Fieldset = make(map[string]*StructField)
		for scope.NextIncludedField() {
			included, err := scope.CurrentIncludedField()
			if err != nil {
				return err
			}
			scope.Fieldset[included.jsonAPIName] = included.StructField
			if err = selectIncludedFieldset(included.Scope); err != nil {
				return err
			}
		}
		scope.ResetIncludedField()
		return nil
	}

	if err = selectIncludedFieldset(presetScope); err != nil {
		return
	}

	// var fieldsetFound bool
	for key, value := range queryParsed {
		if strings.HasPrefix(key, QueryParamFields) {
			var splitted []string
			splitted, err = splitBracketParameter(key[len(QueryParamFields):])
			if err != nil {
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter is of invalid form. %s", err)
				err = errObj
				return
			}

			if len(splitted) != 1 {
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter: '%s' is of invalid form. Nested 'fields' is not supported.", key)
				err = errObj
				return
			}

			collection := splitted[0]
			fieldModel := c.Models.GetByCollection(collection)
			if fieldModel == nil {
				err = fmt.Errorf("Collection: '%s' not found within controller.", collection)
				return
			}

			fieldScope := presetScope.getModelsRootScope(fieldModel)
			if fieldScope == nil {
				err = fmt.Errorf("Collection: '%s' not included into preset. For query: '%s'.", collection, query)
				return
			}

			splitFields := strings.Split(value[0], annotationSeperator)
			errs := fieldScope.buildFieldset(splitFields...)
			if len(errs) != 0 {
				err = fmt.Errorf("Invalid fieldset provided. %v", errs)
				return
			}
			// fieldsetFound = true
		}
	}

	// if !fieldsetFound && check {
	// 	err = fmt.Errorf("The fieldset not found within the query: '%s'.", query)
	// 	return
	// }

	for key, value := range queryParsed {
		switch {
		case key == QueryParamPreset:
			// Preset are already defined
			continue
		case strings.HasPrefix(key, QueryParamFilter):
			// Filter
			var parameters []string
			parameters, err = splitBracketParameter(key[len(QueryParamPreset):])
			if err != nil {
				return
			}

			filterModel := c.Models.GetByCollection(parameters[0])
			if filterModel == nil {
				err = fmt.Errorf("Invalid collection: '%s' for filter parameter within query: %s.", parameters[0], query)
				return
			}

			filterScope := presetScope.getModelsRootScope(filterModel)
			if filterScope == nil {
				err = fmt.Errorf("Invalid collection: '%s' for filter parameter. The collection is not included into preset parameter: '%s' within query: '%s' ", filterModel.collectionType, preset, query)
				return
			}

			filterValues := strings.Split(value[0], annotationSeperator)

			_, errs := buildFilterField(filterScope, filterModel.collectionType,
				filterValues, c, filterModel, c.Flags, parameters[1:]...)
			// _, errs := filterScope.buildFilterfield(filterModel.collectionType, filterValues, filterModel, parameters[1:]...)
			if len(errs) != 0 {
				err = fmt.Errorf("Error while building filter field for collection: '%s' in query: '%s'. Errs: '%s'", filterModel.collectionType, query, errs)
				return
			}

		case strings.HasPrefix(key, QueryParamSort):
			// Sort given collection
			var parameters []string
			parameters, err = splitBracketParameter(key[len(QueryParamSort):])
			if err != nil {
				return
			}
			if len(parameters) > 1 {
				err = fmt.Errorf("Too many sort field parameters within brackets. '%s'", parameters)
				return
			}

			var sortScope *Scope
			if parameters[0] == presetScope.Struct.collectionType {
				sortScope = presetScope
			} else {
				sortModel := c.Models.GetByCollection(parameters[0])
				if sortModel == nil {
					err = fmt.Errorf("Invalid sort collection provided: '%s'.", parameters[0])
					return
				}

				sortScope = presetScope.IncludedScopes[sortModel]
				if sortScope == nil {
					err = fmt.Errorf("The sort collection '%s' is not within the preset parameter in query: '%s'", parameters[0], query)
					return
				}
			}

			var order Order
			sortField := value[0]
			if sortField[0] == '-' {
				order = DescendingOrder
				sortField = sortField[1:]
			}

			if invalidField := newSortField(sortField, order, sortScope); invalidField {
				err = fmt.Errorf("Provided invalid sort field: '%s' for collection: '%s' in query: '%s'.", sortField, sortScope.Struct.collectionType, query)
				return
			}
		case strings.HasPrefix(key, QueryParamPage):
			// Limit given collection
			var parameters []string

			parameters, err = splitBracketParameter(key[len(QueryParamPage):])
			if err != nil {
				return
			}
			if len(parameters) != 2 {
				err = fmt.Errorf("Too many bracket parameters in page[limit] parameter within query: '%s'. The parameter takes only limit and collection bracket parameters.", query)
				return
			}

			var paginateType PaginationParameter
			switch parameters[0] {
			case "limit":
				paginateType = limitPaginateParam
			case "offset":
				paginateType = offsetPaginateParam
			case "number":
				paginateType = numberPaginateParam
			case "size":
				paginateType = sizePaginateParam
			default:
				err = fmt.Errorf("Invalid paginate type parameter: '%s' within query: '%s'", parameters[0], query)
				return
			}

			pageModel := c.Models.GetByCollection(parameters[1])
			if pageModel == nil {
				err = fmt.Errorf("Provided invalid collection for pagination: '%s' within query: '%s'", parameters[1], query)
				return
			}

			pageScope := presetScope.getModelsRootScope(pageModel)
			if pageScope == nil {
				err = fmt.Errorf("Provided collection: '%s' within query: '%s' is not included into preset parameter.", parameters[1], query)
				return
			}

			if errObj := pageScope.preparePaginatedValue(key+"["+parameters[0]+"]", value[0], int(paginateType)); errObj != nil {
				err = errObj
				return
			}
		case strings.HasPrefix(key, QueryParamFields):
			continue
		}
	}
	presetScope.copyPresetParameters()

	// If the prepared scope is not checked (preset)
	// Then check if the value range is consistent
	if !check {
		/**

		TO DO:

		*/

	}

	return
}

func (c *Controller) checkModelRelationships(model *ModelStruct) (err error) {
	for _, rel := range model.relationships {
		val := c.Models.Get(rel.relatedModelType)
		if val == nil {
			err = fmt.Errorf("Model: %v, not precalculated but is used in relationships for: %v field in %v model.", rel.relatedModelType, rel.fieldName, model.modelType.Name())
			return err
		}
		rel.relatedStruct = val
	}
	return
}

func (c *Controller) displaySupportedLanguages() []string {
	namer := display.Tags(language.English)
	var names []string = make([]string, len(c.Locale.Tags()))
	for i, lang := range c.Locale.Tags() {
		names[i] = fmt.Sprintf("%s - '%s'", namer.Name(lang), lang.String())
	}
	return names
}

func (c *Controller) getModelStruct(model interface{}) (*ModelStruct, error) {
	if model == nil {
		return nil, errors.New("No model provided.")
	}
	modelType := reflect.ValueOf(model).Type()
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	mStruct := c.Models.Get(modelType)
	if mStruct == nil {
		return nil, fmt.Errorf("Unmapped model provided: %s", modelType.Name())
	}
	return mStruct, nil
}

func (c *Controller) log() unilogger.LeveledLogger {
	if c.logger == nil {
		basic := unilogger.NewBasicLogger(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
		basic.SetLevel(unilogger.INFO)
		c.logger = basic
	}
	return c.logger
}

// func (c *Controller) registerOperators()

func (c *Controller) setIDFilter(req *http.Request, scope *Scope,
) (errs []*ErrorObject, err error) {
	var id string
	id, err = getID(req, scope.Struct)
	if err != nil {
		return
	}
	errs = scope.setPrimaryFilterfield(c, id)
	if len(errs) > 0 {
		return
	}
	return
}

func (c *Controller) setFilterValues(
	f *FilterField,
	collection string,
	values []string,
	op FilterOperator,
) (errs []*ErrorObject) {
	var (
		er     error
		errObj *ErrorObject

		opInvalid = func() {
			errObj = ErrUnsupportedQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.jsonAPIName)
			errs = append(errs, errObj)
		}
	)

	if op > OpEndsWith {
		errObj = ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported by the server.", op)
	}

	t := f.getDereferencedType()
	// create new FilterValue
	fv := new(FilterValues)
	fv.Operator = op

	// Add and check all values for given field type
	switch f.fieldType {
	case Primary:
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if op > OpLessEqual {
				opInvalid()
			}
		case reflect.String:
			if !op.isBasic() {
				opInvalid()
			}
		}
		for _, value := range values {
			fieldValue := reflect.New(t).Elem()
			er = setPrimaryField(value, fieldValue)
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for primary field in collection: '%s'. %s. ", collection, er)
				errs = append(errs, errObj)
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.Values = append(f.Values, fv)

		// if it is of integer type check which kind of it
	case Attribute:
		switch t.Kind() {
		case reflect.String:
		default:
			if op.isStringOnly() {
				opInvalid()
			}
		}
		if f.isLanguage() {
			switch op {
			case OpIn, OpEqual, OpNotIn, OpNotEqual:
				for i, value := range values {
					tag, err := language.Parse(value)
					if err != nil {
						switch v := err.(type) {
						case language.ValueError:
							errObj := ErrLanguageNotAcceptable.Copy()
							errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s', is not a valid language. Cannot recognize subfield: '%s'.", value, f.GetJSONAPIName(),
								collection, v.Subtag())
							errs = append(errs, errObj)
							continue
						default:
							errObj := ErrInvalidQueryParameter.Copy()
							errObj.Detail = fmt.Sprintf("The value: '%v' for the '%s' filter field within the collection '%s' is not syntetatically valid.", value, f.GetJSONAPIName(), collection)
							errs = append(errs, errObj)
							continue
						}
					}
					if op == OpEqual {
						var confidence language.Confidence
						tag, _, confidence = c.Matcher.Match(tag)
						if confidence <= language.Low {
							errObj := ErrLanguageNotAcceptable.Copy()
							errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s' does not match any supported langauges. The server supports following langauges: %s", value, f.GetJSONAPIName(), collection,
								strings.Join(c.displaySupportedLanguages(), annotationSeperator),
							)
							errs = append(errs, errObj)
							return
						}
					}
					b, _ := tag.Base()
					values[i] = b.String()

				}
			default:
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided operator: '%s' for the language field is not acceptable", op.String())
				errs = append(errs, errObj)
				return
			}
		}
		for _, value := range values {
			fieldValue := reflect.New(t).Elem()
			er = setAttributeField(value, fieldValue)
			if er != nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Invalid filter value for the attribute field: '%s' for collection: '%s'. %s.", f.jsonAPIName, collection, er)
				errs = append(errs, errObj)
				continue
			}
			fv.Values = append(fv.Values, fieldValue.Interface())
		}

		f.Values = append(f.Values, fv)

	}
	return
}

func (c *Controller) setIncludedLangaugeFilters(
	scope *Scope,
	languages string,
) ([]*ErrorObject, error) {
	tags, errObjects := buildLanguageTags(languages)
	if len(errObjects) > 0 {
		return errObjects, nil
	}
	tag, _, confidence := c.Matcher.Match(tags...)
	if confidence <= language.Low {
		// language not supported
		errObj := ErrLanguageNotAcceptable.Copy()
		errObj.Detail = fmt.Sprintf("Provided languages: '%s' are not supported. This document supports following languages: %s",
			languages,
			strings.Join(c.displaySupportedLanguages(), ","),
		)
		return []*ErrorObject{errObj}, nil
	}
	scope.queryLanguage = tag

	var setLanguage func(scope *Scope) error
	//setLanguage sets language filter field for all included fields
	setLanguage = func(scope *Scope) error {
		defer func() {
			scope.ResetIncludedField()
		}()
		if scope.UseI18n() {
			base, _ := tag.Base()
			scope.SetLanguageFilter(base.String())
		}
		for scope.NextIncludedField() {
			field, err := scope.CurrentIncludedField()
			if err != nil {
				return err
			}
			if err = setLanguage(field.Scope); err != nil {
				return err
			}
		}
		return nil
	}

	if err := setLanguage(scope); err != nil {
		return errObjects, err
	}
	return errObjects, nil
}

func (c *Controller) supportI18n() bool {
	if c.Locale == nil || c.Matcher == nil {
		return false
	}

	if len(c.Locale.Tags()) == 0 {
		return false
	}
	return true
}
