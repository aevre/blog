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
		{ResourceLink, ActionUpdate},
		{ResourceLink, ActionDelete},
	},

	accountmodel.UserRoleReader: {
		{ResourcePost, ActionRead},
	},
}
