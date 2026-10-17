package midi

import "slices"

// DrumMap maps GM channel-10 note numbers to instrument group names.
// Groups are intentionally broad so the scorer can match across
// slight kit variations (e.g. note 35 vs 36 both = "Bass Drum").
var DrumMap = map[int]string{
	// Bass Drum
	35: "Bass Drum",
	36: "Bass Drum",
	// Snare
	37: "Snare", // Side Stick
	38: "Snare", // Acoustic Snare
	40: "Snare", // Electric Snare
	// Hi-Hat
	42: "Hi-Hat", // Closed Hi-Hat
	44: "Hi-Hat", // Pedal Hi-Hat
	46: "Hi-Hat", // Open Hi-Hat
	// Crash
	49: "Crash", // Crash Cymbal 1
	57: "Crash", // Crash Cymbal 2
	// Ride
	51: "Ride", // Ride Cymbal 1
	53: "Ride", // Ride Bell
	59: "Ride", // Ride Cymbal 2
	// Toms (four distinct levels by pitch)
	50: "Tom (high)",  // High Tom
	47: "Tom (mid)",   // Low-Mid Tom
	48: "Tom (mid)",   // Hi-Mid Tom
	43: "Tom (low)",   // High Floor Tom
	45: "Tom (low)",   // Low Tom
	41: "Tom (floor)", // Low Floor Tom
}

// Instruments lists every instrument group a note can be mapped to, in
// display order. The frontend mirrors this list in src/lib/drums.ts.
var Instruments = []string{
	"Bass Drum", "Snare", "Hi-Hat", "Crash", "Ride",
	"Tom (high)", "Tom (mid)", "Tom (low)", "Tom (floor)",
	"Other",
}

// legacyInstruments are accepted in custom mappings for backward
// compatibility but no longer offered for new mappings.
var legacyInstruments = []string{"Tom"}

// IsValidInstrument reports whether name is an accepted mapping target.
func IsValidInstrument(name string) bool {
	return slices.Contains(Instruments, name) || slices.Contains(legacyInstruments, name)
}

// InstrumentGroup returns the group name for a given GM drum note number.
// Returns "Other" for notes not in the standard drum kit mapping.
func InstrumentGroup(noteNumber int) string {
	if name, ok := DrumMap[noteNumber]; ok {
		return name
	}
	return "Other"
}
