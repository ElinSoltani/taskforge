package dto

import (
	"github.com/taskforge/taskforge/domain/model"
)

func (d *Job) ToDomain() *model.Job {
	if d == nil {
		return nil
	}
	return &model.Job{
		ID:             d.ID,
		JobType:        d.JobType,
		Payload:        d.Payload,
		Status:         model.JobStatus(d.Status),
		RunAt:          d.RunAt,
		MaxAttempts:    d.MaxAttempts,
		AttemptCount:   d.AttemptCount,
		TimeoutSeconds: d.TimeoutSeconds,
		IdempotencyKey: d.IdempotencyKey,
		CorrelationID:  d.CorrelationID,
		LastError:      d.LastError,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func (d *Job) FromDomain(m *model.Job) {
	if m == nil {
		return
	}
	d.ID = m.ID
	d.JobType = m.JobType
	d.Payload = m.Payload
	d.Priority = 2
	d.Status = string(m.Status)
	d.RunAt = m.RunAt
	d.MaxAttempts = m.MaxAttempts
	d.AttemptCount = m.AttemptCount
	d.TimeoutSeconds = m.TimeoutSeconds
	d.IdempotencyKey = m.IdempotencyKey
	d.CorrelationID = m.CorrelationID
	d.LastError = m.LastError
	d.CreatedAt = m.CreatedAt
	d.UpdatedAt = m.UpdatedAt
}
