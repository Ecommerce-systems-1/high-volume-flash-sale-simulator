package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	RedisAddr    string
	PostgresDSN  string
	DefaultStock int
	SaleDuration int
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "7860"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		PostgresDSN:  getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/flashsale?sslmode=disable"),
		DefaultStock: getEnvInt("DEFAULT_STOCK", 100),
		SaleDuration: getEnvInt("SALE_DURATION", 300),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}