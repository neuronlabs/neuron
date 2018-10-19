package jsonapi

import (
	"context"
	"fmt"
	"github.com/kucjac/uni-db"
	"github.com/kucjac/uni-logger"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"time"
)

var (
	IErrScopeNoValue         = errors.New("No value provided within scope.")
	IErrPresetInvalidScope   = errors.New("Pressetting invalid scope value.")
	IErrPresetNoValues       = errors.New("Preset no values")
	IErrInvalidValueType     = errors.New("Trying to preset values of invalid type.")
	IErrInvalidScopeType     = errors.New("Invalid scope type. Available values are slice of pointers to struct or pointer to struct")
	IErrValueNotValid        = errors.New("Value not valid.")
	IErrModelHandlerNotFound = errors.New("Model Handler not found.")
)

type Handler struct {
	// controller
	Controller *Controller

	// Logger
	log unilogger.LeveledLogger

	// Repositories
	DefaultRepository Repository

	// DBErrMgr database error manager
	DBErrMgr *ErrorManager

	// Supported Languages
	SupportedLanguages []language.Tag

	// LanguageMatcher matches the possible language
	LanguageMatcher language.Matcher

	// Validators validate given
	CreateValidator *validator.Validate
	PatchValidator  *validator.Validate

	// ModelHandlers
	ModelHandlers map[reflect.Type]*ModelHandler
}

// NewHandler creates new handler on the base of
func NewHandler(
	c *Controller,
	log unilogger.LeveledLogger,
	DBErrMgr *ErrorManager,
) *Handler {
	if DBErrMgr == nil {
		DBErrMgr = NewDBErrorMgr()
	}
	h := &Handler{
		Controller:      c,
		log:             log,
		DBErrMgr:        DBErrMgr,
		ModelHandlers:   make(map[reflect.Type]*ModelHandler),
		CreateValidator: validator.New(),
		PatchValidator:  validator.New(),
	}

	h.CreateValidator.SetTagName("create")
	h.PatchValidator.SetTagName("patch")

	// Register name func
	h.CreateValidator.RegisterTagNameFunc(JSONAPITagFunc)
	h.PatchValidator.RegisterTagNameFunc(JSONAPITagFunc)

	return h
}

// AddModelHandlers adds the model handlers for given JSONAPI Handler.
// If there are handlers with the same type the funciton returns error.
func (h *Handler) AddModelHandlers(models ...*ModelHandler) error {
	for _, model := range models {
		if _, ok := h.ModelHandlers[model.ModelType]; ok {
			err := fmt.Errorf("ModelHandler of type: '%s' is already inside the Handler", model.ModelType.Name())
			return err
		}
		h.ModelHandlers[model.ModelType] = model
	}
	return nil
}

func (h *Handler) AddPrecheckFilters(
	scope *Scope,
	req *http.Request,
	rw http.ResponseWriter,
	filters ...*PresetFilter,
) (ok bool) {
	for _, filter := range filters {
		h.log.Debugf("Adding precheck filter: %s", filter.GetFieldName())

		value := req.Context().Value(filter.Key)
		if value == nil {
			continue
		}

		if err := h.SetPresetFilterValues(filter.FilterField, value); err != nil {
			h.log.Errorf("Error while setting values for filter field. Model: %v, Filterfield: %v. Error: %v", scope.Struct.GetType().Name(), filter.GetFieldName(), err)
			h.MarshalInternalError(rw)
			return false
		}
		if err := scope.AddFilterField(filter.FilterField); err != nil {
			h.log.Errorf("Cannot add filter field to root scope in get related field. %v", err)
			h.MarshalInternalError(rw)
			return false
		}
	}
	return true
}

func (h *Handler) AddPrecheckPairFilters(
	scope *Scope,
	model *ModelHandler,
	endpoint *Endpoint,
	req *http.Request,
	rw http.ResponseWriter,
	pairs ...*PresetPair,
) (fine bool) {
	for _, presetPair := range pairs {
		presetScope, presetField := presetPair.GetPair()
		if presetPair.Key != nil {
			if !h.getPrecheckFilter(presetPair.Key, presetScope, req, model) {
				continue
			}
		}
		values, err := h.GetPresetValues(presetScope, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if hErr.Code == HErrNoValues {
					if endpoint.Type == List {
						h.MarshalScope(scope, rw, req)
						return
					}
					if presetErr := presetPair.Error; presetErr != nil {
						// if preset err is ErrorObject marshal and return it
						if errObj, ok := presetErr.(*ErrorObject); ok {
							h.MarshalErrors(rw, errObj)
							return
						}
						// otherwise log
						h.log.Errorf("Preset error while prechecking model: %v on path: %v. %v", model.ModelType.Name(), req.URL.Path, presetErr)
						h.MarshalInternalError(rw)
						return

					}
					errObj := ErrInvalidInput.Copy()
					h.MarshalErrors(rw, errObj)
					return
				}
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
			// if handleHandlerError has warning
			continue
		}

		if err := h.SetPresetFilterValues(presetField, values...); err != nil {
			h.log.Errorf("Error while preseting filter for model: '%s'. '%s'", model.ModelType.Name(), err)
			h.MarshalInternalError(rw)
			return
		}

		if err := scope.AddFilterField(presetField); err != nil {
			h.log.Debugf("Cannot add filter field: %v to the model: %v", presetField.GetFieldName(), scope.Struct.GetType().Name())
			h.MarshalInternalError(rw)
			return
		}
	}
	return true
}

