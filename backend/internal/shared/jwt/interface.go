package jwt

import (
	accountmodel "blog-server/internal/account/model"

	"github.com/google/uuid"
)

type Jwt interface {
	GenerateAccessToken(userID uint, role accountmodel.UserRole, sessionID uuid.UUID) (string, error)
	GenerateRefreshToken(userID uint, role accountmodel.UserRole, sessionID uuid.UUID) (string, error)
	GenerateAllTokens(userID uint, role accountmodel.UserRole, sessionID uuid.UUID) (string, string, error)

	Validate(token string) (bool, error)
	Parse(token string) (*Claims, error)
}
