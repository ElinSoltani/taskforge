package repository

import (
	"context"

	"github.com/taskforge/taskforge/domain/model"
)

func (r *Repository) EnsureGroup(ctx context.Context) (err error) {
	return r.queue.EnsureGroup(ctx)
}

func (r *Repository) Enqueue(ctx context.Context, msg model.QueueMessage) (err error) {
	return r.queue.Enqueue(ctx, msg)
}

func (r *Repository) Dequeue(ctx context.Context) (msg *model.ConsumedMessage, err error) {
	return r.queue.Dequeue(ctx)
}

func (r *Repository) Ack(ctx context.Context, stream, messageID string) (err error) {
	return r.queue.Ack(ctx, stream, messageID)
}

func (r *Repository) PingQueue(ctx context.Context) (err error) {
	return r.queue.Ping(ctx)
}
