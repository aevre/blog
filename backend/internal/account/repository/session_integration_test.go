package repository

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"blog-server/internal/platform/config"
	"blog-server/internal/platform/database"
	logger "blog-server/internal/platform/logging"
	"blog-server/migration"

	"github.com/google/uuid"
)

func TestSessionsIntegration(t *testing.T) {
	name := os.Getenv("TEST_POSTGRES_DB")
	if name == "" {
		t.Skip("set TEST_POSTGRES_DB to a disposable database")
	}
	cfg := &config.Config{Database: config.DatabaseConfig{Host: "/run/postgresql", Port: 5432, User: os.Getenv("USER"), Name: name}}
	ds, err := database.NewDataStore(cfg, logger.NewLogger(cfg))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ds.Close() })
	ctx := context.Background()
	if err := migration.Up(ctx, ds.DB()); err != nil {
		t.Fatal(err)
	}
	u, err := ds.Client(ctx).User.Create().SetUsername("session-test").SetEmail(uuid.NewString() + "@test.local").SetPassword("hash").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	s := NewSessionRepo(ds)
	id := uuid.New()
	expires := time.Now().Add(time.Hour)
	if err := s.Create(ctx, id, u.ID, "old", expires); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var winners atomic.Int64
	for range 16 {
		wg.Go(func() {
			_, _, ok, err := s.Rotate(ctx, "old", uuid.NewString(), expires)
			if err != nil {
				t.Error(err)
			}
			if ok {
				winners.Add(1)
			}
		})
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("successful rotations=%d", winners.Load())
	}
	var revoked bool
	if err := ds.DB().QueryRowContext(ctx, "SELECT revoked_at IS NOT NULL FROM account_sessions WHERE id=$1", id).Scan(&revoked); err != nil {
		t.Fatal(err)
	}
	if !revoked {
		t.Fatal("reuse did not revoke session")
	}
	if err := s.Create(ctx, uuid.New(), u.ID, "other", expires); err != nil {
		t.Fatal(err)
	}
	_, _, ok, err := s.Rotate(ctx, "other", "next", expires)
	if err != nil || !ok {
		t.Fatalf("other session affected: %v %v", ok, err)
	}
	if err := s.Revoke(ctx, "next"); err != nil {
		t.Fatal(err)
	}
	_, _, ok, err = s.Rotate(ctx, "next", "last", expires)
	if err != nil || ok {
		t.Fatalf("logout failed: %v %v", ok, err)
	}

	liveID := uuid.New()
	if err := s.Create(ctx, liveID, u.ID, "live-token", expires); err != nil {
		t.Fatal(err)
	}
	if _, err := ds.DB().ExecContext(ctx, `INSERT INTO account_refresh_tokens(hash,session_id,expires_at,used_at)
		SELECT 'expired-' || n, $1, now()-interval '1 hour', now()-interval '2 hours'
		FROM generate_series(1,501) n`, liveID); err != nil {
		t.Fatal(err)
	}
	expiredID := uuid.New()
	if err := s.Create(ctx, expiredID, u.ID, "expired-session", time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	// A token can expire while its session remains active after a later rotation.
	_, _, ok, err = s.Rotate(ctx, "expired-1", "must-not-exist", expires)
	if err != nil || ok {
		t.Fatalf("expired token accepted: %v %v", ok, err)
	}
	if err := s.CleanupExpired(ctx); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := ds.DB().QueryRowContext(ctx, "SELECT count(*) FROM account_refresh_tokens WHERE expires_at <= now()").Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatalf("batch limit not respected: remaining=%d, want 2", remaining)
	}
	if err := s.CleanupExpired(ctx); err != nil {
		t.Fatal(err)
	}
	var exists bool
	if err := ds.DB().QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM account_sessions WHERE id=$1)", expiredID).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expired empty session was not deleted")
	}
	if err := ds.DB().QueryRowContext(ctx, "SELECT count(*) FROM account_refresh_tokens WHERE session_id=$1", liveID).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 1 {
		t.Fatalf("active session history not cleaned: tokens=%d", remaining)
	}
	_, _, ok, err = s.Rotate(ctx, "live-token", "live-next", expires)
	if err != nil || !ok {
		t.Fatalf("cleanup affected active token: %v %v", ok, err)
	}
	_, _, ok, err = s.Rotate(ctx, "live-token", "reuse", expires)
	if err != nil || ok {
		t.Fatalf("cleanup lost unexpired reuse detection: %v %v", ok, err)
	}
}
