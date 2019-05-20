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

// HandleGetRelationship returns the handler func to the get relationship scope
// for specified relationship field
func (h *Handler) HandleGetRelationship(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[GET-RELATIONSHIP] Begins for model: '%s'", m.Type())
		defer func() { log.Debugf("[GET-RELATIONSHIP] Finished for model: '%s'", m.Type()) }()

		// Prepare Context for the scopes
		ctx := context.WithValue(req.Context(), internal.ControllerCtxKey, h.c)
		rootScope, errs, err := (*ictrl.Controller)(h.c).QueryBuilder().BuildScopeRelationship(
			ctx,
			(*models.ModelStruct)(m),
			req.URL,
			(*models.ModelStruct)(m).Flags(), h.c.Flags,
		)
		// Err defines the internal error
		if err != nil {
			log.Errorf("BuildScopeRelationship for model: '%s' failed: %v", m.Type(), err)
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

		if err := (*scope.Scope)(rootScope).Get(); err != nil {
			log.Debugf("Getting the RootScope failed: %v", err)
			h.handleDBError(req, err, rw)
			return
		}

		relScope, err := rootScope.GetRelationshipScope()
		if err != nil {
			log.Errorf("Error while Getting RelatinoshipScope for model: %v", m.Type())
			h.internalError(req, rw)
			return
		}

		h.marshalScope((*scope.Scope)(relScope), req, rw)
	})
}
