package api

import (
	"net/http"
	"testing"

	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
)

func TestTrackMix(t *testing.T) {
	e := newSharingEnv(t)
	// UpsertUser returns the existing "owner" user created by the env.
	owner, err := dbpkg.UpsertUser(e.db, "github", "owner", "owner@example.com", "owner", "")
	if err != nil {
		t.Fatal(err)
	}
	pilot := &models.PilotTrack{TrackID: e.trackID, Name: "band", Filename: "band.mp3", FilePath: "band.mp3", FileSize: 1, MimeType: "audio/mpeg"}
	if err := dbpkg.InsertPilotTrack(e.db, pilot, owner.ID); err != nil {
		t.Fatal(err)
	}
	mixPath := "/api/v1/tracks/" + itoa(e.trackID) + "/mix"

	body := map[string]any{"metronome_volume": 0.5, "pilots": []any{map[string]any{"id": pilot.ID, "volume": 0.3}}}
	if code := e.do("PUT", mixPath, "owner", body, nil); code != http.StatusNoContent {
		t.Fatalf("save mix: got %d", code)
	}
	var tr models.TrackWithNotes
	e.do("GET", "/api/v1/tracks/"+itoa(e.trackID), "owner", nil, &tr)
	if tr.MetronomeVolume == nil || *tr.MetronomeVolume != 0.5 {
		t.Fatalf("metronome volume: %v", tr.MetronomeVolume)
	}
	var list struct {
		Pilots []models.PilotTrack `json:"pilots"`
	}
	e.do("GET", "/api/v1/tracks/"+itoa(e.trackID)+"/pilots", "owner", nil, &list)
	if pilots := list.Pilots; len(pilots) != 1 || pilots[0].Volume == nil || *pilots[0].Volume != 0.3 {
		t.Fatalf("pilot volume: %+v", list.Pilots)
	}

	// null clears a level again.
	if code := e.do("PUT", mixPath, "owner", map[string]any{"metronome_volume": nil}, nil); code != http.StatusNoContent {
		t.Fatalf("clear mix: got %d", code)
	}
	tr = models.TrackWithNotes{}
	e.do("GET", "/api/v1/tracks/"+itoa(e.trackID), "owner", nil, &tr)
	if tr.MetronomeVolume != nil {
		t.Fatalf("metronome volume not cleared: %v", *tr.MetronomeVolume)
	}

	if code := e.do("PUT", mixPath, "owner", map[string]any{"metronome_volume": 1.5}, nil); code != http.StatusBadRequest {
		t.Fatalf("out of range: got %d, want 400", code)
	}
	if code := e.do("PUT", mixPath, "other", body, nil); code != http.StatusNotFound {
		t.Fatalf("someone else's track: got %d, want 404", code)
	}
}

func TestKitMix(t *testing.T) {
	e := newSharingEnv(t)
	var kit models.KitProfile
	if code := e.do("POST", "/api/v1/kits", "owner", map[string]any{"name": "Roland", "mappings": []any{}}, &kit); code != http.StatusCreated {
		t.Fatalf("create kit: got %d", code)
	}
	var saved models.KitProfile
	if code := e.do("PUT", "/api/v1/kits/"+itoa(kit.ID)+"/mix", "owner", map[string]any{"metronome_volume": 0.4, "pilot_volume": 0.8}, &saved); code != http.StatusOK {
		t.Fatalf("save kit mix: got %d", code)
	}
	if saved.MetronomeVolume == nil || *saved.MetronomeVolume != 0.4 || saved.PilotVolume == nil || *saved.PilotVolume != 0.8 {
		t.Fatalf("kit mix: %+v", saved)
	}
	if code := e.do("PUT", "/api/v1/kits/"+itoa(kit.ID)+"/mix", "other", map[string]any{"metronome_volume": 0.4}, nil); code != http.StatusNotFound {
		t.Fatalf("someone else's kit: got %d, want 404", code)
	}
	if code := e.do("PUT", "/api/v1/kits/"+itoa(kit.ID)+"/mix", "owner", map[string]any{"pilot_volume": -1}, nil); code != http.StatusBadRequest {
		t.Fatalf("out of range: got %d, want 400", code)
	}
}
