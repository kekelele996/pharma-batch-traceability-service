package config

import (
	"os"
	"time"
)

type Config struct {
	Port        string
	Env         string
	BulkTimeout time.Duration
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18090"
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	timeout := time.Duration(0)
	if raw := os.Getenv("BULK_VERIFY_TIMEOUT"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			timeout = d
		}
	}
	return Config{Port: port, Env: env, BulkTimeout: timeout}
}
