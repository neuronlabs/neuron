package store

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	ErrStore         = errors.New("store")
	ErrValueNotFound = errors.Wrap(ErrStore, "value not foud")
	ErrInternal      = errors.Wrap(errors.ErrInternal, "store")
)
