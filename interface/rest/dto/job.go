package dto

import "encoding/json"

const (
	maxJobTypeLen       = 128
	maxCorrelationIDLen = 256
	maxIdempotencyKeyLen = 256
)

// CreateJobRequest is the HTTP body for POST /v1/jobs.
type CreateJobRequest struct {
	JobType       string          `json:"job_type"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID *string         `json:"correlation_id"`
}

// CreateJobHeaders carries transport-level fields for job creation.
type CreateJobHeaders struct {
	IdempotencyKey string
}

// GetJobParams is the path/query input for GET /v1/jobs/:id.
type GetJobParams struct {
	ID string `uri:"id"`
}

// JobResponse is the HTTP representation of a job.
type JobResponse struct {
	ID             string          `json:"id"`
	JobType        string          `json:"job_type"`
	Status         string          `json:"status"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	AttemptCount   int             `json:"attempt_count"`
	MaxAttempts    int             `json:"max_attempts"`
	TimeoutSeconds int             `json:"timeout_seconds"`
	CorrelationID  *string         `json:"correlation_id,omitempty"`
	LastError      *string         `json:"last_error,omitempty"`
	RunAt          string          `json:"run_at,omitempty"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
	Links          JobLinks        `json:"links"`
}

// CreateJobResponse wraps a newly accepted job.
type CreateJobResponse struct {
	Job        JobResponse `json:"job"`
	Duplicate  bool        `json:"duplicate"`
}

// JobLinks provides discoverable related routes.
type JobLinks struct {
	Self string `json:"self"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError describes a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
