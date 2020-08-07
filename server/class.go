package server

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	ClassInternal errors.Class

	MjrServer         errors.Major
	ClassOptions      errors.Class
	ClassURIParameter errors.Class
	// MnrHeader errors.Minor
	MnrHeader                  errors.Minor
	ClassNotAcceptable         errors.Class
	ClassUnsupportedHeader     errors.Class
	ClassMissingRequiredHeader errors.Class
	ClassHeaderValue           errors.Class
)

func init() {
	ClassInternal = errors.MustNewMajorClass(errors.MjrInternal)

	MjrServer = errors.MustNewMajor()
	ClassOptions = errors.MustNewMajorClass(MjrServer)
	ClassURIParameter = errors.MustNewMajorClass(MjrServer)

	MnrHeader = errors.MustNewMinor(MjrServer)
	ClassNotAcceptable = errors.MustNewMinorClass(MjrServer, MnrHeader)
	ClassUnsupportedHeader = errors.MustNewMinorClass(MjrServer, MnrHeader)
	ClassMissingRequiredHeader = errors.MustNewMinorClass(MjrServer, MnrHeader)
	ClassHeaderValue = errors.MustNewMinorClass(MjrServer, MnrHeader)
}
