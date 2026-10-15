package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/simu/drum-trainer/backend/models"
)

// UpsertUser finds a user by (provider, providerID) and updates their profile
// fields, or inserts a new user if not found. Returns the user with their DB id.
func UpsertUser(db *DB, provider, providerID, email, name, avatarURL string) (*models.User, error) {
	var user models.User
	var ct scanTime

	err := db.QueryRow(
		db.Rebind(`SELECT id, provider, email, name, avatar_url, created_at
		 FROM users WHERE provider = ? AND provider_id = ?`),
		provider, providerID,
	).Scan(&user.ID, &user.Provider, &user.Email, &user.Name, &user.AvatarURL, &ct)

	if err == sql.ErrNoRows {
		err = db.QueryRow(
			db.Rebind(`INSERT INTO users (provider, provider_id, email, name, avatar_url)
			 VALUES (?, ?, ?, ?, ?) RETURNING id, provider, email, name, avatar_url, created_at`),
			provider, providerID, email, name, avatarURL,
		).Scan(&user.ID, &user.Provider, &user.Email, &user.Name, &user.AvatarURL, &ct)
		if err != nil {
			return nil, fmt.Errorf("insert user: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	} else {
		// Update mutable fields (name, email, avatar may change on the provider side).
		if _, err := db.Exec(
			db.Rebind(`UPDATE users SET email = ?, name = ?, avatar_url = ? WHERE id = ?`),
			email, name, avatarURL, user.ID,
		); err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
		user.Email = email
		user.Name = name
		user.AvatarURL = avatarURL
	}

	user.CreatedAt = ct.T
	return &user, nil
}

// GetUserByID fetches a user by primary key.
func GetUserByID(db *DB, id int64) (*models.User, error) {
	var user models.User
	var ct scanTime
	err := db.QueryRow(
		db.Rebind(`SELECT id, provider, email, name, avatar_url, created_at FROM users WHERE id = ?`),
		id,
	).Scan(&user.ID, &user.Provider, &user.Email, &user.Name, &user.AvatarURL, &ct)
	if err != nil {
		return nil, err
	}
	user.CreatedAt = ct.T
	return &user, nil
}

// CreateSession inserts a new session token for a user with a 30-day expiry.
func CreateSession(db *DB, token string, userID int64) error {
	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	_, err := db.Exec(
		db.Rebind(`INSERT INTO user_sessions (token, user_id, expires_at) VALUES (?, ?, ?)`),
		token, userID, expiresAt,
	)
	return err
}

// ValidateSession checks the token exists and is not expired, and returns the
// associated user. Returns sql.ErrNoRows if the session is invalid or expired.
func ValidateSession(db *DB, token string) (*models.User, error) {
	var user models.User
	var ct scanTime
	err := db.QueryRow(
		db.Rebind(`SELECT u.id, u.provider, u.email, u.name, u.avatar_url, u.created_at
		 FROM user_sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE s.token = ? AND s.expires_at > ?`),
		token, time.Now().UTC().Format(time.RFC3339),
	).Scan(&user.ID, &user.Provider, &user.Email, &user.Name, &user.AvatarURL, &ct)
	if err != nil {
		return nil, err
	}
	user.CreatedAt = ct.T
	return &user, nil
}

// DeleteSession removes a session token from the database.
func DeleteSession(db *DB, token string) error {
	_, err := db.Exec(db.Rebind(`DELETE FROM user_sessions WHERE token = ?`), token)
	return err
}

// DeleteExpiredSessions removes all session tokens whose expiry has passed and
// returns how many were deleted.
func DeleteExpiredSessions(db *DB) (int64, error) {
	res, err := db.Exec(
		db.Rebind(`DELETE FROM user_sessions WHERE expires_at <= ?`),
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
