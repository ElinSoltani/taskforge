package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/taskforge/taskforge/repository"
)

// RunScheduler periodically re-enqueues jobs whose backoff has elapsed.
func RunScheduler(ctx context.Context, repo *repository.Repository, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("retry scheduler started", "interval", interval.String())

	for {
		select {
		case <-ctx.Done():
			slog.Info("retry scheduler stopping")
			return
		case <-ticker.C:
			if err := processDueRetries(ctx, repo); err != nil {
				slog.Warn("retry scheduler tick failed", "error", err)
			}
		}
	}
}

func processDueRetries(ctx context.Context, repo *repository.Repository) error {
	jobs, err := repo.ListDueForRetry(ctx, 50)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := repo.RequeueDueJob(ctx, job); err != nil {
			slog.Warn("requeue due job failed", "job_id", job.ID, "error", err)
			continue
		}
		slog.Debug("requeued due job", "job_id", job.ID, "job_type", job.JobType)
	}
	return nil
}
