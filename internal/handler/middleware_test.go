package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	parseToken := func(token string) (uuid.UUID, error) {
		if token == "good-token" {
			return uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil
		}
		return uuid.Nil, errors.New("bad token")
	}

	r := gin.New()
	r.Use(RequestLogger(discardLogger()))
	authed := r.Group("/", AuthMiddleware(parseToken))
	authed.GET("/me", func(c *gin.Context) {
		id, ok := UserIDFromContext(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id.String()})
	})

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic abc", http.StatusUnauthorized},
		{"bad token", "Bearer bad-token", http.StatusUnauthorized},
		{"good token", "Bearer good-token", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %s)", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestRequestLoggerSetsRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestLogger(discardLogger()))
	r.GET("/health", func(c *gin.Context) {
		id, ok := RequestIDFromContext(c)
		if !ok || id == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no request id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
	}
	headerID := w.Header().Get("X-Request-ID")
	if headerID == "" {
		t.Fatal("X-Request-ID header is missing")
	}
	if _, err := uuid.Parse(headerID); err != nil {
		t.Fatalf("X-Request-ID %q is not a uuid: %v", headerID, err)
	}
}
