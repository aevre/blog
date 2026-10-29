package service

import (
	"context"
	"testing"
	"time"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/account/repository"
	"blog-server/internal/platform/config"
	"blog-server/internal/shared/jwt"
)

type sessionStub struct {
	current string
	used    map[string]bool
	revoked bool
}

func (s *sessionStub) Create(_ context.Context, _ string, _ uint, hash string, _ time.Time) error {
	s.current = hash
	s.used = map[string]bool{}
	return nil
}

func (s *sessionStub) Rotate(_ context.Context, old, next string, _ time.Time) (uint, string, bool, error) {
	if s.revoked {
		return 0, "", false, nil
	}
	if s.used[old] {
		s.revoked = true
		return 0, "", false, nil
	}
	if s.current != old {
		return 0, "", false, nil
	}
	s.used[old] = true
	s.current = next
	return 1, "session", true, nil
}
func (s *sessionStub) Revoke(context.Context, string) error { s.revoked = true; return nil }

type sessionRepoStub struct{ repository.UserRepo }

func (*sessionRepoStub) GetAuthByID(context.Context, uint) (*accountmodel.UserAuth, error) {
	return &accountmodel.UserAuth{ID: 1, Role: accountmodel.UserRoleReader}, nil
}

func TestSessionRotationAndLogout(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret", Issuer: "test", AccessExpiration: time.Minute, RefreshExpiration: time.Hour}}
	store := &sessionStub{}
	svc := &authService{cfg: cfg, sessions: store, userRepo: &sessionRepoStub{}}
	j := jwt.New(cfg.JWT)
	token, err := j.GenerateRefreshToken(1, accountmodel.UserRoleReader, "session")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := parseRefreshToken(j, token)
	if err != nil || claims.ExpiresAt == nil {
		t.Fatalf("invalid refresh JWT: %v", err)
	}
	ctx := context.Background()
	if err := svc.createSession(ctx, 1, "session", token); err != nil {
		t.Fatal(err)
	}
	if store.current == token {
		t.Fatal("raw refresh token stored")
	}
	_, rotated, err := svc.RefreshAccessToken(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RefreshAccessToken(ctx, token); err == nil {
		t.Fatal("reuse accepted")
	}
	if _, _, err := svc.RefreshAccessToken(ctx, rotated); err == nil {
		t.Fatal("reused session active")
	}
	if err := svc.Logout(ctx, rotated); err != nil {
		t.Fatal(err)
	}
}
