package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	domainerror "github.com/taskforge/taskforge/domain/error"
	"github.com/taskforge/taskforge/interface/rest/dto"
)

func writeBindError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
}

func writeValidationError(c *gin.Context, err error) {
	var ve dto.ValidationErrors
	if errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation failed",
			Details: ve.Details,
		})
		return
	}
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
}

func writeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerror.ErrJobNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domainerror.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domainerror.ErrInvalidTransition):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal error"})
	}
}
