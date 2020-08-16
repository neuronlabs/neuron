package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"time"

	"github.com/neuronlabs/neuron/store"
)

// Tokener is the interface used for the authorization with the token.
type Tokener interface {
	// InspectToken extracts claims from the token.
	InspectToken(ctx context.Context, token string) (claims Claims, err error)
	// Token creates the token for provided options.
	Token(account Account, options ...TokenOption) (Token, error)
	// RevokeToken revokes provided 'token'
	RevokeToken(ctx context.Context, token string) error
}

// Token is the authorization token structure.
type Token struct {
	// AccessToken is the string access token.
	AccessToken string
	// RefreshToken defines the token.
	RefreshToken string
	// ExpiresIn defines the expiration time for given access token.
	ExpiresIn int
	// TokenType defines the token type.
	TokenType string
}

// TokenOptions is the options used to create the token.
type TokenOptions struct {
	// ExpirationTime is the expiration time of the token.
	ExpirationTime time.Duration
	// RefreshExpirationTime is the expiration time for refresh token
	RefreshExpirationTime time.Duration
}

// TokenOption is the token options changer function.
type TokenOption func(o *TokenOptions)

// TokenExpirationTime sets the expiration time for the token.
func TokenExpirationTime(d time.Duration) TokenOption {
	return func(o *TokenOptions) {
		o.ExpirationTime = d
	}
}

// TokenRefreshExpirationTime sets the expiration time for the token.
func TokenRefreshExpirationTime(d time.Duration) TokenOption {
	return func(o *TokenOptions) {
		o.RefreshExpirationTime = d
	}
}

// AccessClaims is an interface used for the access token claims. It should store the whole user account.
type AccessClaims interface {
	Claims
	GetAccount() Account
}

// Claims is an interface used for the tokens.
type Claims interface {
	ExpiresIn() int64
	Valid() error
}

// RefreshClaims is an interface used for the refresh token claims. It should store string - user account id.
type RefreshClaims interface {
	GetAccountID() string
	Claims
}

// SigningMethod is an interface used for signing and verify the string.
// This interface is equal to the Signing method of github.com/dgrijalva/jwt-go.
type SigningMethod interface {
	Verify(signingString, signature string, key interface{}) error
	Sign(signingString string, key interface{}) (string, error)
	Alg() string
}

// TokenerOptions are the options that defines the settings for the Tokener.
type TokenerOptions struct {
	// Store is a store used for some authenticator implementations.
	Store store.Store
	// Secret is the authorization secret.
	Secret []byte
	// RsaPrivateKey is used for encoding the token using RSA methods.
	RsaPrivateKey *rsa.PrivateKey
	// EcdsaPrivateKey is used for encoding the token using ECDSA methods.
	EcdsaPrivateKey *ecdsa.PrivateKey
	// TokenExpiration is the default token expiration time.
	TokenExpiration time.Duration
	// RefreshTokenExpiration is the default refresh token expiration time,.
	RefreshTokenExpiration time.Duration
	// SigningMethod is the token signing method.
	SigningMethod SigningMethod
}

// TokenerOption is a function that sets the TokenerOptions.
type TokenerOption func(o *TokenerOptions)

// TokenerSecret is an option that sets Secret in the auth options.
func TokenerSecret(secret []byte) TokenerOption {
	return func(o *TokenerOptions) {
		o.Secret = secret
	}
}

// TokenerRsaPrivateKey is an option that sets RsaPrivateKey in the auth options.
func TokenerRsaPrivateKey(key *rsa.PrivateKey) TokenerOption {
	return func(o *TokenerOptions) {
		o.RsaPrivateKey = key
	}
}

// TokenerEcdsaPrivateKey is an option that sets EcdsaPrivateKey in the auth options.
func TokenerEcdsaPrivateKey(key *ecdsa.PrivateKey) TokenerOption {
	return func(o *TokenerOptions) {
		o.EcdsaPrivateKey = key
	}
}

// TokenerTokenExpiration is an option that sets TokenExpiration in the auth options.
func TokenerTokenExpiration(op time.Duration) TokenerOption {
	return func(o *TokenerOptions) {
		o.TokenExpiration = op
	}
}

// TokenerRefreshTokenExpiration is an option that sets RefreshTokenExpiration in the auth options.
func TokenerRefreshTokenExpiration(op time.Duration) TokenerOption {
	return func(o *TokenerOptions) {
		o.RefreshTokenExpiration = op
	}
}

// TokenerSigningMethod is an option that sets SigningMethod in the auth options.
func TokenerSigningMethod(op SigningMethod) TokenerOption {
	return func(o *TokenerOptions) {
		o.SigningMethod = op
	}
}
