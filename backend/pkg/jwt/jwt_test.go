package jwt

import (
	"testing"
	"time"

	"blog-server/config"
	"blog-server/entity"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func testJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:            "test-secret",
		Issuer:            "test",
		AccessExpiration:  time.Minute,
		RefreshExpiration: time.Hour,
	}
}

func TestGeneratedTokensHaveDistinctTypes(t *testing.T) {
	j := New(config.JWTConfig{
		Secret:            "test-secret",
		Issuer:            "test",
		AccessExpiration:  time.Minute,
		RefreshExpiration: time.Hour,
	})

	access, refresh, err := j.GenerateAllTokens(1, entity.UserRoleReader, "session-1")
	if err != nil {
		t.Fatal(err)
	}
	accessClaims, err := j.Parse(access)
	if err != nil {
		t.Fatal(err)
	}
	refreshClaims, err := j.Parse(refresh)
	if err != nil {
		t.Fatal(err)
	}
	if accessClaims.TokenType != TokenTypeAccess {
		t.Fatalf("access token type = %q", accessClaims.TokenType)
	}
	if refreshClaims.TokenType != TokenTypeRefresh {
		t.Fatalf("refresh token type = %q", refreshClaims.TokenType)
	}

	if accessClaims.SessionID != "session-1" {
		t.Fatalf("access token session ID = %q, want %q", accessClaims.SessionID, "session-1")
	}

	if refreshClaims.SessionID != "session-1" {
		t.Fatalf("refresh token session ID = %q, want %q", refreshClaims.SessionID, "session-1")
	}
}

func TestGeneratedRefreshTokenAreUnique(t *testing.T) {
	j := New(testJWTConfig())

	sid := uuid.NewString()

	_, first, err := j.GenerateAllTokens(1, entity.UserRoleReader, sid)
	if err != nil {
		t.Fatal(err)
	}

	_, second, err := j.GenerateAllTokens(1, entity.UserRoleReader, sid)
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Fatal("consecutive refresh tokens are identical")
	}

	firstClaims, err := j.Parse(first)
	if err != nil {
		t.Fatal(err)
	}

	secondClaims, err := j.Parse(second)
	if err != nil {
		t.Fatal(err)
	}

	firstTokenID := firstClaims.RegisteredClaims.ID
	secondTokenID := secondClaims.RegisteredClaims.ID

	if firstTokenID == "" {
		t.Fatal("first refresh token has an empty jti")
	}

	if secondTokenID == "" {
		t.Fatal("second refresh token has an empty jti")
	}

	if firstTokenID == secondTokenID {
		t.Fatalf("refresh token jti value are identical: %q", firstTokenID)
	}
}

func TestGeneratedTokenPreservesUserClaims(t *testing.T) {
	j := New(testJWTConfig())
	access, err := j.GenerateAccessToken(42, entity.UserRoleAdmin, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}

	claims, err := j.Parse(access)
	if err != nil {
		t.Fatal(err)
	}

	if claims.ID != 42 {
		t.Fatalf("user ID = %d", claims.ID)
	}

	if claims.Role != entity.UserRoleAdmin {
		t.Fatalf("role = %q, want %q", claims.Role, entity.UserRoleAdmin)
	}

	if claims.RegisteredClaims.ID == "" {
		t.Fatal("token has empty jti")
	}
}

func TestParseRejectsUnexpectedIssuer(t *testing.T) {
	expectedConfig := testJWTConfig()
	otherConfig := expectedConfig
	otherConfig.Issuer = "other-issuer"

	token, err := New(otherConfig).GenerateAccessToken(1, entity.UserRoleReader, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}

	assertRejected(t, New(expectedConfig), token)
}

func TestParseRejectsUnexpectedSigningMethod(t *testing.T) {
	cfg := testJWTConfig()
	now := time.Now()

	claims := Claims{
		ID:        1,
		Role:      entity.UserRoleReader,
		TokenType: TokenTypeAccess,
		SessionID: uuid.NewString(),
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        "test-jti",
			Issuer:    cfg.Issuer,
			Subject:   "user token",
			ExpiresAt: jwtv5.NewNumericDate(now.Add(time.Minute)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
		},
	}

	token := signTestToken(t, jwtv5.SigningMethodHS384, claims, cfg.Secret)

	assertRejected(t, New(cfg), token)
}

func TestParseRejectsTokenWithoutExpiration(t *testing.T) {
	cfg := testJWTConfig()
	now := time.Now()

	claims := Claims{
		ID:        1,
		Role:      entity.UserRoleReader,
		TokenType: TokenTypeAccess,
		SessionID: uuid.NewString(),
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        "test-jti",
			Issuer:    cfg.Issuer,
			Subject:   "user token",
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
		},
	}

	token := signTestToken(t, jwtv5.SigningMethodHS256, claims, cfg.Secret)

	assertRejected(t, New(cfg), token)
}

func signTestToken(
	t *testing.T,
	method jwtv5.SigningMethod,
	claims jwtv5.Claims,
	secret string,
) string {
	t.Helper()

	token, err := jwtv5.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	return token
}

func assertRejected(t *testing.T, j Jwt, token string) {
	t.Helper()

	if _, err := j.Parse(token); err == nil {
		t.Fatal("Parse accepted an invalid token")
	}

	valid, err := j.Validate(token)
	if err == nil {
		t.Fatal("Validate returned no error for an invalid token")
	}

	if valid {
		t.Fatal("Validate returned true for an invalid token")
	}
}
