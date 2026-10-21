package midi

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/simu/drum-trainer/backend/models"
	gomidi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// ExportPPQ is the resolution of exported files.
const ExportPPQ = 480

// exportNoteTicks is how long each exported hit lasts. Drum sounds ignore the
// note length, but a note-off is still expected; a 32nd note is plenty.
const exportNoteTicks = ExportPPQ / 8

// WriteDrumMIDI renders drum notes as a type-0 Standard MIDI File on channel
// 10 with a single tempo and time signature. Note times are milliseconds at
// `bpm`, as stored for a track. Only the given name is written as metadata.
func WriteDrumMIDI(name string, notes []models.NoteEvent, bpm float64, timeSigNumerator int) ([]byte, error) {
	if bpm <= 0 {
		return nil, fmt.Errorf("invalid tempo %v", bpm)
	}
	if timeSigNumerator < 1 || timeSigNumerator > 32 {
		timeSigNumerator = 4
	}
	msPerTick := 60000 / bpm / ExportPPQ

	type event struct {
		tick int64
		off  bool // note-offs first at the same tick, so repeated hits retrigger
		msg  []byte
	}
	events := make([]event, 0, 2*len(notes))
	for _, n := range notes {
		if n.NoteNumber < 0 || n.NoteNumber > 127 {
			continue
		}
		tick := int64(math.Round(math.Max(0, n.TimeMs) / msPerTick))
		vel := n.Velocity
		if vel < 1 || vel > 127 {
			vel = 100
		}
		key := uint8(n.NoteNumber)
		events = append(events,
			event{tick, false, gomidi.NoteOn(GMDrumChannel, key, uint8(vel))},
			event{tick + exportNoteTicks, true, gomidi.NoteOff(GMDrumChannel, key)},
		)
	}
	sort.SliceStable(events, func(a, b int) bool {
		if events[a].tick != events[b].tick {
			return events[a].tick < events[b].tick
		}
		return events[a].off && !events[b].off
	})

	var tr smf.Track
	if name != "" {
		tr.Add(0, smf.MetaTrackSequenceName(name))
	}
	tr.Add(0, smf.MetaMeter(uint8(timeSigNumerator), 4))
	tr.Add(0, smf.MetaTempo(bpm))
	var last int64
	for _, e := range events {
		tr.Add(uint32(e.tick-last), e.msg)
		last = e.tick
	}
	tr.Close(0)

	s := smf.New()
	s.TimeFormat = smf.MetricTicks(ExportPPQ)
	if err := s.Add(tr); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if _, err := s.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
