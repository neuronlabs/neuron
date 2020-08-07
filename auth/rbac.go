package auth

import (
	"context"

	"github.com/neuronlabs/neuron/database"
)

// RoleAuthorizer is the role-based access control authorization.
type RoleAuthorizer interface {
	// CreateRole creates a 'role'. An additional optional description might be provided for given role.
	CreateRole(ctx context.Context, role, description string) (Role, error)
	// FindRoles finds all stored roles. An optional argument(s) might be provided
	// to specify the resource id(s) for which the roles should be taken.
	FindRoles(ctx context.Context, options ...RoleFindOption) (roles []Role, err error)
	// GetRoles gets stored roles for provided accountID.
	GetRoles(ctx context.Context, accountID interface{}) (roles []Role, err error)
	// DeleteRole removes the 'role'
	DeleteRole(ctx context.Context, role string) error

	// GrantRole grants given 'role' access to given 'scope'.
	GrantRole(ctx context.Context, db database.DB, role, scope string) error
	// RevokeRole revokes access to given 'scope' for the 'role'.
	RevokeRole(ctx context.Context, db database.DB, role, scope string) error

	// AddRole adds the role to the given accountID.
	AddRole(ctx context.Context, db database.DB, accountID interface{}, role string) error
	// ClearRoles clears all roles for given account.
	ClearRoles(ctx context.Context, db database.DB, accountID interface{}) error
	// RemoveRole removes the role for given account.
	RemoveRole(ctx context.Context, db database.DB, accountID interface{}, role string) error
	// SetRoles clears and sets the role for the account with id.
	SetRoles(ctx context.Context, db database.DB, accountID interface{}, roles ...string) error
}

// Role is the interface used for the roles.
type Role interface {
	RoleName() string
}

// RoleFindOptions are is the structure used to find the
type RoleFindOptions struct {
	AccountIDs []interface{}
	Scopes     []string
}

type RoleFindOption func(o *RoleFindOptions)

func RolesWithAccountIDs(accountIDs ...interface{}) RoleFindOption {
	return func(o *RoleFindOptions) {
		o.AccountIDs = append(o.AccountIDs, accountIDs...)
	}
}

func RolesWithScopes(scopes ...string) RoleFindOption {
	return func(o *RoleFindOptions) {
		o.Scopes = append(o.Scopes, scopes...)
	}
}
