package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

func TestKitProfiles(t *testing.T) {
	db := openTestDB(t)
	alice, _ := UpsertUser(db, "github", "1", "a@example.com", "A", "")
	bob, _ := UpsertUser(db, "github", "2", "b@example.com", "B", "")

	if _, ok, err := ActiveKitMapping(db, alice.ID); err != nil || ok {
		t.Fatalf("expected no active kit initially, got ok=%v err=%v", ok, err)
	}

	td17, err := CreateKitProfile(db, alice.ID, "TD-17", []models.NoteMapping{
		{NoteNumber: 40, Instrument: "Snare"}, {NoteNumber: 26, Instrument: "Hi-Hat"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !td17.Active || len(td17.Mappings) != 2 {
		t.Fatalf("first kit should be active with 2 mappings, got %+v", td17)
	}
	second, err := CreateKitProfile(db, alice.ID, "Practice pad", nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.Active {
		t.Error("second kit should not replace the active one")
	}

	mapping, ok, err := ActiveKitMapping(db, alice.ID)
	if err != nil || !ok || mapping[26] != "Hi-Hat" {
		t.Fatalf("active mapping = %v (ok=%v, err=%v)", mapping, ok, err)
	}

	// Update replaces name and mappings.
	if err := UpdateKitProfile(db, td17.ID, alice.ID, "TD-17 KVX", []models.NoteMapping{{NoteNumber: 22, Instrument: "Hi-Hat"}}); err != nil {
		t.Fatal(err)
	}
	updated, _ := GetKitProfile(db, td17.ID, alice.ID)
	if updated.Name != "TD-17 KVX" || len(updated.Mappings) != 1 || updated.Mappings[0].NoteNumber != 22 {
		t.Fatalf("update not applied: %+v", updated)
	}

	// Another user cannot touch these kits.
	if err := UpdateKitProfile(db, td17.ID, bob.ID, "x", nil); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("update by other user: %v", err)
	}
	if err := SetActiveKitProfile(db, bob.ID, &td17.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("activate other user's kit: %v", err)
	}
	if err := DeleteKitProfile(db, td17.ID, bob.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("delete by other user: %v", err)
	}

	// Switching and clearing the active kit.
	if err := SetActiveKitProfile(db, alice.ID, &second.ID); err != nil {
		t.Fatal(err)
	}
	kits, _ := ListKitProfiles(db, alice.ID)
	if kits[0].Active || !kits[1].Active {
		t.Errorf("active flags wrong after switch: %+v", kits)
	}
	if err := SetActiveKitProfile(db, alice.ID, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := ActiveKitMapping(db, alice.ID); ok {
		t.Error("expected no active kit after clearing")
	}

	// Deleting the active kit clears it.
	if err := SetActiveKitProfile(db, alice.ID, &td17.ID); err != nil {
		t.Fatal(err)
	}
	if err := DeleteKitProfile(db, td17.ID, alice.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := ActiveKitMapping(db, alice.ID); ok {
		t.Error("deleting the active kit should clear it")
	}
	if kits, _ := ListKitProfiles(db, alice.ID); len(kits) != 1 {
		t.Errorf("expected 1 kit left, got %d", len(kits))
	}
}
