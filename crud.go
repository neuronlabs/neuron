package jsonapi

import (
	"context"
	"github.com/kucjac/uni-db"
	"net/http"
	"reflect"
)

func (h *Handler) getRelationshipScope(
	ctx context.Context,
	scope *Scope,
	relField *StructField,
	relPrim reflect.Value,
) *Scope {
	relScope := newRootScope(relField.relatedStruct)

	relScope.ctx = ctx
	relScope.logger = scope.logger
	prim := relPrim.Interface()

	sl, err := convertToSliceInterface(&prim)
	if err != nil {
		relScope.setIDFilterValues(prim)

	} else {

		relScope.setIDFilterValues(sl...)
	}

	return relScope
}

func (h *Handler) patchRelationScope(
	ctx context.Context,
	scope *Scope,
	primVal, relPrim reflect.Value,
	relField *StructField,
) (err error) {
	relScope := h.getRelationshipScope(ctx, scope, relField, relPrim)
	relScope.newValueSingle()

	h.log.Debugf("patchRelationScope relField: %v, relScope: %v.", relField, relScope.Struct)

	// add foreignkey as the unmarshaled field
	// get the value of the relationshipID

	fkValue, err := relScope.getFieldValue(relField.relationship.ForeignKey)
	if err != nil {
		h.log.Errorf("Unexpected error occurred. relScope.getFieldValue failed. Err: %v, Scope: %#v, relScope: %#v, relation: %#v", err, scope, relScope, relField.relationship)
		return
	}

	if fkValue.Type() != primVal.Type() {
		h.log.Errorf("CreateHandler. Unexpected error occurred. Fun: Patch BelongsTo Relationship. ForeignKey value is of different type than the primary of the model. Model: %v, FK: %v PK: %v", fkValue.Type(), primVal.Type())
		return
	}

	fkValue.Set(primVal)
	repo := h.GetRepositoryByType(relField.relatedModelType)
	return repo.Patch(relScope)
}

func (h *Handler) PatchRelated(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
	return h.EndpointForbidden(model, PatchRelated)
}

// PatchRelationship is a handler function that is used to update
// relationships independently at URLs from relationship links.
// Reference:
//	- http://jsonapi.org/format/#crud-updating-relationships

func (h *Handler) PatchRelationship(model *ModelHandler, endpoint *Endpoint) http.HandlerFunc {
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
func (h *Handler) UpdateToOneRelationships(
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
func (h *Handler) UpdateToManyRelationships(
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
func (h *Handler) PatchToManyRelationships(
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
func (h *Handler) CreateToManyRelationships(
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
func (h *Handler) DeleteToManyRelationships(
	model *ModelHandler,
	endpoint *Endpoint,
) http.HandlerFunc {
	return h.EndpointForbidden(model, endpoint.Type)
}

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
