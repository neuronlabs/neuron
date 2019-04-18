package handler

import (
	"compress/flate"
	"compress/gzip"
	"github.com/kucjac/jsonapi/config"
	ctrl "github.com/kucjac/jsonapi/controller"
	"github.com/kucjac/jsonapi/encoding/jsonapi"
	"github.com/kucjac/jsonapi/errors"
	"github.com/kucjac/jsonapi/internal"
	ictrl "github.com/kucjac/jsonapi/internal/controller"
	ipagination "github.com/kucjac/jsonapi/internal/query/paginations"
	iscope "github.com/kucjac/jsonapi/internal/query/scope"
	"github.com/kucjac/jsonapi/log"
	"github.com/kucjac/jsonapi/query/pagination"
	"github.com/kucjac/jsonapi/query/scope"
	"github.com/kucjac/uni-db"
	"io"
	"sort"
	"strings"
	// ictrl "github.com/kucjac/jsonapi/internal/controller"
	// "github.com/kucjac/jsonapi/mapping"
	"net/http"
)

const (
	unsetStatus int = 0
)

// Handler is the structure that allows the Gateway service to handle API CRUD operations
type Handler struct {
	c                *ctrl.Controller
	ListPagination   *pagination.Pagination
	CompressionLevel int

	octetTypes [256]octetType
}

// New creates new route handler for the gateway
func New(c *ctrl.Controller, cfg *config.RouterConfig) *Handler {
	h := &Handler{
		c:                c,
		CompressionLevel: cfg.CompressionLevel,
	}
	if cfg.DefaultPagination != nil {
		p := ipagination.NewFromConfig(cfg.DefaultPagination)
		if err := p.Check(); err != nil {
			log.Errorf("Invalid default pagination provided for the handler.")
		} else {
			h.ListPagination = (*pagination.Pagination)(p)
			log.Debugf("Handler set default Pagination to: %s", h.ListPagination)
		}
	}

	for c := 0; c < 256; c++ {
		var t octetType
		isCtl := c <= 31 || c == 127
		isChar := 0 <= c && c <= 127
		isSeparator := strings.IndexRune(" \t\"(),/:;<=>?@[]\\{}", rune(c)) >= 0
		if strings.IndexRune(" \t\r\n", rune(c)) >= 0 {
			t |= isSpace
		}
		if isChar && !isCtl && !isSeparator {
			t |= isToken
		}
		h.octetTypes[c] = t
	}

	return h
}

func (h *Handler) getErrorsStatus(errs ...*errors.ApiError) int {
	if len(errs) > 0 {
		return errs[0].IntStatus()
	}
	return 500
}

func (h *Handler) internalError(req *http.Request, rw http.ResponseWriter, info ...string) {
	err := errors.ErrInternalError.Copy()
	if len(info) > 0 {
		err.Detail = strings.Join(info, ".")
	}
	h.marshalErrors(req, rw, http.StatusInternalServerError, err)
}

func (h *Handler) handleDBError(req *http.Request, err error, rw http.ResponseWriter) {
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
		log.Errorf("Unspecified Repository error: %v", e)
		h.internalError(req, rw)
	}
}

// marshalErrors marshals the api errors into the response writer
// if the status is not set it gets the status from the provided errors
func (h *Handler) marshalErrors(
	req *http.Request,
	rw http.ResponseWriter,
	status int,
	errs ...*errors.ApiError,
) {
	h.setContentType(rw)
	if status == 0 {
		status = h.getErrorsStatus(errs...)
	}
	rw.WriteHeader(status)

	var encoding encodingType
	encodings := h.parseAccept(req.Header, "Accept-Encoding")
	sort.Sort(encodings)

	for _, accept := range encodings.accepts {
		switch accept.value {
		case "gzip":
			encoding = gzipped
			break
		case "defalte":
			encoding = deflated
			break
		}
	}
	var payloadWriter io.Writer

	// TODO: set encoding level in config
	if encoding != noencoding {
		var err error
		switch encoding {
		case gzipped:
			payloadWriter, err = gzip.NewWriterLevel(rw, h.CompressionLevel)
		case deflated:
			payloadWriter, err = flate.NewWriter(rw, h.CompressionLevel)
		}
		if err != nil {
			h.internalError(req, rw)
			return
		}
	} else {
		payloadWriter = rw
	}
	jsonapi.MarshalErrors(payloadWriter, errs...)
}

