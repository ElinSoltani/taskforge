package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain"
)

type JobService struct {
	repo  domain.JobRepository
	queue domain.Queue
}

func NewJobService(repo domain.JobRepository, queue domain.Queue) *JobService {
	return &JobService{repo: repo, queue: queue}
}

func (s *JobService) Create(ctx context.Context, in domain.CreateJobInput) (*domain.Job, bool, error) {
	if in.JobType == "" {
		return nil, false, domain.ErrInvalidInput
	}
	if len(in.Payload) == 0 {
		in.Payload = json.RawMessage(`{}`)
	}
	if !json.Valid(in.Payload) {
		return nil, false, domain.ErrInvalidInput
	}

	if in.IdempotencyKey != nil && *in.IdempotencyKey != "" {
		existing, err := s.repo.GetByIdempotencyKey(ctx, *in.IdempotencyKey)
		if err == nil {
			return existing, true, nil
		}
		if !errors.Is(err, domain.ErrJobNotFound) {
			return nil, false, err
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, false, err
	}
	now := time.Now().UTC()
	job := &domain.Job{
		ID:             id,
		JobType:        in.JobType,
		Payload:        in.Payload,
		Status:         domain.StatusPending,
		RunAt:          now,
		MaxAttempts:    5,
		TimeoutSeconds: 300,
		IdempotencyKey: in.IdempotencyKey,
		CorrelationID:  in.CorrelationID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, job); err != nil {
		if in.IdempotencyKey != nil {
			existing, gerr := s.repo.GetByIdempotencyKey(ctx, *in.IdempotencyKey)
			if gerr == nil {
				return existing, true, nil
			}
		}
		return nil, false, err
	}

	if err := s.queue.Enqueue(ctx, domain.QueueMessage{JobID: job.ID, JobType: job.JobType}); err != nil {
		return nil, false, err
	}
	if err := s.repo.UpdateStatus(ctx, job.ID, domain.StatusQueued); err != nil {
		return nil, false, err
	}
	job.Status = domain.StatusQueued
	return job, false, nil
}

func (s *JobService) Get(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	return s.repo.GetByID(ctx, id)
}
