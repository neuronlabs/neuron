package jsonapi

import (
	"errors"
	"fmt"
	"net/http"
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

	// IncludeNestedLimit is a maximum value for nested includes (i.e. IncludeNestedLimit = 1
	// allows ?include=posts.comments but does not allow ?include=posts.comments.author)
	IncludeNestedLimit int

	// UseLinks is a flag that defines if the response should contain links objects
	UseLinks bool
}

// New Creates raw *jsonapi.Controller with no limits and links.
func New() *Controller {
	return &Controller{
		APIURLBase:         "/",
		Models:             newModelMap(),
		ErrorLimitMany:     1,
		ErrorLimitSingle:   1,
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
		APIURLBase:         "/api",
		Models:             newModelMap(),
		ErrorLimitMany:     5,
		ErrorLimitSingle:   2,
		IncludeNestedLimit: 1,
		UseLinks:           true,
	}
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
	scope = newScope(mStruct)

	scope.IncludedScopes = make(map[*ModelStruct]*Scope)

	scope.maxNestedLevel = c.IncludeNestedLimit
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

			filterScope := scope.getIncludedScope(colModel)
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

			fieldsScope := scope.getIncludedScope(fieldModel)
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

	return
}

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

	relatedField, ok := mStruct.relationships[related]
	if !ok {
		// invalid query parameter
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid related field name: '%s', for the collection: '%s'", related, mStruct.collectionType)
		errs = append(errs, errObj)
		return
	}
	fmt.Sprintf("%s, %s", id, relatedField.GetFieldName())

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

	fmt.Sprintf("%s", id)

	relatedField, ok := mStruct.relationships[relationship]
	if !ok {
		// invalid query parameter
		errObj := ErrInvalidQueryParameter.Copy()
		errObj.Detail = fmt.Sprintf("Provided invalid relationship field name: '%s', for the collection: '%s'", relationship, mStruct.collectionType)
		errs = append(errs, errObj)
		return
	}

	scope = newScope(mStruct)
	scope.IncludedScopes = make(map[*ModelStruct]*Scope)
	includedScope := scope.createIncludedScope(relatedField.relatedStruct)
	scope.IncludedFields = []*IncludeField{newIncludeField(relatedField, includedScope)}

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
		// id           string
	)
	// id, err = getID(req, mStruct)
	// if err != nil {
	// 	errs = append(errs, ErrInternalError.Copy())
	// 	return
	// }

	q := req.URL.Query()

	scope = newScope(mStruct)
	errs, err = c.setIDFilter(req, scope)
	if err != nil {
		return
	}
	if len(errs) > 0 {
		return
	}

	scope.maxNestedLevel = c.IncludeNestedLimit

	// errorObjects = scope.setPrimaryFilterfield(id)
	// if len(errorObjects) != 0 {
	// 	errObj = ErrInternalError.Copy()
	// 	errs = append(errs, errObj)
	// 	return
	// }

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

			fieldsetScope := scope.getIncludedScope(fieldsetModel)
			if fieldsetScope == nil {
				errObj = ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("The fields parameter collection: '%s' is not included in the query.", collection)
				addErrors(errObj)
				continue
			}
			splitValues := strings.Split(values[0], annotationSeperator)

			errorObjects = fieldsetScope.buildFieldset(splitValues...)
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
	return newScope(mStruct), nil
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

// SetIDFilter the method that gets the ID value from the request path, for given scope model.
// then prepares the primary filter field for given id value.
// if an internal error occurs returns an 'error'.
// if user error occurs returns array of *ErrorObject's.
func (c *Controller) SetIDFilter(req *http.Request, scope *Scope,
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