func (h *Handler) marshalScope(
	s *iscope.Scope,
	req *http.Request,
	rw http.ResponseWriter,
) {
	h.setContentType(rw)
	payload, err := (*ictrl.Controller)(h.c).MarshalScope(s)
	if err != nil {
		log.Errorf("[REQ-SCOPE-ID][%s] Marshaling Scope failed. Err: %v", (*scope.Scope)(s).ID().String(), err)
		h.internalError(req, rw)
		return
	}

	var encoding encodingType
	encodings := h.parseAccept(req.Header, "Accept-Encoding")
	sort.Sort(encodings)

	for _, accept := range encodings.accepts {
		switch accept.value {
		case "gzip":
			encoding = gzipped
			break
		case "defalte":
			encoding = deflated
			break
		}
	}
	var payloadWriter io.Writer

	// TODO: set encoding level in config
	if encoding != noencoding {
		switch encoding {
		case gzipped:
			payloadWriter, err = gzip.NewWriterLevel(rw, h.CompressionLevel)
		case deflated:
			payloadWriter, err = flate.NewWriter(rw, h.CompressionLevel)
		}

		if err != nil {
			h.internalError(req, rw)
			return
		}
	} else {
		payloadWriter = rw
	}

	if err = ictrl.MarshalPayload(payloadWriter, payload); err != nil {
		log.Errorf("Marshaling payload failed: '%v'", err)
		h.internalError(req, rw)
	}

}

type encodingType int8

const (
	noencoding encodingType = iota
	gzipped
	deflated
)

type octetType byte

const (
	isToken octetType = 1 << iota
	isSpace
)

type acceptSpec struct {
	value string
	q     float64
}

type acceptSorter struct {
	accepts []acceptSpec
}

// Len implements sort.Interface
func (a *acceptSorter) Len() int {
	return len(a.accepts)
}

// Swap implements sort.Interface
func (a *acceptSorter) Swap(i, j int) {
	a.accepts[i], a.accepts[j] = a.accepts[j], a.accepts[i]
}

// Less implements sort.Interface
func (a *acceptSorter) Less(i, j int) bool {
	return a.by(&a.accepts[i], &a.accepts[j])
}

func (a *acceptSorter) by(a1, a2 *acceptSpec) bool {
	return a1.q > a2.q
}

func (h *Handler) parseAccept(header http.Header, key string) *acceptSorter {
	var specs []acceptSpec
loop:
	for _, s := range header[key] {
		for {
			var spec acceptSpec
			spec.value, s = h.expectTokenSlash(s)
			if spec.value == "" {
				continue loop
			}
			spec.q = 1.0
			s = h.skipSpace(s)
			if strings.HasPrefix(s, ";") {
				s = h.skipSpace(s[1:])
				if !strings.HasPrefix(s, "q=") {
					continue loop
				}
				spec.q, s = h.expectQuality(s[2:])
				if spec.q < 0.0 {
					continue loop
				}
			}
			specs = append(specs, spec)
			s = h.skipSpace(s)
			if !strings.HasPrefix(s, ",") {
				continue loop
			}
			s = h.skipSpace(s[1:])
		}
	}
	return &acceptSorter{accepts: specs}
}

func (h *Handler) expectQuality(s string) (q float64, rest string) {
	switch {
	case len(s) == 0:
		return -1, ""
	case s[0] == '0':
		q = 0
	case s[0] == '1':
		q = 1
	default:
		return -1, ""
	}
	s = s[1:]
	if !strings.HasPrefix(s, ".") {
		return q, s
	}
	s = s[1:]
	i := 0
	n := 0
	d := 1
	for ; i < len(s); i++ {
		b := s[i]
		if b < '0' || b > '9' {
			break
		}
		n = n*10 + int(b) - '0'
		d *= 10
	}
	return q + float64(n)/float64(d), s[i:]
}

func (h *Handler) expectTokenSlash(s string) (token, rest string) {
	i := 0
	for ; i < len(s); i++ {
		b := s[i]
		if (h.octetTypes[b]&isToken == 0) && b != '/' {
			break
		}
	}
	return s[:i], s[i:]
}

func (h *Handler) skipSpace(s string) (rest string) {
	i := 0
	for ; i < len(s); i++ {
		if h.octetTypes[s[i]]&isSpace == 0 {
			break
		}
	}
	return s[i:]
}

func (h *Handler) setContentType(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", internal.MediaType)
}
