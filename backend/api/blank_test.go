package api

import (
	"net/http"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

func TestCreateBlankTrack(t *testing.T) {
	e := newSharingEnv(t)
	// The owner's kit sends 26 for its hi-hat; created tracks take that over.
	var kit models.KitProfile
	e.do("POST", "/api/v1/kits", "owner", map[string]any{"name": "Roland", "mappings": []any{map[string]any{"note_number": 26, "instrument": "Hi-Hat"}}}, &kit)

	body := map[string]any{"name": "  Paradiddles ", "bpm": 90, "time_sig_numerator": 4, "bars": 2, "notes": []any{
		map[string]any{"note_number": 26, "time_ms": 0, "velocity": 90},
		map[string]any{"note_number": 36, "time_ms": 0, "velocity": 100},
		map[string]any{"note_number": 38, "time_ms": 9000, "velocity": 100}, // past the 2 bars (5333 ms): dropped
	}}
	var got models.TrackWithNotes
	if code := e.do("POST", "/api/v1/tracks/blank", "owner", body, &got); code != http.StatusCreated {
		t.Fatalf("create: got %d", code)
	}
	if got.Name != "Paradiddles" || got.Origin != models.TrackOriginEditor || got.NoteCount != 2 || got.BPM != 90 {
		t.Fatalf("created track: %+v", got.Track)
	}
	if want := 2 * 4 * 60000 / 90.0; got.DurationMs != want {
		t.Fatalf("duration = %v, want %v", got.DurationMs, want)
	}
	if got.Notes[0].Instrument != "Hi-Hat" && got.Notes[1].Instrument != "Hi-Hat" {
		t.Fatalf("kit mapping not applied: %+v", got.Notes)
	}
	var mapping struct {
		Mappings []models.NoteMapping `json:"mappings"`
	}
	e.do("GET", "/api/v1/tracks/"+itoa(got.ID)+"/mapping", "owner", nil, &mapping)
	found := false
	for _, m := range mapping.Mappings {
		found = found || (m.NoteNumber == 26 && m.Instrument == "Hi-Hat")
	}
	if !found {
		t.Fatalf("track mapping should include the kit's pad 26: %+v", mapping.Mappings)
	}

	// No uploaded file, but the drum part exports.
	base := "/api/v1/tracks/" + itoa(got.ID)
	if code := e.get(base+"/file", "owner").Code; code != http.StatusNotFound {
		t.Fatalf("original file of a created track: got %d, want 404", code)
	}
	if code := e.get(base+"/export.mid", "owner").Code; code != http.StatusOK {
		t.Fatalf("export of a created track: got %d", code)
	}

	for _, bad := range []map[string]any{
		{"name": "", "bpm": 90, "time_sig_numerator": 4, "bars": 2},
		{"name": "x", "bpm": 10, "time_sig_numerator": 4, "bars": 2},
		{"name": "x", "bpm": 90, "time_sig_numerator": 0, "bars": 2},
		{"name": "x", "bpm": 90, "time_sig_numerator": 4, "bars": 0},
		{"name": "x", "bpm": 20, "time_sig_numerator": 16, "bars": 999},
		{"name": "x", "bpm": 90, "time_sig_numerator": 4, "bars": 2, "notes": []any{map[string]any{"note_number": 300, "time_ms": 0, "velocity": 1}}},
	} {
		if code := e.do("POST", "/api/v1/tracks/blank", "owner", bad, nil); code != http.StatusBadRequest {
			t.Errorf("invalid %v: got %d, want 400", bad, code)
		}
	}
}
