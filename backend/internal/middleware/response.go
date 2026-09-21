package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Error payload shape shared with the handlers so every failure looks the same.
type errorBody struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, errorBody{Success: false, Error: message})
}

// AbortInternal logs an unexpected error with the request path and answers
// with a generic 500. Handlers use the api package's equivalent; this one is
// for the middleware chain.
func AbortInternal(c *gin.Context, logger *slog.Logger, err error) {
	logger.Error("request failed",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"error", err.Error(),
	)
	abort(c, 500, "internal server error")
}
