package controller

import (
	"github.com/kucjac/jsonapi/pkg/config"
	"github.com/kucjac/jsonapi/pkg/db-manager"
	"github.com/kucjac/jsonapi/pkg/flags"
	"github.com/kucjac/jsonapi/pkg/i18n"
	"github.com/kucjac/jsonapi/pkg/internal/models"
	"github.com/kucjac/jsonapi/pkg/internal/query"
	"github.com/kucjac/jsonapi/pkg/internal/query/filters"
	"github.com/kucjac/jsonapi/pkg/internal/repositories"
	"github.com/kucjac/jsonapi/pkg/log"

	"github.com/pkg/errors"

	"github.com/kucjac/jsonapi/pkg/internal/namer"
	"github.com/kucjac/uni-logger"

	// "golang.org/x/text/language"
	// "golang.org/x/text/language/display"
	"gopkg.in/go-playground/validator.v9"
	// "net/http"
	// "net/url"

	// "strconv"
	"strings"
)

var (
	validate          *validator.Validate = validator.New()
	defaultController *Controller
)

// Controller is the data structure that is responsible for controlling all the models
// within single API
type Controller struct {
	// Config is the configuration struct for the controller
	Config *config.ControllerConfig

	// Namer defines the function strategy how the model's and it's fields are being named
	NamerFunc namer.Namer

	// Flags defines the controller config flags
	Flags *flags.Container

	// StrictUnmarshalMode if set to true, the incoming data cannot contain
	// any unknown fields
	StrictUnmarshalMode bool

	// queryBuilderis the controllers query builder
	queryBuilder *query.Builder

	// i18nSup defines the i18n support for the provided controller
	i18nSup *i18n.Support

	// schemas is a mapping for the model schemas
	schemas *models.ModelSchemas

	// repositories contains mapping between the model's and it's repositories
	repositories *repositories.RepositoryContainer

	// operators
	operators *filters.OperatorContainer

	// errMgr error manager for the repositories
	errMgr *dbmanager.ErrorManager

	// Validators
	// CreateValidator is used as a validator for the Create processes
	CreateValidator *validator.Validate

	//PatchValidator is used as a validator for the Patch processes
	PatchValidator *validator.Validate
}

// New Creates raw *jsonapi.Controller with no limits and links.
func New(cfg *config.ControllerConfig, logger unilogger.LeveledLogger) (*Controller, error) {

	c, err := newController(cfg)
	if err != nil {
		return nil, err
	}

	if logger != nil {
		log.SetLogger(logger)
	}

	return c, nil
}

func SetDefault(c *Controller) {
	defaultController = c
}

// Default creates new *jsonapi.Controller with preset limits:
// Controller has also set the FlagUseLinks flag to true.
func Default() *Controller {
	if defaultController == nil {
		c, err := newController(DefaultConfig)
		if err != nil {
			panic(err)
		}

		defaultController = c
	}

	return defaultController
}

func newController(cfg *config.ControllerConfig) (*Controller, error) {
	c := &Controller{
		operators:       filters.NewOpContainer(),
		Flags:           flags.New(),
		CreateValidator: validator.New(),
		PatchValidator:  validator.New(),
	}

	err := c.setConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "setConfig failed.")
	}
	if cfg.I18n != nil {
		c.i18nSup, err = i18n.New(cfg.I18n)
		if err != nil {
			return nil, errors.Wrap(err, "i18n.New failed.")
		}
	}

	// create model schemas
	c.schemas, err = models.NewModelSchemas(
		c.NamerFunc,
		c.Config.Builder.IncludeNestedLimit,
		c.Config.ModelSchemas,
		c.Config.DefaultSchema,
		c.Config.DefaultRepository,
		c.Flags,
	)
	if err != nil {
		return nil, err
	}

	// create repository container
	c.repositories = repositories.NewRepoContainer()

	c.queryBuilder, err = query.NewBuilder(c.schemas, c.Config.Builder, c.operators, c.i18nSup)
	if err != nil {
		return nil, errors.Wrap(err, "query.NewBuilder failed")
	}

	// create error manager
	c.errMgr = dbmanager.NewDBErrorMgr()

	return c, nil
}

// DBManager gets the database error manager
func (c *Controller) DBManager() *dbmanager.ErrorManager {
	return c.errMgr
}

// QueryBuilder returns the controller query builder
func (c *Controller) QueryBuilder() *query.Builder {
	return c.queryBuilder
}

// SetLogger sets the logger for the controller operations
func (c *Controller) SetLogger(logger unilogger.LeveledLogger) {
	log.SetLogger(logger)
}

