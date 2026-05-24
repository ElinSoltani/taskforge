package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/taskforge/taskforge/domain/model"
)

// MarkDead transitions the job to dead in Postgres, then publishes to the DLQ stream.
// Postgres is authoritative; DLQ enqueue failures are logged but do not roll back dead status.
func (r *Repository) MarkDead(ctx context.Context, job *model.Job, lastError string) error {
	if err := r.jobs.MarkDead(ctx, job.ID, lastError); err != nil {
		return err
	}

	job.ApplyDead(lastError)
	dlqMsg := model.NewDLQMessage(job, lastError)

	if err := r.queue.EnqueueDLQ(ctx, dlqMsg); err != nil {
		slog.Error("dlq enqueue failed after mark dead",
			"job_id", job.ID,
			"job_type", job.JobType,
			"error", err,
		)
		return fmt.Errorf("job marked dead but dlq enqueue failed: %w", err)
	}

	slog.Info("job moved to dead letter queue",
		"job_id", job.ID,
		"job_type", job.JobType,
		"attempt", job.AttemptCount,
		"max_attempts", job.MaxAttempts,
	)
	return nil
}
