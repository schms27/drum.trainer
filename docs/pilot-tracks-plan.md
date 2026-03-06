# Pilot Tracks — Implementation Plan

A user can attach one or more audio "pilot tracks" to any of their drum tracks. During
a practice session the pilots play back in sync with the MIDI target, each individually
mutable/unmutable, and automatically time-scaled when the user has set a practice BPM.

---

## The two hard problems

### 1 — Synchronisation

The MIDI recorder captures hits relative to `performance.now()` at session start. The
Web Audio API scheduler works on `AudioContext.currentTime`, a separate high-resolution
clock. They cannot be compared directly, but they *can* be tied together at one shared
moment: the "Start" button click.

**Solution**: In the `click` handler for "Start":

```
sessionStartPerfNow  = performance.now()          // MIDI reference
await audioCtx.resume()                           // un-suspend after user gesture
startAudioContextTime = audioCtx.currentTime      // Web Audio reference
source.start(startAudioContextTime)               // schedule audio at that instant
```

Both clocks start from a single synchronised instant, so audio and MIDI remain
frame-accurate for the entire session. No polling, no drift correction needed.

AudioContext must be **created on page load** (it starts in `suspended` state, which
is fine for decoding) and **resumed on the Start click** (browser policy: audio context
activation requires a user gesture). Buffer decoding with `decodeAudioData` works
while the context is suspended.

### 2 — Tempo scaling

The session page already computes:

```
timeScale = track.bpm / currentBpm        // > 1 when slowed down
```

MIDI hit times are divided by `timeScale` before scoring, meaning the target notes
appear stretched in time by `timeScale`. The audio pilot must stretch by the exact
same factor.

**Chosen approach: `AudioBufferSourceNode.playbackRate`**

```
source.playbackRate.value = currentBpm / track.bpm   // = 1 / timeScale
```

This is the reciprocal of `timeScale` — slowing the MIDI to 75% sets `playbackRate`
to 0.75.

**Trade-off**: `playbackRate < 1` lowers pitch along with tempo (a 75% playback rate
sounds a minor third lower). This is a known Web Audio API limitation.

**Practical impact**:
- Click tracks / metronome pilots: **unaffected** — pitch of a click is irrelevant.
- Full-band recordings: audible pitch shift. Acceptable for learning parts; annoying
  for critical listening.

