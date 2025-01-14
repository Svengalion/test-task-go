package config

import (
	"os"
)

type Config struct {
	Port               string
	PostgresDSN        string
	ExternalAPIBaseURL string
}

func New() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		PostgresDSN:        getEnv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/dbname?sslmode=disable"),
		ExternalAPIBaseURL: getEnv("EXTERNAL_API_BASE_URL", "http://localhost:9000"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
