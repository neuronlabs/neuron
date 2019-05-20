package handler

import (
	"context"
	"github.com/neuronlabs/neuron/internal"
	ictrl "github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/scope"
	"net/http"
)

// HandleGetRelated is the query handler function for getting the related object for the given
// model. The related field must be a relationship.
func (h *Handler) HandleGetRelated(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[GET-RELATED] Begins for model: %s", m.Type())

		defer func() { log.Debugf("[GET-RELATED] Finished for model: %s", m.Type()) }()

		ctx := context.WithValue(req.Context(), internal.ControllerCtxKey, h.c)

		// Build the root scope with the related included field
		rootScope, errs, err := (*ictrl.Controller)(h.c).QueryBuilder().BuildScopeRelated(
			// Context
			ctx,
			// Model
			(*models.ModelStruct)(m),
			// Query
			req.URL,
			// Flags
			(*models.ModelStruct)(m).Flags(), h.c.Flags,
		)
		// Err defines the internal error
		if err != nil {
			log.Errorf("BuildScopeRelated for model: '%s' failed: %v", m.Type(), err)
			log.Debugf("URL: '%s'", req.URL.String())
			h.internalError(req, rw)
			return
		}

		// Check ClientSide errors
		if len(errs) > 0 {
			log.Debugf("ClientSide Errors. URL: %s. %v", req.URL.String(), errs)
			h.marshalErrors(req, rw, unsetStatus, errs...)
			return
		}

		// check if the related field is included into the scope's value
		if len(rootScope.IncludedFields()) != 1 {
			log.Errorf("GetRelated: RootScope doesn't have any included fields. Model: '%s', Query: '%s'", m.Type(), req.URL.String())
			h.internalError(req, rw)
			return
		}

		log.Debugf("Succesfully created RootScope with related field: '%s'", rootScope.IncludedFields()[0].ApiName())

		//	Get the root scope values
		if err := (*scope.Scope)(rootScope).Get(); err != nil {
			h.handleDBError(req, err, rw)
			return
		}

		includedScopes := rootScope.IncludedScopes()
		if len(includedScopes) != 1 {
			log.Errorf("GetRelated query for model: '%s'. The rootScope doesn't have any included scopes. Query: %s", m.Type(), req.URL.String())
			h.internalError(req, rw)
			return
		}

		relatedScope := includedScopes[0]
		log.Debugf("Marshaling relatedScope.")

		h.marshalScope((*scope.Scope)(relatedScope), req, rw)
	})
}
