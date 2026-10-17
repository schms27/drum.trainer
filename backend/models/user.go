package models

import "time"

// User represents an authenticated user account.
type User struct {
	ID        int64  `json:"id"`
	Provider  string `json:"provider"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	// SharingBanned is set once the user has collected too many strikes for
	// sharing infringing tracks.
	SharingBanned bool `json:"sharing_banned"`
	// IsAdmin is derived from config.AdminEmails on each request, never stored.
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}
