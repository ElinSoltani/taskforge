package domainerror

import "errors"

var (
	ErrJobNotFound       = errors.New("job not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidTransition = errors.New("invalid status transition")
)
