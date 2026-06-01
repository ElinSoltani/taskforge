package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/enum"
)

// DLQMessage is published to the dead-letter Redis stream for ops and replay tooling.
// PostgreSQL job status `dead` remains the source of truth.
type DLQMessage struct {
	JobID         uuid.UUID `json:"job_id"`
	JobType       string    `json:"job_type"`
	Attempt       int       `json:"attempt"`
	MaxAttempts   int       `json:"max_attempts"`
	LastError     string    `json:"last_error"`
	DeadAt        int64     `json:"dead_at"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

// NewDLQMessage builds a DLQ payload from a job marked dead.
func NewDLQMessage(job *Job, lastError string) DLQMessage {
	msg := DLQMessage{
		JobID:       job.ID,
		JobType:     job.JobType,
		Attempt:     job.AttemptCount,
		MaxAttempts: job.MaxAttempts,
		LastError:   lastError,
		DeadAt:      time.Now().UTC().Unix(),
	}
	if job.CorrelationID != nil {
		msg.CorrelationID = *job.CorrelationID
	}
	return msg
}

// ApplyDead updates in-memory job fields after a successful dead transition.
func (j *Job) ApplyDead(lastError string) {
	now := time.Now().UTC()
	j.Status = enum.JobStatusDead
	j.LastError = &lastError
	j.UpdatedAt = now
}
