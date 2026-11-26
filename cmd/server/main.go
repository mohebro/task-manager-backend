package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mohebro/task-manager-backend/internal/config"
	"github.com/mohebro/task-manager-backend/internal/handler"
	"github.com/mohebro/task-manager-backend/internal/repository"
	"github.com/mohebro/task-manager-backend/internal/router"
	"github.com/mohebro/task-manager-backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	users := repository.NewPostgresUserRepository(pool)
	authService := service.NewAuthService(users, cfg.JWTSecret, cfg.JWTTTL)
	authHandler := handler.NewAuthHandler(authService, users)

	engine := router.New(authHandler)
	log.Printf("listening on :%s", cfg.HTTPPort)
	if err := engine.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
