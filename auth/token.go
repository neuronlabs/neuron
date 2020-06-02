package auth

import (
	"time"
)

// Tokener is the interface used for the authorization with the token.
type Tokener interface {
	// InspectToken extracts accountID from the token.
	InspectToken(token string) (accountID string, err error)
	// Token creates the token for provided options.
	Token(options ...TokenOption) (*Token, error)
}

// Token is the authorization token structure.
type Token struct {
	// AccountIdentifier is the account identifier.
	AccountIdentifier string
	// AccessToken is the string access token.
	AccessToken string
	// RefreshToken defines the token
	RefreshToken string
	// CreatedAt is the creation time of the token.
	CreatedAt time.Time
	// ExpiresAt is the expiration time of the token.
	ExpiresAt time.Time
}

// TokenOption is the options used to create the token.
type TokenOptions struct {
	// AccountIdentifier is the account identifier (email, account id etc.)
	AccountID string
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
func TokenAccountID(id string) TokenOption {
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