// MustGetModelStruct gets (concurrently safe) the model struct from the cached model Map
// panics if the model does not exists in the map.
func (c *Controller) MustGetModelStruct(model interface{}) *models.ModelStruct {
	mStruct, err := c.getModelStruct(model)
	if err != nil {
		panic(err)
	}
	return mStruct
}

// // NewScope creates new scope for given model
// func (c *Controller) NewScope(model interface{}) (*scope.Scope, error) {
// 	mStruct, err := c.GetModelStruct(model)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sc := scope.New(mStruct, c.Config.IncludeNestedLimit)

// 	return sc, nil
// }

// func (c *Controller) NewRootScope(model interface{}) (*scope.Scope, error) {
// 	mStruct, err := c.GetModelStruct(model)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sc := scope.NewRootScope(mStruct, c.Config.IncludeNestedLimit)

// 	return sc, nil
// }

// RegisterRepositories registers multiple repositories.
// Returns error if the repository were already registered
func (c *Controller) RegisterRepositories(repos ...repositories.Repository) error {
	for _, repo := range repos {
		if err := c.repositories.RegisterRepository(repo); err != nil {
			log.Error("RegisterRepository '%s' failed. %v", repo.RepositoryName(), err)
			return err
		}
	}
	return nil
}

// RegisterRepository registers the repository
func (c *Controller) RegisterRepository(repo repositories.Repository) error {
	return c.repositories.RegisterRepository(repo)
}

// RegisterModels precomputes provided models, making it easy to check
// models relationships and  attributes.
func (c *Controller) RegisterModels(models ...interface{}) error {
	if err := c.schemas.RegisterModels(models...); err != nil {
		return err
	}

	for _, schema := range c.schemas.Schemas() {
		for _, mStruct := range schema.Models() {
			if err := c.repositories.MapModel(mStruct); err != nil {
				log.Errorf("Mapping model: %v to repository failed.", mStruct.Type().Name())
				return err
			}
		}
	}

	return nil
}

// RepositoryByName returns the repository by the provided name.
// If the repository doesn't exists it returns nil value and false boolean
func (c *Controller) RepositoryByName(name string) (repositories.Repository, bool) {
	return c.repositories.RepositoryByName(name)
}

// RepositoryByModel returns the repository for the provided model.
// If the repository doesn't exists it returns 'nil' value and 'false' boolean.
func (c *Controller) RepositoryByModel(model *models.ModelStruct) (repositories.Repository, bool) {
	return c.repositories.RepositoryByModel(model)
}

// GetModelStruct returns the ModelStruct for provided model
// Returns error if provided model does not exists in the PrecomputedMap
func (c *Controller) GetModelStruct(model interface{}) (*models.ModelStruct, error) {
	return c.getModelStruct(model)
}

