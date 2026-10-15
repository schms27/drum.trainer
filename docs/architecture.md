# Drum Trainer — Architecture & Implementation Plan

## Overview

A browser-based drum training application. Users upload MIDI files as target tracks, play along using a connected e-drum kit captured via the **Web MIDI API**, and receive precision feedback on timing, hit rate, and per-instrument accuracy. Results are persisted server-side in SQLite.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend framework | SvelteKit 2 + Svelte 5 + TypeScript |
| Styling | Tailwind CSS v4 |
| Charts | D3 v7 (SVG histogram + timeline) |
| MIDI in browser | Web MIDI API (`navigator.requestMIDIAccess`) — Chrome/Edge only |
| Backend language | Go 1.22+ |
| HTTP router | `github.com/go-chi/chi/v5` |
| MIDI parsing | `gitlab.com/gomidi/midi/v2` |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO required) |

---

## Project Structure

```
drum.trainer/
├── Makefile                        # make dev runs both frontend + backend
├── docs/
│   ├── architecture.md             # This file
│   ├── api.md                      # Full REST API reference
│   └── gm-drum-map.md              # GM MIDI drum note → instrument group
├── backend/
│   ├── go.mod
│   ├── main.go                     # Entry point, wires server
│   ├── config/
│   │   └── config.go               # Config struct (port, DB path, upload dir)
│   ├── db/
│   │   ├── db.go                   # sql.Open + migrations runner
│   │   ├── migrations/
│   │   │   └── 001_initial.sql     # All CREATE TABLE statements
│   │   ├── tracks.go               # Track CRUD queries
│   │   └── sessions.go             # Session + score CRUD queries
│   ├── api/
│   │   ├── router.go               # chi router, middleware, route mounting
│   │   ├── tracks.go               # POST /tracks, GET /tracks, GET /tracks/:id, DELETE /tracks/:id
│   │   ├── sessions.go             # POST /sessions, GET /sessions, GET /sessions/:id
│   │   └── middleware.go           # CORS, JSON content-type, request logging
│   ├── midi/
│   │   ├── parser.go               # SMF file → []NoteEvent (tick→ms, ch10 only)
│   │   └── drummap.go              # GM drum map: note number → instrument group name
│   ├── scoring/
│   │   └── scorer.go               # Greedy alignment algorithm + metric computation
│   ├── storage/
│   │   └── files.go                # Save/read uploaded .mid files from disk
│   └── models/
│       ├── note.go                 # NoteEvent, RecordedHit structs
│       ├── track.go                # Track, TrackMetadata structs
│       └── session.go              # Session, ScoreResult, InstrumentBreakdown structs
└── frontend/
    ├── package.json
    ├── svelte.config.js
    ├── vite.config.ts              # Vite proxy: /api/* → localhost:8080
    ├── tsconfig.json
    ├── static/
    │   └── favicon.ico
    └── src/
        ├── app.html
        ├── app.css                 # @import "tailwindcss"
        ├── lib/
        │   ├── api/
        │   │   ├── client.ts       # fetch wrapper with base URL + error handling
        │   │   ├── tracks.ts       # uploadTrack(), getTrack(), listTracks()
        │   │   └── sessions.ts     # submitSession(), getSession(), listSessions()
        │   ├── midi/
        │   │   ├── midiAccess.ts   # requestMIDIAccess, enumerate inputs, handle statechange
        │   │   └── recorder.ts     # RecordedHit[], startRecording(), stopRecording()
        │   ├── stores/
        │   │   ├── trackStore.ts   # writable<Track[]>, selectedTrack
        │   │   └── sessionStore.ts # writable<SessionResult | null>
        │   ├── types/
        │   │   ├── track.ts        # Track, NoteEvent TypeScript interfaces
        │   │   ├── session.ts      # RecordedHit, SessionResult, InstrumentBreakdown
        │   │   └── midi.ts         # MIDIInputDevice, MIDI status byte constants
        │   └── components/
        │       ├── layout/
        │       │   ├── Header.svelte
        │       │   └── Nav.svelte
        │       ├── upload/
        │       │   └── MidiUpload.svelte       # Drag-and-drop + file input, upload progress
        │       ├── tracks/
        │       │   ├── TrackList.svelte         # Fetches + renders all tracks
        │       │   └── TrackCard.svelte         # Name, duration, note count, "Practice" button
        │       ├── recorder/
        │       │   ├── MidiDeviceSelector.svelte   # Enumerate MIDI inputs
        │       │   ├── RecordingControls.svelte     # Start / Stop / Reset state machine
        │       │   └── LiveEventFeed.svelte         # Rolling last-20 hit display
        │       └── results/
        │           ├── ScoreSummary.svelte          # Hit rate %, avg offset ms, ghost % cards
        │           ├── InstrumentTable.svelte        # Per-instrument breakdown table
        │           ├── TimingHistogram.svelte        # D3 SVG ±200ms histogram (20 bins)
        │           └── TimelineView.svelte           # SVG: target dots vs actual hit dots
        └── routes/
            ├── +layout.svelte                       # Header + Nav wrapper
            ├── +page.svelte                         # Home: upload + track list
            ├── session/
            │   └── [trackId]/
            │       ├── +page.svelte                 # Recording page
            │       └── +page.ts                     # load(): fetch track data
            └── results/
                └── [sessionId]/
                    ├── +page.svelte                 # Results display page
                    └── +page.ts                     # load(): fetch session + score
```

