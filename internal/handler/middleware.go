package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// userIDKey is the context key for the authenticated user id.
// It is unexported so only this package can read or write it.
type contextKey string

const userIDKey contextKey = "userID"

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
