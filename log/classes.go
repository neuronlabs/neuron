package log

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrLogger is the major logger error classification.
	MjrLogger errors.Major
	// ClassInvalidLogger is the classification for the invalid logger.
	ClassInvalidLogger errors.Class
)

func init() {
	MjrLogger = errors.MustNewMajor()
	ClassInvalidLogger = errors.MustNewMajorClass(MjrLogger)
}
