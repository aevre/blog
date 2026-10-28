package scheduler

import (
	"context"
	"sync"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"scheduler",
		fx.Provide(
			NewScheduler,
		),
		fx.Invoke(
			runJobsLifecycle,
		),
	)
}

func runJobsLifecycle(lc fx.Lifecycle, scheduler *Scheduler) {
	var cancel context.CancelFunc
	var jobs sync.WaitGroup

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			jobCtx, c := context.WithCancel(context.Background())
			cancel = c
			jobs.Go(func() {
				scheduler.Start(jobCtx)
			})
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cancel == nil {
				return nil
			}
			cancel()
			done := make(chan struct{})
			go func() {
				jobs.Wait()
				close(done)
			}()
			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}
