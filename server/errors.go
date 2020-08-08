package server

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	ErrInternal = errors.Wrap(errors.ErrInternal, "server")

	ErrServer        = errors.New("server")
	ErrServerOptions = errors.Wrap(ErrServer, "options")
	ErrURIParameter  = errors.Wrap(ErrServer, "uri parameter")
	// ErrHeader errors.Minor
	ErrHeader                = errors.Wrap(ErrServer, "header")
	ErrHeaderNotAcceptable   = errors.Wrap(ErrHeader, "not acceptable")
	ErrUnsupportedHeader     = errors.Wrap(ErrHeader, "unsupported")
	ErrMissingRequiredHeader = errors.Wrap(ErrHeader, "missing required")
	ErrHeaderValue           = errors.Wrap(ErrHeader, "value")
)
