package filter

import (
	"github.com/neuronlabs/neuron/errors"
)

// MnrFilter is the minor error classification related to the filters.
var (
	MjrFilter errors.Major
	// ClassFilterField is the error classification for the filter fields.
	ClassFilterField errors.Class
	// ClassFilterFormat is the error classification for the filter format.
	ClassFilterFormat errors.Class
	// ClassFilterCollection is the error classification for the filter collection.
	ClassFilterCollection errors.Class
	// ClassFilterValues is the error classification for the filter values.
	ClassFilterValues errors.Class

	// ClassInternal is an internal error classification.
	ClassInternal errors.Class
)

func init() {
	MjrFilter = errors.MustNewMajor()
	ClassFilterField = errors.MustNewMajorClass(MjrFilter)
	ClassFilterFormat = errors.MustNewMajorClass(MjrFilter)
	ClassFilterCollection = errors.MustNewMajorClass(MjrFilter)

	ClassInternal = errors.MustNewMajorClass(errors.MjrInternal)
}
