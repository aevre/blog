package authz

import (
	accountmodel "blog-server/internal/account/model"
)

// RolePermissions maps user roles to their permissions.
// Keys are accountmodel.UserRole values: accountmodel.UserRoleAdmin, accountmodel.UserRoleReader
var rolePermissions = map[accountmodel.UserRole][]Permission{
	accountmodel.UserRoleAdmin: {
		{ResourcePost, ActionCreate},
		{ResourcePost, ActionRead},
		{ResourcePost, ActionUpdate},
		{ResourcePost, ActionDelete},

		{ResourceLink, ActionCreate},
		{ResourceLink, ActionRead},
		{ResourceLink, ActionUpdate},
		{ResourceLink, ActionDelete},
		{ResourceLink, ActionReview},
	},

	accountmodel.UserRoleReader: {
		{ResourcePost, ActionRead},

		{ResourceLink, ActionCreate},
		{ResourceLink, ActionRead},
		{ResourceLink, ActionUpdate},
	},
}

func can(role accountmodel.UserRole, resource Resource, action Action) bool {
	for _, p := range rolePermissions[role] {
		if p.Resource == resource && p.Action == action {
			return true
		}
	}

	return false
}
