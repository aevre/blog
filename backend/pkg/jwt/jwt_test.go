package jwt

import (
	"testing"
	"time"

	"blog-server/config"
	"blog-server/entity"
)

func TestGeneratedTokensHaveDistinctTypes(t *testing.T) {
	j := New(config.JWTConfig{
		Secret:            "test-secret",
		Issuer:            "test",
		AccessExpiration:  time.Minute,
		RefreshExpiration: time.Hour,
	})

	access, refresh, err := j.GenerateAllTokens(1, entity.UserRoleReader)
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
}
