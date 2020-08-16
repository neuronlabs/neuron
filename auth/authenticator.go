package auth

import (
	"github.com/neuronlabs/neuron/store"
)

// Authenticator is the interface used to authenticate the username and password.
type Authenticator interface {
	// HashAndSetPassword creates a password hash and stores it within given account.
	// If a model implements SaltSetter this function should set the salt also.
	HashAndSetPassword(account Account, password *Password) error
	// ComparePassword hash the 'password' (with optional salt) and compare with stored password hash.
	ComparePassword(account Account, password string) error
}

// AuthenticatorOptions are the authentication service options.
type AuthenticatorOptions struct {
	// Store is a store used for some authenticator implementations.
	Store store.Store
	// AccountModel is the option that defines account model for the authenticator.
	AccountModel Account
	// BCryptCost is an option that defines the cost of given password.
	BCryptCost int
	// AuthenticateMethod is a method used for authentication.
	AuthenticateMethod AuthenticateMethod
	// SaltLength is the length of the salt.
	SaltLength int
}

// AuthenticatorOption is a function used to set authentication options.
type AuthenticatorOption func(o *AuthenticatorOptions)

// AuthenticateMethod is a method of authentication used by the authenticator
type AuthenticateMethod int

const (
	// BCrypt is a bcrypt password hashing method
	BCrypt AuthenticateMethod = iota
	// MD5 is a md5 password hashing method.
	MD5
	// SHA256 is a sha256 password hashing method.
	SHA256
	// SHA512 is a sha512 password hashing method.
	SHA512
)

// AuthenticatorStore is an option that sets Store in the option.
func AuthenticatorStore(op store.Store) AuthenticatorOption {
	return func(o *AuthenticatorOptions) {
		o.Store = op
	}
}

// AuthenticatorAccountModel is an option that sets AccountModel in the auth options.
func AuthenticatorAccountModel(op Account) AuthenticatorOption {
	return func(o *AuthenticatorOptions) {
		o.AccountModel = op
	}
}

// AuthenticatorBCryptCost is an option that sets BCryptCost in the auth options.
func AuthenticatorBCryptCost(op int) AuthenticatorOption {
	return func(o *AuthenticatorOptions) {
		o.BCryptCost = op
	}
}

// AuthenticatorMethod is an option that sets AuthenticateMethod in the auth options.
func AuthenticatorMethod(op AuthenticateMethod) AuthenticatorOption {
	return func(o *AuthenticatorOptions) {
		o.AuthenticateMethod = op
	}
}

// AuthenticatorSaltLength is an option that sets SaltLength in the auth options.
func AuthenticatorSaltLength(op int) AuthenticatorOption {
	return func(o *AuthenticatorOptions) {
		o.SaltLength = op
	}
}