// NewFilterField creates new filter field based on the fieldFilter argument and provided values
//	Arguments:
//		@fieldFilter - jsonapi field name for provided model the field type must be of the same as
//				the last element of the preset. By default the filter operator is of 'in' type.
//				It must be of form: filter[collection][field]([operator]|([subfield][operator])).
//				The operator or subfield are not required.
//		@values - the values of type matching the filter field
func (c *Controller) NewFilterField(fieldFilter string, values ...interface{},
) (filter *filters.FilterField, err error) {

	// if valid := strings.HasPrefix(fieldFilter, internal.QueryParamFilter); !valid {
	// 	err = fmt.Errorf("Invalid field filter argument provided: '%s'. The argument should be composed as: 'filter[collection][field]([subfield][operator])|([operator]).", fieldFilter)
	// 	return
	// }

	// var splitted []string
	// splitted, err = internal.SplitBracketParameter(fieldFilter[len(QueryParamFilter):])
	// if err != nil {
	// 	return
	// }

	// mStruct := c.Models.GetByCollection(splitted[0])
	// if mStruct == nil {
	// 	err = fmt.Errorf("The model for collection: '%s' is not precomputed within controller. Cannot clreate new filter field.", splitted[0])
	// 	return
	// }

	// var (
	// 	structField *mapping.StructField
	// 	operator    query.Operator
	// )
	// handle3and4 := func() {
	// 	structField = mStruct.relationships[splitted[1]]
	// 	if structField == nil {
	// 		err = fmt.Errorf("Invalid field name provided in fieldFilter: '%s'.", fieldFilter)
	// 		return
	// 	}
	// 	filter = &FilterField{StructField: structField}

	// 	if splitted[2] == "id" {
	// 		structField = filter.relatedStruct.primary
	// 	} else {
	// 		if structField = filter.relatedStruct.attributes[splitted[2]]; structField == nil {
	// 			err = fmt.Errorf("Invalid subfield name provided: '%s' for the query: '%s'.", splitted[2], fieldFilter)
	// 			return
	// 		}
	// 	}
	// 	subfieldFilter := &FilterField{StructField: structField}

	// 	if len(splitted) == 4 {
	// 		var ok bool
	// 		operator, ok = operatorsValue[splitted[3]]
	// 		if !ok {
	// 			err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[3], fieldFilter)
	// 			return
	// 		}
	// 	} else {
	// 		operator = OpIn
	// 	}

	// 	fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}
	// 	for i, value := range values {
	// 		t := reflect.TypeOf(value)
	// 		if t == subfieldFilter.refStruct.Type {
	// 			fv.Values[i] = value
	// 		} else {
	// 			err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, subfieldFilter.refStruct.Type)
	// 			return
	// 		}
	// 	}

	// 	subfieldFilter.Values = append(subfieldFilter.Values, fv)

	// 	// Add subfield
	// 	filter.Nested = append(filter.Nested, subfieldFilter)
	// }

	// handle2and3 := func() {
	// 	if splitted[1] == "id" {
	// 		structField = mStruct.primary
	// 	} else {
	// 		structField = mStruct.attributes[splitted[1]]
	// 		if structField == nil {
	// 			if structField = mStruct.relationships[splitted[1]]; structField != nil {
	// 				if len(splitted) == 3 {
	// 					handle3and4()
	// 				} else {
	// 					err = fmt.Errorf("The relationship field: '%s' in the filter argument must specify the subfield. Filter: '%s'", splitted[1], fieldFilter)
	// 				}
	// 			} else {
	// 				err = fmt.Errorf("Invalid field name: '%s' for the collection: '%s'.", splitted[1], splitted[0])
	// 			}
	// 			return
	// 		}
	// 	}
	// 	filter = &FilterField{StructField: structField}
	// 	var ok bool
	// 	if len(splitted) == 3 {
	// 		operator, ok = operatorsValue[splitted[2]]
	// 		if !ok {
	// 			err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[2], fieldFilter)
	// 			return
	// 		}
	// 	} else {
	// 		operator = OpIn
	// 	}
	// 	fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}

	// 	for i, value := range values {
	// 		t := reflect.TypeOf(value)
	// 		if t == filter.refStruct.Type {
	// 			fv.Values[i] = value
	// 		} else {
	// 			err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, filter.refStruct.Type)
	// 			return
	// 		}
	// 	}
	// 	filter.Values = append(filter.Values, fv)
	// }

	// switch len(splitted) {
	// case 2, 3:
	// 	handle2and3()
	// case 4:
	// 	handle3and4()
	// default:
	// 	err = fmt.Errorf("The filter argument: '%s' is of invalid form.", fieldFilter)
	// 	return
	// }
	return
}

func (c *Controller) getModelStruct(model interface{}) (*models.ModelStruct, error) {
	if model == nil {
		return nil, errors.New("Nil model provided.")
	}

	mStruct, err := c.schemas.GetModelStruct(model)
	if err != nil {
		return nil, err
	}

	return mStruct, nil
}

// setConfig sets and validates provided config
func (c *Controller) setConfig(cfg *config.ControllerConfig) error {
	if cfg == nil {
		return errors.New("Nil config provided")
	}

	cfg.NamingConvention = strings.ToLower(cfg.NamingConvention)

	if err := validate.Struct(cfg); err != nil {
		return errors.Wrap(err, "Validate config failed.")
	}

	c.Config = cfg

	// set naming convention
	switch cfg.NamingConvention {
	case "kebab":
		c.NamerFunc = namer.NamingKebab
	case "camel":
		c.NamerFunc = namer.NamingCamel
	case "lowercamel":
		c.NamerFunc = namer.NamingLowerCamel
	case "snake":
		c.NamerFunc = namer.NamingSnake
	}

	if cfg.DefaultSchema == "" {
		cfg.DefaultSchema = "api"
	}

	if cfg.CreateValidatorAlias == "" {
		cfg.CreateValidatorAlias = "create"
	}
	c.CreateValidator.SetTagName(cfg.CreateValidatorAlias)

	if cfg.PatchValidatorAlias == "" {
		cfg.PatchValidatorAlias = "patch"
	}
	c.PatchValidator.SetTagName(cfg.PatchValidatorAlias)

	return nil
}

// // GetAndSetIDFilter is the method that gets the ID value from the request path, for given scope's
// // model. Then it sets the id query.for given scope.
// // If the url is constructed incorrectly it returns an internal error.
// func (c *Controller) GetAndSetIDFilter(req *http.Request, scope *scope.Scope) error {
// 	c.log().Debug("GetAndSetIDFIlter")

// 	id, err := getID(req, scope.Struct)
// 	if err != nil {
// 		return err
// 	}

// 	scope.setIDFilterValues(id)

// 	return nil
// }

