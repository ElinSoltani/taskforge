package repository

import (
	drep "github.com/taskforge/taskforge/domain/repository"
)

type Repository struct {
	jobs  drep.JobStore
	queue drep.JobQueue
}

func NewRepository(jobs drep.JobStore, queue drep.JobQueue) *Repository {
	return &Repository{jobs: jobs, queue: queue}
}
