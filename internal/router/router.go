package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/mohebro/task-manager-backend/internal/handler"
)

// New builds the Gin engine with the auth routes wired in.
// /health stays public; /auth/me needs a Bearer token.
// API docs live at /swagger/index.html.
//
// gin.New (not gin.Default) is used on purpose: the default engine
// logs every request in its own text format, and this project logs
// requests once, as JSON, through RequestLogger instead.
func New(auth *handler.AuthHandler, logger *slog.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(handler.RequestLogger(logger))

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