// func (c *Controller) GetAndSetID(req *http.Request, scope *scope.Scope) (prim interface{}, err error) {
// 	id, err := getID(req, scope.Struct)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if scope.Value == nil {
// 		return nil, IErrNoValue
// 	}

// 	val := reflect.ValueOf(scope.Value)
// 	if val.IsNil() {
// 		return nil, IErrScopeNoValue
// 	}

// 	if val.Kind() == reflect.Ptr {
// 		val = val.Elem()
// 	} else {
// 		return nil, IErrInvalidType
// 	}

// 	c.log().Debugf("Setting newPrim value")
// 	newPrim := reflect.New(scope.Struct.primary.refStruct.Type).Elem()

// 	err = setPrimaryField(id, newPrim)
// 	if err != nil {
// 		errObj := ErrInvalidQueryParameter.Copy()
// 		errObj.Detail = "Provided invalid id value within the url."
// 		return nil, errObj
// 	}

// 	primVal := val.FieldByIndex(scope.Struct.primary.refStruct.Index)

// 	if reflect.DeepEqual(primVal.Interface(), reflect.Zero(scope.Struct.primary.refStruct.Type).Interface()) {
// 		primVal.Set(newPrim)
// 	} else {
// 		c.log().Debugf("Checking the values")
// 		if newPrim.Interface() != primVal.Interface() {
// 			errObj := ErrInvalidQueryParameter.Copy()
// 			errObj.Detail = "Provided invalid id value within the url. The id value doesn't match the primary field within the root object."
// 			return nil, errObj
// 		}
// 	}

// 	v := reflect.ValueOf(scope.Value)
// 	if v.Kind() != reflect.Ptr {
// 		return nil, IErrInvalidType
// 	}
// 	return primVal.Interface(), nil
// }

// // GetCheckSetIDFilter the method that gets the ID value from the request path, for given scope
// // model.then prepares the primary filter field for given id value.
// // if an internal error occurs returns an 'error'.
// // if user error occurs returns array of *aerrors.ApiError's.
// func (c *Controller) GetSetCheckIDFilter(req *http.Request, scope *scope.Scope,
// ) (errs []*aerrors.ApiError, err error) {
// 	return c.setIDFilter(req, scope)
// }

// func (c *Controller) NewFilterFieldWithForeigns(fieldFilter string, values ...interface{},
// ) (filter *filters.FilterField, err error) {

// 	if valid := strings.HasPrefix(fieldFilter, internal.QueryParamFilter); !valid {
// 		err = fmt.Errorf("Invalid field filter argument provided: '%s'. The argument should be composed as: 'filter[collection][field]([subfield][operator])|([operator]).", fieldFilter)
// 		return
// 	}

// 	var splitted []string
// 	splitted, err = internal.SplitBracketParameter(fieldFilter[len(QueryParamFilter):])
// 	if err != nil {
// 		return
// 	}

// 	mStruct := c.Models.GetByCollection(splitted[0])
// 	if mStruct == nil {
// 		err = fmt.Errorf("The model for collection: '%s' is not precomputed within controller. Cannot clreate new filter field.", splitted[0])
// 		return
// 	}

// 	var (
// 		structField *mapping.StructField
// 		operator    *scope.Operator
// 	)
// 	handle3and4 := func() {
// 		structField = mStruct.relationships[splitted[1]]
// 		if structField == nil {
// 			err = fmt.Errorf("Invalid field name provided in fieldFilter: '%s'.", fieldFilter)
// 			return
// 		}
// 		filter = &FilterField{StructField: structField}

// 		if splitted[2] == "id" {
// 			structField = filter.relatedStruct.primary
// 		} else {
// 			if structField = filter.relatedStruct.attributes[splitted[2]]; structField == nil {
// 				if structField = filter.relatedStruct.foreignKeys[splitted[2]]; structField == nil {
// 					err = fmt.Errorf("Invalid subfield name provided: '%s' for the query: '%s'.", splitted[2], fieldFilter)
// 					return
// 				}
// 			}
// 		}
// 		subfieldFilter := &FilterField{StructField: structField}

// 		if len(splitted) == 4 {
// 			var ok bool
// 			operator, ok = c.Operators[splitted[3]]
// 			if !ok {
// 				err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[3], fieldFilter)
// 				return
// 			}
// 		} else {
// 			operator = OpIn
// 		}

// 		fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}
// 		for i, value := range values {
// 			t := reflect.TypeOf(value)
// 			if t == subfieldFilter.refStruct.Type {
// 				fv.Values[i] = value
// 			} else {
// 				err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, subfieldFilter.refStruct.Type)
// 				return
// 			}
// 		}

// 		subfieldFilter.Values = append(subfieldFilter.Values, fv)

// 		// Add subfield
// 		filter.Nested = append(filter.Nested, subfieldFilter)
// 	}

