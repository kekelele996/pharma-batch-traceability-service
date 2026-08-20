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
	return Config{Port: port, Env: env, BulkTimeout: 0}
}
