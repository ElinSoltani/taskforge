package handler

import (
	"context"

	"github.com/taskforge/taskforge/domain/model"
)

type JobHandler interface {
	Execute(ctx context.Context, job *model.Job) (err error)
}
