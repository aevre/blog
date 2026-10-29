package service

import (
	"context"
	"errors"
	"testing"
	"time"

	contentmodel "blog-server/internal/content/model"
	"blog-server/internal/content/repository"
	"blog-server/internal/platform/config"
	"blog-server/internal/platform/logging"
)

type atomicStoreStub struct {
	values map[string]int64
	fail   bool
}

func (s *atomicStoreStub) Incr(context.Context, string) (int64, error) { return 0, nil }
func (s *atomicStoreStub) IncrBy(_ context.Context, key string, value int64) (int64, error) {
	if s.fail {
		return 0, errors.New("redis unavailable")
	}
	s.values[key] += value
	return s.values[key], nil
}
func (s *atomicStoreStub) PopBatch(context.Context, []string) (map[string]string, error) {
	return nil, nil
}

type detailRepoStub struct {
	repository.PostRepo
	visible bool
	reads   int
}

func (s *detailRepoStub) PublishedVersion(context.Context, uint) (time.Time, error) {
	if !s.visible {
		return time.Time{}, errors.New("not published")
	}
	return time.Unix(1, 0), nil
}
func (s *detailRepoStub) GetPublishedByID(context.Context, uint) (*contentmodel.Post, error) {
	s.reads++
	return &contentmodel.Post{ID: 1, Title: "published"}, nil
}

type detailStoreStub struct {
	atomicStoreStub
	value string
	fail  bool
}

func (s *detailStoreStub) Get(context.Context, string) (string, error) {
	if s.fail {
		return "", errors.New("redis down")
	}
	return s.value, nil
}
func (s *detailStoreStub) Set(_ context.Context, _, v string, _ time.Duration) error {
	s.value = v
	return nil
}
func (s *detailStoreStub) Delete(context.Context, string) error { return nil }
func (s *detailStoreStub) Scan(context.Context, string, uint64, int64) ([]string, uint64, error) {
	return nil, 0, nil
}

func TestDetailCache(t *testing.T) {
	repo := &detailRepoStub{visible: true}
	store := &detailStoreStub{}
	svc := &postService{pr: repo, rc: store, log: logger.NewLogger(&config.Config{})}
	for range 2 {
		if _, err := svc.GetPostByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	}
	if repo.reads != 1 {
		t.Fatalf("detail reads=%d", repo.reads)
	}
	repo.visible = false
	if _, err := svc.GetPostByID(context.Background(), 1); err == nil {
		t.Fatal("cached draft exposed")
	}
	repo.visible = true
	store.fail = true
	if _, err := svc.GetPostByID(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if repo.reads != 2 {
		t.Fatalf("Redis failure did not fall back: %d", repo.reads)
	}
}

func TestRestoreViewCounts(t *testing.T) {
	store := &atomicStoreStub{values: map[string]int64{"views:1": 2}}
	errs := restoreViewCounts(context.Background(), store, map[uint]int64{1: 3}, map[uint]string{1: "views:1"})
	if len(errs) != 0 {
		t.Fatalf("restore errors: %v", errs)
	}
	if got := store.values["views:1"]; got != 5 {
		t.Fatalf("restored count = %d, want 5", got)
	}
}