func (h *Handler) CheckPrecheckValues(
	scope *Scope,
	filter *FilterField,
) (err error) {
	if scope.Value == nil {
		h.log.Errorf("Provided no value for the scope of type: '%s'", scope.Struct.GetType().Name())
		return IErrScopeNoValue
	}

	checkSingle := func(single reflect.Value) bool {
		field := single.Field(filter.GetFieldIndex())
		h.log.Debugf("Checking field: %v", single.Type().Field(filter.GetFieldIndex()).Name)
		if len(filter.Relationships) > 0 {
			relatedIndex := filter.Relationships[0].GetFieldIndex()

			switch filter.GetFieldKind() {
			case RelationshipSingle:
				if field.IsNil() {
					err = IErrPresetNoValues
					return false
				}
				if field.Kind() == reflect.Ptr {
					field = field.Elem()
				}
				relatedField := field.Field(relatedIndex)
				ok := h.checkValues(filter.Relationships[0].Values[0], relatedField)
				if !ok {
					h.log.Debug("Values does not match: %v", filter.Relationships[0].Values[0], relatedField.Interface())
				}
				return ok
			case RelationshipMultiple:
				for i := 0; i < field.Len(); i++ {
					fieldElem := field.Index(i)
					if fieldElem.Kind() == reflect.Ptr {
						fieldElem = fieldElem.Elem()
					}
					relatedField := fieldElem.Field(relatedIndex)
					if ok := h.checkValues(filter.Relationships[0].Values[0], relatedField); !ok {
						h.log.Debug("Values does not match: %v", filter.Relationships[0].Values[0], relatedField.Interface())
						return false
					}
				}
				return true
			default:
				h.log.Errorf("Invalid filter field kind for field: '%s'. Within model: '%s'.", filter.GetFieldName(), scope.Struct.GetType().Name())
				err = IErrInvalidValueType
				return false
			}
		} else {
			return h.checkValues(filter.Values[0], field)
		}
	}

	v := reflect.ValueOf(scope.Value)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			single := v.Index(i)
			single = single.Elem()
			if ok := checkSingle(single); !ok {
				return IErrValueNotValid
			}
			if err != nil {
				return
			}
		}
	} else if v.Kind() != reflect.Ptr {
		return IErrInvalidScopeType
	} else {
		v = v.Elem()
		if ok := checkSingle(v); !ok {
			return IErrValueNotValid
		}
	}
	return
}

// GetRepositoryByType returns the repository by provided model type.
// If no modelHandler is found within the handler - then the default repository would be
// set.
func (h *Handler) GetRepositoryByType(model reflect.Type) (repo Repository) {
	return h.getModelRepositoryByType(model)
}

// Exported method to get included values for given scope
func (h *Handler) GetIncluded(
	scope *Scope,
	rw http.ResponseWriter,
	req *http.Request,
) (correct bool) {
	// if the scope is the root and there is no included scopes return fast.
	if scope.IsRoot() && len(scope.IncludedScopes) == 0 {
		return true
	}

	if err := scope.SetCollectionValues(); err != nil {
		h.log.Errorf("Setting collection values for the scope of type: %v failed. Err: %v", scope.Struct.GetType(), err)
		h.MarshalInternalError(rw)
		return
	}
	// h.log.Debugf("After setting collection values for: %v", scope.Struct.GetType())

	// h.log.Debug(scope.GetCollectionScope().IncludedValues)

	// Iterate over included fields
	for scope.NextIncludedField() {
		// Get next included field
		includedField, err := scope.CurrentIncludedField()
		if err != nil {
			h.log.Error(err)
			h.MarshalInternalError(rw)
			return
		}

		// Get the primaries from the scope.collection primaries
		missing, err := includedField.GetMissingPrimaries()
		if err != nil {
			h.log.Errorf("While getting missing objects for: '%v'over included field an error occured: %v", includedField.GetFieldName(), err)
			h.MarshalInternalError(rw)
			return
		}

		if len(missing) > 0 {
			// h.log.Debugf("There are: '%d' missing values in get Included.", len(missing))
			includedField.Scope.SetIDFilters(missing...)
			// h.log.Debugf("Created ID Filters: '%v'", includedField.Scope.PrimaryFilters)

			includedRepo := h.GetRepositoryByType(includedField.Scope.Struct.GetType())

			// Get NewMultipleValue
			includedField.Scope.NewValueMany()

			if errObj := h.HookBeforeReader(includedField.Scope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}

			dbErr := includedRepo.List(includedField.Scope)
			if dbErr != nil {
				h.manageDBError(rw, dbErr)
				return
			}

			if includedField.Scope == nil {
				h.log.Errorf("Model: %v, Included repository for model: %v. List() set scope.Value to nil.", scope.Struct.modelType.Name(), includedField.relatedStruct.modelType.Name())
				h.MarshalInternalError(rw)
				return
			}

			if errObj := h.HookAfterReader(includedField.Scope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}

			err := h.getForeginRelationships(includedField.Scope.ctx, includedField.Scope)
			if err != nil {
				h.manageDBError(rw, err)
				return
			}

			if correct = h.GetIncluded(includedField.Scope, rw, req); !correct {
				return
			}
		}
	}
	scope.ResetIncludedField()
	return true
}

func (h *Handler) EndpointForbidden(
	model *ModelHandler,
	endpoint EndpointType,
) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		mStruct := h.Controller.Models.Get(model.ModelType)
		if mStruct == nil {
			h.log.Errorf("Invalid model provided. The Controller does not contain provided model type within ModelMap. Model: '%s'", model.ModelType)
			h.MarshalInternalError(rw)
			return
		}
		errObj := ErrEndpointForbidden.Copy()
		errObj.Detail = fmt.Sprintf("Server does not allow '%s' operation, at given URI: '%s' for the collection: '%s'.", endpoint.String(), req.URL.Path, mStruct.GetCollectionType())
		h.MarshalErrors(rw, errObj)
	}

}