// 	handle2and3 := func() {
// 		if splitted[1] == "id" {
// 			structField = mStruct.primary
// 		} else {
// 			structField = mStruct.attributes[splitted[1]]
// 			if structField == nil {
// 				structField = mStruct.foreignKeys[splitted[1]]
// 				if structField == nil {
// 					if structField = mStruct.relationships[splitted[1]]; structField != nil {
// 						if len(splitted) == 3 {
// 							handle3and4()
// 						} else {
// 							err = fmt.Errorf("The relationship field: '%s' in the filter argument must specify the subfield. Filter: '%s'", splitted[1], fieldFilter)
// 						}
// 					} else {
// 						err = fmt.Errorf("Invalid field name: '%s' for the collection: '%s'.", splitted[1], splitted[0])
// 					}
// 					return
// 				}

// 			}
// 		}
// 		filter = &FilterField{StructField: structField}
// 		var ok bool
// 		if len(splitted) == 3 {
// 			operator, ok = operatorsValue[splitted[2]]
// 			if !ok {
// 				err = fmt.Errorf("Invalid operator provided: '%s' for the field filter: '%s'.", splitted[2], fieldFilter)
// 				return
// 			}
// 		} else {
// 			operator = OpIn
// 		}
// 		fv := &FilterValues{Operator: operator, Values: make([]interface{}, len(values))}

// 		for i, value := range values {
// 			t := reflect.TypeOf(value)
// 			if t == filter.refStruct.Type {
// 				fv.Values[i] = value
// 			} else {
// 				err = fmt.Errorf("Invalid value type provided for filter: '%s', '%v' and should be: '%s'", fieldFilter, t, filter.refStruct.Type)
// 				return
// 			}
// 		}
// 		filter.Values = append(filter.Values, fv)
// 	}

// 	switch len(splitted) {
// 	case 2, 3:
// 		handle2and3()
// 	case 4:
// 		handle3and4()
// 	default:
// 		err = fmt.Errorf("The filter argument: '%s' is of invalid form.", fieldFilter)
// 		return
// 	}
// 	return
// }

// func (c *Controller) buildPreparedPair(
// 	query, fieldFilter string,
// 	check bool,
// ) (presetScope *scope.Scope, filter *scope.FilterField) {
// 	var err error
// 	defer func() {
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()

// 	filter, err = c.NewFilterField(fieldFilter)
// 	if err != nil {
// 		err = fmt.Errorf("Invalid field filter provided for presetPair: '%s'. %v", fieldFilter, err)
// 		return
// 	}

// 	// mStruct := filter.mStruct

// 	// Parse the query
// 	var queryParsed url.Values
// 	queryParsed, err = url.ParseQuery(query)
// 	if err != nil {
// 		return
// 	}

// 	preset := queryParsed.Get(QueryParamPreset)
// 	if preset == "" {
// 		err = fmt.Errorf("Invalid query: '%s' . No preset parameter", query)
// 		return
// 	}

// 	presets := strings.Split(preset, annotationSeperator)
// 	if len(presets) != 1 {
// 		err = fmt.Errorf("Multiple presets are not allowed. Query: '%s'", query)
// 		return
// 	}

// 	presetValues := strings.Split(presets[0], annotationNestedSeperator)
// 	// if len(presetValues) <= 1 {
// 	// 	err = fmt.Errorf("Preset scope should contain at least two parameter values within its preset. Query: '%s'", query)
// 	// 	return
// 	// }

// 	rootModel := c.Models.GetByCollection(presetValues[0])
// 	if rootModel == nil {
// 		err = fmt.Errorf("Invalid query parameter: '%s'. No model found for the collection: '%s'. ", query, presetValues[0])
// 		return
// 	}

// 	presetScope = newRootScope(rootModel)
// 	presetScope.maxNestedLevel = 10000
// 	presetScope.Fieldset = nil

// 	if len(presetValues) > 1 {
// 		if errs := presetScope.buildIncludeList(strings.Join(presetValues[1:], annotationNestedSeperator)); len(errs) > 0 {
// 			err = fmt.Errorf("Invalid preset values. %s", errs)
// 			return
// 		}
// 	}

// 	var selectIncludedFieldset func(scope *scope.Scope) error

// 	selectIncludedFieldset = func(scope *scope.Scope) error {
// 		scope.Fieldset = make(map[string]*mapping.StructField)
// 		for scope.NextIncludedField() {
// 			included, err := scope.CurrentIncludedField()
// 			if err != nil {
// 				return err
// 			}
// 			scope.Fieldset[included.jsonAPIName] = included.StructField
// 			if err = selectIncludedFieldset(included.Scope); err != nil {
// 				return err
// 			}
// 		}
// 		scope.ResetIncludedField()
// 		return nil
// 	}

