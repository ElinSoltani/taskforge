package rest

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type readiness interface {
	Ping(ctx context.Context) error
}

func NewRouter(h *Handler, pg readiness, rq readiness) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), correlationMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ready", func(c *gin.Context) {
		if err := pg.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"postgres": err.Error()})
			return
		}
		if err := rq.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"redis": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	v1 := r.Group("/v1")
	v1.POST("/jobs", h.CreateJob)
	v1.GET("/jobs/:id", h.GetJob)
	return r
}

func correlationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Correlation-ID")
		if id == "" {
			id = c.GetHeader("X-Request-ID")
		}
		if id != "" {
			c.Request = c.Request.WithContext(contextWithCorrelation(c.Request.Context(), id))
		}
		c.Next()
	}
}

type corrKey struct{}

func contextWithCorrelation(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, corrKey{}, id)
}
