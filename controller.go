package jsonapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// Controller
type Controller struct {
	// APIURLBase is a url prefix for the resources. (I.e. having APIURLBase = "/api/v1" and
	// resource with collection posts this would lead to links -> /api/v1/posts)
	APIURLBase string

	// Models is a *ModelStruct map that
	Models *ModelMap

	// Pagination is a pagination scope that defines default pagination.
	// If nil then the value would not be included
	Pagination *Pagination

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

	// UseLinks is a flag that defines if the response should contain links objects
	UseLinks bool
}

// New Creates raw *jsonapi.Controller with no limits and links.
func New() *Controller {
	return &Controller{
		Models:             newModelMap(),
		ErrorLimitMany:     1,
		ErrorLimitSingle:   1,
		ErrorLimitRelated:  1,
		IncludeNestedLimit: 1,
	}
}

// Default creates new *jsonapi.Controller with preset limits:
// 	ErrorLimitMany:		5
// 	ErrorLimitSingle:	2
//	IncludeNestedLimit:	1
// Controller has also set the UseLinks flag to true.
func Default() *Controller {
	return &Controller{
		// APIURLBase:         "/",
		Models:             newModelMap(),
		ErrorLimitMany:     5,
		ErrorLimitSingle:   2,
		ErrorLimitRelated:  2,
		IncludeNestedLimit: 1,
		UseLinks:           true,
	}
}

/**

PRESETS

*/

// BuildPresetScope builds the preset scope which should enable the table 'JOIN' feature.
// The query parameter should be URL parseble query ("x1=1&x2=2" etc.)
// The query parameter that are allowed are:
//	- preset=collection.relationfield.relationfield .... - this creates a relation path
//		the last relation field should be of type of model provided as an argument.
//	- filter[collection][field][operator]=value
//	- page[limit][collection] - limit the value of ids within given collection
//	- sort[collection]=field - sorts the collection by provided field. Does not allow nesteds.
// 		@query - url like query that should define how the preset scope should look like. The query
//		allows to set the relation path, filter collections, limit given collection, sort given
//		collection.
//	 	model - is a model that defines target of the preset scope. The last preset relation field //		should be of the same collection as provided model.
func (c *Controller) BuildPresetScope(query string, model interface{},
) (presetScope *Scope) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		err = fmt.Errorf("Provided invalid model: %s", modelType.Kind())
		return
	}

	mStruct := c.Models.Get(modelType)
	if mStruct == nil {

		err = fmt.Errorf("Invalid model provided. The model of type: '%s' is not precomputed within controller. %v", modelType.Name(), c.Models.models)
		return
	}

	var queryParsed url.Values
	queryParsed, err = url.ParseQuery(query)
	if err != nil {
		return
	}

	preset := queryParsed.Get(QueryParamPreset)
	if preset == "" {
		err = fmt.Errorf("Invalid query: '%s' for model: '%s'. No preset parameter", query, modelType.Name())
		return
	}

	presets := strings.Split(preset, annotationSeperator)
	if len(presets) != 1 {
		err = fmt.Errorf("Multiple presets are not allowed. Query: '%s'", query)
		return
	}

	presetValues := strings.Split(presets[0], annotationNestedSeperator)
	if len(presetValues) <= 1 {
		err = fmt.Errorf("Preset scope should contain at least two parameter values within its preset. Query: '%s'", query)
		return
	}

	rootModel := c.Models.GetByCollection(presetValues[0])
	if rootModel == nil {
		err = fmt.Errorf("Invalid query parameter: '%s'. No model found for the collection: '%s'. ", query, presetValues[0])
		return
	}

	presetScope = newRootScope(rootModel)
	presetScope.maxNestedLevel = 10000

	if errs := presetScope.buildIncludeList(strings.Join(presetValues[1:], annotationNestedSeperator)); len(errs) > 0 {
		err = fmt.Errorf("Invalid preset values. %s", errs)
		return
	}

	if presetScope.IncludedScopes[mStruct] == nil {
		err = fmt.Errorf("The model: '%s' collection is not preset within the scope.", modelType.Name())
		return
	}

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

			_, errs := filterScope.buildFilterfield(filterModel.collectionType, filterValues, filterModel, parameters[1:]...)
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
		}
	}
	presetScope.copyPresetParameters()
	return

}

