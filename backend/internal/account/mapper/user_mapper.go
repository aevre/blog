// Package mapper converts account persistence models into domain models.
package mapper

import (
	"blog-server/ent"
	accountmodel "blog-server/internal/account/model"
)

func ToUser(u *ent.User) *accountmodel.User {
	if u == nil {
		return &accountmodel.User{}
	}

	return &accountmodel.User{
		ID:        u.ID,
		UUID:      u.UUID,
		Username:  u.Username,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
