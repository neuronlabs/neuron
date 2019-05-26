package handler

import (
	"compress/flate"
	"compress/gzip"
	"github.com/kucjac/uni-db"
	"github.com/neuronlabs/neuron/config"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/encoding/jsonapi"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/gateway/builder"
	"github.com/neuronlabs/neuron/i18n"
	"github.com/neuronlabs/neuron/internal"
	ipagination "github.com/neuronlabs/neuron/internal/query/paginations"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query"
	"github.com/neuronlabs/neuron/query/pagination"
	"io"
	"net/http"
	"sort"
	"strings"
)

const (
	unsetStatus int = 0
)

// Handler is the structure that allows the Gateway service to handle API CRUD operations
type Handler struct {
	c                *controller.Controller
	ListPagination   *pagination.Pagination
	CompressionLevel int

	octetTypes [256]octetType

	Builder *builder.JSONAPI
}

// New creates new route handler for the gateway
func New(c *controller.Controller, cfg *config.Gateway, isup *i18n.Support) *Handler {

	h := &Handler{
		c:                c,
		Builder:          builder.NewJSONAPI(c, cfg.QueryBuilder, isup),
		CompressionLevel: cfg.Router.CompressionLevel,
	}

	if cfg.Router.DefaultPagination != nil {
		p := ipagination.NewFromConfig(cfg.Router.DefaultPagination)
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
	s *query.Scope,
	req *http.Request,
	rw http.ResponseWriter,
) {
	h.setContentType(rw)

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
	var (
		payloadWriter io.Writer
		err           error
	)

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

	if err := jsonapi.MarshalScopeC(h.c, payloadWriter, s); err != nil {
		log.Errorf("[REQ-SCOPE-ID][%s] Marshaling Scope failed. Err: %v", (*query.Scope)(s).ID().String(), err)
		h.internalError(req, rw)
		return
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
