package auth

import (
	"github.com/neuronlabs/neuron/errors"
)

var (
	// ErrAuth is the account general error.
	ErrAuth = errors.New("Auth")
	// ErrAccountNotFound is the error classification when account is not found.
	ErrAccountNotFound = errors.Wrap(ErrAuth, "account not found")
	// ErrAccountNotValid is the error for invalid accounts.
	ErrAccountNotValid = errors.Wrap(ErrAuth, "account not valid")
	// ErrAccountModelNotDefined is an error that occurs when the account model is not defined.
	ErrAccountModelNotDefined = errors.Wrap(ErrAuth, "account model not defined")
	// ErrAccountAlreadyExists is an error when the account already exists.
	ErrAccountAlreadyExists = errors.Wrap(ErrAuth, "an account with provided username already exists")
	// ErrInternalError is an auth package internal error.
	ErrInternalError = errors.Wrap(errors.ErrInternal, "auth")
	// ErrAuthentication is an error related with authentication.
	ErrAuthentication = errors.Wrap(ErrAuth, "authentication")
	// ErrInvalidUsername is an error for invalid usernames.
	ErrInvalidUsername = errors.Wrap(ErrAuthentication, "invalid username")
	// ErrInvalidPassword is the error classification when provided secret is not valid.
	ErrInvalidPassword = errors.Wrap(ErrAuthentication, "provided invalid secret")
	// ErrNoRequiredOption is the error classification while there is no required option.
	ErrNoRequiredOption = errors.Wrap(ErrAuthentication, "provided no required option")
	// ErrInitialization is the error classification while initializing the structures.
	ErrInitialization = errors.New("auth initialization failed")
	// ErrInvalidSecret is an error for initialization invalid secret.
	ErrInvalidSecret = errors.Wrap(ErrInitialization, "invalid secret")
	// ErrInvalidRSAKey is an error for initialization with an invalid RSA key.
	ErrInvalidRSAKey = errors.Wrap(ErrInitialization, "invalid RSA key")
	// ErrInvalidECDSAKey is an error for initialization with an invalid ECDSA key.
	ErrInvalidECDSAKey = errors.Wrap(ErrInitialization, "invalid ECDSA key")
	// ErrToken is the error for invalid token.
	ErrToken = errors.Wrap(ErrAuthentication, "invalid token")
	// ErrTokenRevoked is the error for invalid token.
	ErrTokenRevoked = errors.Wrap(ErrToken, "revoked")
	// ErrTokenExpired is an error related to expired token.
	ErrTokenExpired = errors.Wrap(ErrToken, "expired")
	// ErrTokenNotValidYet is an error related to the token that is not valid yet.
	ErrTokenNotValidYet = errors.Wrap(ErrToken, "not valid yet")
)

var (
	// ErrAuthorization is the major authorization errors.
	ErrAuthorization = errors.Wrap(ErrAuth, "authorization")
	// ErrAuthorizationScope is an error related to the authorization scope.
	ErrAuthorizationScope = errors.Wrap(ErrAuthorization, "scope")
	// ErrAuthorizationHeader is an error related to authorization header.
	ErrAuthorizationHeader = errors.Wrap(ErrAuthorization, "header")
	// ErrForbidden is the error classification when authorization fails.
	ErrForbidden = errors.Wrap(ErrAuthorization, "forbidden")
	// ErrInvalidRole is the error classification when the role is not valid.
	ErrInvalidRole = errors.Wrap(ErrAuthorization, "invalid role")
	// ErrRoleAlreadyGranted is the error when the role is already granted.
	ErrRoleAlreadyGranted = errors.Wrap(ErrAuthorization, "role already granter")
)
