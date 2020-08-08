package repository

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrRepository is the major error repository classification.
	ErrRepository = errors.New("repository")
	// ErrNotImplements is the error classification for the repositories that doesn't implement some interface.
	ErrNotImplements = errors.Wrap(ErrRepository, "not implements")
	// ErrConnection is the error classification related with repository connection.
	ErrConnection = errors.Wrap(ErrRepository, "connection")
	// ErrAuthorization is the error classification related with repository authorization.
	ErrAuthorization = errors.Wrap(ErrRepository, "authorization")
	// ErrReservedName is the error classification related with using reserved name.
	ErrReservedName = errors.Wrap(ErrRepository, "reserved name")
)
