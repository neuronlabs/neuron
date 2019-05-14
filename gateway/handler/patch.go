package handler

import (
	"context"
	"github.com/neuronlabs/neuron/encoding/jsonapi"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/internal"
	"github.com/neuronlabs/neuron/internal/flags"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/internal/query"
	"github.com/neuronlabs/neuron/internal/query/filters"
	iscope "github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/query/scope"
	"net/http"
)

// HandlePatch handles the patch request for the models
func (h *Handler) HandlePatch(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[PATCH] Begins for model: '%s'", m.Type().String())
		defer func() { log.Debugf("[PATCH] Finished for model: '%s'", m.Type().String()) }()

		// Unmarshal the incoming data into the scope
		s, err := jsonapi.UnmarshalSingleScopeC(h.c, req.Body, m)
		if err != nil {
			switch e := err.(type) {
			case *errors.ApiError:
				h.marshalErrors(req, rw, unsetStatus, e)
			default:
				log.Errorf("Unmarshaling Scope one failed. %v", err)
				h.internalError(req, rw)

			}
			return
		}

		// set controller into scope's context
		ctx := context.WithValue(s.Context(), internal.ControllerIDCtxKey, h.c)
		s.WithContext(ctx)

		// SetFlags for the scope
		(*iscope.Scope)(s).SetFlagsFrom((*models.ModelStruct)(m).Flags(), h.c.Flags)
		id, err := query.GetAndSetID(req, (*iscope.Scope)(s))
		if err != nil {
			if err == internal.IErrInvalidType {
				log.Errorf("Invalid type provided for the GetAndSetID and model: %v", m.Type().String())
				h.internalError(req, rw)
				return
			}

			// ClientSide Error
			var errObj *errors.ApiError
			switch e := err.(type) {
			case *errors.ApiError:
				errObj = e
			default:
				errObj = errors.ErrInvalidQueryParameter.Copy()
				errObj.Detail = "Provided invalid 'id' in the query."
			}

			h.marshalErrors(req, rw, unsetStatus, errObj)
			return
		}

		// Set Primary Field filter to 'id'
		err = (*iscope.Scope)(s).AddFilterField(
			filters.NewFilter(
				(*iscope.Scope)(s).Struct().PrimaryField(),
				filters.NewOpValuePair(
					filters.OpEqual,
					id,
				),
			),
		)
		if err != nil {
			log.Errorf("Adding primary filter field for model: '%s' failed: '%v'", m.Type().String(), err)
			h.internalError(req, rw)
			return
		}

		// Validate the Patch model
		if errs := (*scope.Scope)(s).ValidatePatch(); len(errs) > 0 {
			h.marshalErrors(req, rw, unsetStatus, errs...)
			return
		}

		// Patch the values
		if err := (*scope.Scope)(s).Patch(); err != nil {
			log.Debugf("Patching the scope: '%s' failed. %v", (*scope.Scope)(s).ID().String(), err)
			h.handleDBError(req, err, rw)
			return
		}

		v, ok := (*iscope.Scope)(s).Flags().Get(flags.ReturnPatchContent)
		if ok && v {
			// SetAllFields in the fieldset
			(*iscope.Scope)(s).SetAllFields()
			if err := (*scope.Scope)(s).Get(); err != nil {
				log.Debugf("Getting the Patched scope: '%s' failed. %v", (*scope.Scope)(s).ID().String(), err)
				h.handleDBError(req, err, rw)
				return
			}
			h.marshalScope(s, req, rw)
		} else {
			rw.WriteHeader(http.StatusNoContent)
		}
	})
}