func (h *Handler) GetRelationshipFilters(
	scope *Scope,
	req *http.Request,
	rw http.ResponseWriter,
) error {

	// every relationship filter is for different field
	// replace the filter with the preset values of id field
	// so that the repository should not handle the relationship filter
	for i, relFilter := range scope.RelationshipFilters {
		// Every relationship filter may contain multiple subfilters
		relationshipScope, err := h.Controller.NewScope(reflect.New(relFilter.GetRelatedModelType()).Interface())
		if err != nil {
			// internal
			// model not precomputed
			hErr := newHandlerError(HErrNoModel, "Cannot get new scope.")
			hErr.Model = relFilter.GetRelatedModelStruct()
			return hErr
		}

		relationshipScope.Fieldset = nil

		// Get PresetFilters for the relationship model type
		//	i.e. materials -> storage
		//	storage should have {preset=panel-info.supplier} {filter[panel-info][id]=some-id}
		//

		relModel, ok := h.ModelHandlers[relFilter.GetRelatedModelType()]
		if !ok {
			hErr := newHandlerError(HErrNoModel, "Cannot get model handler")
			hErr.Model = relFilter.GetRelatedModelStruct()
			return hErr
		}
		if relModel.List != nil {
			for _, precheck := range relModel.List.PrecheckPairs {
				precheckScope, precheckField := precheck.GetPair()
				if precheck.Key != nil {
					if !h.getPrecheckFilter(precheck.Key, precheckScope, req, relModel) {
						continue
					}
				}
				values, err := h.GetPresetValues(precheckScope, rw)
				if err != nil {
					if hErr := err.(*HandlerError); hErr != nil {
						return hErr

					} else {
						return err
					}
				}

				if err := h.SetPresetFilterValues(precheckField, values...); err != nil {
					hErr := newHandlerError(HErrValuePreset, err.Error())
					hErr.Field = precheckField.StructField
					return hErr
				}

				if err := relationshipScope.AddFilterField(precheckField); err != nil {
					hErr := newHandlerError(HErrValuePreset, err.Error())
					hErr.Field = precheckField.StructField
					return hErr
				}
			}
		}

		var (
			attrFilter    bool
			primFilter    bool
			foreignFilter bool
		)

		// Get relationship scope filters
		for _, subFieldFilter := range relFilter.Relationships {
			switch subFieldFilter.GetFieldKind() {
			case Primary:
				relationshipScope.PrimaryFilters = append(relationshipScope.PrimaryFilters, subFieldFilter)
				primFilter = true
			case Attribute:
				relationshipScope.AttributeFilters = append(relationshipScope.AttributeFilters, subFieldFilter)
				attrFilter = true
			case ForeignKey:
				relationshipScope.ForeignKeyFilters = append(relationshipScope.ForeignKeyFilters, subFieldFilter)
				foreignFilter = true
			default:
				h.log.Warningf("The subfield of the filter cannot be of relationship filter type. Model: '%s',", scope.Struct.GetType().Name(), req.URL.Path)
			}
		}

		if primFilter && !(attrFilter || foreignFilter) {
			continue
		}

		// Get the relationship scope
		relationshipScope.NewValueMany()

		if errObj := h.HookBeforeReader(relationshipScope); errObj != nil {
			return errObj
		}

		dbErr := h.GetRepositoryByType(relationshipScope.Struct.GetType()).List(relationshipScope)
		if dbErr != nil {
			return dbErr
		}

		if err = h.getForeginRelationships(relationshipScope.Context(), relationshipScope); err != nil {
			return err
		}

		if errObj := h.HookAfterReader(relationshipScope); errObj != nil {
			return errObj
		}

		values, err := relationshipScope.GetPrimaryFieldValues()
		if err != nil {
			h.log.Debugf("GetPrimaryFieldValues error within GetRelationship function. %v", err)
			hErr := newHandlerError(HErrBadValues, err.Error())
			hErr.Model = relFilter.GetRelatedModelStruct()
			return hErr
		}

		if len(values) == 0 {
			hErr := newHandlerError(HErrNoValues, "")
			hErr.Model = relFilter.GetRelatedModelStruct()
			hErr.Scope = relationshipScope
			return hErr
		}

		subField := &FilterField{
			StructField: relFilter.GetRelatedModelStruct().GetPrimaryField(),
			Values:      []*FilterValues{{Operator: OpIn, Values: values}},
		}

		relationFilter := &FilterField{StructField: relFilter.StructField, Relationships: []*FilterField{subField}}

		scope.RelationshipFilters[i] = relationFilter

	}
	return nil
}

// MarshalScope is a handler helper for marshaling the provided scope.
func (h *Handler) MarshalScope(
	scope *Scope,
	rw http.ResponseWriter,
	req *http.Request,
) {
	SetContentType(rw)
	payload, err := h.Controller.MarshalScope(scope)
	if err != nil {
		h.log.Errorf("Error while marshaling scope for model: '%v', for path: '%s', and method: '%s', Error: %s", scope.Struct.GetType(), req.URL.Path, req.Method, err)
		h.errMarshalScope(rw, req)
		return
	}

	if err = MarshalPayload(rw, payload); err != nil {
		h.errMarshalPayload(payload, err, scope.Struct.GetType(), rw, req)
		return
	}
	return

}

