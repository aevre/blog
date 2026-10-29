package jwt

import accountmodel "blog-server/internal/account/model"

type Jwt interface {
	GenerateAccessToken(userID uint, role accountmodel.UserRole, sessionID string) (string, error)
	GenerateRefreshToken(userID uint, role accountmodel.UserRole, sessionID string) (string, error)
	GenerateAllTokens(userID uint, role accountmodel.UserRole, sessionID string) (string, string, error)

	Validate(token string) (bool, error)
	Parse(token string) (*Claims, error)
}
