package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taskforge/taskforge/interface/rest/dto"
	"github.com/taskforge/taskforge/service"
)

type Handler struct {
	jobs    service.JobService
	baseURL string
}

func NewHandler(jobs service.JobService, baseURL string) *Handler {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &Handler{jobs: jobs, baseURL: baseURL}
}

// CreateJob accepts a job submission, validates the REST DTO, maps to domain, and delegates to service.
func (h *Handler) CreateJob(c *gin.Context) {
	var req dto.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}
	if err := req.Validate(); err != nil {
		writeValidationError(c, err)
		return
	}

	headers := &dto.CreateJobHeaders{IdempotencyKey: c.GetHeader("Idempotency-Key")}
	if err := headers.Validate(); err != nil {
		writeValidationError(c, err)
		return
	}

	input := req.ToCreateJobInput(headers)
	job, duplicate, err := h.jobs.Create(c.Request.Context(), input)
	if err != nil {
		writeDomainError(c, err)
		return
	}

	code := http.StatusAccepted
	if duplicate {
		code = http.StatusOK
	}
	c.JSON(code, dto.CreateJobResponseFromDomain(job, duplicate, h.baseURL))
}

// GetJob loads a job by id: validate path param → domain UUID → service → response DTO.
func (h *Handler) GetJob(c *gin.Context) {
	var params dto.GetJobParams
	if err := c.ShouldBindUri(&params); err != nil {
		writeBindError(c, err)
		return
	}
	if err := params.Validate(); err != nil {
		writeValidationError(c, err)
		return
	}

	id, err := dto.ParseJobID(params.ID)
	if err != nil {
		writeValidationError(c, err)
		return
	}

	job, err := h.jobs.Get(c.Request.Context(), id)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.JobResponseFromDomain(job, h.baseURL))
}