// SetLanguages sets the default langauges for given handler.
// Creates the language matcher for given languages.
func (h *Handler) SetLanguages(languages ...language.Tag) {
	h.LanguageMatcher = language.NewMatcher(languages)
	h.SupportedLanguages = languages
}

func (h *Handler) SetDefaultRepo(Repositoriesitory Repository) {
	h.DefaultRepository = Repositoriesitory
}

func (h *Handler) UnmarshalScope(
	model reflect.Type,
	rw http.ResponseWriter,
	req *http.Request,
) *Scope {
	var (
		scope *Scope
		err   error
	)
	scope, err = h.Controller.unmarshalScopeOne(req.Body, model, true)
	if err != nil {
		errObj, ok := err.(*ErrorObject)
		if ok {
			h.MarshalErrors(rw, errObj)
			if errObj.Err != nil {
				h.log.Debugf("Client-Side error. Unmarshal failed. %v", errObj.Err)
			}
			return nil
		}
		h.log.Errorf("Error while unmarshaling: '%v' for path: '%s' and method: %s. Error: %s", model, req.URL.Path, req.Method, err)
		h.MarshalInternalError(rw)
		return nil
	}

	scope.ctx = req.Context()

	return scope
}

func (h *Handler) MarshalInternalError(rw http.ResponseWriter) {
	SetContentType(rw)
	rw.WriteHeader(http.StatusInternalServerError)
	MarshalErrors(rw, ErrInternalError.Copy())
}

func (h *Handler) MarshalErrors(rw http.ResponseWriter, errors ...*ErrorObject) {
	SetContentType(rw)
	if len(errors) > 0 {
		code, err := strconv.Atoi(errors[0].Status)
		if err != nil {
			h.log.Errorf("Status: '%s', for error: %v cannot be converted into http.Status.", errors[0].Status, errors[0])
			h.MarshalInternalError(rw)
			return
		}
		rw.WriteHeader(code)
	} else {
		rw.WriteHeader(http.StatusBadRequest)
	}
	MarshalErrors(rw, errors...)
}

func SetContentType(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", MediaType)
}

func (h *Handler) HandleValidateError(
	model *ModelHandler,
	err error,
	rw http.ResponseWriter,
) {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		h.log.Debug("Invalid Validation Error")
		h.MarshalInternalError(rw)
	}

	// mStruct, err := h.Controller.GetModelStruct(model)
	// if err != nil {
	// 	h.log.Error("Cannot retrieve model from struct.")
	// 	h.MarshalInternalError(rw)
	// }
	vErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		h.MarshalInternalError(rw)
	}

	var errs []*ErrorObject
	for _, fieldError := range vErrors {
		errObj := ErrInvalidJSONFieldValue.Copy()
		errObj.Detail = fmt.Sprintf("Invalid: '%s' for : '%s'", fieldError.ActualTag(), fieldError.Field())
		errs = append(errs, errObj)
	}
	h.MarshalErrors(rw, errs...)
	return
}

// func (h *Handler) checkManyValues(
// 	filterValue *FilterValues,
// 	fieldValues ...reflect.Value,
// ) (ok bool) {

// 	for _, fieldValue := range fieldValues {
// 		switch filterValue.Operator {
// 		case OpIn:
// 			ok = h.checkIn(fieldValue, filterValue.Values...)
// 		case OpNotIn:

// 		case OpEqual:
// 		case OpNotEqual:

// 		}

// 	}

// }

func (h *Handler) checkValues(filterValue *FilterValues, fieldValue reflect.Value) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("Paniced while checking values. '%s'", r)
			ok = false
		}

	}()
	switch filterValue.Operator {
	case OpIn:
		return h.checkIn(fieldValue, filterValue.Values...)
	case OpNotIn:
		return h.checkNotIn(fieldValue, filterValue.Values...)
	case OpEqual:
		return checkEqual(fieldValue, filterValue.Values...)
	case OpNotEqual:
		return !checkEqual(fieldValue, filterValue.Values...)
	case OpLessEqual:
	case OpLessThan:
	case OpGreaterEqual:
	case OpGreaterThan:
	default:
		return false
	}
	return false

}

func (h *Handler) checkIn(fieldValue reflect.Value, values ...interface{}) (ok bool) {
	h.log.Debug("CheckIn")
	if len(values) == 0 {
		return false
	}

	var isTime bool
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}
	if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
		isTime = true
		fieldValue = fieldValue.MethodByName("UnixNano")
	}

	for _, value := range values {
		v := reflect.ValueOf(value)
		if isTime {
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Type() == reflect.TypeOf(time.Time{}) {
				v = v.MethodByName("UnixNano")
			}
		}

		if v.Type() != fieldValue.Type() {
			h.log.Debugf("Invalid type: %v, %v", v.Type(), fieldValue.Type())
			return false
		}

		h.log.Debugf("Comparing: %v, %v", v, fieldValue)
		if ok = fieldValue.Interface() == v.Interface(); ok {
			h.log.Debug("Equal")
			return
		}

		h.log.Debug("Not equal")
		// h.log.Debugf("Comapring Values: %v, %v", v, fieldValue)
		// h.log.Debugf("First: %v, Second: %v", v.Type(), fieldValue.Type())
		// ok = reflect.DeepEqual(v, fieldValue)
		// if ok {
		// h.log.Debug("Equal")
		// return
		// }
		// h.log.Debug("Not equal")
	}
	return
}

func (h *Handler) checkNotIn(fieldValue reflect.Value, values ...interface{}) (ok bool) {

	return !h.checkIn(fieldValue, values...)
}

