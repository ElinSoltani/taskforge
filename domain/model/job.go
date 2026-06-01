package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/enum"
)

type Job struct {
	ID             uuid.UUID
	JobType        string
	Payload        json.RawMessage
	Status         enum.JobStatus
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
