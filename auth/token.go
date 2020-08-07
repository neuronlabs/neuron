package auth

import (
	"context"
	"time"
)

// Tokener is the interface used for the authorization with the token.
type Tokener interface {
	// InspectToken extracts claims from the token.
	InspectToken(token string) (claims interface{}, err error)
	// Token creates the token for provided options.
	Token(ctx context.Context, options ...TokenOption) (Token, error)
	// RefreshToken creates the token based on provided 'refreshToken'.
	RefreshToken(ctx context.Context, refreshToken string) (Token, error)
}

// Token is the authorization token structure.
type Token struct {
	// AccessToken is the string access token.
	AccessToken string
	// RefreshToken defines the token
	RefreshToken string
}

// TokenOption is the options used to create the token.
type TokenOptions struct {
	// AccountIdentifier is the account identifier (email, account id etc.)
	AccountID interface{}
	// RefreshToken is the token used to refresh the new token.
	RefreshToken string
	// ExpirationTime is the expiration time of the token.
	ExpirationTime time.Duration
}

// TokenOption is the token options changer function.
type TokenOption func(o *TokenOptions)

// TokenExpirationTime sets the expiration time for the token.
func TokenExpirationTime(d time.Duration) TokenOption {
	return func(o *TokenOptions) {
		o.ExpirationTime = d
	}
}

// TokenAccountID sets account identifier for the token.
func TokenAccountID(id interface{}) TokenOption {
	return func(o *TokenOptions) {
		o.AccountID = id
	}
}

// TokenRefreshToken sets the refresh token in the options.
func TokenRefreshToken(refresh string) TokenOption {
	return func(o *TokenOptions) {
		o.RefreshToken = refresh
	}
}