func checkEqual(fieldValue reflect.Value, values ...interface{}) (ok bool) {
	if len(values) != 1 {
		return false
	}
	v := reflect.ValueOf(values[0])
	if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
		return reflect.DeepEqual(fieldValue.MethodByName("UnixNano"), v.MethodByName("UnixNano"))
	}
	return reflect.DeepEqual(fieldValue, v)
}

func checkLessEqual(fieldValue reflect.Value, values ...interface{}) (ok bool) {
	// for _, value := range values {
	// 	v := reflect.ValueOf(i)
	// }
	return
}

func checkLessThan(fieldValue reflect.Value, values ...interface{}) (ok bool) {
	// var isTime bool
	// if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
	// 	isTime = true
	// 	fieldValue = fieldValue.MethodByName("UnixNano")
	// }

	// switch fieldValue.Kind() {
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

	// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

	// case reflect.Float32, reflect.Float64:

	// case reflect.String:

	// case reflect.Struct:

	// default:
	// 	h.log.Errorf("Invalid field type for compare: '%s'", )
	// }

	// for _, value := range values {
	// 	v := reflect.ValueOf(value)
	// 	if isTime {
	// 		v = v.FieldByName("UnixNano")
	// 	}

	// }
	return
}

func checkGreaterEqual(fieldValue reflect.Value, values ...interface{}) (ok bool) {
	return
}

func checkContains(fieldValue reflect.Value, values ...interface{}) (ok bool) {
	return
}

func (h *Handler) addPresetFilterToPresetScope(
	presetScope *Scope,
	presetFilter *FilterField,
) bool {
	switch presetFilter.GetFieldKind() {
	case Primary:
		presetScope.PrimaryFilters = append(presetScope.PrimaryFilters, presetFilter)
	case Attribute:
		presetScope.AttributeFilters = append(presetScope.AttributeFilters, presetFilter)
	default:
		h.log.Warningf("PrecheckFilter cannot be of reltionship field filter type.")
		return false
	}
	return true
}

func (h *Handler) getModelRepositoryByType(modelType reflect.Type) (repo Repository) {
	model, ok := h.ModelHandlers[modelType]
	if !ok {
		repo = h.DefaultRepository
	} else {
		repo = model.Repository
		if repo == nil {
			repo = h.DefaultRepository
		}
	}
	return repo
}

// getForeginRelationshiops gets foreign relationships primaries
func (h *Handler) getForeginRelationships(ctx context.Context, scope *Scope) error {

	relations := map[string]*StructField{}

	for name, field := range scope.Fieldset {
		if field.isRelationship() {
			if _, ok := relations[name]; !ok {
				relations[name] = field
			}
		}
	}

	// If the included is not within fieldset, get their primaries also.
	for _, included := range scope.IncludedFields {
		if _, ok := relations[included.jsonAPIName]; !ok {
			relations[included.jsonAPIName] = included.StructField
		}
	}

	v := reflect.ValueOf(scope.Value)
	if v.Kind() == reflect.Ptr {
		scope.IsMany = false
	} else if v.Kind() == reflect.Slice {
		scope.IsMany = true
	}

	if reflect.DeepEqual(scope.Value, reflect.Zero(v.Type())) {
		err := errors.Errorf("getForeignRelatinoships failed. Provided scope with nil value. scope.Struct: %v.", scope.Struct)
		h.log.Error(err)
		return err
	}

	for _, rel := range relations {
		err := h.getForeignRelationshipValue(ctx, scope, v, rel)
		if err != nil {
			return err
		}
	}
	return nil
}

