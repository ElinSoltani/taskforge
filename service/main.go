package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/config"
	"github.com/taskforge/taskforge/domain/enum"
	"github.com/taskforge/taskforge/domain/model"
)

// Repository defines persistence and queue operations required by the service layer.
type Repository interface {
	Create(ctx context.Context, job *model.Job) (err error)
	GetByID(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
	GetByIdempotencyKey(ctx context.Context, key string) (job *model.Job, err error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status enum.JobStatus) (err error)
	Enqueue(ctx context.Context, msg model.QueueMessage) (err error)
}

// JobService is the application API consumed by REST and other entrypoints.
type JobService interface {
	Create(ctx context.Context, job *model.Job) (created *model.Job, duplicate bool, err error)
	Get(ctx context.Context, id uuid.UUID) (job *model.Job, err error)
}

type jobService struct {
	repo           Repository
	maxAttempts    int
	timeoutSeconds int
}

// NewJobService wires the service layer to a repository implementation.
func NewJobService(repo Repository, cfg config.ServiceConfig) JobService {
	return &jobService{
		repo:           repo,
		maxAttempts:    cfg.MaxAttempts,
		timeoutSeconds: cfg.TimeoutSeconds,
	}
}
