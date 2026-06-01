package repository

import (
	"context"

	"github.com/taskforge/taskforge/domain/model"
)

type JobQueue interface {
	EnsureGroup(ctx context.Context) (err error)
	Enqueue(ctx context.Context, msg model.QueueMessage) (err error)
	EnqueueDLQ(ctx context.Context, msg model.DLQMessage) (err error)
	Dequeue(ctx context.Context) (msg *model.ConsumedMessage, err error)
	Ack(ctx context.Context, stream, messageID string) (err error)
	Ping(ctx context.Context) (err error)
}
