package router

import (
	"github.com/gin-gonic/gin"

	"github.com/mohebro/task-manager-backend/internal/handler"
)

// New builds the Gin engine with the auth routes wired in.
// /health stays public; /auth/me needs a Bearer token.
func New(auth *handler.AuthHandler) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authGroup := r.Group("/auth")
	authGroup.POST("/register", auth.Register)
	authGroup.POST("/login", auth.Login)
	authGroup.GET("/me", handler.AuthMiddleware(auth.ParseToken), auth.Me)

	return r
}
