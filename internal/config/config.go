package config

import (
	"os"
	"strings"
)

type Config struct {
	RedisAddr string
	RedisDB   int
	RedisPass string
	HTTPPort  string
}

func Load() Config {
	return Config{
		RedisAddr: strings.TrimSpace(getEnv("REDIS_ADDR", "localhost:6379")),
		RedisDB:   0,
		RedisPass: strings.TrimSpace(getEnv("REDIS_PASS", "")),
		HTTPPort:  strings.TrimSpace(getEnv("HTTP_PORT", "8080")),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
