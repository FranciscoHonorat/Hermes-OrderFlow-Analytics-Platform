package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(ctx context.Context, c *gin.Context) {
	// Perform health check logic here, e.g., check database connection, cache, etc.
	// For simplicity, we'll just return a 200 OK response.
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