// 	if err = selectIncludedFieldset(presetScope); err != nil {
// 		return
// 	}

// 	// var fieldsetFound bool
// 	for key, value := range queryParsed {
// 		if strings.HasPrefix(key, QueryParamFields) {
// 			var splitted []string
// 			splitted, err = splitBracketParameter(key[len(QueryParamFields):])
// 			if err != nil {
// 				errObj := ErrInvalidQueryParameter.Copy()
// 				errObj.Detail = fmt.Sprintf("The fields parameter is of invalid form. %s", err)
// 				err = errObj
// 				return
// 			}

// 			if len(splitted) != 1 {
// 				errObj := ErrInvalidQueryParameter.Copy()
// 				errObj.Detail = fmt.Sprintf("The fields parameter: '%s' is of invalid form. Nested 'fields' is not supported.", key)
// 				err = errObj
// 				return
// 			}

// 			collection := splitted[0]
// 			fieldModel := c.Models.GetByCollection(collection)
// 			if fieldModel == nil {
// 				err = fmt.Errorf("Collection: '%s' not found within controller.", collection)
// 				return
// 			}

// 			fieldScope := presetScope.getModelsRootScope(fieldModel)
// 			if fieldScope == nil {
// 				err = fmt.Errorf("Collection: '%s' not included into preset. For query: '%s'.", collection, query)
// 				return
// 			}

// 			splitFields := strings.Split(value[0], annotationSeperator)
// 			errs := fieldScope.buildFieldset(splitFields...)
// 			if len(errs) != 0 {
// 				err = fmt.Errorf("Invalid fieldset provided. %v", errs)
// 				return
// 			}
// 			// fieldsetFound = true
// 		}
// 	}

// 	// if !fieldsetFound && check {
// 	// 	err = fmt.Errorf("The fieldset not found within the query: '%s'.", query)
// 	// 	return
// 	// }

// 	for key, value := range queryParsed {
// 		switch {
// 		case key == QueryParamPreset:
// 			// Preset are already defined
// 			continue
// 		case strings.HasPrefix(key, QueryParamFilter):
// 			// Filter
// 			var parameters []string
// 			parameters, err = splitBracketParameter(key[len(QueryParamPreset):])
// 			if err != nil {
// 				return
// 			}

// 			filterModel := c.Models.GetByCollection(parameters[0])
// 			if filterModel == nil {
// 				err = fmt.Errorf("Invalid collection: '%s' for filter parameter within query: %s.", parameters[0], query)
// 				return
// 			}

// 			filterScope := presetScope.getModelsRootScope(filterModel)
// 			if filterScope == nil {
// 				err = fmt.Errorf("Invalid collection: '%s' for filter parameter. The collection is not included into preset parameter: '%s' within query: '%s' ", filterModel.collectionType, preset, query)
// 				return
// 			}

// 			filterValues := strings.Split(value[0], annotationSeperator)

// 			_, errs := buildFilterField(filterScope, filterModel.collectionType,
// 				filterValues, c, filterModel, c.Flags, parameters[1:]...)
// 			// _, errs := filterScope.buildFilterfield(filterModel.collectionType, filterValues, filterModel, parameters[1:]...)
// 			if len(errs) != 0 {
// 				err = fmt.Errorf("Error while building filter field for collection: '%s' in query: '%s'. Errs: '%s'", filterModel.collectionType, query, errs)
// 				return
// 			}

// 		case strings.HasPrefix(key, QueryParamSort):
// 			// Sort given collection
// 			var parameters []string
// 			parameters, err = splitBracketParameter(key[len(QueryParamSort):])
// 			if err != nil {
// 				return
// 			}
// 			if len(parameters) > 1 {
// 				err = fmt.Errorf("Too many sort field parameters within brackets. '%s'", parameters)
// 				return
// 			}

// 			var sortScope *scope.Scope
// 			if parameters[0] == presetScope.Struct.collectionType {
// 				sortScope = presetScope
// 			} else {
// 				sortModel := c.Models.GetByCollection(parameters[0])
// 				if sortModel == nil {
// 					err = fmt.Errorf("Invalid sort collection provided: '%s'.", parameters[0])
// 					return
// 				}

// 				sortScope = presetScope.IncludedScopes[sortModel]
// 				if sortScope == nil {
// 					err = fmt.Errorf("The sort collection '%s' is not within the preset parameter in query: '%s'", parameters[0], query)
// 					return
// 				}
// 			}

// 			var order Order
// 			sortField := value[0]
// 			if sortField[0] == '-' {
// 				order = DescendingOrder
// 				sortField = sortField[1:]
// 			}

