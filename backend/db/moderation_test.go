package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

func reportFor(t *testing.T, db *DB, trackID, ownerID int64) *models.Report {
	t.Helper()
	r := &models.Report{TrackID: trackID, TrackName: "groove", OwnerUserID: &ownerID, ShareToken: "tok",
		ReporterName: "Rights Holder", ReporterEmail: "rights@example.com", Kind: models.ReportCopyright,
		WorkIdentified: "Song X", Details: "This is our song.", GoodFaith: true}
	if err := CreateReport(db, r, "hash"); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFingerprintIgnoresStartOffset(t *testing.T) {
	db, owner, _, trackID := sharingFixture(t)
	shifted := &models.Track{Name: "again", Filename: "a.mid", FilePath: "a.mid", FileSize: 1, DurationMs: 2000, NoteCount: 2, PPQ: 480}
	if err := InsertTrack(db, shifted, []models.NoteEvent{
		{NoteNumber: 36, TimeMs: 1000, Velocity: 1, Instrument: "kick"},
		{NoteNumber: 38, TimeMs: 1502, Velocity: 1, Instrument: "snare"},
	}, owner.ID); err != nil {
		t.Fatal(err)
	}
	a, _ := TrackFingerprint(db, trackID)
	b, _ := TrackFingerprint(db, shifted.ID)
	if a == "" || a != b {
		t.Fatalf("fingerprints differ: %q vs %q", a, b)
	}
}

func TestDisablePropagatesAndBlocksResharing(t *testing.T) {
	db, owner, other, trackID := sharingFixture(t)
	if _, err := ShareTrack(db, trackID, owner.ID, ShareEvent{RightsBasis: "original"}); err != nil {
		t.Fatal(err)
	}
	copyID, err := CopyTrack(db, trackID, other.ID)
	if err != nil {
		t.Fatal(err)
	}
	r := reportFor(t, db, trackID, owner.ID)

	res, err := DisableTrackLineage(db, trackID, "Copyright notice from the rights holder", true)
	if err != nil {
		t.Fatal(err)
	}
	if res.TracksAffected != 2 || res.OwnerStrikes != 1 || res.OwnerBanned {
		t.Fatalf("unexpected result: %+v", res)
	}
	for _, c := range []struct{ id, user int64 }{{trackID, owner.ID}, {copyID, other.ID}} {
		tr, _ := GetTrack(db, c.id, c.user)
		if !tr.Disabled || tr.DisabledReason == nil || tr.ShareToken != nil || tr.Visibility != models.VisibilityPrivate {
			t.Fatalf("track %d not disabled: %+v", c.id, tr)
		}
	}
	reports, _ := ListReports(db, models.ReportActioned)
	if len(reports) != 1 || reports[0].ID != r.ID || !reports[0].TrackDisabled || reports[0].Copies != 1 {
		t.Fatalf("report not actioned: %+v", reports)
	}

	// Disabling again doesn't strike twice.
	if res, _ := DisableTrackLineage(db, trackID, "again", true); res.OwnerStrikes != 1 {
		t.Fatalf("second takedown added a strike: %+v", res)
	}

	// Neither the disabled track nor a fresh upload of the same notes can be shared.
	if _, err := ShareTrack(db, trackID, owner.ID, ShareEvent{RightsBasis: "original"}); !errors.Is(err, ErrBlockedContent) {
		t.Fatalf("re-sharing disabled track: got %v", err)
	}
	notes, _ := GetTrackNotes(db, trackID)
	reupload := &models.Track{Name: "renamed", Filename: "r.mid", FilePath: "r.mid", FileSize: 1, DurationMs: 2000, NoteCount: 2, PPQ: 480}
	if err := InsertTrack(db, reupload, notes, owner.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := ShareTrack(db, reupload.ID, owner.ID, ShareEvent{RightsBasis: "original"}); !errors.Is(err, ErrBlockedContent) {
		t.Fatalf("re-upload: got %v, want ErrBlockedContent", err)
	}

	// Restoring makes everything playable and shareable again.
	res, err = RestoreTrackLineage(db, trackID, "Counter-notice accepted", true)
	if err != nil {
		t.Fatal(err)
	}
	if res.TracksAffected != 2 || res.OwnerStrikes != 0 {
		t.Fatalf("unexpected restore result: %+v", res)
	}
	if tr, _ := GetTrack(db, copyID, other.ID); tr.Disabled {
		t.Fatal("copy still disabled after restore")
	}
	if _, err := ShareTrack(db, reupload.ID, owner.ID, ShareEvent{RightsBasis: "original"}); err != nil {
		t.Fatalf("sharing after restore: %v", err)
	}
	if reports, _ := ListReports(db, models.ReportRejected); len(reports) != 1 {
		t.Fatalf("report should be rejected after restore, got %d", len(reports))
	}
	if _, err := RestoreTrackLineage(db, trackID, "", false); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("restoring an active track: got %v", err)
	}
}

func TestDisableReachesCopiesAfterOriginalDeleted(t *testing.T) {
	db, owner, other, trackID := sharingFixture(t)
	ShareTrack(db, trackID, owner.ID, ShareEvent{RightsBasis: "original"})
	copyID, _ := CopyTrack(db, trackID, other.ID)
	reportFor(t, db, trackID, owner.ID)
	if err := DeleteTrack(db, trackID, owner.ID); err != nil {
		t.Fatal(err)
	}

	res, err := DisableTrackLineage(db, trackID, "Copyright notice", true)
	if err != nil {
		t.Fatal(err)
	}
	if res.TracksAffected != 1 || res.OwnerStrikes != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if tr, _ := GetTrack(db, copyID, other.ID); !tr.Disabled {
		t.Fatal("copy of deleted original not disabled")
	}
	// The orphaned copy still can't be shared by its holder.
	if _, err := ShareTrack(db, copyID, other.ID, ShareEvent{RightsBasis: "original"}); err == nil {
		t.Fatal("orphaned copy could be shared")
	}
}

func TestStrikeLimitBansSharing(t *testing.T) {
	db, owner, _, fixtureID := sharingFixture(t)
	// An unrelated shared track, which the ban must unshare.
	if _, err := ShareTrack(db, fixtureID, owner.ID, ShareEvent{RightsBasis: "original"}); err != nil {
		t.Fatal(err)
	}
	var last int64
	for i := range models.StrikeLimit {
		tr := &models.Track{Name: "t", Filename: "t.mid", FilePath: "t.mid", FileSize: 1, DurationMs: 1000, NoteCount: 1, PPQ: 480}
		if err := InsertTrack(db, tr, []models.NoteEvent{{NoteNumber: 36 + i, TimeMs: 0, Velocity: 1, Instrument: "x"}}, owner.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := ShareTrack(db, tr.ID, owner.ID, ShareEvent{RightsBasis: "original"}); err != nil {
			t.Fatal(err)
		}
		last = tr.ID
		if i < models.StrikeLimit-1 {
			if _, err := DisableTrackLineage(db, tr.ID, "notice", true); err != nil {
				t.Fatal(err)
			}
		}
	}
	res, err := DisableTrackLineage(db, last, "notice", true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OwnerBanned || res.OwnerStrikes != models.StrikeLimit {
		t.Fatalf("expected ban: %+v", res)
	}
	u, _ := GetUserByID(db, owner.ID)
	if !u.SharingBanned {
		t.Fatal("user not flagged as banned")
	}
	var shared int
	db.QueryRow(`SELECT COUNT(*) FROM tracks WHERE user_id = ? AND visibility <> 'private'`, owner.ID).Scan(&shared)
	if shared != 0 {
		t.Fatalf("%d tracks still shared after ban", shared)
	}
}

func TestRejectReport(t *testing.T) {
	db, owner, _, trackID := sharingFixture(t)
	r := reportFor(t, db, trackID, owner.ID)
	if err := RejectReport(db, r.ID, "Not a copyrighted work"); err != nil {
		t.Fatal(err)
	}
	if err := RejectReport(db, r.ID, ""); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("rejecting twice: got %v", err)
	}
	open, _ := ListReports(db, models.ReportOpen)
	all, _ := ListReports(db, "")
	if len(open) != 0 || len(all) != 1 || all[0].ResolutionNote == nil || all[0].ResolvedAt == nil {
		t.Fatalf("unexpected reports: open=%d all=%+v", len(open), all)
	}
}
