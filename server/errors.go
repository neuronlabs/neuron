package server

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrInternal is an internal server error.
	ErrInternal = errors.Wrap(errors.ErrInternal, "server")

	// ErrServer is an error related with server.
	ErrServer = errors.New("server")
	// ErrServerOptions is an error related with server options.
	ErrServerOptions = errors.Wrap(ErrServer, "options")
	// ErrURIParameter is an error related with uri parameters.
	ErrURIParameter = errors.Wrap(ErrServer, "uri parameter")
	// ErrHeader errors.Minor
	ErrHeader = errors.Wrap(ErrServer, "header")
	// ErrHeaderNotAcceptable is an error related to not acceptable request header.
	ErrHeaderNotAcceptable = errors.Wrap(ErrHeader, "not acceptable")
	// ErrUnsupportedHeader is an error related to unsupported request header.
	ErrUnsupportedHeader = errors.Wrap(ErrHeader, "unsupported")
	// ErrMissingRequiredHeader is an error related to missing required header.
	ErrMissingRequiredHeader = errors.Wrap(ErrHeader, "missing required")
	// ErrHeaderValue is an error related to request header value.
	ErrHeaderValue = errors.Wrap(ErrHeader, "value")
)
