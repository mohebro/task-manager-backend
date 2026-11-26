package router

import (
	"github.com/gin-gonic/gin"
)

// New builds the Gin engine.
// PR 1 is scaffold only: just /health. Auth, projects and
// tasks are added in later PRs, one at a time.
func New() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}
