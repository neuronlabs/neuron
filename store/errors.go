package store

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrStore is the main store error.
	ErrStore = errors.New("store")
	// ErrRecordNotFound is the error when the value stored with 'key' is not found.
	// This should be implemented by all stores.
	ErrRecordNotFound = errors.Wrap(ErrStore, "record not found")
	// ErrInitialization is the error returned when the store have some issues with initialization.
	ErrInitialization = errors.Wrap(ErrStore, "initialization")
	// ErrInternal is an internal error for the stores.
	ErrInternal = errors.Wrap(errors.ErrInternal, "store")
)