// 			if invalidField := newSortField(sortField, order, sortScope); invalidField {
// 				err = fmt.Errorf("Provided invalid sort field: '%s' for collection: '%s' in query: '%s'.", sortField, sortScope.Struct.collectionType, query)
// 				return
// 			}
// 		case strings.HasPrefix(key, QueryParamPage):
// 			// Limit given collection
// 			var parameters []string

// 			parameters, err = splitBracketParameter(key[len(QueryParamPage):])
// 			if err != nil {
// 				return
// 			}
// 			if len(parameters) != 2 {
// 				err = fmt.Errorf("Too many bracket parameters in page[limit] parameter within query: '%s'. The parameter takes only limit and collection bracket parameters.", query)
// 				return
// 			}

// 			var paginateType PaginationParameter
// 			switch parameters[0] {
// 			case "limit":
// 				paginateType = limitPaginateParam
// 			case "offset":
// 				paginateType = offsetPaginateParam
// 			case "number":
// 				paginateType = numberPaginateParam
// 			case "size":
// 				paginateType = sizePaginateParam
// 			default:
// 				err = fmt.Errorf("Invalid paginate type parameter: '%s' within query: '%s'", parameters[0], query)
// 				return
// 			}

// 			pageModel := c.Models.GetByCollection(parameters[1])
// 			if pageModel == nil {
// 				err = fmt.Errorf("Provided invalid collection for pagination: '%s' within query: '%s'", parameters[1], query)
// 				return
// 			}

// 			pageScope := presetScope.getModelsRootScope(pageModel)
// 			if pageScope == nil {
// 				err = fmt.Errorf("Provided collection: '%s' within query: '%s' is not included into preset parameter.", parameters[1], query)
// 				return
// 			}

// 			if errObj := pageScope.preparePaginatedValue(key+"["+parameters[0]+"]", value[0], int(paginateType)); errObj != nil {
// 				err = errObj
// 				return
// 			}
// 		case strings.HasPrefix(key, QueryParamFields):
// 			continue
// 		}
// 	}
// 	presetScope.copyPresetParameters()

// 	// If the prepared scope is not checked (preset)
// 	// Then check if the value range is consistent
// 	if !check {
// 		/**

// 		TO DO:

// 		*/

// 	}

// 	return
// }

// func (c *Controller) checkModelRelationships(model *mapping.ModelStruct) (err error) {
// 	for _, rel := range model.relationships {
// 		val := c.Models.Get(rel.relatedModelType)
// 		if val == nil {
// 			err = fmt.Errorf("Model: %v, not precalculated but is used in relationships for: %v field in %v model.", rel.relatedModelType, rel.fieldName, model.modelType.Name())
// 			return err
// 		}
// 		rel.relatedStruct = val
// 	}
// 	return
// }

// func (c *Controller) registerOperators()

// func (c *Controller) setIDFilter(req *http.Request, scope *scope.Scope,
// ) (errs []*aerrors.ApiError, err error) {
// 	var id string
// 	id, err = getID(req, scope.Struct)
// 	if err != nil {
// 		return
// 	}
// 	errs = scope.setPrimaryFilterfield(c, id)
// 	if len(errs) > 0 {
// 		return
// 	}
// 	return
// }

// func (c *Controller) setFilterValues(
// 	f *scope.FilterField,
// 	collection string,
// 	values []string,
// 	op query.Operator,
// ) (errs []*aerrors.ApiError) {
// 	var (
// 		er     error
// 		errObj *aerrors.ApiError

// 		opInvalid = func() {
// 			errObj = ErrUnsupportedQueryParameter.Copy()
// 			errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported for the field: '%s'.", op, f.jsonAPIName)
// 			errs = append(errs, errObj)
// 		}
// 	)

// 	if op > OpEndsWith {
// 		errObj = ErrInvalidQueryParameter.Copy()
// 		errObj.Detail = fmt.Sprintf("The filter operator: '%s' is not supported by the server.", op)
// 	}

// 	t := f.getDereferencedType()
// 	// create new FilterValue
// 	fv := new(FilterValues)
// 	fv.Operator = op

// 	// Add and check all values for given field type
// 	switch f.fieldType {
// 	case Primary:
// 		switch t.Kind() {
// 		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
// 			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 			if op > OpLessEqual {
// 				opInvalid()
// 			}
// 		case reflect.String:
// 			if !op.isBasic() {
// 				opInvalid()
// 			}
// 		}
// 		for _, value := range values {
// 			fieldValue := reflect.New(t).Elem()
// 			er = setPrimaryField(value, fieldValue)
// 			if er != nil {
// 				errObj = ErrInvalidQueryParameter.Copy()
// 				errObj.Detail = fmt.Sprintf("Invalid filter value for primary field in collection: '%s'. %s. ", collection, er)
// 				errs = append(errs, errObj)
// 			}
// 			fv.Values = append(fv.Values, fieldValue.Interface())
// 		}

