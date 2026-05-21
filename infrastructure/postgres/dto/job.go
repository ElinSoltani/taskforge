package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	tableName      struct{}        `pg:"jobs"`
	ID             uuid.UUID       `pg:"id,pk"`
	JobType        string          `pg:"job_type,notnull"`
	Payload        json.RawMessage `pg:"payload,notnull"`
	Priority       int             `pg:"priority,notnull"`
	Status         string          `pg:"status,notnull"`
	RunAt          time.Time       `pg:"run_at,notnull"`
	MaxAttempts    int             `pg:"max_attempts,notnull"`
	AttemptCount   int             `pg:"attempt_count,notnull"`
	TimeoutSeconds int             `pg:"timeout_seconds,notnull"`
	IdempotencyKey *string         `pg:"idempotency_key"`
	CorrelationID  *string         `pg:"correlation_id"`
	LastError      *string         `pg:"last_error"`
	CreatedAt      time.Time       `pg:"created_at,notnull"`
	UpdatedAt      time.Time       `pg:"updated_at,notnull"`
	StartedAt      *time.Time      `pg:"started_at"`
	FinishedAt     *time.Time      `pg:"finished_at"`
}
