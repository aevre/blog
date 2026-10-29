// Package jobs provider scheduler jobs
package jobs

import (
	"context"
	"time"

	"blog-server/internal/content/service"
	"blog-server/internal/platform/logging"
)

func StartViewFlushJob(ctx context.Context, svc service.PostService, log logger.Logger) {
	for {
		now := time.Now()
		next := now.Truncate(time.Hour).Add(time.Hour)
		wait := time.Until(next)

		ticker := time.NewTimer(wait)
		select {
		case <-ticker.C:
			if err := svc.FlushViewCountToDB(ctx); err != nil {
				log.Error("flush post view counts failed",
					logger.String("module", "scheduler"),
					logger.String("job", "view_count_flush"),
					logger.Err(err),
				)
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
