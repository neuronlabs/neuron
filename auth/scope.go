package auth

import (
	"context"
)

// Scope is an interface that defines authorization scope.
type Scope interface {
	ScopeName() string
}

// RoleScoper is an interface for authorizators that allows to set and get scopes.
type RoleScoper interface {
	// ListRoleScopes lists the scopes for provided options.
	ListRoleScopes(ctx context.Context, options ...ListScopeOption) ([]Scope, error)
	// ClearRoleScopes clears the scopes for provided roles/accounts.
	ClearRoleScopes(ctx context.Context, roles ...Role) error
	// GrantRoleScope grants roles/accounts access for given scope.
	GrantRoleScope(ctx context.Context, role Role, scope Scope) error
	// RevokeRoleScope revokes the roles/accounts access for given scope.
	RevokeRoleScope(ctx context.Context, role Role, scope Scope) error
}

// ListScopeOptions are the options used for listing the
type ListScopeOptions struct {
	Limit, Offset int
	Role          Role
}

// ListScopeOption is an option function that changes list scope options.
type ListScopeOption func(o *ListScopeOptions)
