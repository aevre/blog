// Package scheduler
package scheduler

import (
	"context"
	"sync"

	"blog-server/internal/account/repository"
	contentservice "blog-server/internal/content/service"
	linkservice "blog-server/internal/link/service"
	logger "blog-server/internal/platform/logging"
	"blog-server/internal/scheduler/jobs"
)

type Scheduler struct {
	postService contentservice.PostService
	linkService linkservice.LinkService
	log         logger.Logger
	sessions    *repository.SessionRepo
}

func NewScheduler(log logger.Logger, postService contentservice.PostService, linkService linkservice.LinkService, sessions *repository.SessionRepo) *Scheduler {
	return &Scheduler{postService, linkService, log, sessions}
}

func (s *Scheduler) Start(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		jobs.StartSessionCleanupJob(ctx, s.sessions, s.log)
	}()
	go func() {
		defer wg.Done()
		jobs.StartViewFlushJob(ctx, s.postService, s.log)
	}()
	go func() {
		defer wg.Done()
		jobs.StartCheckLinkStatusJob(ctx, s.linkService, s.log)
	}()
	wg.Wait()
}
