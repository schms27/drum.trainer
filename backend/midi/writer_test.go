package midi

import (
	"math"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

func TestWriteDrumMIDIRoundTrip(t *testing.T) {
	// 96 BPM in 3/4: a beat is 625 ms.
	notes := []models.NoteEvent{
		{NoteNumber: 36, TimeMs: 0, Velocity: 110},
		{NoteNumber: 42, TimeMs: 0, Velocity: 80},
		{NoteNumber: 38, TimeMs: 625, Velocity: 100},
		{NoteNumber: 38, TimeMs: 640, Velocity: 0}, // flam, invalid velocity → default
		{NoteNumber: 49, TimeMs: 1875, Velocity: 127},
	}
	data, err := WriteDrumMIDI("My groove", notes, 96, 3)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseMIDIFile(data, "export.mid", GMDrumChannel)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "My groove" || got.BPM != 96 || got.TimeSigNumerator != 3 || got.PPQ != ExportPPQ {
		t.Fatalf("metadata: name=%q bpm=%v sig=%d ppq=%d", got.Name, got.BPM, got.TimeSigNumerator, got.PPQ)
	}
	if len(got.Notes) != len(notes) {
		t.Fatalf("got %d notes, want %d", len(got.Notes), len(notes))
	}
	for i, n := range got.Notes {
		if n.NoteNumber != notes[i].NoteNumber || math.Abs(n.TimeMs-notes[i].TimeMs) > 1 {
			t.Errorf("note %d: got #%d at %.1f ms, want #%d at %.1f ms", i, n.NoteNumber, n.TimeMs, notes[i].NoteNumber, notes[i].TimeMs)
		}
	}
	if got.Notes[3].Velocity != 100 {
		t.Errorf("invalid velocity should export as 100, got %d", got.Notes[3].Velocity)
	}
}

func TestWriteDrumMIDIRejectsBadTempo(t *testing.T) {
	if _, err := WriteDrumMIDI("x", nil, 0, 4); err == nil {
		t.Fatal("expected an error for tempo 0")
	}
}
