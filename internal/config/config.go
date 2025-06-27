package config

import (
	"os"
)

type Config struct {
	RedisAddr string
	RedisDB   int
	RedisPass string
	HTTPPort  string
}

func Load() Config {
	return Config{
		RedisAddr: getEnv("REDIS_ADDR", "localhosyt:6379"),
		RedisDB:   0,
		RedisPass: getEnv("REDIS_PASS", ""),
		HTTPPort:  getEnv("HTTP_PORT", "8080"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
