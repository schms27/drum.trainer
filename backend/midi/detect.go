package midi

import (
	"bytes"
	"errors"
	"fmt"

	"gitlab.com/gomidi/midi/v2/smf"
)

// GMDrumChannel is MIDI channel 10 (0-indexed), where General MIDI puts drums.
const GMDrumChannel uint8 = 9

// ErrNoDrumChannel is returned when no channel looks like a drum part.
var ErrNoDrumChannel = errors.New("could not find a drum part in this file — choose the MIDI channel manually")

const (
	// minDrumNotes is how many notes a channel needs to be considered.
	minDrumNotes = 8
	// minDrumShare is the fraction of a channel's notes that must be standard
	// drum notes (see DrumMap) for it to be picked as the drum part.
	minDrumShare = 0.6
)

// DetectDrumChannel picks the channel (0-indexed) holding the drum part:
// channel 10 when it has any notes, otherwise the channel with the most
// notes among those that consist mostly of standard drum notes.
func DetectDrumChannel(data []byte) (uint8, error) {
	s, err := smf.ReadFrom(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("parse SMF: %w", err)
	}

	var total, drumLike [16]int
	for _, track := range s.Tracks {
		for _, ev := range track {
			var ch, key, vel uint8
			if !ev.Message.GetNoteOn(&ch, &key, &vel) || vel == 0 {
				continue
			}
			total[ch]++
			if _, ok := DrumMap[int(key)]; ok {
				drumLike[ch]++
			}
		}
	}

	if total[GMDrumChannel] > 0 {
		return GMDrumChannel, nil
	}

	best, bestCount := -1, 0
	for ch := 0; ch < 16; ch++ {
		if total[ch] < minDrumNotes {
			continue
		}
		if float64(drumLike[ch])/float64(total[ch]) < minDrumShare {
			continue
		}
		if total[ch] > bestCount {
			best, bestCount = ch, total[ch]
		}
	}
	if best < 0 {
		return 0, ErrNoDrumChannel
	}
	return uint8(best), nil
}