---

## Data Flow

### 1. Upload MIDI Target

```
Browser                      Go Backend                        Disk + SQLite
  │                               │                                │
  │ POST /api/v1/tracks           │                                │
  │ (multipart: file.mid) ───────>│                                │
  │                               │── smf.ReadFrom(data) ─────────>│
  │                               │── s.TimeAt(tick) for each note │
  │                               │   (handles multi-tempo maps)   │
  │                               │── filter channel 9 (= ch 10)   │
  │                               │── INSERT INTO tracks           │
  │                               │── INSERT INTO track_notes ────>│
  │                               │── save .mid to UPLOAD_DIR ────>│
  │<── 201 { id, notes[], ... } ──│                                │
```

### 2. Recording Session

```
E-drum ──MIDI──> Browser (Chrome/Edge)
                     │
                     │ navigator.requestMIDIAccess()
                     │ input.onmidimessage = handler
                     │
                     │ For each NoteOn (status & 0xF0 == 0x90, velocity > 0):
                     │   timeMs = event.timeStamp - sessionStartTime
                     │   hits.push({ note_number, time_ms, velocity })
                     │
                     │ User clicks "Stop"
                     │
                     │ POST /api/v1/sessions
                     │ { track_id, started_at, ended_at, hits[] }
                     │──────────────────────────────────────────> Go Backend
                                                                      │
                                                                      │ Load target notes from DB
                                                                      │ Run scoring algorithm
                                                                      │ INSERT INTO sessions
                                                                      │ INSERT INTO session_hits
                                                                      │ INSERT INTO session_scores
                                                                      │
                     <── 201 { session, score } ────────────────────────
                     │
                     │ Navigate to /results/[sessionId]
```

### 3. Results Display

```
Browser loads /results/[sessionId]
  │
  │ +page.ts load(): GET /api/v1/sessions/:id
  │<── { session, score: { hit_rate, by_instrument, matched_hits[] } }
  │
  │ Renders:
  │   ScoreSummary    ← hit_rate, ghost_rate, avg_offset_ms
  │   InstrumentTable ← score.by_instrument[]
  │   TimingHistogram ← score.matched_hits[].offset_ms → D3 histogram
  │   TimelineView    ← target notes + actual hits overlaid on time axis
```

---

## Scoring Algorithm

Located in `backend/scoring/scorer.go`.

### Match Window
Each target note gets its own window: half the distance to its nearest neighbouring target in the same instrument group (ignoring notes at the exact same time), clamped to `[25ms, 200ms]`. An isolated note accepts hits up to ±200ms away; in a fast passage the window stops at the midpoint between notes, so a hit is judged against the note it was closest to.

### Algorithm: Closest-First Matching with Instrument Grouping

Matching is done per instrument group (never across groups — a bass drum hit cannot match a snare target). This is critical for accuracy.

```
For each instrument group G:
  targets  = target notes in G, sorted by time_ms
  recorded = recorded hits in G, sorted by time_ms
  usedRecorded = {}

  candidates = every (T, R) with |R.time_ms - T.time_ms| <= window(T)
  For each candidate, closest first:
    If neither T nor R is matched yet:
      pair (T, R), offset_ms = R.time_ms - T.time_ms
      usedRecorded.add(R)

  Any T left unmatched → MISS

  Any R in recorded not in usedRecorded → GHOST

Aggregate metrics (global + per-group):
  hit_rate      = matched_count / target_count
  miss_rate     = missed_count  / target_count
  ghost_rate    = ghost_count   / recorded_count
  avg_offset_ms = mean(offset_ms of all matched pairs)
  std_offset_ms = stddev(offset_ms of all matched pairs)
```

**Offset sign convention:** negative = early (hit before target), positive = late (hit after target).

### Instrument Groups (GM Channel 10)

| Group | Note Numbers |
|-------|-------------|
| Bass Drum | 35, 36 |
| Snare | 37, 38, 40 |
| Hi-Hat | 42, 44, 46 |
| Crash | 49, 57 |
| Ride | 51, 53, 59 |
| Tom | 41, 43, 45, 47, 48, 50 |
| Other | anything else on channel 10 |

---

## MIDI Parsing Details

### Tick-to-Millisecond Conversion

The `smf.ReadFrom()` function from `gitlab.com/gomidi/midi/v2/smf` returns an `smf.SMF` value. Use `s.TimeAt(absTicks int64) int64` which returns microseconds from the start of the piece, **correctly accounting for all tempo changes** (including mid-song tempo maps). Convert: `timeMs = float64(s.TimeAt(absTick)) / 1000.0`.

Do **not** use a naive single-BPM calculation (`tick / ppq * (60000 / bpm)`) — it will be wrong for any piece with tempo changes.

