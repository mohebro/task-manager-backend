package main

import (
	"log"
	"os"

	"github.com/mohebro/task-manager-backend/internal/router"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	engine := router.New()
	log.Printf("listening on :%s", port)
	if err := engine.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