func (c *Controller) BuildScopeList(req *http.Request, model interface{},
) (scope *Scope, errs []*ErrorObject, err error) {
	// Get ModelStruct
	var mStruct *ModelStruct
	mStruct, err = c.getModelStruct(model)
	if err != nil {
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

	scope.IncludedScopes = make(map[*ModelStruct]*Scope)

	scope.maxNestedLevel = c.IncludeNestedLimit
	scope.collectionScope = scope
	scope.IsMany = true

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

			_, errorObjects = filterScope.buildFilterfield(collection, splitValues, colModel, splitted[1:]...)
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

	// Copy the filters for the included fields
	scope.copyIncludedBoundaries()

	return
}

// BuildScopeSingle builds the scope for given request and model.
// It gets and sets the ID from the 'http' request.
func (c *Controller) BuildScopeSingle(req *http.Request, model interface{},
) (scope *Scope, errs []*ErrorObject, err error) {
	// get model type
	// Get ModelStruct
	var mStruct *ModelStruct
	mStruct, err = c.getModelStruct(model)
	if err != nil {
		return
	}
	var (
		addErrors = func(errObjects ...*ErrorObject) {
			errs = append(errs, errObjects...)
			scope.currentErrorCount += len(errObjects)
		}
		errObj       *ErrorObject
		errorObjects []*ErrorObject
	)

	q := req.URL.Query()

	scope = newRootScope(mStruct)
	errs, err = c.setIDFilter(req, scope)
	if err != nil {
		return
	}
	if len(errs) > 0 {
		return
	}

	scope.maxNestedLevel = c.IncludeNestedLimit

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

			_, errorObjects = filterScope.buildFilterfield(collection, splitValues, colModel, splitted[1:]...)
			addErrors(errorObjects...)
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

// BuildScopeRelated builds the scope for the related
func (c *Controller) BuildScopeRelated(req *http.Request, root interface{},
) (scope *Scope, errs []*ErrorObject, err error) {
	var mStruct *ModelStruct
	mStruct, err = c.getModelStruct(root)
	if err != nil {
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
		kind: rootKind,
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

	errs = scope.setPrimaryFilterfield(id)
	if len(errs) > 0 {
		return
	}

	scope.Fieldset[related] = relationField
	scope.IncludedScopes = make(map[*ModelStruct]*Scope)

	// preset related scope
	includedField := scope.getOrCreateIncludeField(relationField)
	includedField.Scope.kind = relatedKind

	scope.copyIncludedBoundaries()

	return
}

func (c *Controller) BuildScopeRelationship(req *http.Request, root interface{},
) (scope *Scope, errs []*ErrorObject, err error) {
	var mStruct *ModelStruct
	mStruct, err = c.getModelStruct(root)
	if err != nil {
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
		kind: rootKind,
	}
	scope.collectionScope = scope

	// set primary field filter
	errs = scope.setPrimaryFilterfield(id)
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
	scope.IncludedScopes = make(map[*ModelStruct]*Scope)
	scope.createModelsRootScope(relationField.relatedStruct)
	scope.IncludedScopes[relationField.relatedStruct].Fieldset = nil

	// preset relationship scope
	includedField := scope.getOrCreateIncludeField(relationField)
	includedField.Scope.kind = relationshipKind

	scope.copyIncludedBoundaries()

	return
}

// GetModelStruct returns the ModelStruct for provided model
// Returns error if provided model does not exists in the PrecomputedMap
func (c *Controller) GetModelStruct(model interface{}) (*ModelStruct, error) {
	return c.getModelStruct(model)
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
	return scope, nil
}

// PrecomputeModels precomputes provided models, making it easy to check
// models relationships and  attributes.
func (c *Controller) PrecomputeModels(models ...interface{}) error {
	var err error
	if c.Models == nil {
		c.Models = newModelMap()
	}

	for _, model := range models {
		err = buildModelStruct(model, c.Models)
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

	return nil
}

func (c *Controller) SetAPIURL(url string) error {
	// manage the url
	c.APIURLBase = url
	return nil
}

// GetAndSetIDFilter is the method that gets the ID value from the request path, for given scope's
// model. Then it sets the id filters for given scope.
// If the url is constructed incorrectly it returns an internal error.
func (c *Controller) GetAndSetIDFilter(req *http.Request, scope *Scope) error {
	id, err := getID(req, scope.Struct)
	if err != nil {
		return err
	}
	scope.setIDFilterValues(id)
	return nil
}

// GetCheckSetIDFilter the method that gets the ID value from the request path, for given scope
// model.then prepares the primary filter field for given id value.
// if an internal error occurs returns an 'error'.
// if user error occurs returns array of *ErrorObject's.
func (c *Controller) GetSetCheckIDFilter(req *http.Request, scope *Scope,
) (errs []*ErrorObject, err error) {
	return c.setIDFilter(req, scope)
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

func (c *Controller) setIDFilter(req *http.Request, scope *Scope,
) (errs []*ErrorObject, err error) {
	var id string
	id, err = getID(req, scope.Struct)
	if err != nil {
		return
	}
	errs = scope.setPrimaryFilterfield(id)
	if len(errs) > 0 {
		return
	}
	return
}
