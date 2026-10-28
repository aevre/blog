// Package jwt
package jwt

import (
	"fmt"
	"time"

	"blog-server/config"
	"blog-server/entity"
	"blog-server/pkg/errx"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type jwtx struct {
	cfg config.JWTConfig
}

func (j *jwtx) GenerateAccessToken(userID uint, role entity.UserRole) (string, error) {
	return j.generateToken(userID, role, TokenTypeAccess, j.cfg.AccessExpiration)
}

func (j *jwtx) GenerateRefreshToken(userID uint, role entity.UserRole) (string, error) {
	return j.generateToken(userID, role, TokenTypeRefresh, j.cfg.RefreshExpiration)
}

func (j *jwtx) GenerateAllTokens(userID uint, role entity.UserRole) (string, string, error) {
	access, err := j.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", err
	}

	refresh, err := j.GenerateRefreshToken(userID, role)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (j *jwtx) Parse(tokenStr string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenStr, &Claims{}, func(token *jwtv5.Token) (any, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errx.New(
				errx.CodeInternalError,
				fmt.Errorf("unexpected signing method: %v", token.Header["alg"]),
			)
		}
		return []byte(j.cfg.Secret), nil
	})
	if err != nil {
		return nil, errx.New(errx.CodeUnauthorized, fmt.Errorf("invalid token: %w", err))
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errx.New(errx.CodeUnauthorized, fmt.Errorf("invalid token claims"))
	}

	return claims, nil
}

func (j *jwtx) Validate(tokenStr string) (bool, error) {
	_, err := jwtv5.Parse(tokenStr, func(token *jwtv5.Token) (any, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errx.New(
				errx.CodeInternalError,
				fmt.Errorf("unexpected signing method: %v", token.Header["alg"]),
			)
		}
		return []byte(j.cfg.Secret), nil
	})
	if err != nil {
		return false, errx.New(errx.CodeUnauthorized, err)
	}

	return true, nil
}

func (j *jwtx) generateToken(userID uint, role entity.UserRole, tokenType string, expires time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		ID:        userID,
		Role:      role,
		TokenType: tokenType,
		ExpiresAt: jwtv5.NewNumericDate(now.Add(expires)),
		IssuedAt:  jwtv5.NewNumericDate(now),
		NotBefore: jwtv5.NewNumericDate(now),
		Issuer:    j.cfg.Issuer,
		Subject:   "user token",
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(j.cfg.Secret))
	if err != nil {
		return "", errx.New(errx.CodeInternalError, err)
	}

	return signed, nil
}

func New(cfg config.JWTConfig) Jwt {
	return &jwtx{cfg: cfg}
}
