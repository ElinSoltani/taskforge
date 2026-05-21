package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/model"
)

type JobStore interface {
	Create(ctx context.Context, job *model.Job) (err error)
	GetByID(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
	GetByIdempotencyKey(ctx context.Context, key string) (job *model.Job, err error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.JobStatus) (err error)
	Claim(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
	Complete(ctx context.Context, id uuid.UUID) (err error)
}
