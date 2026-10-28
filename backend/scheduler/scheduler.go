// Package scheduler
package scheduler

import (
	"context"
	"sync"

	"blog-server/logger"
	"blog-server/scheduler/jobs"
	"blog-server/service"
)

type Scheduler struct {
	postService service.PostService
	linkService service.LinkService
	log         logger.Logger
}

func NewScheduler(log logger.Logger, postService service.PostService, linkService service.LinkService) *Scheduler {
	return &Scheduler{postService, linkService, log}
}

func (s *Scheduler) Start(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)
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
