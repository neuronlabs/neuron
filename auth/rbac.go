package auth

// RBAC is the role-based access control authorization.
type RBAC interface {
	// CreateRole creates a 'role'. An additional optional description might be provided for given role.
	CreateRole(role, description string) error
	// FindRoles finds all stored roles. An optional argument(s) might be provided
	// to specify the resource id(s) for which the roles should be taken.
	FindRoles(options ...RoleFindOption) (roles []string, err error)
	// DeleteRole removes the 'role'
	DeleteRole(role string) error

	// GrantRole grants given 'role' access to given 'resourceID'.
	GrantRole(role, resourceID string) error
	// RevokeRole revokes access to given 'resource' for the 'role'.
	RevokeRole(role, resourceID string) error

	// AddRole adds the role to the given accountID.
	AddRole(accountID, role string) error
	// ClearRoles clears all roles for given account.
	ClearRoles(accountID string) error
	// RemoveRole removes the role for given account.
	RemoveRole(accountID, role string) error
	// SetRole clears and sets the role for the account with id.
	SetRole(accountID string, roles ...string) error
}

type RoleFindOptions struct {
	AccountIDs   []string
	ResourcesIDs []string
}

type RoleFindOption func(o *RoleFindOptions)

func RolesWithAccountIDs(accountIDs ...string) RoleFindOption {
	return func(o *RoleFindOptions) {
		o.AccountIDs = append(o.AccountIDs, accountIDs...)
	}
}

func RolesWithResourcesIDs(resourcesIDs ...string) RoleFindOption {
	return func(o *RoleFindOptions) {
		o.ResourcesIDs = append(o.ResourcesIDs, resourcesIDs...)
	}
}
