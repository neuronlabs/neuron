package handler

import (
	"context"
	"github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/jsonapi/flags"
	"github.com/kucjac/jsonapi/internal"
	ictrl "github.com/kucjac/jsonapi/internal/controller"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/internal/query"
	"github.com/kucjac/jsonapi/internal/query/filters"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/jsonapi/mapping"
	"github.com/kucjac/jsonapi/query/scope"
	"net/http"
)

// HandlePatch handles the patch request for the models
func (h *Handler) HandlePatch(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[PATCH] Begins for model: '%s'", m.Type().String())
		defer func() { log.Debugf("[PATCH] Finished for model: '%s'", m.Type().String()) }()

		// Unmarshal the incoming data into the scope
		s, err := (*ictrl.Controller)(h.c).UnmarshalScopeOne(
			req.Body,
			(*models.ModelStruct)(m),
			true,
		)
		if err != nil {
			switch e := err.(type) {
			case *errors.ApiError:
				h.marshalErrors(rw, unsetStatus, e)
			default:
				log.Errorf("Unmarshaling Scope one failed. %v", err)
				h.internalError(rw)

			}
			return
		}

		// set controller into scope's context
		ctx := context.WithValue(s.Context(), internal.ControllerIDCtxKey, h.c)
		s.WithContext(ctx)

		// SetFlags for the scope
		s.SetFlagsFrom((*models.ModelStruct)(m).Flags(), h.c.Flags)
		id, err := query.GetAndSetID(req, s)
		if err != nil {
			if err == internal.IErrInvalidType {
				log.Errorf("Invalid type provided for the GetAndSetID and model: %v", m.Type().String())
				h.internalError(rw)
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

			h.marshalErrors(rw, unsetStatus, errObj)
			return
		}

		// Set Primary Field filter to 'id'
		err = s.AddFilterField(
			filters.NewFilter(
				s.Struct().PrimaryField(),
				filters.NewOpValuePair(
					filters.OpEqual,
					id,
				),
			),
		)
		if err != nil {
			log.Errorf("Adding primary filter field for model: '%s' failed: '%v'", m.Type().String(), err)
			h.internalError(rw)
			return
		}

		// Validate the Patch model
		if errs := (*scope.Scope)(s).ValidatePatch(); len(errs) > 0 {
			h.marshalErrors(rw, unsetStatus, errs...)
			return
		}

		// Patch the values
		if err := (*scope.Scope)(s).Patch(); err != nil {
			log.Debugf("Patching the scope: '%s' failed. %v", (*scope.Scope)(s).ID().String(), err)
			h.handleDBError(err, rw)
			return
		}

		v, ok := s.Flags().Get(flags.ReturnPatchContent)
		if ok && v {
			// SetAllFields in the fieldset
			s.SetAllFields()
			if err := (*scope.Scope)(s).Get(); err != nil {
				log.Debugf("Getting the Patched scope: '%s' failed. %v", (*scope.Scope)(s).ID().String(), err)
				h.handleDBError(err, rw)
				return
			}
			h.marshalScope(s, rw)
		} else {
			rw.WriteHeader(http.StatusNoContent)
		}
	})
}
