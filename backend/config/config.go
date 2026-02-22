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
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		DBPath:          getEnv("DB_PATH", "./data/drum_trainer.db"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		UploadDir:       getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxUploadSizeMB: getEnvInt64("MAX_UPLOAD_SIZE_MB", 16),
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