### Channel Filter

MIDI channel 10 is represented as channel index **9** (0-indexed) in the binary stream (`status & 0x0F == 9`). Only process NoteOn events on this channel. NoteOn with velocity 0 is a NoteOff — discard it.

---

## Browser MIDI Recording

```typescript
// frontend/src/lib/midi/recorder.ts

let sessionStartTime: number; // performance.now() at "Start" click

function onMidiMessage(event: MIDIMessageEvent): void {
  const [status, note, velocity] = event.data;
  const isNoteOn = (status & 0xF0) === 0x90 && velocity > 0;

  if (isNoteOn) {
    // event.timeStamp is DOMHighResTimeStamp — same clock as performance.now()
    const timeMs = event.timeStamp - sessionStartTime;
    hits.push({ note_number: note, time_ms: timeMs, velocity });
  }
}
```

`MIDIMessageEvent.timeStamp` uses the same high-resolution clock as `performance.now()`, giving sub-millisecond precision.

### Recording State Machine

```
idle ──(device selected)──> armed ──(Start clicked)──> recording ──(Stop clicked)──> stopped
                                                                  ↘──(Reset)──> idle
```

On transition to `stopped`: call `stopRecording()` → `submitSession()` → navigate to `/results/[sessionId]`.

---

## Development Workflow

### Prerequisites
- Go 1.22+
- Node.js 20+ with `pnpm`
- Chrome or Edge (Web MIDI API)

### Running

```bash
# Terminal 1 — Go backend on http://localhost:8080
cd backend && go run ./main.go

# Terminal 2 — SvelteKit frontend on http://localhost:5173
cd frontend && pnpm dev

# Or use the Makefile:
make dev   # runs both with make -j2
```

Vite proxies `/api/*` → `http://localhost:8080`, eliminating CORS issues in development.

### Environment Variables (backend)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `./data/drum_trainer.db` | SQLite file path |
| `UPLOAD_DIR` | `./data/uploads` | Uploaded .mid file storage |
| `MAX_UPLOAD_SIZE_MB` | `16` | Maximum upload size |

### Production Build

```bash
# Build SvelteKit to static output (SPA, ssr = false)
cd frontend && npm run build   # outputs to frontend/build/

# Build the Go API server
cd backend && go build -o drum-trainer .
```

The frontend uses `@sveltejs/adapter-static` with `fallback: 'index.html'`. In
Docker, the frontend image is plain nginx serving `frontend/build/`
(`frontend/nginx.conf`: SPA fallback, long-lived caching for
`/_app/immutable/`). The outer proxy (repo-root `nginx.conf`) routes `/api/` to
the backend and everything else to the frontend container.

---

## Go Dependencies (`backend/go.mod`)

```
github.com/go-chi/chi/v5       v5.2.x   — HTTP router
github.com/go-chi/cors         v1.2.x   — CORS middleware
gitlab.com/gomidi/midi/v2      v2.3.x   — SMF MIDI file parsing
modernc.org/sqlite             v1.33.x  — Pure-Go SQLite (no CGO)
```

## Frontend Dependencies (`frontend/package.json`)

```
svelte                    ^5.0
@sveltejs/kit             ^2.16
@sveltejs/adapter-static  ^3.0
@sveltejs/vite-plugin-svelte ^5.0
@tailwindcss/vite         ^4.0
tailwindcss               ^4.0
d3                        ^7.9
typescript                ^5.7
vite                      ^6.0
svelte-check              ^4.0
```

---

## Implementation Order

### Phase 1 — Backend Foundation

1. `models/*.go` — All Go structs (NoteEvent, Track, Session, ScoreResult, etc.)
2. `db/migrations/001_initial.sql` + `db/db.go` — Schema + connection setup
3. `midi/drummap.go` — GM drum map constant map
4. `midi/parser.go` — SMF → []NoteEvent *(unit test with a sample .mid file)*
5. `storage/files.go` — Save/read uploaded files
6. `db/tracks.go` + `db/sessions.go` — CRUD operations
7. `scoring/scorer.go` — Alignment algorithm *(unit test in isolation)*
8. `api/router.go` + `api/tracks.go` + `api/sessions.go` — HTTP handlers
9. `main.go` — Wire everything together

### Phase 2 — Frontend

1. `lib/types/*.ts` — TypeScript interfaces matching Go response shapes
2. `lib/api/client.ts` + `tracks.ts` + `sessions.ts` — API call wrappers
3. `+layout.svelte` + `components/layout/*` — App shell
4. `MidiUpload.svelte` + home `+page.svelte` — Upload flow
5. `lib/midi/midiAccess.ts` + `recorder.ts` — Web MIDI integration
6. `session/[trackId]/` page + recorder components
7. `results/[sessionId]/` page + all result components

### Phase 3 — Polish

1. D3 histogram with color zones (green ±20ms, yellow ±50ms, red beyond)
2. SVG timeline overlay with hover tooltips
3. Error states, loading skeletons, empty states
4. Production: Go serves SvelteKit static build
