package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config holds everything the server reads from the environment.
// JWT_SECRET is required: the server refuses to start without it,
// so a demo token can never be signed with an empty secret.
type Config struct {
	DatabaseURL string
	HTTPPort    string
	JWTSecret   string
	JWTTTL      time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		HTTPPort:    os.Getenv("HTTP_PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTTTL:      24 * time.Hour,
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if v := os.Getenv("JWT_TTL_HOURS"); v != "" {
		hours, err := strconv.Atoi(v)
		if err != nil || hours <= 0 {
			return nil, errors.New("JWT_TTL_HOURS must be a positive number of hours")
		}
		cfg.JWTTTL = time.Duration(hours) * time.Hour
	}
	return cfg, nil
}
