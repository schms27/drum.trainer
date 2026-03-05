package models

import "time"

// User represents an authenticated user account.
type User struct {
	ID        int64     `json:"id"`
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}
