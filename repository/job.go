package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/enum"
	"github.com/taskforge/taskforge/domain/model"
)

func (r *Repository) Create(ctx context.Context, job *model.Job) (err error) {
	return r.jobs.Create(ctx, job)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (job *model.Job, err error) {
	return r.jobs.GetByID(ctx, id)
}

func (r *Repository) GetByIdempotencyKey(ctx context.Context, key string) (job *model.Job, err error) {
	return r.jobs.GetByIdempotencyKey(ctx, key)
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status enum.JobStatus) (err error) {
	return r.jobs.UpdateStatus(ctx, id, status)
}

func (r *Repository) Claim(ctx context.Context, id uuid.UUID) (job *model.Job, err error) {
	return r.jobs.Claim(ctx, id)
}

func (r *Repository) Complete(ctx context.Context, id uuid.UUID) (err error) {
	return r.jobs.Complete(ctx, id)
}
