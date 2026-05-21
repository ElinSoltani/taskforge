package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	domainerror "github.com/taskforge/taskforge/domain/error"
	"github.com/taskforge/taskforge/domain/model"
)

func (s *jobService) Create(ctx context.Context, in model.CreateJobInput) (*model.Job, bool, error) {
	if in.JobType == "" {
		return nil, false, domainerror.ErrInvalidInput
	}
	if len(in.Payload) == 0 {
		in.Payload = json.RawMessage(`{}`)
	}
	if !json.Valid(in.Payload) {
		return nil, false, domainerror.ErrInvalidInput
	}

	if in.IdempotencyKey != nil && *in.IdempotencyKey != "" {
		existing, err := s.repo.GetByIdempotencyKey(ctx, *in.IdempotencyKey)
		if err == nil {
			return existing, true, nil
		}
		if !errors.Is(err, domainerror.ErrJobNotFound) {
			return nil, false, err
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, false, err
	}
	now := time.Now().UTC()
	job := &model.Job{
		ID:             id,
		JobType:        in.JobType,
		Payload:        in.Payload,
		Status:         model.JobStatusPending,
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

	if err := s.repo.Enqueue(ctx, model.QueueMessage{JobID: job.ID, JobType: job.JobType}); err != nil {
		return nil, false, err
	}
	if err := s.repo.UpdateStatus(ctx, job.ID, model.JobStatusQueued); err != nil {
		return nil, false, err
	}
	job.Status = model.JobStatusQueued
	return job, false, nil
}

func (s *jobService) Get(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	return s.repo.GetByID(ctx, id)
}
