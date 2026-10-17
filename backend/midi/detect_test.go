package midi_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/simu/drum-trainer/backend/midi"
	gm "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// buildSMF writes a one-track file containing the given notes per channel.
func buildSMF(t *testing.T, notes map[uint8][]uint8) []byte {
	t.Helper()
	var tr smf.Track
	for ch, keys := range notes {
		for _, k := range keys {
			tr.Add(120, gm.NoteOn(ch, k, 100))
		}
	}
	tr.Close(0)
	s := smf.New()
	if err := s.Add(tr); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := s.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func repeat(keys []uint8, n int) []uint8 {
	var out []uint8
	for len(out) < n {
		out = append(out, keys...)
	}
	return out[:n]
}

func TestDetectDrumChannel_PrefersChannel10(t *testing.T) {
	data := buildSMF(t, map[uint8][]uint8{
		9: {36, 38},
		1: repeat([]uint8{36, 38, 42}, 50), // drum-like notes on another channel
	})
	ch, err := midi.DetectDrumChannel(data)
	if err != nil || ch != 9 {
		t.Fatalf("got (%d, %v), want channel index 9", ch, err)
	}
}

func TestDetectDrumChannel_FindsDrumsOnOtherChannel(t *testing.T) {
	data := buildSMF(t, map[uint8][]uint8{
		0: repeat([]uint8{60, 62, 64, 65, 67}, 40), // melody
		3: repeat([]uint8{36, 42, 38, 42}, 32),     // drums on channel 4
		5: repeat([]uint8{36, 70, 72, 74, 76}, 20), // mostly non-drum notes
	})
	ch, err := midi.DetectDrumChannel(data)
	if err != nil || ch != 3 {
		t.Fatalf("got (%d, %v), want channel index 3", ch, err)
	}
}

func TestDetectDrumChannel_NoDrums(t *testing.T) {
	data := buildSMF(t, map[uint8][]uint8{0: repeat([]uint8{60, 64, 67}, 30)})
	if _, err := midi.DetectDrumChannel(data); !errors.Is(err, midi.ErrNoDrumChannel) {
		t.Fatalf("got %v, want ErrNoDrumChannel", err)
	}
}
