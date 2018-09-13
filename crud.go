package jsonapi

import (
	"fmt"
	"github.com/kucjac/uni-db"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"reflect"
	"strings"
)

// Create returns http.HandlerFunc that creates new 'model' entity within it's repository.
//
// I18n
// It supports i18n of the model. The language in the request is being checked
// if the value provided is supported by the server. If the match is confident
// the language is converted.
//
// Correctly Response with status '201' Created.
func (h *JSONAPIHandler) Create(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		SetContentType(rw)
		scope := h.UnmarshalScope(model.ModelType, rw, req)
		if scope == nil {
			return
		}

		h.log.Debugf("Create scope value: %+v", scope.Value)

		/**

		CREATE: LANGUAGE

		*/
		// // if the model is i18n-ready control it's language field value
		// if scope.UseI18n() {
		// 	lang, ok := h.CheckValueLanguage(scope, rw)
		// 	if !ok {
		// 		return
		// 	}
		// 	h.HeaderContentLanguage(rw, lang)
		// }

		/**

		CREATE: PRESET PAIRS

		*/
		for _, pair := range endpoint.PresetPairs {
			presetScope, presetField := pair.GetPair()
			if pair.Key != nil {
				if !h.getPresetFilter(pair.Key, presetScope, req, model) {
					continue
				}
			}

			values, err := h.GetPresetValues(presetScope, rw)
			if err != nil {
				if hErr := err.(*HandlerError); hErr != nil {
					if hErr.Code == HErrNoValues {
						errObj := ErrInsufficientAccPerm.Copy()
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
				continue
			}

			if err := h.PresetScopeValue(scope, presetField, values...); err != nil {
				h.log.Errorf("Cannot preset value while creating model: '%s'.'%s'", model.ModelType.Name(), err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		CREATE: PRESET FILTERS

		*/
		for _, filter := range endpoint.PresetFilters {
			value := req.Context().Value(filter.Key)
			if value != nil {
				if err := h.SetPresetFilterValues(filter.FilterField, value); err != nil {
					h.log.Errorf("Cannot preset value for the Create.PresetFilter. Model %v. Field %v. Value: %v", model.ModelType.Name(), filter.StructField.GetFieldName(), value)
					h.MarshalInternalError(rw)
					return
				}
				if err := h.PresetScopeValue(scope, filter.FilterField, value); err != nil {
					h.log.Errorf("Cannot preset value for the model: '%s'. FilterField: %v. Error: %v", model.ModelType.Name(), filter.GetFieldName(), err)
					h.MarshalInternalError(rw)
					return
				}
			}
		}

		/**

		CREATE: VALIDATE MODEL

		*/

		if err := h.CreateValidator.Struct(scope.Value); err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				errObj := ErrInvalidJSONFieldValue.Copy()
				h.MarshalErrors(rw, errObj)
				return
			}

			validateErrors, ok := err.(validator.ValidationErrors)
			if !ok || ok && len(validateErrors) == 0 {
				h.log.Debugf("Unknown error type while validating. %v", err)
				h.MarshalErrors(rw, ErrInvalidJSONFieldValue.Copy())
				return
			}

			var errs []*ErrorObject
			for _, verr := range validateErrors {
				tag := verr.Tag()

				var errObj *ErrorObject
				if tag == "required" {
					// if field is required but and the field tag is empty
					if verr.Field() == "" {
						errObj = ErrInternalError.Copy()
						h.MarshalErrors(rw, errObj)
						return
					}
					errObj = ErrMissingRequiredJSONField.Copy()
					errObj.Detail = fmt.Sprintf("The field: %s, is required.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else if tag == "isdefault" {
					if verr.Field() == "" {
						errObj = ErrInternalError.Copy()
						h.MarshalErrors(rw, errObj)
						return
					}
					errObj = ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The field: '%s' must be empty.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else if strings.HasPrefix(tag, "len") {
					if verr.Field() == "" {
						errObj = ErrInternalError.Copy()
						h.MarshalErrors(rw, errObj)
						return
					}
					errObj = ErrInvalidJSONFieldValue.Copy()
					errObj.Detail = fmt.Sprintf("The value of the field: %s is of invalid length.", verr.Field())
					errs = append(errs, errObj)
					continue
				} else {
					errObj = ErrInvalidJSONFieldValue.Copy()
					if verr.Field() != "" {
						errObj.Detail = fmt.Sprintf("Invalid value for the field: '%s'.", verr.Field())
					}
					errs = append(errs, errObj)
					continue
				}
			}
			h.MarshalErrors(rw, errs...)
			return
		}

		/**

		CREATE: PRECHECK PAIRS

		*/

		// for _, pair := range endpoint.PrecheckPairs {
		// 	presetScope, presetField := pair.GetPair()
		// 	if pair.Key != nil {
		// 		if !h.getPrecheckFilter(pair.Key, presetScope, req, model) {
		// 			continue
		// 		}
		// 	}

		// 	values, err := h.GetPresetValues(presetScope, rw)
		// 	if err != nil {
		// 		if hErr := err.(*HandlerError); hErr != nil {
		// 			if hErr.Code == HErrNoValues {
		// 				errObj := ErrInsufficientAccPerm.Copy()
		// 				h.MarshalErrors(rw, errObj)
		// 				return
		// 			}
		// 			if !h.handleHandlerError(hErr, rw) {
		// 				return
		// 			}
		// 		} else {
		// 			h.log.Error(err)
		// 			h.MarshalInternalError(rw)
		// 			return
		// 		}
		// 		continue
		// 	}

		// 	if err := h.SetPresetFilterValues(presetField, values...); err != nil {
		// 		h.log.Error("Cannot preset values to the filter value. %s", err)
		// 		h.MarshalInternalError(rw)
		// 		return
		// 	}

		// 	if err := h.CheckPrecheckValues(scope, presetField); err != nil {
		// 		h.log.Debugf("Precheck value error: '%s'", err)
		// 		if err == IErrValueNotValid {
		// 			errObj := ErrInvalidJSONFieldValue.Copy()
		// 			errObj.Detail = "One of the field values are not valid."
		// 			h.MarshalErrors(rw, errObj)
		// 		} else {
		// 			h.MarshalInternalError(rw)
		// 		}
		// 		return
		// 	}
		// }

		/**

		CREATE: PRECHECK FILTERS

		*/

		// for _, filter := range endpoint.PrecheckFilters {
		// 	value := req.Context().Value(filter.Key)
		// 	if value != nil {
		// 		if err := h.SetPresetFilterValues(filter.FilterField, value); err != nil {
		// 			h.log.Errorf("Cannot preset value for the Create.PresetFilter. Model %v. Field %v. Value: %v", model.ModelType.Name(), filter.StructField.GetFieldName(), value)
		// 		}

		// 		if err := scope.AddFilterField(filter.FilterField); err != nil {
		// 			h.log.Error(err)
		// 			h.MarshalInternalError(rw)
		// 			return
		// 		}
		// 	}
		// }

		/**

		CREATE: RELATIONSHIP FILTERS

		*/
		err := h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		CREATE: HOOK BEFORE

		*/

		beforeCreater, ok := scope.Value.(HookBeforeCreator)
		if ok {
			if err = beforeCreater.JSONAPIBeforeCreate(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.MarshalInternalError(rw)
				return
			}
		}

		repo := h.GetRepositoryByType(model.ModelType)

		/**

		CREATE: REPOSITORY CREATE

		*/
		if dbErr := repo.Create(scope); dbErr != nil {
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		CREATE: HOOK AFTER

		*/
		afterCreator, ok := scope.Value.(HookAfterCreator)
		if ok {
			if err = afterCreator.JSONAPIAfterCreate(scope); err != nil {
				// the value should not be created?
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Debugf("Error in HookAfterCreator: %v", err)
				h.MarshalInternalError(rw)
				return
			}
		}

		rw.WriteHeader(http.StatusCreated)
		h.MarshalScope(scope, rw, req)
	}
}

// Get returns a http.HandlerFunc that gets single entity from the "model's"
// repository.
func (h *JSONAPIHandler) Get(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		SetContentType(rw)

		/**

		GET: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeSingle(req, reflect.New(model.ModelType).Interface(), nil)
		if err != nil {
			h.log.Error(err)
			h.MarshalInternalError(rw)
			return
		}

		if errs != nil {
			h.MarshalErrors(rw, errs...)
			return
		}
		scope.NewValueSingle()

		/**

		GET: PRECHECK PAIR

		*/
		if !h.AddPrecheckPairFilters(scope, model, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		GET: PRECHECK FILTERS

		*/

		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		/**

		GET: HOOK BEFORE

		*/

		if errObj := h.HookBeforeReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		GET: RELATIONSHIP FILTERS

		*/
		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
		}

		repo := h.GetRepositoryByType(model.ModelType)
		// Set NewSingleValue for the scope

		/**

		GET: REPOSITORY GET

		*/
		dbErr := repo.Get(scope)
		if dbErr != nil {
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		GET: HOOK AFTER

		*/
		if errObj := h.HookAfterReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		GET: GET INCLUDED FIELDS

		*/
		if correct := h.GetIncluded(scope, rw, req); !correct {
			return
		}

		h.MarshalScope(scope, rw, req)
		return
	}
}

// GetRelated returns a http.HandlerFunc that returns the related field for the 'root' model
// It prepares the scope rooted with 'root' model with some id and gets the 'related' field from
// url. Related field must be a relationship, otherwise an error would be returned.
// The handler gets the root and the specific related field 'id' from the repository
// and then gets the related object from it's repository.
// If no error occurred an related object is being returned
func (h *JSONAPIHandler) GetRelated(root *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[root.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		SetContentType(rw)

		/**

		GET RELATED: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeRelated(req, reflect.New(root.ModelType).Interface())
		if err != nil {
			h.log.Errorf("An internal error occurred while building related scope for model: '%v'. %v", root.ModelType, err)
			h.MarshalInternalError(rw)
			return
		}
		if len(errs) > 0 {
			h.MarshalErrors(rw, errs...)
			return
		}

		scope.NewValueSingle()

		/**

		GET RELATED: PRECHECK PAIR

		*/
		if !h.AddPrecheckPairFilters(scope, root, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		GET RELATED: PRECHECK FILTERS

		*/
		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		/**

		  GET RELATED: HOOK BEFORE READ

		*/

		if errObj := h.HookBeforeReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		GET RELATED:  RELATIONSHIP FILTERS

		*/

		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
		}

		// Get root repository
		rootRepository := h.GetRepositoryByType(root.ModelType)
		// Get the root for given id
		// Select the related field inside

		/**

		GET RELATED: REPOSITORY GET ROOT

		*/
		dbErr := rootRepository.Get(scope)
		if dbErr != nil {
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		  GET RELATED: ROOT HOOK AFTER READ

		*/
		if errObj := h.HookAfterReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		GET RELATED: BUILD RELATED SCOPE

		*/

		relatedScope, err := scope.GetRelatedScope()
		if err != nil {
			h.log.Errorf("Error while getting Related Scope: %v", err)
			h.MarshalInternalError(rw)
			return
		}

		// if there is any primary filter
		if relatedScope.Value != nil && len(relatedScope.PrimaryFilters) != 0 {

			relatedRepository := h.GetRepositoryByType(relatedScope.Struct.GetType())

			/**

			  GET RELATED: HOOK BEFORE READER

			*/
			if errObj := h.HookBeforeReader(relatedScope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}

			// SELECT METHOD TO GET
			if relatedScope.IsMany {
				h.log.Debug("The related scope isMany.")
				dbErr = relatedRepository.List(relatedScope)
			} else {
				h.log.Debug("The related scope isSingle.")
				h.log.Debugf("The value of related scope: %+v", relatedScope.Value)
				h.log.Debugf("Fieldset %+v", relatedScope.Fieldset)
				dbErr = relatedRepository.Get(relatedScope)
			}
			if dbErr != nil {
				h.manageDBError(rw, dbErr)
				return
			}

			/**

			HOOK AFTER READER

			*/
			if errObj := h.HookAfterReader(relatedScope); errObj != nil {
				h.MarshalErrors(rw, errObj)
				return
			}
		}
		h.MarshalScope(relatedScope, rw, req)
		return

	}
}

// GetRelationship returns a http.HandlerFunc that returns in the response the relationship field
// for the root model
func (h *JSONAPIHandler) GetRelationship(root *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[root.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}

		/**

		GET RELATIONSHIP: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeRelationship(req, reflect.New(root.ModelType).Interface())
		if err != nil {
			h.log.Error(err)
			h.MarshalInternalError(rw)
			return
		}
		if len(errs) > 0 {
			h.MarshalErrors(rw, errs...)
			return
		}
		scope.NewValueSingle()

		/**

		  GET RELATIONSHIP: PRECHECK PAIR

		*/
		if !h.AddPrecheckPairFilters(scope, root, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		GET RELATIONSHIP: PRECHECK FILTERS

		*/

		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		/**

		  GET RELATIONSHIP: ROOT HOOK BEFORE READ

		*/

		if errObj := h.HookBeforeReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		GET RELATIONSHIP: GET RELATIONSHIP FILTERS

		*/

		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  GET RELATIONSHIP: GET ROOT FROM REPOSITORY

		*/

		rootRepository := h.GetRepositoryByType(scope.Struct.GetType())
		dbErr := rootRepository.Get(scope)
		if dbErr != nil {
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		  GET RELATIONSHIP: ROOT HOOK AFTER READ

		*/

		if errObj := h.HookAfterReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		  GET RELATIONSHIP: GET RELATIONSHIP SCOPE

		*/
		relationshipScope, err := scope.GetRelationshipScope()
		if err != nil {
			h.log.Errorf("Error while getting RelationshipScope for model: %v. %v", scope.Struct.GetType(), err)
			h.MarshalInternalError(rw)
			return
		}

		/**

		  GET RELATIONSHIP: MARSHAL SCOPE

		*/
		h.MarshalScope(relationshipScope, rw, req)
	}
}

// List returns a http.HandlerFunc that response with the model's entities taken
// from it's repository.
// QueryParameters:
//	- filter - filter parameter must be followed by the collection name within brackets
// 		i.e. '[collection]' and the field scoped for the filter within brackets, i.e. '[id]'
//		i.e. url: http://myapiurl.com/api/blogs?filter[blogs][id]=4
func (h *JSONAPIHandler) List(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		SetContentType(rw)

		/**

		  LIST: BUILD SCOPE

		*/
		scope, errs, err := h.Controller.BuildScopeList(req, reflect.New(model.ModelType).Interface())
		if err != nil {
			h.log.Error(err)
			h.MarshalInternalError(rw)
			return
		}
		if len(errs) > 0 {
			h.MarshalErrors(rw, errs...)
			return
		}
		scope.NewValueMany()

		/**

		  LIST: PRECHECK PAIRS

		*/
		if !h.AddPrecheckPairFilters(scope, model, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		  LIST: PRECHECK FILTERS

		*/

		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}

		/**

		  LIST: HOOK BEFORE READER

		*/

		if errObj := h.HookBeforeReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		  LIST: GET RELATIONSHIP FILTERS

		*/
		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if hErr.Code == HErrNoValues {
					scope.NewValueMany()
					h.MarshalScope(scope, rw, req)
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
		}

		repo := h.GetRepositoryByType(model.ModelType)

		/**

		  LIST: INCLUDE COUNT

		  Include count into meta data
		*/
		if endpoint.FlagMetaCountList {
			scope.PageTotal = true
		}

		/**

		  LIST: DEFAULT PAGINATION

		*/

		if endpoint.PresetPaginate != nil && scope.Pagination == nil {
			scope.Pagination = endpoint.PresetPaginate
		}

		/**

		  LIST: DEFAULT SORT

		*/
		if len(endpoint.PresetSort) != 0 {
			scope.Sorts = append(endpoint.PresetSort, scope.Sorts...)
		}

		/**

		  LIST: LIST FROM REPOSITORY

		*/
		dbErr := repo.List(scope)
		if dbErr != nil {
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		  LIST: HOOK AFTER READ

		*/
		if errObj := h.HookAfterReader(scope); errObj != nil {
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		  LIST: GET INCLUDED

		*/
		if correct := h.GetIncluded(scope, rw, req); !correct {
			return
		}

		/**

		  LIST: MARSHAL SCOPE

		*/
		h.MarshalScope(scope, rw, req)
		return
	}
}

// Patch the patch endpoint is used to patch given entity.
// It accepts only the models that matches the provided model Handler.
// If the incoming model
// PRESETTING:
//	- Preset values using PresetScope
//	- Precheck values using PrecheckScope
func (h *JSONAPIHandler) Patch(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}
		// UnmarshalScope from the request body.
		SetContentType(rw)

		/**

		  PATCH: UNMARSHAL SCOPE

		*/
		scope := h.UnmarshalScope(model.ModelType, rw, req)
		if scope == nil {
			return
		}

		/**

		  PATCH: GET and SET ID

		  Set the ID for given model's scope
		*/

		err := h.Controller.GetAndSetID(req, scope)
		if err != nil {
			if err != IErrInvalidType {
				h.MarshalInternalError(rw)
				return
			}
			errObj := ErrInvalidQueryParameter.Copy()
			errObj.Detail = "Provided invalid id parameter."
			h.MarshalErrors(rw, errObj)
			return
		}

		/**

		  PATCH: PRESET PAIRS

		*/
		for _, presetPair := range endpoint.PresetPairs {
			presetScope, presetField := presetPair.GetPair()
			if presetPair.Key != nil {
				if !h.getPresetFilter(presetPair.Key, presetScope, req, model) {
					continue
				}
			}
			values, err := h.GetPresetValues(presetScope, rw)
			if err != nil {
				if hErr := err.(*HandlerError); hErr != nil {
					if hErr.Code == HErrNoValues {
						errObj := ErrInsufficientAccPerm.Copy()
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
				continue
			}

			if err := h.PresetScopeValue(scope, presetField, values...); err != nil {
				h.log.Errorf("Cannot preset value while creating model: '%s'.'%s'", model.ModelType.Name(), err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  PATCH: PRESET FILTERS

		*/

		if !h.SetPresetFilters(scope, model, req, rw, endpoint.PresetFilters...) {
			return
		}

		/**

		  PATCH: PRECHECK PAIRS

		*/

		/**

		CREATE: PRECHECK PAIRS

		*/

		// for _, pair := range endpoint.PrecheckPairs {
		// 	presetScope, presetField := pair.GetPair()
		// 	if pair.Key != nil {
		// 		if !h.getPrecheckFilter(pair.Key, presetScope, req, model) {
		// 			continue
		// 		}
		// 	}

		// 	values, err := h.GetPresetValues(presetScope, rw)
		// 	if err != nil {
		// 		if hErr := err.(*HandlerError); hErr != nil {
		// 			if hErr.Code == HErrNoValues {
		// 				errObj := ErrInsufficientAccPerm.Copy()
		// 				h.MarshalErrors(rw, errObj)
		// 				return
		// 			}
		// 			if !h.handleHandlerError(hErr, rw) {
		// 				return
		// 			}
		// 		} else {
		// 			h.log.Error(err)
		// 			h.MarshalInternalError(rw)
		// 			return
		// 		}
		// 		continue
		// 	}

		// 	if err := h.SetPresetFilterValues(presetField, values...); err != nil {
		// 		h.log.Error("Cannot preset values to the filter value. %s", err)
		// 		h.MarshalInternalError(rw)
		// 		return
		// 	}

		// 	if err := h.CheckPrecheckValues(scope, presetField); err != nil {
		// 		h.log.Debugf("Precheck value error: '%s'", err)
		// 		if err == IErrValueNotValid {
		// 			errObj := ErrInvalidJSONFieldValue.Copy()
		// 			errObj.Detail = "One of the field values are not valid."
		// 			h.MarshalErrors(rw, errObj)
		// 		} else {
		// 			h.MarshalInternalError(rw)
		// 		}
		// 		return
		// 	}
		// }
		// if !h.AddPrecheckPairFilters(scope, model, endpoint, req, rw, endpoint.PrecheckPairs...) {
		// 	return
		// }

		/**

		  PATCH: PRECHECK FILTERS

		*/

		// if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
		// 	return
		// }

		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
		}

		// Get the Repository for given model
		repo := h.GetRepositoryByType(model.ModelType)

		/**

		  PATCH: GET MODIFIED RESULT
			The default hierarchy is that at first it checks wether the
			endpoint contains the
		*/
		if endpoint.FlagReturnPatchContent != nil {
			scope.FlagReturnPatchContent = *endpoint.FlagReturnPatchContent
		} else if h.Controller.FlagReturnPatchContent {
			scope.FlagReturnPatchContent = true
		}

		/**

		  PATCH: HOOK BEFORE PATCH

		*/
		if beforePatcher, ok := scope.Value.(HookBeforePatcher); ok {
			if err = beforePatcher.JSONAPIBeforePatch(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Errorf("Error in HookBeforePatch for model: %v. Error: %v", model.ModelType.Name(), err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  PATCH: REPOSITORY PATCH

		*/
		// Use Patch Method on given model's Repository for given scope.
		if dbErr := repo.Patch(scope); dbErr != nil {
			if dbErr.Compare(unidb.ErrNoResult) && endpoint.HasPrechecks() {
				errObj := ErrInsufficientAccPerm.Copy()
				errObj.Detail = "Given object is not available for this account or it does not exists."
				h.MarshalErrors(rw, errObj)
				return
			}
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		  PATCH: HOOK AFTER PATCH

		*/
		if afterPatcher, ok := scope.Value.(HookAfterPatcher); ok {
			if err = afterPatcher.JSONAPIAfterPatch(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Errorf("Error in HookAfterPatcher for model: %v. Error: %v", model.ModelType.Name(), err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  PATCH: MARSHAL RESULT

		*/
		if scope.FlagReturnPatchContent {
			h.MarshalScope(scope, rw, req)
		} else {
			rw.WriteHeader(http.StatusNoContent)
		}
		return
	}
}

func (h *JSONAPIHandler) PatchRelated(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return h.EndpointForbidden(model, PatchRelated)
}

// PatchRelationship is a handler function that is used to update
// relationships independently at URLs from relationship links.
// Reference:
//	- http://jsonapi.org/format/#crud-updating-relationships

func (h *JSONAPIHandler) PatchRelationship(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	/**

	 */
	return h.EndpointForbidden(model, PatchRelationship)
}

// UpdateToOneRelationships
// A server must respond to PATCH request to a URL a to-one relationship
// link as describe:
//	- a resource identifier object corresponding to the new related
//	resource.
//	- null, to remove the relationship.
func (h *JSONAPIHandler) UpdateToOneRelationships(
	model *ModelHandler,
	endpoint *Endpoint,
) http.HandlerFunc {

	return h.EndpointForbidden(model, endpoint.Type)
}

// UpdateToManyRelationships is a handler function that
// updates the relationship for the provided model.
// A server must respond to PATCH, POST, DELETE request to a URL from to-many
// relationship-link as described below:
// For all request types, the body MUST contain a data member whose value is
// an empty array or an array of resource identifier objects.
func (h *JSONAPIHandler) UpdateToManyRelationships(
	model *ModelHandler,
	endpoint *Endpoint,
) http.HandlerFunc {
	return h.EndpointForbidden(model, endpoint.Type)
}

// PatchToManyRelationships is a handler function used to update to-many
// relationships. If a client makes a PATCH request to a URL from a to-many
// relationship link, the server MUST either completely replace every member
// of the relationship, return an appropiate error response if some resources
// can not be found or accessed, or return a 403 Forbidden response if complete
// replacement is not allowed by the server.
// The request with "data" as an array of typed id's would replace
// every 'relationship' for some Model.
// If the request contains a "data" that is an empty array, it should
// clear every 'relationship' for the model.
func (h *JSONAPIHandler) PatchToManyRelationships(
	model *ModelHandler,
	endpoint *Endpoint,
) http.HandlerFunc {
	return h.EndpointForbidden(model, endpoint.Type)
}

// CreateToManyRelationships is a handler function that is used to handle
// the 'POST' on the relationship link.
// The server Must add the specified members to the relationship unless
// they are already present. If a given type and id is already in the
// relationship, the server MUST NOT add it again.
// If all of the specified resources can be added to, or are already
// present in the relationship, then the server MUST return a succesful
// response
func (h *JSONAPIHandler) CreateToManyRelationships(
	model *ModelHandler,
	endpoint *Endpoint,
) http.HandlerFunc {
	return h.EndpointForbidden(model, endpoint.Type)
}

// DeleteToManyRelationships is a handler function that is used as
// 'DELETE' method on the relationship link.
// If the client makes a DELETE request to a URL from a relationship link
// The server MUST delete the specified members from the relationship or
// return a '403 Forbidden' response. If all of the specified resources are
// ableto be removed from or are already missing from, the relationship then
// the server MUST return a succesful response.
func (h *JSONAPIHandler) DeleteToManyRelationships(
	model *ModelHandler,
	endpoint *Endpoint,
) http.HandlerFunc {
	return h.EndpointForbidden(model, endpoint.Type)
}

func (h *JSONAPIHandler) Delete(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := h.ModelHandlers[model.ModelType]; !ok {
			h.MarshalInternalError(rw)
			return
		}

		/**

		  DELETE: BUILD SCOPE

		*/
		// Create a scope for given delete handler
		scope, err := h.Controller.NewScope(reflect.New(model.ModelType).Interface())
		if err != nil {
			h.log.Errorf("Error while creating scope: '%v' for model: '%v'", err, reflect.TypeOf(model))
			h.MarshalInternalError(rw)
			return
		}

		/**

		  DELETE: GET ID FILTER

		*/
		// Set the ID for given model's scope
		errs, err := h.Controller.GetSetCheckIDFilter(req, scope)
		if err != nil {
			h.errSetIDFilter(scope, err, rw, req)
			return
		}

		if len(errs) > 0 {
			h.MarshalErrors(rw, errs...)
			return
		}

		/**

		  DELETE: PRECHECK PAIRS

		*/

		if !h.AddPrecheckPairFilters(scope, model, endpoint, req, rw, endpoint.PrecheckPairs...) {
			return
		}

		/**

		  DELETE: PRECHECK FILTERS

		*/

		if !h.AddPrecheckFilters(scope, req, rw, endpoint.PrecheckFilters...) {
			return
		}
		/**

		  DELETE: GET RELATIONSHIP FILTERS

		*/
		err = h.GetRelationshipFilters(scope, req, rw)
		if err != nil {
			if hErr := err.(*HandlerError); hErr != nil {
				if !h.handleHandlerError(hErr, rw) {
					return
				}
			} else {
				h.log.Error(err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  DELETE: HOOK BEFORE DELETE

		*/
		scope.NewValueSingle()
		if deleteBefore, ok := scope.Value.(HookBeforeDeleter); ok {
			if err = deleteBefore.JSONAPIBeforeDelete(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Errorf("Unknown error in Hook Before Delete. Path: %v. Error: %v", req.URL.Path, err)
				h.MarshalInternalError(rw)
				return
			}
		}

		/**

		  DELETE: REPOSITORY DELETE

		*/
		repo := h.GetRepositoryByType(model.ModelType)
		if dbErr := repo.Delete(scope); dbErr != nil {
			if dbErr.Compare(unidb.ErrNoResult) && endpoint.HasPrechecks() {
				errObj := ErrInsufficientAccPerm.Copy()
				errObj.Detail = "Given object is not available for this account or it does not exists."
				h.MarshalErrors(rw, errObj)
				return
			}
			h.manageDBError(rw, dbErr)
			return
		}

		/**

		  DELETE: HOOK AFTER DELETE

		*/
		if afterDeleter, ok := scope.Value.(HookAfterDeleter); ok {
			if err = afterDeleter.JSONAPIAfterDelete(scope); err != nil {
				if errObj, ok := err.(*ErrorObject); ok {
					h.MarshalErrors(rw, errObj)
					return
				}
				h.log.Errorf("Error of unknown type during Hook After Delete. Path: %v. Error %v", req.URL.Path, err)
				h.MarshalInternalError(rw)
				return
			}
		}

		rw.WriteHeader(http.StatusNoContent)
	}
}
