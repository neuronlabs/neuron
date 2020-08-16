package auth

import (
	"context"
)

// Verifier is the interface used to authorize resources.
type Verifier interface {
	// Authorize if the is allowed to access the resource. The resourceID is a unique resource identifier.
	Verify(ctx context.Context, account Account, options ...VerifyOption) error
}

// VerifyOptions is the structure contains authorize query options.
type VerifyOptions struct {
	AllowedRoles    []Role
	DisallowedRoles []Role
	Scopes          []Scope
}

// VerifyOption is an option used for the verification.
type VerifyOption func(o *VerifyOptions)

// VerifyAllowedRoles sets allowed roles for the verify options.
func VerifyAllowedRoles(allowedRoles ...Role) VerifyOption {
	return func(o *VerifyOptions) {
		o.AllowedRoles = append(o.AllowedRoles, allowedRoles...)
	}
}

// VerifyDisallowedRoles sets disallowed roles for the verify options.
func VerifyDisallowedRoles(disallowedRoles ...Role) VerifyOption {
	return func(o *VerifyOptions) {
		o.DisallowedRoles = append(o.DisallowedRoles, disallowedRoles...)
	}
}

// VerifyScopes sets the verify options scopes.
func VerifyScopes(scopes ...Scope) VerifyOption {
	return func(o *VerifyOptions) {
		o.Scopes = append(o.Scopes, scopes...)
	}
}
