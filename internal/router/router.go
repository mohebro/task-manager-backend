package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/mohebro/task-manager-backend/internal/handler"
)

// New builds the Gin engine with the auth routes wired in.
// /health stays public; /auth/me needs a Bearer token.
// API docs live at /swagger/index.html.
func New(auth *handler.AuthHandler) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authGroup := r.Group("/auth")
	authGroup.POST("/register", auth.Register)
	authGroup.POST("/login", auth.Login)
	authGroup.GET("/me", handler.AuthMiddleware(auth.ParseToken), auth.Me)

	return r
}
