package redis

import "errors"

var (
	ErrInvalidRedisConfig = errors.New("invalid redis config")
	ErrRedisNotReady      = errors.New("redis not ready")
)
