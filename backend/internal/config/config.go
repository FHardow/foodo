package config

import (
	"encoding/base64"
	"fmt"
	"os"
)

type Config struct {
	Env               string
	Port              string
	DSN               string
	KeycloakURL       string
	KeycloakRealm     string
	CORSOrigin        string
	TelegramBotToken  string
	TelegramChatID    string
	VAPIDPublicKey    string
	VAPIDPrivateKey   string
	VAPIDSubject      string
	PushEncryptionKey string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:               getEnv("ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		KeycloakURL:       getEnv("KEYCLOAK_URL", "http://localhost:8180"),
		KeycloakRealm:     getEnv("KEYCLOAK_REALM", "foodo"),
		CORSOrigin:        os.Getenv("CORS_ORIGIN"),
		TelegramBotToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:    os.Getenv("TELEGRAM_CHAT_ID"),
		VAPIDPublicKey:    os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:   os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDSubject:      os.Getenv("VAPID_SUBJECT"),
		PushEncryptionKey: os.Getenv("PUSH_ENCRYPTION_KEY"),
	}

	cfg.DSN = buildDSN()
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database configuration is required")
	}

	if cfg.PushEncryptionKey == "" {
		return nil, fmt.Errorf("PUSH_ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(cfg.PushEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("PUSH_ENCRYPTION_KEY must be valid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("PUSH_ENCRYPTION_KEY must decode to 32 bytes, got %d", len(key))
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
		getEnv("DB_NAME", "foodo"),
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
