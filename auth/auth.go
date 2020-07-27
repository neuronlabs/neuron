package auth

import (
	"context"
	"time"
)

// Authorizer is the interface used to authorize resources.
type Authorizer interface {
	// Verify if the 'accountID' is allowed to access the resource. The resourceID is a unique resource identifier.
	Verify(ctx context.Context, accountID, resource interface{}) error
}

// Authenticator is the interface used to authenticate the username and password.
type Authenticator interface {
	Authenticate(ctx context.Context, username, password string) (accountID interface{}, err error)
}

// Options are the authorization service options.
type Options struct {
	PasswordCost int
	// Secret is the authorization secret.
	Secret string
	// PublicKey is used for decoding the token public key.
	PublicKey string
	// PrivateKey is used for encoding the token private key.
	PrivateKey string
	// TokenExpiration is the default token expiration time.
	TokenExpiration time.Duration
	// RefreshTokenExpiration is the default refresh token expiration time.
	RefreshTokenExpiration time.Duration
}

// Option is a function used to set authentication options.
type Option func(o *Options)

type accountKey struct{}

// CtxWithAccountID stores account id in the context.
func CtxWithAccountID(ctx context.Context, accountID interface{}) context.Context {
	return context.WithValue(ctx, accountKey{}, accountID)
}

// CtxAccountID gets the account id from the context 'ctx'. If the context doesn't contain account id the
// function returns empty string.
func CtxAccountID(ctx context.Context) interface{} {
	return ctx.Value(accountKey{})
}
