package model

import "github.com/google/uuid"

type QueueMessage struct {
	JobID   uuid.UUID `json:"job_id"`
	JobType string    `json:"job_type"`
}
