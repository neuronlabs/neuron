package auth

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrAuthorization is the major authorization errors.
	MjrAuthorization errors.Major

	// ClassForbidden is the error classification when authorization fails.
	ClassForbidden errors.Class
	// ClassTokenExpired is the error classification when the token expired.
	ClassTokenExpired errors.Class
	// ClassInvalidRole is the error classification when the role is not valid.
	ClassInvalidRole errors.Class
	// ClassInvalidSecret is the error classification when provided secret is not valid.
	ClassInvalidSecret errors.Class
	// ClassAccountNotFound is the error classification when account is not found.
	ClassAccountNotFound errors.Class
)

func init() {
	MjrAuthorization = errors.MustNewMajor()

	ClassForbidden = errors.MustNewMajorClass(MjrAuthorization)
	ClassTokenExpired = errors.MustNewMajorClass(MjrAuthorization)
	ClassInvalidRole = errors.MustNewMajorClass(MjrAuthorization)
	ClassInvalidSecret = errors.MustNewMajorClass(MjrAuthorization)
}
