package service

import (
	"context"
	"testing"
	"time"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/platform/config"
	"blog-server/internal/shared/jwt"
)

func TestRefreshRejectsInvalidTokensBeforeStorage(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "secret", Issuer: "test", AccessExpiration: time.Minute, RefreshExpiration: time.Hour}}
	j := jwt.New(cfg.JWT)
	access, err := j.GenerateAccessToken(1, accountmodel.UserRoleReader, "session")
	if err != nil {
		t.Fatal(err)
	}
	expiredCfg := cfg.JWT
	expiredCfg.RefreshExpiration = -time.Hour
	expired, err := jwt.New(expiredCfg).GenerateRefreshToken(1, accountmodel.UserRoleReader, "session")
	if err != nil {
		t.Fatal(err)
	}
	otherCfg := cfg.JWT
	otherCfg.Secret = "other"
	forged, err := jwt.New(otherCfg).GenerateRefreshToken(1, accountmodel.UserRoleReader, "session")
	if err != nil {
		t.Fatal(err)
	}
	// A nil store makes any database access for an invalid JWT fail this test.
	svc := &authService{cfg: cfg}
	for _, token := range []string{"opaque-token", access, expired, forged} {
		if _, _, err := svc.RefreshAccessToken(context.Background(), token); err == nil {
			t.Fatal("invalid refresh token accepted")
		}
	}
}
