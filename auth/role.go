package auth

import (
	"context"

	"github.com/neuronlabs/neuron/query"
)

// Roler is the role-based access control authorization.
type Roler interface {
	// FindRoles list all roles for
	FindRoles(ctx context.Context, options ...ListRoleOption) ([]Role, error)
	// ClearRoles clears all roles for given account.
	ClearRoles(ctx context.Context, account Account) error
	// GrantRole grants given 'role' access to given 'scope'.
	GrantRole(ctx context.Context, account Account, role Role) error
	// RevokeRole revokes access to given 'scope' for the 'role'.
	RevokeRole(ctx context.Context, account Account, role Role) error
}

// ListRoleOptions are the options for listing the roles.
type ListRoleOptions struct {
	SortByHierarchy bool
	SortOrder       query.SortOrder
	Limit, Offset   int
	Account         Account
}

// ListRoleOption is the option
type ListRoleOption func(o *ListRoleOptions)

// Role is the interface used for the roles.
type Role interface {
	RoleName() string
}

// HierarchicalRole is an interface for the
type HierarchicalRole interface {
	Role
	HierarchyValue() int
}
