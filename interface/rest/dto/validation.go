package dto

import (
	"encoding/json"
	"strings"
)

// ValidationErrors collects multiple field errors from one request.
type ValidationErrors struct {
	Details []ValidationError
}

func (v ValidationErrors) Error() string {
	if len(v.Details) == 0 {
		return "validation failed"
	}
	return v.Details[0].Message
}

func (r *CreateJobRequest) Validate() error {
	var details []ValidationError

	jobType := strings.TrimSpace(r.JobType)
	if jobType == "" {
		details = append(details, ValidationError{Field: "job_type", Message: "job_type is required"})
	} else if len(jobType) > maxJobTypeLen {
		details = append(details, ValidationError{Field: "job_type", Message: "job_type must be at most 128 characters"})
	}

	if len(r.Payload) > 0 && !json.Valid(r.Payload) {
		details = append(details, ValidationError{Field: "payload", Message: "payload must be valid JSON"})
	}

	if r.CorrelationID != nil {
		cid := strings.TrimSpace(*r.CorrelationID)
		if cid == "" {
			details = append(details, ValidationError{Field: "correlation_id", Message: "correlation_id cannot be empty when provided"})
		} else if len(cid) > maxCorrelationIDLen {
			details = append(details, ValidationError{Field: "correlation_id", Message: "correlation_id must be at most 256 characters"})
		}
	}

	if len(details) > 0 {
		return ValidationErrors{Details: details}
	}
	return nil
}

func (h *CreateJobHeaders) Validate() error {
	if h == nil {
		return nil
	}
	key := strings.TrimSpace(h.IdempotencyKey)
	if key == "" {
		return nil
	}
	if len(key) > maxIdempotencyKeyLen {
		return ValidationErrors{Details: []ValidationError{{
			Field:   "Idempotency-Key",
			Message: "Idempotency-Key must be at most 256 characters",
		}}}
	}
	return nil
}

func (p *GetJobParams) Validate() error {
	id := strings.TrimSpace(p.ID)
	if id == "" {
		return ValidationErrors{Details: []ValidationError{{
			Field:   "id",
			Message: "job id is required",
		}}}
	}
	return nil
}
