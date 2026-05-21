package worker

import (
	domainhandler "github.com/taskforge/taskforge/domain/handler"
)

// DefaultHandlers returns built-in job handlers for worker registration.
func DefaultHandlers() map[string]domainhandler.JobHandler {
	return map[string]domainhandler.JobHandler{
		"ping": PingHandler{},
		"fail": FailHandler{},
	}
}
