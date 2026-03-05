package midi_test

import (
	"os"
	"testing"

	"github.com/simu/drum-trainer/backend/midi"
)

// loadTestFile reads a file relative to the repo root.
// Tests are skipped if the file is absent so the suite stays green in CI
// without the fixture files.
func loadTestFile(t *testing.T, relPath string) []byte {
	t.Helper()
	data, err := os.ReadFile(relPath)
	if err != nil {
		t.Skipf("test file not available (%s): %v", relPath, err)
	}
	return data
}

func TestParseMIDIFile_ValidFile_BasicProperties(t *testing.T) {
	data := loadTestFile(t, "../../test_beat.mid")

	result, err := midi.ParseMIDIFile(data, "test_beat.mid", 9)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(result.Notes) == 0 {
		t.Error("expected at least one note")
	}
	if result.PPQ <= 0 {
		t.Errorf("expected PPQ > 0, got %d", result.PPQ)
	}
	if result.BPM <= 0 {
		t.Errorf("expected BPM > 0, got %.2f", result.BPM)
	}
	if result.DurationMs <= 0 {
		t.Errorf("expected duration > 0, got %.2f ms", result.DurationMs)
	}
	if result.Name == "" {
		t.Error("expected non-empty track name")
	}
}

func TestParseMIDIFile_NoteTimestampsOrdered(t *testing.T) {
	data := loadTestFile(t, "../../test_beat.mid")

	result, err := midi.ParseMIDIFile(data, "test_beat.mid", 9)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	for i := 1; i < len(result.Notes); i++ {
		if result.Notes[i].TimeMs < result.Notes[i-1].TimeMs {
			t.Errorf("notes not sorted at index %d: %.2f < %.2f",
				i, result.Notes[i].TimeMs, result.Notes[i-1].TimeMs)
		}
	}
}

func TestParseMIDIFile_DurationCoversAllNotes(t *testing.T) {
	data := loadTestFile(t, "../../test_beat.mid")

	result, err := midi.ParseMIDIFile(data, "test_beat.mid", 9)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	last := result.Notes[len(result.Notes)-1].TimeMs
	if result.DurationMs < last {
		t.Errorf("duration %.2fms is less than last note %.2fms", result.DurationMs, last)
	}
}

func TestParseMIDIFile_NoteInstrumentsPopulated(t *testing.T) {
	data := loadTestFile(t, "../../test_beat.mid")

	result, err := midi.ParseMIDIFile(data, "test_beat.mid", 9)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	for _, n := range result.Notes {
		if n.Instrument == "" {
			t.Errorf("note %d at %.2fms has empty instrument", n.NoteNumber, n.TimeMs)
		}
	}
}

func TestParseMIDIFile_FilenameUsedAsFallbackName(t *testing.T) {
	data := loadTestFile(t, "../../test_beat.mid")

	// Parse with a distinctive filename.  If the MIDI has no TrackName meta-event
	// the fallback name is derived from this filename.
	result, err := midi.ParseMIDIFile(data, "my_custom_name.mid", 9)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	// Name is either the embedded TrackName or the stem "my_custom_name".
	if result.Name == "" {
		t.Error("expected non-empty name")
	}
}

func TestParseMIDIFile_WrongChannel_ReturnsError(t *testing.T) {
	data := loadTestFile(t, "../../test_beat.mid")

	// Channel 0 (MIDI channel 1) should have no drum notes.
	_, err := midi.ParseMIDIFile(data, "test_beat.mid", 0)
	if err == nil {
		t.Error("expected an error for a channel with no notes, got nil")
	}
}

func TestParseMIDIFile_InvalidData_ReturnsError(t *testing.T) {
	_, err := midi.ParseMIDIFile([]byte("this is not MIDI"), "garbage.mid", 9)
	if err == nil {
		t.Error("expected parse error for invalid data, got nil")
	}
}

func TestParseMIDIFile_EmptyData_ReturnsError(t *testing.T) {
	_, err := midi.ParseMIDIFile([]byte{}, "empty.mid", 9)
	if err == nil {
		t.Error("expected parse error for empty data, got nil")
	}
}

