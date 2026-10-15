package db

import (
	"testing"
	"time"
)

func TestDeleteExpiredSessions(t *testing.T) {
	db := openTestDB(t)

	user, err := UpsertUser(db, "github", "1", "a@example.com", "A", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := CreateSession(db, "live", user.ID); err != nil {
		t.Fatal(err)
	}
	expired := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO user_sessions (token, user_id, expires_at) VALUES (?, ?, ?)`, "old", user.ID, expired); err != nil {
		t.Fatal(err)
	}

	n, err := DeleteExpiredSessions(db)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("deleted %d sessions, want 1", n)
	}
	if _, err := ValidateSession(db, "live"); err != nil {
		t.Fatalf("live session should still be valid: %v", err)
	}
}
