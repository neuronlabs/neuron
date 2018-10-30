package jsonapi

import (
	"context"
	"github.com/pkg/errors"
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
	fk := relField.relationship.ForeignKey

	// get the value of the relationshipID
	fkValue, err := relScope.getFieldValue(fk)
	if err != nil {
		h.log.Errorf("Unexpected error occurred. relScope.getFieldValue failed. Err: %v, Scope: %#v, relScope: %#v, relation: %#v", err, scope, relScope, relField.relationship)
		return
	}

	if fkValue.Type() != primVal.Type() {
		err = errors.Errorf("CreateHandler. Unexpected error occurred. Fun: Patch BelongsTo Relationship. ForeignKey value is of different type than the primary of the model. Model: %v, FK: %v PK: %v", fkValue.Type(), primVal.Type())
		h.log.Error(err)
		return
	}

	fkValue.Set(primVal)

	relScope.SelectedFields = append(relScope.SelectedFields, fk)
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
