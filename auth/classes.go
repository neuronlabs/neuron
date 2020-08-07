package auth

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// MjrAuthorization is the major authorization errors.
	MjrAuthorization errors.Major
	ClassScope       errors.Class

	// ClassInternal is the errors classification for internal auth errors.
	ClassInternal errors.Class
	// ClassForbidden is the error classification when authorization fails.
	ClassForbidden errors.Class
	// ClassTokenExpired is the error classification when the token expired.
	MnrToken          errors.Minor
	ClassInvalidToken errors.Class
	ClassTokenExpired errors.Class
	// ClassInvalidRole is the error classification when the role is not valid.
	ClassInvalidRole errors.Class
	// ClassInvalidSecret is the error classification when provided secret is not valid.
	ClassInvalidSecret errors.Class
	// ClassAccountNotFound is the error classification when account is not found.
	ClassAccountNotFound     errors.Class
	ClassAuthorizationHeader errors.Class
	// ClassInitialization is the error classification while initializing the structures.
	ClassInitialization errors.Class
	// ClassNoRequiredOption is the error classification while there is no required option.
	ClassNoRequiredOption errors.Class
)

func init() {
	ClassInternal = errors.MustNewMajorClass(errors.MjrInternal)

	MjrAuthorization = errors.MustNewMajor()
	ClassForbidden = errors.MustNewMajorClass(MjrAuthorization)
	ClassScope = errors.MustNewMajorClass(MjrAuthorization)
	ClassInvalidRole = errors.MustNewMajorClass(MjrAuthorization)
	ClassInvalidSecret = errors.MustNewMajorClass(MjrAuthorization)
	ClassAuthorizationHeader = errors.MustNewMajorClass(MjrAuthorization)
	ClassInitialization = errors.MustNewMajorClass(MjrAuthorization)
	ClassNoRequiredOption = errors.MustNewMajorClass(MjrAuthorization)

	MnrToken = errors.MustNewMinor(MjrAuthorization)
	ClassInvalidToken = errors.MustNewMinorClass(MjrAuthorization, MnrToken)
	ClassTokenExpired = errors.MustNewMinorClass(MjrAuthorization, MnrToken)

}
