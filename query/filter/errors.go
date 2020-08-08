package filter

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrFilter is the minor error classification related to the filters.
	ErrFilter = errors.New("filter")
	// ErrFilterField is the error classification for the filter fields.
	ErrFilterField = errors.Wrap(ErrFilter, "field")
	// ErrFilterFormat is the error classification for the filter format.
	ErrFilterFormat = errors.Wrap(ErrFilter, "format")
	// ErrFilterCollection is the error classification for the filter collection.
	ErrFilterCollection = errors.Wrap(ErrFilter, "collection")
	// ErrFilterValues is the error classification for the filter values.
	ErrFilterValues = errors.Wrap(ErrFilter, "values")
)