// 		f.Values = append(f.Values, fv)

// 		// if it is of integer type check which kind of it
// 	case Attribute:
// 		switch t.Kind() {
// 		case reflect.String:
// 		default:
// 			if op.isStringOnly() {
// 				opInvalid()
// 			}
// 		}
// 		if f.isLanguage() {
// 			switch op {
// 			case OpIn, OpEqual, OpNotIn, OpNotEqual:
// 				for i, value := range values {
// 					tag, err := language.Parse(value)
// 					if err != nil {
// 						switch v := err.(type) {
// 						case language.ValueError:
// 							errObj := ErrLanguageNotAcceptable.Copy()
// 							errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s', is not a valid language. Cannot recognize subfield: '%s'.", value, f.GetJSONAPIName(),
// 								collection, v.Subtag())
// 							errs = append(errs, errObj)
// 							continue
// 						default:
// 							errObj := ErrInvalidQueryParameter.Copy()
// 							errObj.Detail = fmt.Sprintf("The value: '%v' for the '%s' filter field within the collection '%s' is not syntetatically valid.", value, f.GetJSONAPIName(), collection)
// 							errs = append(errs, errObj)
// 							continue
// 						}
// 					}
// 					if op == OpEqual {
// 						var confidence language.Confidence
// 						tag, _, confidence = c.Matcher.Match(tag)
// 						if confidence <= language.Low {
// 							errObj := ErrLanguageNotAcceptable.Copy()
// 							errObj.Detail = fmt.Sprintf("The value: '%s' for the '%s' filter field within the collection '%s' does not match any supported langauges. The server supports following langauges: %s", value, f.GetJSONAPIName(), collection,
// 								strings.Join(c.displaySupportedLanguages(), annotationSeperator),
// 							)
// 							errs = append(errs, errObj)
// 							return
// 						}
// 					}
// 					b, _ := tag.Base()
// 					values[i] = b.String()

// 				}
// 			default:
// 				errObj := ErrInvalidQueryParameter.Copy()
// 				errObj.Detail = fmt.Sprintf("Provided operator: '%s' for the language field is not acceptable", op.String())
// 				errs = append(errs, errObj)
// 				return
// 			}
// 		}
// 		for _, value := range values {
// 			fieldValue := reflect.New(t).Elem()
// 			er = setAttributeField(value, fieldValue)
// 			if er != nil {
// 				errObj = ErrInvalidQueryParameter.Copy()
// 				errObj.Detail = fmt.Sprintf("Invalid filter value for the attribute field: '%s' for collection: '%s'. %s.", f.jsonAPIName, collection, er)
// 				errs = append(errs, errObj)
// 				continue
// 			}
// 			fv.Values = append(fv.Values, fieldValue.Interface())
// 		}

// 		f.Values = append(f.Values, fv)

// 	}
// 	return
// }

// func (c *Controller) setIncludedLangaugeFilters(
// 	scope *scope.Scope,
// 	languages string,
// ) ([]*aerrors.ApiError, error) {
// 	tags, errObjects := buildLanguageTags(languages)
// 	if len(errObjects) > 0 {
// 		return errObjects, nil
// 	}
// 	tag, _, confidence := c.Matcher.Match(tags...)
// 	if confidence <= language.Low {
// 		// language not supported
// 		errObj := ErrLanguageNotAcceptable.Copy()
// 		errObj.Detail = fmt.Sprintf("Provided languages: '%s' are not supported. This document supports following languages: %s",
// 			languages,
// 			strings.Join(c.displaySupportedLanguages(), ","),
// 		)
// 		return []*aerrors.ApiError{errObj}, nil
// 	}
// 	scope.queryLanguage = tag

// 	var setLanguage func(scope *scope.Scope) error
// 	//setLanguage sets language filter field for all included fields
// 	setLanguage = func(scope *scope.Scope) error {
// 		defer func() {
// 			scope.ResetIncludedField()
// 		}()
// 		if scope.UseI18n() {
// 			base, _ := tag.Base()
// 			scope.SetLanguageFilter(base.String())
// 		}
// 		for scope.NextIncludedField() {
// 			field, err := scope.CurrentIncludedField()
// 			if err != nil {
// 				return err
// 			}
// 			if err = setLanguage(field.Scope); err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	}

// 	if err := setLanguage(scope); err != nil {
// 		return errObjects, err
// 	}
// 	return errObjects, nil
// }

// func (c *Controller) supportI18n() bool {
// 	if c.Locale == nil || c.Matcher == nil {
// 		return false
// 	}

// 	if len(c.Locale.Tags()) == 0 {
// 		return false
// 	}
// 	return true
// }
