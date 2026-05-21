package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
	JobStatusDead      JobStatus = "dead"
)

type Job struct {
	ID             uuid.UUID
	JobType        string
	Payload        json.RawMessage
	Status         JobStatus
	RunAt          time.Time
	MaxAttempts    int
	AttemptCount   int
	TimeoutSeconds int
	IdempotencyKey *string
	CorrelationID  *string
	LastError      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (j *Job) HasAttemptsRemaining() bool {
	return j.AttemptCount < j.MaxAttempts
}

type CreateJobInput struct {
	JobType        string
	Payload        json.RawMessage
	IdempotencyKey *string
	CorrelationID  *string
}

type QueueMessage struct {
	JobID   uuid.UUID `json:"job_id"`
	JobType string    `json:"job_type"`
}

type ConsumedMessage struct {
	Stream    string
	MessageID string
	Payload   QueueMessage
}
