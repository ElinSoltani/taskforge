package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/model"
)

// Repository defines persistence and queue operations required by the service layer.
type Repository interface {
	Create(ctx context.Context, job *model.Job) (err error)
	GetByID(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
	GetByIdempotencyKey(ctx context.Context, key string) (job *model.Job, err error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.JobStatus) (err error)
	Enqueue(ctx context.Context, msg model.QueueMessage) (err error)
}

// JobService is the application API consumed by REST and other entrypoints.
type JobService interface {
	Create(ctx context.Context, in model.CreateJobInput) (job *model.Job, duplicate bool, err error)
	Get(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
}

type jobService struct {
	repo Repository
}

// NewJobService wires the service layer to a repository implementation.
func NewJobService(repo Repository) JobService {
	return &jobService{repo: repo}
}
