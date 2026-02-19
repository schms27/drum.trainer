// gen_test_midi generates a simple 4-bar drum loop MIDI file for testing
// the drum trainer app.
//
// Pattern (4/4, 120 BPM, 4 bars):
//   Beat 1 : Kick (36) + Hi-Hat closed (42)
//   Beat 1+: Hi-Hat closed (42)  [8th-note subdivision]
//   Beat 2 : Snare (38) + Hi-Hat closed (42)
//   Beat 2+: Hi-Hat closed (42)
//   Beat 3 : Kick (36) + Hi-Hat closed (42)
//   Beat 3+: Hi-Hat closed (42)
//   Beat 4 : Snare (38) + Hi-Hat closed (42)
//   Beat 4+: Hi-Hat closed (42)
//
// Usage:
//   go run ./tools/gen_test_midi  (from drum.trainer/ root)
//   → writes test_beat.mid to the current directory
package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

// writeVLQ writes v as a MIDI variable-length quantity.
func writeVLQ(b *bytes.Buffer, v uint32) {
	if v == 0 {
		b.WriteByte(0)
		return
	}
	var tmp [4]byte
	n := 0
	for v > 0 {
		tmp[n] = byte(v & 0x7F)
		v >>= 7
		n++
	}
	// Write most-significant groups first, all with continuation bit set except the last.
	for i := n - 1; i >= 0; i-- {
		if i > 0 {
			b.WriteByte(tmp[i] | 0x80)
		} else {
			b.WriteByte(tmp[i])
		}
	}
}

// addEv writes a MIDI event (delta + raw bytes) to b.
func addEv(b *bytes.Buffer, delta uint32, data []byte) {
	writeVLQ(b, delta)
	b.Write(data)
}

// mkChunk builds a MIDI chunk: 4-byte ID + 4-byte big-endian length + data.
func mkChunk(id string, data []byte) []byte {
	var out bytes.Buffer
	out.WriteString(id)
	if err := binary.Write(&out, binary.BigEndian, uint32(len(data))); err != nil {
		panic(err)
	}
	out.Write(data)
	return out.Bytes()
}

func main() {
	const (
		ppq   = uint32(480) // ticks per quarter note
		ch    = byte(9)     // MIDI channel 10 (0-indexed) — GM drums
		kick  = byte(36)    // Bass Drum 1
		snare = byte(38)    // Acoustic Snare
		hh    = byte(42)    // Closed Hi-Hat
		bars  = uint32(4)
	)

	q := ppq          // quarter note  = 480 ticks
	e := ppq / 2      // eighth note   = 240 ticks
	barLen := q * 4   // one 4/4 bar   = 1920 ticks
	noteDur := uint32(10) // percussive note-off after 10 ticks

	on := func(note, vel byte) []byte { return []byte{0x90 | ch, note, vel} }
	off := func(note byte) []byte { return []byte{0x80 | ch, note, 0} }

	// ── Track 0: Tempo & Time Signature ──────────────────────────────────────
	var t0 bytes.Buffer
	// Set Tempo: 500 000 μs/beat = 120 BPM  →  0x07A120
	addEv(&t0, 0, []byte{0xFF, 0x51, 0x03, 0x07, 0xA1, 0x20})
	// Time Signature: 4/4 (numerator=4, denominator=2²=4, 24 MIDI clocks, 8 32nds)
	addEv(&t0, 0, []byte{0xFF, 0x58, 0x04, 0x04, 0x02, 0x18, 0x08})
	// End of Track
	addEv(&t0, 0, []byte{0xFF, 0x2F, 0x00})

	// ── Track 1: Drum events (channel 10) ────────────────────────────────────
	type event struct {
		abs  uint32
		data []byte
	}

	var evs []event

	for bar := uint32(0); bar < bars; bar++ {
		o := bar * barLen

		// Beat 1: Kick + HH (downbeat)
		evs = append(evs,
			event{o + 0, on(kick, 100)},
			event{o + 0, on(hh, 90)},
			event{o + noteDur, off(kick)},
			event{o + noteDur, off(hh)},
		)
		// Beat 1+: HH (8th)
		evs = append(evs,
			event{o + e, on(hh, 65)},
			event{o + e + noteDur, off(hh)},
		)
		// Beat 2: Snare + HH
		evs = append(evs,
			event{o + q, on(snare, 100)},
			event{o + q, on(hh, 90)},
			event{o + q + noteDur, off(snare)},
			event{o + q + noteDur, off(hh)},
		)
		// Beat 2+: HH (8th)
		evs = append(evs,
			event{o + q + e, on(hh, 65)},
			event{o + q + e + noteDur, off(hh)},
		)
		// Beat 3: Kick + HH
		evs = append(evs,
			event{o + 2*q, on(kick, 95)},
			event{o + 2*q, on(hh, 90)},
			event{o + 2*q + noteDur, off(kick)},
			event{o + 2*q + noteDur, off(hh)},
		)
		// Beat 3+: HH (8th)
		evs = append(evs,
			event{o + 2*q + e, on(hh, 65)},
			event{o + 2*q + e + noteDur, off(hh)},
		)
		// Beat 4: Snare + HH
		evs = append(evs,
			event{o + 3*q, on(snare, 100)},
			event{o + 3*q, on(hh, 90)},
			event{o + 3*q + noteDur, off(snare)},
			event{o + 3*q + noteDur, off(hh)},
		)
		// Beat 4+: HH (8th)
		evs = append(evs,
			event{o + 3*q + e, on(hh, 65)},
			event{o + 3*q + e + noteDur, off(hh)},
		)
	}

	// Convert absolute ticks to delta times and write to buffer.
	var t1 bytes.Buffer
	prev := uint32(0)
	for _, ev := range evs {
		addEv(&t1, ev.abs-prev, ev.data)
		prev = ev.abs
	}
	// End of Track at end of last bar
	addEv(&t1, bars*barLen-prev, []byte{0xFF, 0x2F, 0x00})

	// ── MIDI File Header ─────────────────────────────────────────────────────
	var hdr [6]byte
	binary.BigEndian.PutUint16(hdr[0:], 1)           // format 1 (multi-track)
	binary.BigEndian.PutUint16(hdr[2:], 2)           // 2 tracks
	binary.BigEndian.PutUint16(hdr[4:], uint16(ppq)) // time division

	var out bytes.Buffer
	out.Write(mkChunk("MThd", hdr[:]))
	out.Write(mkChunk("MTrk", t0.Bytes()))
	out.Write(mkChunk("MTrk", t1.Bytes()))

	outPath := "test_beat.mid"
	if err := os.WriteFile(outPath, out.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote %s  (%d bytes, %d bars, 120 BPM, PPQ=%d)",
		outPath, out.Len(), bars, ppq)
	log.Printf("Pattern: kick on beats 1+3, snare on beats 2+4, hi-hat on every 8th note")
}
