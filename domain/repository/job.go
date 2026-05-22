package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/enum"
	"github.com/taskforge/taskforge/domain/model"
)

type JobStore interface {
	Create(ctx context.Context, job *model.Job) (err error)
	GetByID(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
	GetByIdempotencyKey(ctx context.Context, key string) (job *model.Job, err error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status enum.JobStatus) (err error)
	Claim(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
	Complete(ctx context.Context, id uuid.UUID) (err error)
	ScheduleRetry(ctx context.Context, id uuid.UUID, lastError string, runAt time.Time) (err error)
	MarkDead(ctx context.Context, id uuid.UUID, lastError string) (err error)
	ListDueForRetry(ctx context.Context, limit int) (jobs []*model.Job, err error)
	MarkQueuedIfDue(ctx context.Context, id uuid.UUID) (updated bool, err error)
}
