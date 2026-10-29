package jwt

import (
	accountmodel "blog-server/internal/account/model"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	ID        uint                  `json:"id"`
	Role      accountmodel.UserRole `json:"role"`
	TokenType string                `json:"tokenType"`
	SessionID string                `json:"sid"`
	jwt.RegisteredClaims
}
