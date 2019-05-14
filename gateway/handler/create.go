package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/kucjac/uni-db"
	"github.com/neuronlabs/neuron/encoding/jsonapi"
	"github.com/neuronlabs/neuron/internal"
	iscope "github.com/neuronlabs/neuron/internal/query/scope"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query/scope"

	// ctrl "github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/errors"
	ictrl "github.com/neuronlabs/neuron/internal/controller"
	"github.com/neuronlabs/neuron/internal/models"
	"github.com/neuronlabs/neuron/mapping"
	"net/http"
	// ers "github.com/pkg/errors"
)

func (h *Handler) HandleCreate(m *mapping.ModelStruct) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Debugf("[CREATE] begins for model: '%s'", m.Type().String())
		defer func() { log.Debugf("[CREATE] finished for model: '%s'", m.Type().String()) }()

		ic := (*ictrl.Controller)(h.c)
		im := (*models.ModelStruct)(m)

		// Unmarshal the incoming data

		s, err := jsonapi.UnmarshalManyScopeC(h.c, req.Body, m)
		if err != nil {
			switch e := err.(type) {
			case *errors.ApiError:
				log.Debugf("UnmarshalScope One failed: %v", e)
				h.marshalErrors(req, rw, unsetStatus, e)
			default:
				log.Errorf("HandlerCreate UnmarshalScope for model: '%v' internal error: %v", m.Type().String(), e)
				h.internalError(req, rw)
			}
			return
		}

		// set controller into scope's context
		ctx := context.WithValue(s.Context(), internal.ControllerIDCtxKey, h.c)
		s.WithContext(ctx)

		log.Debugf("[REQ-SCOPE-ID] %s", (*scope.Scope)(s).ID().String())

		// SetFlags

		is := (*iscope.Scope)(s)
		is.SetFlagsFrom((*models.ModelStruct)(m).Flags(), ic.Flags)

		primVal, err := is.GetFieldValue(im.PrimaryField())
		if err != nil {
			log.Errorf("Getting PrimaryField value failed for the model: '%s'", m.Type().String())
			h.internalError(req, rw)
			return
		}

		// Check if the primary field was unmarshaled
		if is.IsPrimaryFieldSelected() {

			// the handler allows only the 'UUID' generated client-id
			if im.AllowClientID() {
				_, err := uuid.Parse(primVal.String())
				if err != nil {
					log.Debugf("Client Generated PrimaryValue is not a valid UUID. %v", err)
					e := errors.ErrInvalidJSONFieldValue.Copy()
					e.Detail = "Client-Generated ID must be a valid UUID"
					h.marshalErrors(req, rw, unsetStatus, e)
					return
				}
			} else {
				log.Debugf("Client Generated ID not allowed for the model: '%s'", im.Type().String())
				e := errors.ErrInvalidJSONFieldValue.Copy()
				e.Detail = "Client-Generated ID is not allowed for this model."
				h.marshalErrors(req, rw, unsetStatus, e)
				return
			}
		}

		// Validate the input entry
		errs := s.ValidateCreate()
		if len(errs) > 0 {
			h.marshalErrors(req, rw, unsetStatus, errs...)
			return
		}

		// Create the scope's value
		if err = s.Create(); err != nil {
			switch e := err.(type) {
			case *unidb.Error:
				log.Debugf("DBError is *unidb.Error: %v", e)

				// get prototype for the provided error
				proto, er := e.GetPrototype()
				if er != nil {
					log.Errorf("*unidb.Error.GetPrototype (%#v) failed. %v", e, err)
					h.internalError(req, rw)
					return
				}

				// if the error is unspecified or it is internal error marshal as internal error
				if proto == unidb.ErrUnspecifiedError || proto == unidb.ErrInternalError {
					log.Errorf("*unidb.ErrUnspecified. Message: %v", e.Message)
					h.internalError(req, rw)
					return
				}

				// handle the db error
				errObj, err := h.c.DBErrorMapper().Handle(e)
				if err != nil {
					log.Errorf("DBManager Handle failed for error: %v. Err: %v", e, err)
					h.internalError(req, rw)
					return
				}

				h.marshalErrors(req, rw, errObj.IntStatus(), errObj)
			case *errors.ApiError:
				log.Debugf("Create failed: %v", e)
				h.marshalErrors(req, rw, unsetStatus, e)
			default:
				log.Errorf("Unspecified error after create: %v", e)
				h.internalError(req, rw)
			}
			return
		}

		rw.WriteHeader(http.StatusCreated)
		h.setContentType(rw)
		h.marshalScope(s, req, rw)
	})
}
