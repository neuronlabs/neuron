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
	// ErrInternal is an internal error for the stores.
	ErrInternal = errors.Wrap(errors.ErrInternal, "store")
)
