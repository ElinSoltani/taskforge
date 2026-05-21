package domain

import (
	"context"

	"github.com/google/uuid"
)

type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*Job, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	Claim(ctx context.Context, id uuid.UUID) (*Job, error)
	Complete(ctx context.Context, id uuid.UUID) error
}

type Queue interface {
	EnsureGroup(ctx context.Context) error
	Enqueue(ctx context.Context, msg QueueMessage) error
	Dequeue(ctx context.Context) (*ConsumedMessage, error)
	Ack(ctx context.Context, stream, messageID string) error
	Ping(ctx context.Context) error
}

type ConsumedMessage struct {
	Stream    string
	MessageID string
	Payload   QueueMessage
}

type JobHandler interface {
	Execute(ctx context.Context, job *Job) error
}
