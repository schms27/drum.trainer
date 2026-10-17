package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

// sharingFixture creates an owner with one track (two notes, one mapping
// override, one pilot) and a second user to receive the share.
func sharingFixture(t *testing.T) (db *DB, owner, other *models.User, trackID int64) {
	t.Helper()
	db = openTestDB(t)
	owner, err := UpsertUser(db, "github", "1", "a@example.com", "A", "")
	if err != nil {
		t.Fatal(err)
	}
	other, err = UpsertUser(db, "github", "2", "b@example.com", "B", "")
	if err != nil {
		t.Fatal(err)
	}
	track := &models.Track{Name: "groove", Filename: "groove.mid", FilePath: "abc.mid", FileSize: 10,
		DurationMs: 2000, NoteCount: 2, PPQ: 480, BPM: 120, TimeSigNumerator: 4}
	notes := []models.NoteEvent{
		{NoteNumber: 36, TimeMs: 0, Velocity: 100, Instrument: "kick"},
		{NoteNumber: 38, TimeMs: 500, Velocity: 90, Instrument: "snare"},
	}
	if err := InsertTrack(db, track, notes, owner.ID); err != nil {
		t.Fatal(err)
	}
	if err := SetTrackMapping(db, track.ID, []models.NoteMapping{{NoteNumber: 38, Instrument: "tom_high"}}); err != nil {
		t.Fatal(err)
	}
	if err := InsertPilotTrack(db, &models.PilotTrack{TrackID: track.ID, Name: "band", Filename: "band.mp3",
		FilePath: "band.mp3", FileSize: 1, MimeType: "audio/mpeg"}, owner.ID); err != nil {
		t.Fatal(err)
	}
	return db, owner, other, track.ID
}

func TestShareAndResolve(t *testing.T) {
	db, owner, _, trackID := sharingFixture(t)

	token, err := ShareTrack(db, trackID, owner.ID, ShareEvent{RightsBasis: "original", TermsVer: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 22 {
		t.Fatalf("token %q has length %d, want 22", token, len(token))
	}
	again, err := ShareTrack(db, trackID, owner.ID, ShareEvent{RightsBasis: "licensed", TermsVer: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	if again != token {
		t.Fatal("re-sharing should keep the existing token")
	}

	got, ref, err := GetSharedTrack(db, token)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != trackID || ref.OwnerID != owner.ID || *got.RightsBasis != "licensed" || got.Visibility != models.VisibilityUnlisted {
		t.Fatalf("unexpected shared track: %+v", got)
	}

	var events int
	db.QueryRow(`SELECT COUNT(*) FROM track_share_events WHERE track_id = ?`, trackID).Scan(&events)
	if events != 2 {
		t.Fatalf("got %d share events, want 2", events)
	}
}

func TestShareRequiresOwnership(t *testing.T) {
	db, _, other, trackID := sharingFixture(t)
	if _, err := ShareTrack(db, trackID, other.ID, ShareEvent{RightsBasis: "original"}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("got %v, want sql.ErrNoRows", err)
	}
}

func TestRotateAndUnshare(t *testing.T) {
	db, owner, _, trackID := sharingFixture(t)

	if _, err := RotateShareToken(db, trackID, owner.ID, ShareEvent{}); !errors.Is(err, ErrNotShared) {
		t.Fatalf("rotating a private track: got %v, want ErrNotShared", err)
	}
	old, _ := ShareTrack(db, trackID, owner.ID, ShareEvent{RightsBasis: "original"})
	fresh, err := RotateShareToken(db, trackID, owner.ID, ShareEvent{})
	if err != nil {
		t.Fatal(err)
	}
	if fresh == old {
		t.Fatal("rotate returned the same token")
	}
	if _, _, err := GetSharedTrack(db, old); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("old token still resolves: %v", err)
	}
	if _, _, err := GetSharedTrack(db, fresh); err != nil {
		t.Fatalf("new token does not resolve: %v", err)
	}

	if err := UnshareTrack(db, trackID, owner.ID, ShareEvent{}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := GetSharedTrack(db, fresh); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unshared track still resolves: %v", err)
	}
	tr, _ := GetTrack(db, trackID, owner.ID)
	if tr.Visibility != models.VisibilityPrivate || tr.ShareToken != nil {
		t.Fatalf("track not private after unshare: %+v", tr)
	}
}

func TestCopyTrack(t *testing.T) {
	db, owner, other, trackID := sharingFixture(t)
	if _, err := db.Exec(`UPDATE tracks SET practice_bpm = 90 WHERE id = ?`, trackID); err != nil {
		t.Fatal(err)
	}

	copyID, err := CopyTrack(db, trackID, other.ID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := CopyTrack(db, trackID, other.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again != copyID {
		t.Fatal("copying twice should return the existing copy")
	}

	c, err := GetTrack(db, copyID, other.ID)
	if err != nil {
		t.Fatal(err)
	}
	if c.SourceTrackID == nil || *c.SourceTrackID != trackID {
		t.Fatalf("source_track_id = %v, want %d", c.SourceTrackID, trackID)
	}
	if c.FilePath != "" {
		t.Fatalf("copy must not reference the original file, got %q", c.FilePath)
	}
	if c.PracticeBPM != nil {
		t.Fatal("practice BPM should not be copied")
	}
	if c.Visibility != models.VisibilityPrivate || c.Name != "groove" {
		t.Fatalf("unexpected copy: %+v", c)
	}

	notes, _ := GetTrackNotes(db, copyID)
	if len(notes) != 2 {
		t.Fatalf("copy has %d notes, want 2", len(notes))
	}
	mapping, _ := GetTrackMapping(db, copyID)
	if mapping[38] != "tom_high" {
		t.Fatalf("mapping not copied: %v", mapping)
	}
	pilots, _ := ListPilotFilePaths(db, copyID)
	if len(pilots) != 0 {
		t.Fatal("pilot tracks must not be copied")
	}

	// A copy cannot be re-shared.
	if _, err := ShareTrack(db, copyID, other.ID, ShareEvent{RightsBasis: "original"}); !errors.Is(err, ErrCopyNotShareable) {
		t.Fatalf("sharing a copy: got %v, want ErrCopyNotShareable", err)
	}

	// Deleting the original leaves the copy intact.
	if err := DeleteTrack(db, trackID, owner.ID); err != nil {
		t.Fatal(err)
	}
	c, err = GetTrack(db, copyID, other.ID)
	if err != nil {
		t.Fatalf("copy gone after original was deleted: %v", err)
	}
	if c.SourceTrackID != nil {
		t.Fatal("source_track_id should be cleared when the original is deleted")
	}
}
