package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

// RequestLogger writes one structured line per request and echoes a request
// id so a log line can be matched to a client-side report.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Header(requestIDHeader, requestID)

		c.Next()

		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(started).Milliseconds(),
		}
		if id, ok := UserID(c); ok {
			attrs = append(attrs, "user_id", id)
		}
		logger.Info("request", attrs...)
	}
}

func newRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "unavailable"
	}
	return hex.EncodeToString(buf[:])
}

// Recover turns a handler panic into a logged 500 in the standard envelope.
func Recover(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		logger.Error("request panicked",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", recovered,
		)
		abort(c, 500, "internal server error")
	})
}
