package log

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrLogger is the major logger error classification.
	ErrLogger = errors.New("logger")
	// ErrInvalidLogger is the classification for the invalid logger.
	ErrInvalidLogger = errors.New("invalid")
)
