package db

import (
	"errors"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

// addTrack inserts a one-note track (notes differ per call so fingerprints don't collide).
func addTrack(t *testing.T, db *DB, userID int64, name string, note int) int64 {
	t.Helper()
	tr := &models.Track{Name: name, Filename: name + ".mid", FilePath: name + ".mid", FileSize: 1, DurationMs: 1000, NoteCount: 1, PPQ: 480, BPM: 100, TimeSigNumerator: 4}
	if err := InsertTrack(db, tr, []models.NoteEvent{{NoteNumber: note, TimeMs: 0, Velocity: 1, Instrument: "x"}}, userID); err != nil {
		t.Fatal(err)
	}
	return tr.ID
}

func TestPublicRequiresListableRights(t *testing.T) {
	db, owner, _, trackID := sharingFixture(t)
	_, err := ShareTrack(db, trackID, owner.ID, ShareEvent{Visibility: models.VisibilityPublic, RightsBasis: "transcription"})
	if !errors.Is(err, ErrNotListable) {
		t.Fatalf("public transcription: got %v, want ErrNotListable", err)
	}
	if _, err := ShareTrack(db, trackID, owner.ID, ShareEvent{Visibility: models.VisibilityUnlisted, RightsBasis: "transcription"}); err != nil {
		t.Fatalf("unlisted transcription should be allowed: %v", err)
	}
}

func TestListPublicTracks(t *testing.T) {
	db, owner, other, _ := sharingFixture(t)
	share := func(id int64, vis string) {
		if _, err := ShareTrack(db, id, owner.ID, ShareEvent{Visibility: vis, RightsBasis: "original"}); err != nil {
			t.Fatal(err)
		}
	}
	funk := addTrack(t, db, owner.ID, "Funk Groove", 40)
	rock := addTrack(t, db, owner.ID, "Rock Beat", 41)
	odd := addTrack(t, db, owner.ID, "100%_odd", 42)
	unlisted := addTrack(t, db, owner.ID, "Secret Groove", 43)
	share(funk, models.VisibilityPublic)
	share(rock, models.VisibilityPublic)
	share(odd, models.VisibilityPublic)
	share(unlisted, models.VisibilityUnlisted)
	if _, err := CopyTrack(db, rock, other.ID); err != nil {
		t.Fatal(err)
	}

	all, more, err := ListPublicTracks(db, "", SortNewest, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 || more {
		t.Fatalf("got %d public tracks (more=%v), want 3", len(all), more)
	}

	popular, _, _ := ListPublicTracks(db, "", SortPopular, 10, 0)
	if popular[0].Name != "Rock Beat" || popular[0].Copies != 1 {
		t.Fatalf("popular first = %+v", popular[0])
	}

	groove, _, _ := ListPublicTracks(db, "GROOVE", SortNewest, 10, 0)
	if len(groove) != 1 || groove[0].Name != "Funk Groove" {
		t.Fatalf("search: %+v", groove)
	}
	// LIKE wildcards in the query are literal.
	if pct, _, _ := ListPublicTracks(db, "%", SortNewest, 10, 0); len(pct) != 1 || pct[0].Name != "100%_odd" {
		t.Fatalf("search for %%: %+v", pct)
	}

	page1, more, _ := ListPublicTracks(db, "", SortNewest, 2, 0)
	page2, more2, _ := ListPublicTracks(db, "", SortNewest, 2, 2)
	if len(page1) != 2 || !more || len(page2) != 1 || more2 {
		t.Fatalf("paging: %d (more=%v), %d (more=%v)", len(page1), more, len(page2), more2)
	}

	// Going back to unlisted, and moderation, both remove a track from Explore.
	share(funk, models.VisibilityUnlisted)
	if _, err := DisableTrackLineage(db, rock, "notice", false); err != nil {
		t.Fatal(err)
	}
	rest, _, _ := ListPublicTracks(db, "", SortNewest, 10, 0)
	if len(rest) != 1 || rest[0].Name != "100%_odd" {
		t.Fatalf("after unlisting and takedown: %+v", rest)
	}
	var listedAt *string
	db.QueryRow(`SELECT listed_at FROM tracks WHERE id = ?`, funk).Scan(&listedAt)
	if listedAt != nil {
		t.Fatal("listed_at should be cleared when a track stops being public")
	}
}
