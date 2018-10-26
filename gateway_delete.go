package jsonapi

import (
	"github.com/kucjac/uni-db"
	"net/http"
	"reflect"
)

func (h *Handler) Delete(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
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

		h.setScopeFlags(scope, endpoint, model)

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
		err = repo.Delete(scope)
		if err != nil {
			if dbErr, ok := err.(*unidb.Error); ok {
				if dbErr.Compare(unidb.ErrNoResult) && endpoint.HasPrechecks() {
					errObj := ErrInsufficientAccPerm.Copy()
					errObj.Detail = "Given object is not available for this account or it does not exists."
					h.MarshalErrors(rw, errObj)
					return
				}
				h.manageDBError(rw, dbErr)
				return
			} else if errObj, ok := err.(*ErrorObject); ok {
				h.MarshalErrors(rw, errObj)
				return
			}
			h.log.Errorf("Repository Delete, unknown error occurred: %v. Scope: %#v", err, scope)
			h.MarshalInternalError(rw)
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
