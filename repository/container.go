package repository

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrRepository is the major error repository classification.
	MjrRepository errors.Major
	// ClassNotImplements is the error classification for the repositories that doesn't implement some interface.
	ClassNotImplements errors.Class
	// ClassConnection is the error classification related with repository connection.
	ClassConnection errors.Class
	// ClassAuthorization is the error classification related with repository authorization.
	ClassAuthorization errors.Class
	// ClassReservedName is the error classification related with using reserved name.
	ClassReservedName errors.Class
	// ClassService is the error related with the repository service.
	ClassService errors.Class
)

func init() {
	// Initialize error classes.
	MjrRepository = errors.MustNewMajor()
	ClassNotImplements = errors.MustNewMajorClass(MjrRepository)
	ClassAuthorization = errors.MustNewMajorClass(MjrRepository)
	ClassConnection = errors.MustNewMajorClass(MjrRepository)
	ClassReservedName = errors.MustNewMajorClass(MjrRepository)
	ClassService = errors.MustNewMajorClass(MjrRepository)
}