**Upgrade path** (not in this phase): integrate a client-side phase vocoder (e.g.
[`soundtouchjs`](https://github.com/cutterbl/SoundTouchJS) or
[`rubberband-wasm`](https://github.com/jamesremuscat/rubberband-wasm)) via an
AudioWorklet to achieve pitch-preserving time stretching without a server FFmpeg
dependency. The `PilotPlayer` class (see below) is designed to swap the playback
engine without touching the rest of the code.

---

## Data model

### New DB migration: `009_add_pilot_tracks.sql` (sqlite + postgres variants)

```sql
CREATE TABLE pilot_tracks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    filename    TEXT    NOT NULL,   -- original upload filename (display only)
    file_path   TEXT    NOT NULL,   -- absolute path on disk
    file_size   INTEGER NOT NULL,
    mime_type   TEXT    NOT NULL,   -- "audio/mpeg", "audio/wav", "audio/ogg", etc.
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_pilot_tracks_track ON pilot_tracks(track_id);
```

No `duration_ms` column: duration is not needed server-side. The browser determines
it after `decodeAudioData`.

No `practice_bpm` column: tempo scaling is computed at runtime from the drum track's
`bpm` / `practice_bpm` ratio; no data need be stored.

---

## Backend

### New config env var

| Variable | Default | Purpose |
|---|---|---|
| `MAX_PILOT_SIZE_MB` | `150` | Upload limit for audio files |

### File storage

Pilot audio files are stored in the same `UPLOAD_DIR` as MIDI files (or a sub-
directory `UPLOAD_DIR/pilots/` to keep them separate). `storage.UniqueFilename` is
reused as-is.

### Accepted MIME types

Server validates the `Content-Type` of the upload (and optionally inspects the first
few bytes for magic numbers):

```
audio/mpeg          (.mp3)
audio/wav           (.wav)
audio/ogg           (.ogg)
audio/aac           (.aac / .m4a)
audio/flac          (.flac)
audio/x-m4a         (.m4a)
```

The browser's Web Audio API `decodeAudioData` is the ultimate decoder — server-side
MIME validation is just a sanity check to reject obvious non-audio uploads.

### New DB queries: `db/pilots.go`

```go
InsertPilotTrack(db, pilot *models.PilotTrack) (int64, error)
ListPilotTracks(db, trackID, userID int64) ([]models.PilotTrack, error)
GetPilotTrack(db, id, trackID, userID int64) (*models.PilotTrack, error)
DeletePilotTrack(db, id, trackID, userID int64) error
```

All reads/writes scope by `user_id` to prevent cross-user access.

### New model: `models/pilot.go`

```go
type PilotTrack struct {
    ID        int64     `json:"id"`
    TrackID   int64     `json:"track_id"`
    Name      string    `json:"name"`
    Filename  string    `json:"filename"`
    FileSize  int64     `json:"file_size"`
    MimeType  string    `json:"mime_type"`
    CreatedAt time.Time `json:"created_at"`
    // FilePath intentionally omitted from JSON
}
```

### New API handlers: `api/pilots.go`

| Method | Route | Description |
|---|---|---|
| `POST` | `/api/v1/tracks/{id}/pilots` | Upload a pilot audio file (`multipart/form-data`, fields: `file`, `name`) |
| `GET` | `/api/v1/tracks/{id}/pilots` | List all pilots for a track |
| `DELETE` | `/api/v1/tracks/{id}/pilots/{pilotId}` | Delete a pilot (also removes file from disk) |
| `GET` | `/api/v1/tracks/{id}/pilots/{pilotId}/audio` | Stream the audio file to the browser |

All routes are inside the existing `RequireAuth` group.

**Audio streaming handler** (`GET .../audio`): reads the file from disk and writes it
to the response with:
- `Content-Type` set to the stored `mime_type`
- `Content-Length` set to the file size
- `Accept-Ranges: bytes` to support partial requests (needed for `<audio>` seek and
  some browser buffering strategies)
- Uses `http.ServeContent` / `io.Copy` — no full file load into memory

---

## Frontend

### New API module: `lib/api/pilots.ts`

```ts
type PilotTrack = { id: number; track_id: number; name: string; filename: string;
                    file_size: number; mime_type: string; created_at: string; };

uploadPilot(trackId: number, name: string, file: File,
            onProgress: (pct: number) => void): Promise<PilotTrack>   // uses XHR (like MIDI upload)
listPilots(trackId: number): Promise<PilotTrack[]>
deletePilot(trackId: number, pilotId: number): Promise<void>
pilotAudioUrl(trackId: number, pilotId: number): string               // just builds the URL string
```

### New audio engine: `lib/audio/PilotPlayer.ts`

Central class that owns the `AudioContext` and manages all pilot playback. Kept
completely separate from MIDI logic.

```ts
class PilotPlayer {
    // Created once on page load. Starts suspended (browser policy).
    private ctx: AudioContext

    // Decoded buffers — loaded during page load / before session starts.
    private buffers: Map<number, AudioBuffer>     // pilotId → buffer
    private muted:   Set<number>                  // pilotIds currently muted

    // Live during a running session only.
    private sources:   Map<number, AudioBufferSourceNode>
    private gainNodes: Map<number, GainNode>

    // ── Lifecycle ──────────────────────────────────────────────────────────
    async load(pilotId: number, url: string): Promise<void>
    // Fetches audio, decodes via ctx.decodeAudioData, stores in buffers.
    // Safe to call while ctx is suspended. Shows a loading state externally.

    // ── Session control ────────────────────────────────────────────────────
    async startAll(playbackRate: number): Promise<void>
    // Resumes ctx (browser gesture requirement), then for each loaded buffer:
    //   source = ctx.createBufferSource()
    //   source.playbackRate.value = playbackRate
    //   gain  = ctx.createGain()
    //   gain.gain.value = muted.has(id) ? 0 : 1
    //   source → gain → ctx.destination
    //   source.start(ctx.currentTime)      ← all start at the same instant
    // Returns the AudioContext start time (for potential future sync use).

    stopAll(): void
    // Calls source.stop() on all live sources.

    // ── Per-pilot controls (safe to call during playback) ──────────────────
    setMuted(pilotId: number, muted: boolean): void
    // Flips gain.gain.value between 0 and 1 with a tiny 10ms ramp to avoid clicks:
    //   gain.gain.linearRampToValueAtTime(target, ctx.currentTime + 0.01)

    isLoaded(pilotId: number): boolean
    isMuted(pilotId: number):  boolean

    dispose(): void   // stop all, close AudioContext
}
```

**Why `linearRampToValueAtTime` for mute?** Instantaneous gain change to 0 causes a
click (DC offset discontinuity). A 10 ms ramp is inaudible as a fade but eliminates
the click.

### Updated session page: `session/[trackId]/+page.svelte`

**Load phase** (`+page.ts`): fetch both `track` and `pilots` in parallel.

**Page load** (before user interacts):
1. Create a `PilotPlayer` instance.
2. Call `player.load(pilot.id, pilotAudioUrl(...))` for each pilot concurrently.
3. Track loading state — show a "Loading audio…" indicator; disable "Start" until all
   pilots are decoded.

**BPM change** (`handleBpmChange`): no audio action needed (audio hasn't started).
The `playbackRate` is computed fresh when "Start" is clicked.

**Start**:
```ts
const playbackRate = currentBpm / track.bpm   // 1.0 if no practice BPM
await player.startAll(playbackRate)
// ... existing MIDI recording start ...
```
Both happen in the same synchronous click handler (after the `await` for
`ctx.resume()`, which is near-instantaneous). Sub-millisecond sync is achievable.

**Stop**: `player.stopAll()` alongside existing MIDI stop.

**Reset**: `player.stopAll()`, then re-start is handled by creating new
`AudioBufferSourceNode`s on the next `startAll()` call (a source node can only be
started once — `PilotPlayer` always creates fresh nodes in `startAll`).

**Mute/unmute toggle button** (per pilot, visible during recording):
```ts
player.setMuted(pilotId, !player.isMuted(pilotId))
```

### New component: `lib/components/session/PilotControls.svelte`

Displays the list of pilots with:
- Pilot name
- Mute/unmute button (speaker icon, toggles on click)
- Loading indicator per pilot (spinner while buffer is decoding)

Shown on the session page between the Metronome and Recording Controls sections.

### Track management: pilot upload UI

On the track's home page (`+page.svelte` track card or a new track detail page),
add a "Pilot tracks" section per track with:
- List of existing pilots (name, file size, delete button)
- Upload form: file picker + name input + upload button with progress bar (reuse the
  XHR pattern from `MidiUpload.svelte`)

The upload section lives on the main home page within an expandable panel on each
`TrackCard`, or on a new dedicated `/tracks/:id` detail page. The latter is cleaner
if track management grows further.

---

## Sequence diagram — session with pilots

```
Page load
  +page.ts:  GET /tracks/:id  ──────────────────────────→ track + notes
  +page.ts:  GET /tracks/:id/pilots ────────────────────→ [pilot1, pilot2]
  +page.svelte: PilotPlayer.load(pilot1, "/audio") ─────→ fetch + decode (async)
  +page.svelte: PilotPlayer.load(pilot2, "/audio") ─────→ fetch + decode (async)
  [Start button disabled until all buffers decoded]

User clicks Start
  PilotPlayer.startAll(currentBpm / track.bpm)
    └─ ctx.resume()                         ← un-suspend AudioContext
    └─ source1.start(ctx.currentTime)       ← pilot1 starts
    └─ source2.start(ctx.currentTime)       ← pilot2 starts (same tick)
  sessionStartTime = performance.now()      ← MIDI reference point
  [MIDI recording begins]

User clicks mute on pilot1
  gain1.linearRampToValueAtTime(0, ctx.currentTime + 0.01)

User clicks Stop
  PilotPlayer.stopAll()
  [MIDI recording stops, hits submitted]
```

---

## What is NOT in this plan

- **Server-side time stretching** (FFmpeg): not needed; browser `playbackRate` covers
  the practice tempo range well. Add later if high-quality pitch preservation is
  required.
- **Waveform visualisation**: show a waveform of the pilot during playback. Nice to
  have, non-trivial (OfflineAudioContext rendering). Add in a later phase.
- **Per-pilot volume control**: a gain slider rather than a binary mute. Simple to add
  (just change `gain.gain.value` to a 0–1 float), deferred for now.
- **Looping**: if the pilot is shorter than the drum track, loop it. Use
  `source.loop = true`. Trivial to add, deferred.

---

## Suggested implementation order

1. DB migration (`009_add_pilot_tracks.sql`)
2. `models/pilot.go`, `db/pilots.go`
3. `api/pilots.go` + wire routes in `router.go` + `config.go` (`MAX_PILOT_SIZE_MB`)
4. `lib/api/pilots.ts`
5. `lib/audio/PilotPlayer.ts`
6. `PilotControls.svelte` component (mute UI, loading state)
7. Update `session/[trackId]/+page.ts` to fetch pilots
8. Update `session/[trackId]/+page.svelte` to instantiate `PilotPlayer`, integrate
   start/stop, show `PilotControls`
9. Pilot upload UI (on track management page or home page track card)
