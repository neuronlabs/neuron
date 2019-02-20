package handler

import (
	"github.com/kucjac/jsonapi/pkg/config"
	ctrl "github.com/kucjac/jsonapi/pkg/controller"
	"github.com/kucjac/jsonapi/pkg/encoding/jsonapi"
	"github.com/kucjac/jsonapi/pkg/errors"
	"github.com/kucjac/jsonapi/pkg/internal"
	ictrl "github.com/kucjac/jsonapi/pkg/internal/controller"
	ipagination "github.com/kucjac/jsonapi/pkg/internal/query/paginations"
	iscope "github.com/kucjac/jsonapi/pkg/internal/query/scope"
	"github.com/kucjac/jsonapi/pkg/log"
	"github.com/kucjac/jsonapi/pkg/query/pagination"
	"github.com/kucjac/jsonapi/pkg/query/scope"
	"github.com/kucjac/uni-db"
	"strings"
	// ictrl "github.com/kucjac/jsonapi/pkg/internal/controller"
	// "github.com/kucjac/jsonapi/pkg/mapping"
	"net/http"
)

const (
	unsetStatus int = 0
)

// GatewayHandler is the structure that allows the Gateway service to handle API CRUD operations
type Handler struct {
	c              *ctrl.Controller
	ListPagination *pagination.Pagination
}

// New creates new route handler for the gateway
func New(c *ctrl.Controller, defaultPagination *config.Pagination) *Handler {
	h := &Handler{c: c}
	if defaultPagination != nil {
		p := ipagination.NewFromConfig(defaultPagination)
		if err := p.Check(); err != nil {
			log.Errorf("Invalid default pagination provided for the handler.")
		} else {
			h.ListPagination = (*pagination.Pagination)(p)
			log.Debugf("Handler set default Pagination to: %s", h.ListPagination)
		}
	}
	return h
}

func (h *Handler) getErrorsStatus(errs ...*errors.ApiError) int {
	if len(errs) > 0 {
		return errs[0].IntStatus()
	}
	return 500
}

func (h *Handler) internalError(rw http.ResponseWriter, info ...string) {
	err := errors.ErrInternalError.Copy()
	if len(info) > 0 {
		err.Detail = strings.Join(info, ".")
	}
	h.marshalErrors(rw, http.StatusInternalServerError, err)
}

func (h *Handler) handleDBError(err error, rw http.ResponseWriter) {
	switch e := err.(type) {
	case *unidb.Error:
		log.Debugf("DBError is *unidb.Error: %v", e)

		// get prototype for the provided error
		proto, er := e.GetPrototype()
		if er != nil {
			log.Errorf("*unidb.Error.GetPrototype (%#v) failed. %v", e, err)
			h.internalError(rw)
			return
		}

		// if the error is unspecified or it is internal error marshal as internal error
		if proto == unidb.ErrUnspecifiedError || proto == unidb.ErrInternalError {
			log.Errorf("*unidb.ErrUnspecified. Message: %v", e.Message)
			h.internalError(rw)
			return
		}

		// handle the db error
		errObj, err := h.c.DBManager().Handle(e)
		if err != nil {
			log.Errorf("DBManager Handle failed for error: %v. Err: %v", e, err)
			h.internalError(rw)
			return
		}

		h.marshalErrors(rw, errObj.IntStatus(), errObj)
	case *errors.ApiError:
		log.Debugf("Create failed: %v", e)
		h.marshalErrors(rw, unsetStatus, e)
	default:
		log.Errorf("Unspecified Repository error: %v", e)
		h.internalError(rw)
	}
}

// marshalErrors marshals the api errors into the response writer
// if the status is not set it gets the status from the provided errors
func (h *Handler) marshalErrors(
	rw http.ResponseWriter,
	status int,
	errs ...*errors.ApiError,
) {
	h.setContentType(rw)
	if status == 0 {
		status = h.getErrorsStatus(errs...)
	}
	rw.WriteHeader(status)
	jsonapi.MarshalErrors(rw, errs...)
}

func (h *Handler) marshalScope(
	s *iscope.Scope,
	rw http.ResponseWriter,
) {
	h.setContentType(rw)
	payload, err := (*ictrl.Controller)(h.c).MarshalScope(s)
	if err != nil {
		log.Errorf("[REQ-SCOPE-ID][%s] Marshaling Scope failed. Err: %v", (*scope.Scope)(s).ID().String(), err)
		h.internalError(rw)
		return
	}
	if err = ictrl.MarshalPayload(rw, payload); err != nil {
		log.Errorf("Marshaling payload failed: '%v'", err)
		h.internalError(rw)
	}
}

func (h *Handler) setContentType(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", internal.MediaType)
}
