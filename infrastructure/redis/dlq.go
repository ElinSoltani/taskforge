package redis

import (
	"context"
	"encoding/json"

	goredis "github.com/redis/go-redis/v9"
	"github.com/taskforge/taskforge/domain/model"
)

func (r *redisClient) EnqueueDLQ(ctx context.Context, msg model.DLQMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = r.db.XAdd(ctx, &goredis.XAddArgs{
		Stream: cfg.DLQStream,
		Values: map[string]interface{}{
			"message": string(data),
			"job_id":  msg.JobID.String(),
		},
	}).Result()
	return err
}