// getForeignRelationshipValue gets the value for the given relationship
func (h *Handler) getForeignRelationshipValue(
	ctx context.Context,
	scope *Scope,
	v reflect.Value,
	rel *StructField,
) error {
	// Check the relationships
	if rel.relationship == nil {
		h.log.Debugf("Empty relationship at Field: %s, Model: %s.", rel.fieldName, rel.mStruct.modelType.Name())
		return nil
	}

	switch rel.relationship.Kind {
	case RelHasOne:
		// if relationship is synced get the relations primaries from the related repository
		// if scope value is many use List
		//
		// if scope value is single use Get
		//
		// the scope should contain only foreign key as fieldset
		// Create new relation scope
		var (
			op           FilterOperator
			filterValues []interface{}
		)

		sync := rel.relationship.Sync
		if sync != nil && !*sync {
			return nil
		}

		if !scope.IsMany {
			// Value already checked if nil

			v = v.Elem()
			op = OpEqual

			primVal := v.FieldByIndex(scope.Struct.primary.refStruct.Index)
			prim := primVal.Interface()
			if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
				return nil
			}

			filterValues = []interface{}{prim}
		} else {
			op = OpIn
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.IsNil() {
					continue
				}
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}

				primVal := elem.FieldByIndex(scope.Struct.primary.refStruct.Index)
				prim := primVal.Interface()
				if reflect.DeepEqual(prim, reflect.Zero(primVal.Type()).Interface()) {
					continue
				}
				filterValues = append(filterValues, prim)
			}

			if len(filterValues) == 0 {
				return nil
			}
		}

		fk := rel.relationship.ForeignKey
		relatedScope := &Scope{
			Struct: rel.relatedStruct,
			Fieldset: map[string]*StructField{
				fk.jsonAPIName: fk,
			},
			currentIncludedFieldIndex: -1,
			ctx:    ctx,
			logger: scope.logger,
		}
		relatedScope.collectionScope = relatedScope

		// set filterfield
		filterField := &FilterField{
			StructField: rel.relationship.ForeignKey,
			Values: []*FilterValues{
				{
					Operator: op,
					Values:   filterValues,
				},
			},
		}

		relatedScope.ForeignKeyFilters = append(relatedScope.ForeignKeyFilters, filterField)

		relatedRepo := h.GetRepositoryByType(relatedScope.Struct.modelType)
		if scope.IsMany {
			relatedScope.newValueMany()
			err := relatedRepo.List(relatedScope)
			if err != nil {
				return err
			}

			// iterate over relatedScope values and match the fk's with scope's primaries.
			relVal := reflect.ValueOf(relatedScope.Value)
			for i := 0; i < relVal.Len(); i++ {
				elem := relVal.Index(i)
				if elem.IsNil() {
					continue
				}

				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}

				fkVal := elem.FieldByIndex(fk.refStruct.Index)
				pkVal := elem.FieldByIndex(relatedScope.Struct.primary.refStruct.Index)

				pk := pkVal.Interface()
				if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
					h.log.Debugf("Relationship HasMany. Elem value with Zero Value for foreign key. Elem: %#v, FKName: %s", elem.Interface(), fk.fieldName)
					continue
				}

			rootLoop:
				for j := 0; j < v.Len(); j++ {
					rootElem := v.Index(j)
					if rootElem.Kind() == reflect.Ptr {
						rootElem = rootElem.Elem()
					}

					rootPrim := rootElem.FieldByIndex(scope.Struct.primary.refStruct.Index)
					if rootPrim.Interface() == fkVal.Interface() {
						rootElem.FieldByIndex(rel.refStruct.Index).Set(relVal.Index(i))
						break rootLoop
					}
				}

			}

		} else {
			relatedScope.newValueSingle()
			err := relatedRepo.Get(relatedScope)
			if err != nil {
				return err
			}
			v.FieldByIndex(rel.refStruct.Index).Set(reflect.ValueOf(relatedScope.Value))
		}

		// After getting the values set the scope relation value into the value of the relatedscope
	case RelHasMany:
		// if relationship is synced get it from the related repository
		//
		// Use a relatedRepo.List method
		//
		// the scope should contain only foreign key as fieldset
		var (
			op           FilterOperator
			filterValues []interface{}
			primMap      map[interface{}]int
			primIndex    = scope.Struct.primary.refStruct.Index
		)
		sync := rel.relationship.Sync
		if sync != nil && !*sync {
			return nil
		}

		if !scope.IsMany {
			v = v.Elem()
			op = OpEqual
			pkVal := v.FieldByIndex(primIndex)
			pk := pkVal.Interface()
			if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
				err := errors.Errorf("Empty Scope Primary Value")
				h.log.Debugf("Err: %v. pk:%v, Scope: %#v", err, pk, scope)
				return err
			}

			filterValues = append(filterValues, pk)
		} else {
			primMap = map[interface{}]int{}
			op = OpIn
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.IsNil() {
					continue
				}

				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}

				pkVal := elem.FieldByIndex(primIndex)
				pk := pkVal.Interface()

				// if the value is Zero like continue
				if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
					continue
				}

				primMap[pk] = i

				filterValues = append(filterValues, pk)
			}
		}

		relatedScope := &Scope{
			Struct:                    rel.relatedStruct,
			Fieldset:                  map[string]*StructField{},
			currentIncludedFieldIndex: -1,
			ctx:    ctx,
			logger: scope.logger,
		}
		relatedScope.collectionScope = relatedScope

		fk := rel.relationship.ForeignKey

		// set filterfield
		filterField := &FilterField{
			StructField: fk,
			Values: []*FilterValues{
				{
					Operator: op,
					Values:   filterValues,
				},
			},
		}
		relatedScope.ForeignKeyFilters = append(relatedScope.ForeignKeyFilters, filterField)

		// set fieldset
		relatedScope.Fieldset[fk.jsonAPIName] = fk
		relatedScope.newValueMany()

		relRepo := h.GetRepositoryByType(relatedScope.Struct.modelType)
		err := relRepo.List(relatedScope)
		if err != nil {
			h.log.Debugf("relationRepo List failed. Scope: %#v, Err: %v", scope, err)
			return err
		}

		relVal := reflect.ValueOf(relatedScope.Value)
		for i := 0; i < relVal.Len(); i++ {
			elem := relVal.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}

			// get the foreignkey value
			fkVal := elem.FieldByIndex(fk.refStruct.Index)
			// pkVal := elem.FieldByIndex(relatedScope.Struct.primary.refStruct.Index)

			var scopeElem reflect.Value
			if scope.IsMany {
				// foreign key in relation would be a primary key within root scope
				j, ok := primMap[fkVal.Interface()]
				if !ok {
					h.log.Debugf("Foreign Key not found.")
					continue
				}
				// rootPK should be at index 'j'
				scopeElem = v.Index(j)
				if scopeElem.Kind() == reflect.Ptr {
					scopeElem = scopeElem.Elem()
				}
			} else {
				scopeElem = v
			}

			elemRelVal := scopeElem.FieldByIndex(rel.refStruct.Index)
			elemRelVal.Set(reflect.Append(elemRelVal, relVal.Index(i)))

		}

	case RelMany2Many:
		// if the relationship is synced get the values from the relationship
		sync := rel.relationship.Sync
		if sync != nil && *sync {
			// when sync get the relationships value from the backreference relationship
			var (
				filterValues = []interface{}{}
				primMap      map[interface{}]int
				primIndex    = scope.Struct.primary.refStruct.Index
				op           FilterOperator
			)

			if !scope.IsMany {
				// set isMany to false just to see the difference

				op = OpEqual

				// Derefernce the pointer
				v = v.Elem()

				pkVal := v.FieldByIndex(primIndex)
				pk := pkVal.Interface()

				if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
					err := errors.Errorf("Error no primary valoue set for the scope value.  Scope %#v. Value: %+v", scope, scope.Value)
					return err
				}

				filterValues = append(filterValues, pk)

			} else {

				// Operator is 'in'
				op = OpIn

				// primMap contains the index of the proper scope value for
				//given primary key value
				primMap = map[interface{}]int{}
				for i := 0; i < v.Len(); i++ {
					elem := v.Index(i)
					if elem.IsNil() {
						continue
					}

					if elem.Kind() == reflect.Ptr {
						elem = elem.Elem()
					}

					pkVal := elem.FieldByIndex(primIndex)
					pk := pkVal.Interface()

					// if the value is Zero like continue
					if reflect.DeepEqual(pk, reflect.Zero(pkVal.Type()).Interface()) {
						continue
					}

					primMap[pk] = i

					filterValues = append(filterValues, pk)
				}
			}

			backReference := rel.relationship.BackReferenceField

			backRefRelPrimary := backReference.relatedStruct.primary

			filterField := &FilterField{
				StructField: backReference,
				Relationships: []*FilterField{
					{
						StructField: backRefRelPrimary,
						Values: []*FilterValues{
							{
								Operator: op,
								Values:   filterValues,
							},
						},
					},
				},
			}

			relatedScope := &Scope{
				Struct: rel.relatedStruct,
				Fieldset: map[string]*StructField{
					backReference.jsonAPIName: backReference,
				},
				currentIncludedFieldIndex: -1,
				ctx:    ctx,
				logger: scope.logger,
			}
			relatedScope.collectionScope = relatedScope

			relatedScope.RelationshipFilters = append(relatedScope.RelationshipFilters, filterField)

			relatedScope.newValueMany()

			relatedRepo := h.GetRepositoryByType(relatedScope.Struct.modelType)
			err := relatedRepo.List(relatedScope)
			if err != nil {
				return err
			}

			relVal := reflect.ValueOf(relatedScope.Value)

			// iterate over relatedScoep Value
			for i := 0; i < relVal.Len(); i++ {
				// Get Each element from the relatedScope Value
				relElem := relVal.Index(i)
				if relElem.IsNil() {
					continue
				}

				// Dereference the ptr
				if relElem.Kind() == reflect.Ptr {
					relElem = relElem.Elem()
				}

				relPrimVal := relElem.FieldByIndex(relatedScope.Struct.primary.refStruct.Index)

				// continue if empty
				if reflect.DeepEqual(relPrimVal.Int(), reflect.Zero(relPrimVal.Type()).Interface()) {
					continue
				}

				backRefVal := relElem.FieldByIndex(backReference.refStruct.Index)

				// iterate over backreference elems
			backRefLoop:
				for j := 0; j < backRefVal.Len(); j++ {
					// BackRefPrimary would be root primary
					backRefElem := backRefVal.Index(j)
					if backRefElem.IsNil() {
						continue
					}

					if backRefElem.Kind() == reflect.Ptr {
						backRefElem = backRefElem.Elem()
					}

					backRefPrimVal := backRefElem.FieldByIndex(backRefRelPrimary.refStruct.Index)
					backRefPrim := backRefPrimVal.Interface()

					if reflect.DeepEqual(backRefPrim, reflect.Zero(backRefPrimVal.Type()).Interface()) {
						h.log.Warningf("Backreferecne Relationship Many2Many contains zero valued primary key. Scope: %#v, RelatedScope: %#v. Relationship: %#v", scope, relatedScope, rel)
						continue backRefLoop
					}

					var val reflect.Value
					if scope.IsMany {
						index, ok := primMap[backRefPrim]
						if !ok {
							h.log.Warningf("Relationship Many2Many contains backreference primaries that doesn't match the root scope value. Scope: %#v. Relationship: %v", scope, rel)
							continue backRefLoop
						}

						val = v.Index(index)
					} else {
						val = v
					}

					m2mElem := reflect.New(rel.relatedStruct.modelType)
					m2mElem.FieldByIndex(rel.relatedStruct.primary.refStruct.Index).Set(relPrimVal)

					relationField := val.FieldByIndex(rel.refStruct.Index)
					relationField.Set(reflect.Append(relationField, m2mElem))
				}

			}

		}
	case RelBelongsTo:
		// if scope value is a slice
		// iterate over all entries and transfer all foreign keys into relationship primaries.

		// for single value scope do it once

		sync := rel.relationship.Sync
		if sync != nil && *sync {
			// don't know what to do now
			// Get it from the backreference ? and
			h.log.Warningf("Synced BelongsTo relationship. Scope: %#v, Relation: %#v", scope, rel)
		}

		fkField := rel.relationship.ForeignKey
		if !scope.IsMany {
			v = v.Elem()
			fkVal := v.FieldByIndex(fkField.refStruct.Index)
			if reflect.DeepEqual(fkVal.Interface(), reflect.Zero(fkVal.Type()).Interface()) {
				return nil
			}
			relVal := v.FieldByIndex(rel.refStruct.Index)
			if relVal.Kind() == reflect.Ptr {
				if relVal.IsNil() {
					relVal.Set(reflect.New(relVal.Type().Elem()))
				}
				relVal = relVal.Elem()
			} else if relVal.Kind() != reflect.Struct {
				err := errors.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", scope.Struct, rel)
				h.log.Warning(err)
				return err
			}
			relPrim := relVal.FieldByIndex(rel.relatedStruct.primary.refStruct.Index)
			relPrim.Set(fkVal)

		} else {
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
				fkVal := elem.FieldByIndex(fkField.refStruct.Index)
				if reflect.DeepEqual(fkVal.Interface(), reflect.Zero(fkVal.Type()).Interface()) {
					continue
				}

				relVal := elem.FieldByIndex(rel.refStruct.Index)
				if relVal.Kind() == reflect.Ptr {
					if relVal.IsNil() {
						relVal.Set(reflect.New(relVal.Type().Elem()))
					}
					relVal = relVal.Elem()
				} else if relVal.Kind() != reflect.Struct {
					err := errors.Errorf("Relation Field signed as BelongsTo with unknown field type. Model: %#v, Relation: %#v", scope.Struct, rel)
					h.log.Warning(err)
					return err
				}
				relPrim := relVal.FieldByIndex(rel.relatedStruct.primary.refStruct.Index)
				relPrim.Set(fkVal)
			}
		}
	}
	return nil
}

