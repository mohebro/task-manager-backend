package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	base := map[string]string{
		"DATABASE_URL": "postgres://taskmanager:taskmanager@localhost:5434/taskmanager?sslmode=disable",
		"JWT_SECRET":   "test-secret",
	}

	t.Run("defaults", func(t *testing.T) {
		for k, v := range base {
			t.Setenv(k, v)
		}
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.HTTPPort != "8080" {
			t.Fatalf("port = %q, want 8080", cfg.HTTPPort)
		}
		if cfg.JWTTTL != 24*time.Hour {
			t.Fatalf("ttl = %v, want 24h", cfg.JWTTTL)
		}
		if cfg.AppEnv != "production" {
			t.Fatalf("env = %q, want production (safe default)", cfg.AppEnv)
		}
		if cfg.LogLevel != "info" {
			t.Fatalf("level = %q, want info", cfg.LogLevel)
		}
		if cfg.SwaggerHost != "" {
			t.Fatalf("swagger host = %q, want empty", cfg.SwaggerHost)
		}
	})

	t.Run("dev and log level honored", func(t *testing.T) {
		for k, v := range base {
			t.Setenv(k, v)
		}
		t.Setenv("APP_ENV", "dev")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("JWT_TTL_HOURS", "2")
		t.Setenv("SWAGGER_HOST", "api.example.com:8080")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.AppEnv != "dev" || cfg.LogLevel != "debug" {
			t.Fatalf("env = %q, level = %q", cfg.AppEnv, cfg.LogLevel)
		}
		if cfg.JWTTTL != 2*time.Hour {
			t.Fatalf("ttl = %v, want 2h", cfg.JWTTTL)
		}
		if cfg.SwaggerHost != "api.example.com:8080" {
			t.Fatalf("swagger host = %q", cfg.SwaggerHost)
		}
	})

	cases := []struct {
		name  string
		unset []string
		set   map[string]string
	}{
		{"missing DATABASE_URL", []string{"DATABASE_URL"}, nil},
		{"missing JWT_SECRET", []string{"JWT_SECRET"}, nil},
		{"bad JWT_TTL_HOURS", nil, map[string]string{"JWT_TTL_HOURS": "zero"}},
		{"bad LOG_LEVEL", nil, map[string]string{"LOG_LEVEL": "loud"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range base {
				t.Setenv(k, v)
			}
			for _, k := range tc.unset {
				t.Setenv(k, "")
			}
			for k, v := range tc.set {
				t.Setenv(k, v)
			}
			if _, err := Load(); err == nil {
				t.Fatal("Load should fail")
			}
		})
	}
}
