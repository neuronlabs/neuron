package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/kucjac/jsonapi/internal"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/uni-db"

	// ctrl "github.com/kucjac/jsonapi/controller"
	"github.com/kucjac/jsonapi/errors"
	ictrl "github.com/kucjac/jsonapi/internal/controller"
	"github.com/kucjac/jsonapi/internal/models"
	"github.com/kucjac/jsonapi/mapping"
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
		s, err := ic.UnmarshalScopeOne(req.Body, (*models.ModelStruct)(m), true)
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

		// SetFlags?
		s.SetFlagsFrom((*models.ModelStruct)(m).Flags(), ic.Flags)

		primVal, err := s.GetFieldValue(im.PrimaryField())
		if err != nil {
			log.Errorf("Getting PrimaryField value failed for the model: '%s'", m.Type().String())
			h.internalError(req, rw)
			return
		}

		// Check if the primary field was unmarshaled
		if s.IsPrimaryFieldSelected() {

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
		errs := (*scope.Scope)(s).ValidateCreate()
		if len(errs) > 0 {
			h.marshalErrors(req, rw, unsetStatus, errs...)
			return
		}

		// Create the scope's value
		if err = (*scope.Scope)(s).Create(); err != nil {
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
				errObj, err := h.c.DBManager().Handle(e)
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
