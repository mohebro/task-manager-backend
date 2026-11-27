package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// userIDKey is the context key for the authenticated user id.
// It is unexported so only this package can read or write it.
type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
)

// RequestLogger gives every request an id, logs one JSON line for it
// (method, path, status, duration, request id) and returns the id in
// the X-Request-ID response header, so a log line and a client report
// can be matched later.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Set(string(requestIDKey), requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()
		logger.Info("http request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

// RequestIDFromContext returns the id set by RequestLogger.
func RequestIDFromContext(c *gin.Context) (string, bool) {
	value, ok := c.Get(string(requestIDKey))
	if !ok {
		return "", false
	}
	id, ok := value.(string)
	return id, ok
}

// AuthMiddleware rejects requests without a valid Bearer JWT
// and stores the user id in the request context for handlers.
func AuthMiddleware(parseToken func(string) (uuid.UUID, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, err := parseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(string(userIDKey), userID)
		c.Next()
	}
}

// UserIDFromContext returns the authenticated user id set by AuthMiddleware.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	value, ok := c.Get(string(userIDKey))
	if !ok {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}
