package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain/model"
)

// ToDomain maps a validated REST request into a domain job (creation fields only).
func (r *CreateJobRequest) ToDomain(headers *CreateJobHeaders) *model.Job {
	payload := r.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	var idem *string
	if headers != nil {
		if key := strings.TrimSpace(headers.IdempotencyKey); key != "" {
			idem = &key
		}
	}

	var correlationID *string
	if r.CorrelationID != nil {
		if cid := strings.TrimSpace(*r.CorrelationID); cid != "" {
			correlationID = &cid
		}
	}

	return &model.Job{
		JobType:        strings.TrimSpace(r.JobType),
		Payload:        payload,
		IdempotencyKey: idem,
		CorrelationID:  correlationID,
	}
}

// ParseJobID validates and parses a path job id.
func ParseJobID(raw string) (uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return uuid.Nil, ValidationErrors{Details: []ValidationError{{
			Field: "id", Message: "job id is required",
		}}}
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, ValidationErrors{Details: []ValidationError{{
			Field: "id", Message: "job id must be a valid UUID",
		}}}
	}
	return id, nil
}

func JobResponseFromDomain(job *model.Job, baseURL string) JobResponse {
	if job == nil {
		return JobResponse{}
	}
	self := baseURL + "/v1/jobs/" + job.ID.String()
	return JobResponse{
		ID:             job.ID.String(),
		JobType:        job.JobType,
		Status:         string(job.Status),
		Payload:        job.Payload,
		AttemptCount:   job.AttemptCount,
		MaxAttempts:    job.MaxAttempts,
		TimeoutSeconds: job.TimeoutSeconds,
		CorrelationID:  job.CorrelationID,
		LastError:      job.LastError,
		RunAt:          formatTime(job.RunAt),
		CreatedAt:      formatTime(job.CreatedAt),
		UpdatedAt:      formatTime(job.UpdatedAt),
		Links:          JobLinks{Self: self},
	}
}

func FromDomain(job *model.Job, duplicate bool, baseURL string) CreateJobResponse {
	return CreateJobResponse{
		Job:       JobResponseFromDomain(job, baseURL),
		Duplicate: duplicate,
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
