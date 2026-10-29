package jobs

import (
	"context"
	"time"

	"blog-server/internal/platform/logging"
)

type sessionCleaner interface {
	CleanupExpired(context.Context) error
}

func StartSessionCleanupJob(ctx context.Context, sessions sessionCleaner, log logger.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		batchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := sessions.CleanupExpired(batchCtx)
		cancel()
		if err != nil && ctx.Err() == nil {
			log.Error("clean up expired refresh sessions failed", logger.Err(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
