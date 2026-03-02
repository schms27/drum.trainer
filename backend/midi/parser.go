package midi

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/simu/drum-trainer/backend/models"
	"gitlab.com/gomidi/midi/v2/smf"
)

// ParseResult holds the parsed output of an SMF MIDI file.
type ParseResult struct {
	Notes      []models.NoteEvent
	DurationMs float64
	PPQ        int
	BPM        float64
	Name       string // from TrackName meta event or filename fallback
}

// ParseMIDIFile parses raw SMF MIDI bytes and returns all NoteOn events on the
// given drumChannel (0-indexed, so pass 9 for MIDI channel 10) converted to
// millisecond timestamps. Multi-tempo maps are handled correctly via
// smf.SMF.TimeAt() which accumulates tempo changes.
func ParseMIDIFile(data []byte, filename string, drumChannel uint8) (*ParseResult, error) {
	s, err := smf.ReadFrom(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse SMF: %w", err)
	}

	// Extract PPQ (ticks per quarter note).
	mt, ok := s.TimeFormat.(smf.MetricTicks)
	if !ok {
		return nil, fmt.Errorf("unsupported MIDI time format (only MetricTicks supported)")
	}
	ppq := int(mt)

	var (
		notes       []models.NoteEvent
		trackName   string
		maxTickSeen int64
		bpm         = 120.0 // MIDI default when no tempo event is present
	)

	// Iterate all tracks and accumulate absolute tick positions.
	for _, track := range s.Tracks {
		var absTick int64
		for _, ev := range track {
			absTick += int64(ev.Delta)
			if absTick > maxTickSeen {
				maxTickSeen = absTick
			}

			msg := ev.Message

			// Capture TrackName meta event for a human-readable name.
			if trackName == "" {
				var name string
				if msg.GetMetaTrackName(&name) {
					trackName = name
				}
			}

			// Capture the first tempo event (microseconds per beat → BPM).
			if bpm == 120.0 {
				var t float64
				if msg.GetMetaTempo(&t) && t > 0 {
					bpm = t
				}
			}

			// Filter to the requested drum channel (0-indexed).
			var channel, key, vel uint8
			if !msg.GetNoteOn(&channel, &key, &vel) {
				continue
			}
			if channel != drumChannel {
				continue
			}
			// NoteOn with velocity 0 is a NoteOff — skip it.
			if vel == 0 {
				continue
			}

			// Convert absolute ticks → microseconds → milliseconds.
			// s.TimeAt handles all prior tempo changes correctly.
			timeUs := s.TimeAt(absTick)
			timeMs := float64(timeUs) / 1000.0

			notes = append(notes, models.NoteEvent{
				NoteNumber: int(key),
				TimeMs:     timeMs,
				Velocity:   int(vel),
				Instrument: InstrumentGroup(int(key)),
			})
		}
	}

	if len(notes) == 0 {
		return nil, fmt.Errorf("no notes found on MIDI channel %d in file", int(drumChannel)+1)
	}

	// Sort by time (tracks may interleave).
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].TimeMs < notes[j].TimeMs
	})

	durationMs := float64(s.TimeAt(maxTickSeen)) / 1000.0
	if len(notes) > 0 && notes[len(notes)-1].TimeMs > durationMs {
		durationMs = notes[len(notes)-1].TimeMs
	}

	// Fall back to filename (without extension) if no TrackName found.
	if trackName == "" {
		base := filepath.Base(filename)
		trackName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return &ParseResult{
		Notes:      notes,
		DurationMs: durationMs,
		PPQ:        ppq,
		BPM:        bpm,
		Name:       trackName,
	}, nil
}