func (h *Handler) handleHandlerError(hErr *HandlerError, rw http.ResponseWriter) bool {
	switch hErr.Code {
	case HErrInternal:
		h.log.Error(hErr.Error())
		h.log.Error(string(debug.Stack()))
		h.MarshalInternalError(rw)
		return false
	case HErrAlreadyWritten:
		return false
	case HErrBadValues, HErrNoModel, HErrValuePreset:
		h.log.Error(hErr.Error())
		h.MarshalInternalError(rw)
		return false
	case HErrNoValues:
		errObj := ErrResourceNotFound.Copy()
		h.MarshalErrors(rw, errObj)
		return false
	case HErrWarning:
		h.log.Warning(hErr)
		return true
	}
	return true

}

func (h *Handler) manageDBError(rw http.ResponseWriter, err error) {
	var errObj *ErrorObject
	switch e := err.(type) {
	case *ErrorObject:
		h.log.Debugf("DBError is *ErrorObject: %v", e)
		errObj = e
	case *unidb.Error:
		h.log.Debugf("DBError is *unidb.Error: %v", e)
		proto, er := e.GetPrototype()
		if er != nil {
			h.log.Errorf("*unidb.Error.GetPrototype (%#v) failed. %v", e, err)
			h.MarshalInternalError(rw)
			return
		}

		if proto == unidb.ErrUnspecifiedError || proto == unidb.ErrInternalError {
			h.log.Errorf("*unidb.ErrUnspecified. Message: %v", e.Message)
			h.MarshalInternalError(rw)
			return
		}

		errObj, err = h.DBErrMgr.Handle(e)
		if err != nil {
			h.log.Errorf("DBErrorManager.Handle failed. %v", err)
			h.MarshalInternalError(rw)
			return
		}
	default:
		h.log.Errorf("Handler.ManageDbErrors failed. Unknown error provided %v.", err)
		h.MarshalInternalError(rw)
		return
	}

	h.MarshalErrors(rw, errObj)
	return
}

