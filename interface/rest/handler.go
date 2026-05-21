package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskforge/taskforge/domain"
	"github.com/taskforge/taskforge/service"
)

type Handler struct {
	jobs *service.JobService
}

func NewHandler(jobs *service.JobService) *Handler {
	return &Handler{jobs: jobs}
}

type createJobRequest struct {
	JobType   string          `json:"job_type" binding:"required"`
	Payload   json.RawMessage `json:"payload"`
	CorrelationID *string     `json:"correlation_id"`
}

type jobResponse struct {
	ID        string `json:"id"`
	JobType   string `json:"job_type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req createJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var idem *string
	if k := c.GetHeader("Idempotency-Key"); k != "" {
		idem = &k
	}

	job, duplicate, err := h.jobs.Create(c.Request.Context(), domain.CreateJobInput{
		JobType:        req.JobType,
		Payload:        req.Payload,
		IdempotencyKey: idem,
		CorrelationID:  req.CorrelationID,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	code := http.StatusAccepted
	if duplicate {
		code = http.StatusOK
	}
	c.JSON(code, jobResponse{
		ID:        job.ID.String(),
		JobType:   job.JobType,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) GetJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	job, err := h.jobs.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, jobResponse{
		ID:        job.ID.String(),
		JobType:   job.JobType,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt.Format(time.RFC3339),
	})
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrJobNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
