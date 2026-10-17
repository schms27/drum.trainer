package midi_test

import (
	"testing"

	"github.com/simu/drum-trainer/backend/midi"
)

func TestInstrumentGroup_KnownNotes(t *testing.T) {
	cases := []struct {
		note  int
		group string
	}{
		// Bass Drum
		{35, "Bass Drum"},
		{36, "Bass Drum"},
		// Snare
		{37, "Snare"},
		{38, "Snare"},
		{40, "Snare"},
		// Hi-Hat
		{42, "Hi-Hat"},
		{44, "Hi-Hat"},
		{46, "Hi-Hat"},
		// Crash
		{49, "Crash"},
		{57, "Crash"},
		// Ride
		{51, "Ride"},
		{53, "Ride"},
		{59, "Ride"},
		// Toms
		{50, "Tom (high)"},
		{47, "Tom (mid)"},
		{48, "Tom (mid)"},
		{43, "Tom (low)"},
		{45, "Tom (low)"},
		{41, "Tom (floor)"},
	}

	for _, c := range cases {
		got := midi.InstrumentGroup(c.note)
		if got != c.group {
			t.Errorf("note %d: expected %q, got %q", c.note, c.group, got)
		}
	}
}

func TestInstrumentGroup_UnknownNote(t *testing.T) {
	unknowns := []int{0, 34, 39, 99, 127}
	for _, n := range unknowns {
		got := midi.InstrumentGroup(n)
		if got != "Other" {
			t.Errorf("note %d: expected \"Other\", got %q", n, got)
		}
	}
}

func TestDrumMapGroupsAreValidInstruments(t *testing.T) {
	for note, group := range midi.DrumMap {
		if !midi.IsValidInstrument(group) {
			t.Errorf("note %d maps to %q, which is not in Instruments", note, group)
		}
	}
}

func TestIsValidInstrument(t *testing.T) {
	for _, name := range []string{"Snare", "Other", "Tom"} {
		if !midi.IsValidInstrument(name) {
			t.Errorf("%q should be valid", name)
		}
	}
	for _, name := range []string{"", "snare", "Cowbell"} {
		if midi.IsValidInstrument(name) {
			t.Errorf("%q should be invalid", name)
		}
	}
}
