package config

import "os"

type Config struct {
	Port string
	Env  string
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
	return Config{Port: port, Env: env}
}
