package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env     string
	Port    string
	DSN     string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:  getEnv("ENV", "development"),
		Port: getEnv("PORT", "8080"),
	}

	cfg.DSN = buildDSN()
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database configuration is required")
	}

	return cfg, nil
}

func buildDSN() string {
	// Allow explicit DSN or build from components.
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		return ""
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host,
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "bread_order"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
