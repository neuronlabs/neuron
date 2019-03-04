package handler

import (
	"context"
	"github.com/kucjac/jsonapi/internal"
	ictrl "github.com/kucjac/jsonapi/internal/controller"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/jsonapi/mapping"
	"github.com/kucjac/jsonapi/query/scope"
	"net/http"
)

func (h *Handler) HandleGet(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[GET] begins for model: '%s'", m.Type().String())
		defer func() { log.Debugf("[GET] finished for model: '%s'.", m.Type().String()) }()

		ic := (*ictrl.Controller)(h.c)

		s, errs, err := ic.QueryBuilder().BuildScopeSingle(req.Context(), (*models.ModelStruct)(m), req.URL, nil)
		if err != nil {
			log.Errorf("[GET] Building Scope for the request failed: %v", err)
			h.internalError(rw)
			return
		}
		if len(errs) > 0 {
			log.Debugf("Building Get Scope failed. ClientSide Error: %v", errs)
			h.marshalErrors(rw, unsetStatus, errs...)
			return
		}

		// set controller into scope's context
		ctx := context.WithValue(s.Context(), internal.ControllerIDCtxKey, h.c)
		s.WithContext(ctx)

		log.Debugf("[REQ-SCOPE-ID] %s", (*scope.Scope)(s).ID().String())

		/**

		TO DO:

		- rewrite language filters

		*/

		if err := (*scope.Scope)(s).Get(); err != nil {
			h.handleDBError(err, rw)
			return
		}

		h.marshalScope(s, rw)
	})
}
