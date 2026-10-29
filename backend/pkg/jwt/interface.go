package jwt

import "blog-server/entity"

type Jwt interface {
	GenerateAccessToken(userID uint, role entity.UserRole, sessionID string) (string, error)
	GenerateRefreshToken(userID uint, role entity.UserRole, sessionID string) (string, error)
	GenerateAllTokens(userID uint, role entity.UserRole, sessionID string) (string, string, error)

	Validate(token string) (bool, error)
	Parse(token string) (*Claims, error)
}
