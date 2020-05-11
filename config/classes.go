package config

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrConfig is the major config error classification.
	MjrConfig errors.Major
	// ClassConfigInvalidValue is the errors classification for invalid config values.
	ClassConfigInvalidValue errors.Class
)

func init() {
	MjrConfig = errors.MustNewMajor()
	ClassConfigInvalidValue = errors.MustNewMajorClass(MjrConfig)
}
