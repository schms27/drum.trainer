# Implementation Log

Decisions and findings recorded during Phase 1 + Phase 2 implementation.

---

## Go Backend

### Go toolchain upgrade required by gomidi
`gitlab.com/gomidi/midi/v2 v2.3.22` requires Go ≥ 1.24.2. `go mod tidy` automatically downloaded the newer toolchain. The `go.mod` was updated to `go 1.24.2` by the tool. This is transparent to the developer but worth noting for CI environments that pin a specific Go version.

### `smf.TimeAt()` is the correct tick→ms converter
The `smf.SMF.TimeAt(absTicks int64) int64` method returns microseconds from the piece start and correctly handles all prior tempo changes in the MIDI file. A naive single-BPM formula (`tick / ppq * (60000 / bpm)`) would break for any piece with a tempo change. `TimeAt` is the right approach.

### Channel 10 is index 9 (0-based)
MIDI channel 10 is transmitted with the 0-based channel nibble `9` (i.e. `status & 0x0F == 9`). The gomidi library exposes this as a `uint8` channel value of `9`. Only NoteOn events on channel 9 with velocity > 0 are drum hits.

### NoteOn with velocity 0 = NoteOff
Some MIDI files use NoteOn+vel0 instead of NoteOff for note endings. These are explicitly discarded in `midi/parser.go` to avoid counting them as hits.

### `modernc.org/sqlite` — pure Go, no CGO
Chose `modernc.org/sqlite` over `mattn/go-sqlite3` to avoid requiring a C compiler (`gcc`) on the developer machine and in CI. Works out of the box on Windows, Linux, and macOS with `go build`.

### Single writer connection for SQLite
`db.SetMaxOpenConns(1)` is set in `db/db.go`. SQLite only supports one writer at a time; allowing multiple connections would cause `database is locked` errors under concurrent writes. WAL mode (`PRAGMA journal_mode = WAL`) is applied in the migration to enable concurrent readers while a write is in progress.

### Score JSON blob + normalized rows
Session scores are stored two ways:
1. `score_json TEXT` blob in `session_scores` — fast single-row fetch for the results page.
2. Normalized `session_instrument_scores` rows — for future analytics queries without JSON parsing.

This dual storage is intentional redundancy for read performance.

### Scoring algorithm: instrument groups first, then time proximity
The greedy nearest-neighbour algorithm groups notes by instrument before matching. This prevents a bass drum hit from matching a snare target even if they're close in time. Within each group, a two-pointer style scan advances a `lowerBound` index to skip hits that are definitely too early (delta < −200ms), making the algorithm O(n log n) for sorting + O(n) for scanning.

### `storage.UniqueFilename` uses nanosecond timestamps
Uploaded MIDI files are stored as `<UnixNano>.mid`. This avoids filename collisions for concurrent uploads without requiring a UUID library dependency. The original filename is preserved in the database `filename` column.

### File cleanup on upload failure
If the database insert fails after a successful file save, `h.store.Delete(filePath)` is called as best-effort cleanup in `api/tracks.go`. This prevents orphaned files on disk.

---

## SvelteKit Frontend

### Svelte 5 `$props()` destructuring warning
Destructuring `data` from `$props()` at the top of the script block (e.g. `const { track } = data`) generates a Svelte 5 linter warning: "This reference only captures the initial value." This is a known Svelte 5 reactive-reference gotcha. For pages loaded by `+page.ts` the data doesn't change after initial load, so this is acceptable. The warning does not affect runtime behavior and the build succeeds.

### `{@const}` placement restriction in Svelte 5
`{@const}` blocks must be direct children of `{#if}`, `{#each}`, `{:else}`, etc. — not at the top level of a template. The note-count computation in `session/[trackId]/+page.svelte` was initially written as a top-level `{@const}`, which caused a build error. Fixed by moving it to a `$derived` in the `<script>` block.

### Web MIDI API timing: `event.timeStamp` over `performance.now()`
The recorder uses `event.timeStamp - sessionStartTime` where `sessionStartTime = performance.now()` at recording start. `MIDIMessageEvent.timeStamp` is a `DOMHighResTimeStamp` from the same origin clock as `performance.now()`, so the subtraction is valid and gives sub-millisecond precision relative to when the user clicked "Start".

### Live feed polling vs. event-driven updates
The recording controls component polls `recorder.hits` every 80ms rather than pushing events reactively. This avoids needing a Svelte store or callback for every individual MIDI hit, which at high-speed passages (e.g. hi-hat 16th notes at 160 BPM ≈ 6.25 hits/second) could cause excessive reactive overhead. 80ms polling is imperceptible to the user and keeps the live feed readable.

### XHR for upload progress
The MIDI file upload uses `XMLHttpRequest` (not `fetch`) because the Fetch API does not expose upload progress events. `xhr.upload.addEventListener('progress', ...)` gives the percentage for the progress bar in `MidiUpload.svelte`.

### Vite proxy eliminates CORS in development
`vite.config.ts` proxies `/api/*` to `http://localhost:8080`. This means the browser sees all requests as same-origin during development and no CORS preflight is needed. The Go CORS middleware (`go-chi/cors`) is still configured for production deployments where origins may differ.

### D3 histogram uses `d3.bin()` directly on offset values
`TimingHistogram.svelte` passes the raw `offset_ms` array to `d3.bin().domain([-200, 200]).thresholds(20)`. D3 handles binning, and the SVG is rendered as a `viewBox`-based responsive SVG (scales to any width). Bars are color-coded by proximity to zero: green (±20ms), yellow (±50ms), red (beyond ±50ms).

### Timeline view uses open circles for targets, filled circles for hits
`TimelineView.svelte` renders target notes as indigo outlined circles and matched hits as filled circles color-coded by offset. SVG `<title>` elements provide hover tooltips with exact offset values. The view uses a linear time-to-pixel scale derived from `durationMs / trackWidth`.

### `adapter-node` chosen for production
`@sveltejs/adapter-node` outputs a Node.js server bundle to `frontend/build/`. In production, the Go binary could serve the static SvelteKit output directly (mounting the `build/` directory as a static file handler) for a single-binary deployment. The `adapter-static` option was considered but rejected because SvelteKit's dynamic routes (`[trackId]`, `[sessionId]`) require a server-side handler.

*Superseded:* the app later set `ssr = false` globally, which makes it a pure SPA. Dynamic routes then work with `adapter-static` and `fallback: 'index.html'`: the client-side router resolves `[trackId]` / `[sessionId]` in the browser. The frontend now ships as an nginx image serving static files, with no Node runtime and no `ORIGIN` setting.

---

## Open Items / Future Work

- **Go static file serving**: Mount `frontend/build/` in `api/router.go` so a single `./drum-trainer` binary serves both API and frontend in production.
- **Session history page**: Add a route `/tracks/[id]/history` showing past session summaries with trend charts.
- **Metronome / playback**: Add a Web Audio API metronome or play back the target MIDI in the browser while recording.
- **MIDI channel override**: Some e-drum kits transmit on channel 1 instead of 10. Add a UI toggle to accept hits from any channel.
- **Tempo map export**: Expose the parsed tempo map in the `/tracks/:id` response so the frontend can show a BPM timeline.
