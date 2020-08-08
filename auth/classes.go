package auth

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrAuthorization is the major authorization errors.
	ErrAuthorization = errors.New("authorization")
	// ErrAuthorizationScope is an error related to the authorization scope.
	ErrAuthorizationScope = errors.Wrap(ErrAuthorization, "scope")
	// ErrAuthorizationHeader is an error related to authorization header.
	ErrAuthorizationHeader = errors.Wrap(ErrAuthorization, "invalid or no header")

	// ErrForbidden is the error classification when authorization fails.
	ErrForbidden = errors.Wrap(ErrAuthorization, "forbidden access")
	// ErrTokenExpired is the error classification when the token expired.
	ErrToken = errors.Wrap(ErrAuthorization, "invalid token")
	// ErrTokenExpired is an error related to expired token.
	ErrTokenExpired = errors.Wrap(ErrToken, "expired")
	// ErrInvalidRole is the error classification when the role is not valid.
	ErrInvalidRole = errors.Wrap(ErrAuthorization, "invalid role")

	// Authentication errors.
	ErrAuthentication = errors.New("authentication")
	// ErrInvalidSecret is the error classification when provided secret is not valid.
	ErrInvalidSecret = errors.Wrap(ErrAuthentication, "provided invalid secret")
	// ErrAccountNotFound is the error classification when account is not found.
	ErrAccountNotFound = errors.Wrap(ErrAuthentication, "account not found")

	// ErrInitialization is the error classification while initializing the structures.
	ErrInitialization = errors.New("auth initialization failed")
	// ErrAuthenticationNoRequiredOption is the error classification while there is no required option.
	ErrAuthenticationNoRequiredOption = errors.Wrap(ErrAuthentication, "provided no required option")
)
