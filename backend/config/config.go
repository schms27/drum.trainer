package config

import (
	"os"
	"strconv"
	"strings"
)

// DefaultSessionSecret is the placeholder used when SESSION_SECRET is unset.
const DefaultSessionSecret = "change-me-in-production"

// Config holds all server configuration values.
type Config struct {
	Port            string
	DBPath          string
	DatabaseURL     string
	UploadDir       string
	MaxUploadSizeMB int64
	MaxPilotSizeMB  int64
	MaxSessionHits  int

	// OAuth providers.
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	// Auth.
	SessionSecret string // used to sign CSRF state cookies
	BaseURL       string // public base URL, e.g. "https://drums.example.com"
	DevMode       bool   // relaxes Secure cookie flag for HTTP localhost

	// Moderation. Users signing in with one of these emails can review
	// reports and remove shared tracks.
	AdminEmails []string
	// ContactEmail is published on the terms and copyright pages as the
	// address for copyright notices.
	ContactEmail string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		DBPath:          getEnv("DB_PATH", "./data/drum_trainer.db"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		UploadDir:       getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxUploadSizeMB: getEnvInt64("MAX_UPLOAD_SIZE_MB", 16),
		MaxPilotSizeMB:  getEnvInt64("MAX_PILOT_SIZE_MB", 150),
		MaxSessionHits:  int(getEnvInt64("MAX_SESSION_HITS", 20_000)),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),

		SessionSecret: getEnv("SESSION_SECRET", DefaultSessionSecret),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		DevMode:       os.Getenv("DEV_MODE") == "true",

		AdminEmails:  splitList(os.Getenv("ADMIN_EMAILS")),
		ContactEmail: os.Getenv("CONTACT_EMAIL"),
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

// splitList splits a comma-separated value, trimming and lower-casing each
// entry and dropping empty ones.
func splitList(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		if p := strings.ToLower(strings.TrimSpace(part)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsAdmin reports whether email belongs to a configured admin.
func (c Config) IsAdmin(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false
	}
	for _, a := range c.AdminEmails {
		if a == email {
			return true
		}
	}
	return false
}
