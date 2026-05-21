package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/model"
)

func (r *Repository) ScheduleRetry(ctx context.Context, id uuid.UUID, lastError string, runAt time.Time) (err error) {
	return r.jobs.ScheduleRetry(ctx, id, lastError, runAt)
}

func (r *Repository) MarkDead(ctx context.Context, id uuid.UUID, lastError string) (err error) {
	return r.jobs.MarkDead(ctx, id, lastError)
}

func (r *Repository) ListDueForRetry(ctx context.Context, limit int) (jobs []*model.Job, err error) {
	return r.jobs.ListDueForRetry(ctx, limit)
}

func (r *Repository) MarkQueuedIfDue(ctx context.Context, id uuid.UUID) (updated bool, err error) {
	return r.jobs.MarkQueuedIfDue(ctx, id)
}

// RequeueDueJob promotes a retrying job to queued and enqueues it to Redis.
func (r *Repository) RequeueDueJob(ctx context.Context, job *model.Job) error {
	ok, err := r.MarkQueuedIfDue(ctx, job.ID)
	if err != nil || !ok {
		return err
	}
	if err := r.queue.Enqueue(ctx, model.QueueMessage{JobID: job.ID, JobType: job.JobType}); err != nil {
		_ = r.jobs.ScheduleRetry(ctx, job.ID, err.Error(), time.Now().UTC())
		return err
	}
	return nil
}