func (h *Handler) errSetIDFilter(
	scope *Scope,
	err error,
	rw http.ResponseWriter,
	req *http.Request,
) {
	h.log.Errorf("Error while setting id filter for the path: '%s', and scope: of type '%v'. Error: %v", req.URL.Path, scope.Struct.GetType(), err)
	h.MarshalInternalError(rw)
	return
}

func (h *Handler) errMarshalPayload(
	payload Payloader,
	err error,
	model reflect.Type,
	rw http.ResponseWriter,
	req *http.Request,
) {
	h.log.Errorf("Error while marshaling payload: '%v'. For model: '%v', Path: '%s', Method: '%s', Error: %v", payload, model, req.URL.Path, req.Method, err)
	h.MarshalInternalError(rw)
}

// setScopeFlags sets the flags for given scope based on the endpoint and modelHandler
func (h *Handler) setScopeFlags(scope *Scope, endpoint *Endpoint, model *ModelHandler) {

	// FlagMetaCountList
	if scope.FlagMetaCountList == nil {
		if endpoint.FlagMetaCountList != nil {
			scope.FlagMetaCountList = &(*endpoint.FlagMetaCountList)
		} else if model.FlagMetaCountList != nil {
			scope.FlagMetaCountList = &(*model.FlagMetaCountList)
		} else if h.Controller.FlagMetaCountList != nil {
			scope.FlagMetaCountList = &(*h.Controller.FlagMetaCountList)
		}
	}

	// FlagReturnPatchContent
	if scope.FlagReturnPatchContent == nil {
		if endpoint.FlagReturnPatchContent != nil {
			scope.FlagReturnPatchContent = &(*endpoint.FlagReturnPatchContent)
		} else if model.FlagReturnPatchContent != nil {
			scope.FlagReturnPatchContent = &(*model.FlagReturnPatchContent)
		} else if h.Controller.FlagReturnPatchContent != nil {
			scope.FlagReturnPatchContent = &(*h.Controller.FlagReturnPatchContent)
		}
	}

	// FlagUseLinks
	if scope.FlagUseLinks == nil {
		if endpoint.FlagReturnPatchContent != nil {
			scope.FlagUseLinks = &(*endpoint.FlagUseLinks)
		} else if model.FlagUseLinks != nil {
			scope.FlagUseLinks = &(*model.FlagUseLinks)
		} else if h.Controller.FlagUseLinks != nil {
			scope.FlagUseLinks = &(*h.Controller.FlagUseLinks)
		}
	}

}

func (h *Handler) errMarshalScope(
	rw http.ResponseWriter,
	req *http.Request,
) {
	h.MarshalInternalError(rw)
}
