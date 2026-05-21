package domainerror

// RetryableError signals the worker should schedule another attempt with backoff.
type RetryableError struct {
	Code    string
	Message string
}

func (e *RetryableError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "retryable error"
}

// TerminalError signals the job should move to dead letter without further retries.
type TerminalError struct {
	Code    string
	Message string
}

func (e *TerminalError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "terminal error"
}
