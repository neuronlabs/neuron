package auth

import (
	"context"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/mapping"
)

// Account is an interface for the authenticate account models. It needs to get/set username and password.
type Account interface {
	mapping.Model
	// GetUsername gets the current account username.
	GetUsername() string
	// SetUsername sets the account username.
	SetUsername(username string)
	// GetPasswordHash sets the password hash for the account.
	GetPasswordHash() []byte
	// SetPasswordHash sets the password hash for given account.
	SetPasswordHash(hash []byte)
	// UsernameField gets the account username field name.
	UsernameField() string
	// PasswordHashField gets the hashed password field name.
	PasswordHashField() string
}

// SaltSetter is an interface for Account that sets it's salt field value.
type SaltSetter interface {
	SetSalt(salt []byte)
}

// SaltGetter is an interface for Account that could get it's stored salt value.
type SaltGetter interface {
	GetSalt() []byte
}

// SaltFielder is an interface that gets the salt field name for given account.
type SaltFielder interface {
	SaltField() string
}

type accountKey struct{}

// CtxWithAccount stores account in the context.
func CtxWithAccount(ctx context.Context, account Account) context.Context {
	return context.WithValue(ctx, accountKey{}, account)
}

// CtxGetAccount gets the account from the context 'ctx'
func CtxGetAccount(ctx context.Context) (Account, bool) {
	acc, ok := ctx.Value(accountKey{}).(Account)
	return acc, ok
}

// UsernameValidator is a function used to validate the username for the account.
type UsernameValidator func(username string) error

// DefaultUsernameValidator is the default username validator function.
func DefaultUsernameValidator(username string) error {
	if len(username) < 6 {
		return errors.WrapDetf(ErrInvalidUsername, "too short username").
			WithDetail("a username needs to be at least 6 character long")
	}
	return nil
}
