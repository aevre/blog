package jwt

import (
	"blog-server/entity"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	ID        uint            `json:"id"`
	Role      entity.UserRole `json:"role"`
	TokenType string          `json:"tokenType"`
	jwt.RegisteredClaims
}
