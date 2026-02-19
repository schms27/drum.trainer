# GM Drum Map — Channel 10 Reference

In General MIDI, channel 10 (MIDI channel index 9, 0-based) is reserved for drums. Each MIDI note number maps to a specific drum sound. This app groups those note numbers into logical instrument groups for scoring purposes.

## Instrument Groups Used in Scoring

| Group | Note Numbers | Sound Names |
|-------|-------------|-------------|
| **Bass Drum** | 35, 36 | Acoustic Bass Drum, Bass Drum 1 |
| **Snare** | 37, 38, 40 | Side Stick, Acoustic Snare, Electric Snare |
| **Hi-Hat** | 42, 44, 46 | Closed Hi-Hat, Pedal Hi-Hat, Open Hi-Hat |
| **Crash** | 49, 57 | Crash Cymbal 1, Crash Cymbal 2 |
| **Ride** | 51, 53, 59 | Ride Cymbal 1, Ride Bell, Ride Cymbal 2 |
| **Tom** | 41, 43, 45, 47, 48, 50 | Low Floor Tom, High Floor Tom, Low Tom, Low-Mid Tom, Hi-Mid Tom, High Tom |
| **Other** | Any other ch-10 note | Cowbell (56), Tambourine (54), etc. |

> **Scoring note:** Matching is performed *within* instrument groups only. A bass drum hit recorded by the player will never be matched against a snare target, even if it occurs at the same time.

---

## Full GM Drum Map Reference

| Note # | Sound |
|--------|-------|
| 27 | High Q |
| 28 | Slap |
| 29 | Scratch Push |
| 30 | Scratch Pull |
| 31 | Sticks |
| 32 | Square Click |
| 33 | Metronome Click |
| 34 | Metronome Bell |
| **35** | **Acoustic Bass Drum** |
| **36** | **Bass Drum 1** |
| **37** | **Side Stick** |
| **38** | **Acoustic Snare** |
| 39 | Hand Clap |
| **40** | **Electric Snare** |
| **41** | **Low Floor Tom** |
| **42** | **Closed Hi-Hat** |
| **43** | **High Floor Tom** |
| **44** | **Pedal Hi-Hat** |
| **45** | **Low Tom** |
| **46** | **Open Hi-Hat** |
| **47** | **Low-Mid Tom** |
| **48** | **Hi-Mid Tom** |
| **49** | **Crash Cymbal 1** |
| **50** | **High Tom** |
| **51** | **Ride Cymbal 1** |
| 52 | Chinese Cymbal |
| **53** | **Ride Bell** |
| 54 | Tambourine |
| 55 | Splash Cymbal |
| 56 | Cowbell |
| **57** | **Crash Cymbal 2** |
| 58 | Vibraslap |
| **59** | **Ride Cymbal 2** |
| 60 | High Bongo |
| 61 | Low Bongo |
| 62 | Mute High Conga |
| 63 | Open High Conga |
| 64 | Low Conga |
| 65 | High Timbale |
| 66 | Low Timbale |
| 67 | High Agogo |
| 68 | Low Agogo |
| 69 | Cabasa |
| 70 | Maracas |
| 71 | Short Whistle |
| 72 | Long Whistle |
| 73 | Short Guiro |
| 74 | Long Guiro |
| 75 | Claves |
| 76 | High Wood Block |
| 77 | Low Wood Block |
| 78 | Mute Cuica |
| 79 | Open Cuica |
| 80 | Mute Triangle |
| 81 | Open Triangle |

Bold rows are the ones actively used in drum kit scoring.

---

## Common E-Drum Kit Mappings

Most Roland and Yamaha e-drum kits follow standard GM mappings on channel 10. If a MIDI file uses a non-standard mapping, the "Other" group will catch those notes without breaking the scoring algorithm.

| Kit Component | Typical Note(s) |
|--------------|----------------|
| Bass drum pedal | 36 |
| Snare (center) | 38 |
| Snare (rim) | 40 or 37 |
| Hi-hat (closed) | 42 |
| Hi-hat (pedal) | 44 |
| Hi-hat (open) | 46 |
| Crash left | 49 |
| Crash right | 57 |
| Ride | 51 |
| Ride bell | 53 |
| Tom 1 (high) | 48 or 50 |
| Tom 2 | 47 or 45 |
| Tom 3 (floor) | 43 or 41 |
