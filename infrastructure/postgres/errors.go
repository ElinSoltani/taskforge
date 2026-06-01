package postgres

import "errors"

var (
	ErrInvalidPostgresConfig = errors.New("invalid postgres config")
	ErrInitiatePostgres      = errors.New("initiate postgres error")
	ErrPostgresNotReady      = errors.New("postgres not ready")
)
