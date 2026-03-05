package config

import (
	"os"
	"strconv"
)

// Config holds all server configuration values.
type Config struct {
	Port            string
	DBPath          string
	DatabaseURL     string
	UploadDir       string
	MaxUploadSizeMB int64

	// OAuth providers.
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	// Auth.
	SessionSecret string // used to sign CSRF state cookies
	BaseURL       string // public base URL, e.g. "https://drums.example.com"
	DevMode       bool   // relaxes Secure cookie flag for HTTP localhost
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		DBPath:          getEnv("DB_PATH", "./data/drum_trainer.db"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		UploadDir:       getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxUploadSizeMB: getEnvInt64("MAX_UPLOAD_SIZE_MB", 16),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),

		SessionSecret: getEnv("SESSION_SECRET", "change-me-in-production"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		DevMode:       os.Getenv("DEV_MODE") == "true",
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}
