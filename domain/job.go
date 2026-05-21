package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusCompleted Status = "completed"
	StatusDead    Status = "dead"
)

type Job struct {
	ID             uuid.UUID
	JobType        string
	Payload        json.RawMessage
	Status         Status
	RunAt          time.Time
	MaxAttempts    int
	AttemptCount   int
	TimeoutSeconds int
	IdempotencyKey *string
	CorrelationID  *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
