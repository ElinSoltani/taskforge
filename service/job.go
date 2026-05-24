package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	domainerror "github.com/taskforge/taskforge/domain/error"
	"github.com/taskforge/taskforge/domain/enum"
	"github.com/taskforge/taskforge/domain/model"
)

func (s *jobService) Create(ctx context.Context, in *model.Job) (*model.Job, bool, error) {
	if in == nil || in.JobType == "" {
		return nil, false, domainerror.ErrInvalidInput
	}
	payload := in.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	if !json.Valid(payload) {
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
		Payload:        payload,
		Status:         enum.JobStatusPending,
		RunAt:          now,
		MaxAttempts:    s.maxAttempts,
		TimeoutSeconds: s.timeoutSeconds,
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

	// Mark queued before Redis enqueue so workers never claim while status is still pending.
	if err := s.repo.UpdateStatus(ctx, job.ID, enum.JobStatusQueued); err != nil {
		return nil, false, err
	}
	if err := s.repo.Enqueue(ctx, model.QueueMessage{JobID: job.ID, JobType: job.JobType}); err != nil {
		return nil, false, err
	}
	job.Status = enum.JobStatusQueued
	return job, false, nil
}

func (s *jobService) Get(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	return s.repo.GetByID(ctx, id)
}
