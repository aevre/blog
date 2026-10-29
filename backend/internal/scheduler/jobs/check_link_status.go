package jobs

import (
	"context"
	"time"

	"blog-server/internal/link/service"
	"blog-server/internal/platform/logging"
)

func StartCheckLinkStatusJob(ctx context.Context, svc service.LinkService, log logger.Logger) {
	for {
		now := time.Now()
		next := now.Truncate(time.Hour).Add(time.Hour)
		wait := time.Until(next)

		ticker := time.NewTimer(wait)
		select {
		case <-ticker.C:
			if err := svc.CheckLinkStatus(ctx); err != nil {
				log.Error("check link status failed",
					logger.String("module", "scheduler"),
					logger.String("job", "check_link_status"),
					logger.Err(err),
				)
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
