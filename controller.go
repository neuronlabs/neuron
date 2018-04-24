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
	Pagination *PaginationScope

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

func (c *Controller) BuildScopeMany(req *http.Request, model interface{},
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

	scope = newRootScope(mStruct, true)

	// Get URLQuery
	q := req.URL.Query()

	// Check first included in order to create subscopes
	included, ok := q[QueryParamInclude]
	if ok {
		// build included scopes
		errorObjects = scope.buildIncludedScopes(included...)
		addErrors(errorObjects...)
		if len(errs) > 0 {
			return
		}
	}
	// id is always present
	scope.Fields = append(scope.Fields, scope.Struct.primary)

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

			_, errorObjects = filterScope.
				newFilterScope(splitted[0], splitValues, filterScope.Struct, splitted[1:]...)
			addErrors(errorObjects...)

		case key == QueryParamSort:
			splitted := strings.Split(value[0], annotationSeperator)

			errorObjects = scope.setSortScopes(splitted...)
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

		if scope.currentErrorCount >= c.ErrorLimitMany {
			return
		}
	}
	if scope.PaginationScope != nil {
		er := scope.PaginationScope.check()
		if er != nil {
			errObj = ErrInvalidQueryParameter.Copy()
			errObj.Detail = fmt.Sprintf("Pagination parameters are not valid. %s", er)
			addErrors(errObj)
		}
	}

	// if none fields were set
	if len(scope.Fields) <= 1 {
		scope.Fields = scope.Struct.fields
	}
	return
}

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
		id           string
	)
	id, err = getID(req, mStruct)
	if err != nil {
		errs = append(errs, ErrInternalError.Copy())
		return
	}

	q := req.URL.Query()

	scope = newRootScope(mStruct, true)

	errorObjects = scope.setPrimaryFilterScope(id)
	if len(errorObjects) != 0 {
		errObj = ErrInternalError.Copy()
		errs = append(errs, errObj)
		return
	}

	// Check first included in order to create subscopes
	included, ok := q[QueryParamInclude]
	if ok {
		// build included scopes
		errorObjects = scope.buildIncludedScopes(included...)
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
