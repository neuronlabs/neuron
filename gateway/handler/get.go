package handler

import (
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query"

	"net/http"
)

// HandleGet handles the Get single object request
func (h *Handler) HandleGet(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[GET] begins for model: '%s'", m.Type().String())
		defer func() { log.Debugf("[GET] finished for model: '%s'.", m.Type().String()) }()

		s, errs, err := h.Builder.BuildScopeSingle(req.Context(), (*models.ModelStruct)(m), req.URL, nil)
		if err != nil {
			log.Errorf("[GET] Building Scope for the request failed: %v", err)
			h.internalError(req, rw)
			return
		}
		if len(errs) > 0 {
			log.Debugf("Building Get Scope failed. ClientSide Error: %v", errs)
			h.marshalErrors(req, rw, unsetStatus, errs...)
			return
		}

		// set controller into scope's context
		s.Store[internal.ControllerCtxKey] = h.c

		log.Debugf("[REQ-SCOPE-ID] %s", (*query.Scope)(s).ID().String())

		/**

		TO DO:

		- rewrite language filters

		*/

		if err := (*query.Scope)(s).Get(); err != nil {
			h.handleDBError(req, err, rw)
			return
		}

		h.marshalScope((*query.Scope)(s), req, rw)
	})
}
