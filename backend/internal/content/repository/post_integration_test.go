package repository

import (
	"context"
	"os"
	"testing"

	contentmodel "blog-server/internal/content/model"
	"blog-server/internal/platform/config"
	"blog-server/internal/platform/database"
	logger "blog-server/internal/platform/logging"
	"blog-server/migration"
)

// The named database must be disposable: this test applies migrations.
func TestPostIntegration(t *testing.T) {
	name := os.Getenv("TEST_POSTGRES_DB")
	if name == "" {
		t.Skip("set TEST_POSTGRES_DB to a disposable database")
	}
	cfg := &config.Config{Database: config.DatabaseConfig{Host: "/run/postgresql", Port: 5432, User: os.Getenv("USER"), Name: name, SSLMode: "disable"}}
	ds, err := database.NewDataStore(cfg, logger.NewLogger(cfg))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ds.Close() })
	ctx := context.Background()
	if err := migration.Up(ctx, ds.DB()); err != nil {
		t.Fatal(err)
	}
	u, err := ds.Client(ctx).User.Create().SetUsername("test").SetEmail("integration@example.com").SetPassword("hash").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewPostRepo(ds)
	p, err := repo.Create(ctx, &contentmodel.Post{Title: "published", Content: "body", UserID: u.ID, Status: contentmodel.PostStatusPublish, ReadTimeMinutes: 1})
	if err != nil {
		t.Fatal(err)
	}
	before, err := repo.PublishedVersion(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if owner, err := repo.IsOwner(ctx, u.ID, p.ID); err != nil || !owner {
		t.Fatalf("owner=%v err=%v", owner, err)
	}
	if owner, err := repo.IsOwner(ctx, u.ID+1, p.ID); err != nil || owner {
		t.Fatalf("other owner=%v err=%v", owner, err)
	}
	if err := repo.Update(ctx, &contentmodel.Post{ID: p.ID, Title: "edited"}); err != nil {
		t.Fatal(err)
	}
	after, err := repo.PublishedVersion(ctx, p.ID)
	if err != nil || !after.After(before) {
		t.Fatalf("version did not advance: %v %v %v", before, after, err)
	}
	if err := repo.Update(ctx, &contentmodel.Post{ID: p.ID, Status: contentmodel.PostStatusDraft}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.PublishedVersion(ctx, p.ID); err == nil {
		t.Fatal("draft remains public")
	}
	rollback := ds.WithTx(ctx, func(txctx context.Context) error {
		if _, err := repo.Create(txctx, &contentmodel.Post{Title: "rollback", Content: "body", UserID: u.ID, Status: contentmodel.PostStatusDraft, ReadTimeMinutes: 1}); err != nil {
			return err
		}
		return context.Canceled
	})
	if rollback == nil {
		t.Fatal("expected rollback")
	}
	n, err := ds.Client(ctx).Post.Query().Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("transaction rollback count=%d err=%v", n, err)
	}
}
