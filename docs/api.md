# Drum Trainer — REST API Reference

**Base URL (dev):** `http://localhost:8080/api/v1`
**Base URL (prod):** `/api/v1` (served by the same Go binary)

All requests and responses use `application/json` unless noted. File uploads use `multipart/form-data`.

---

## Tracks

### POST /api/v1/tracks

Upload a MIDI file. The backend parses the drum part's note events, converts ticks to milliseconds (multi-tempo-aware), stores the file on disk, and persists track metadata + notes in the database.

**Request:** `multipart/form-data`

| Field | Type | Description |
|-------|------|-------------|
| `file` | File | `.mid` file, max 16 MB |
| `channel` | string, optional | MIDI channel of the drum part: `1`–`16`, or `auto` to detect it (channel 10 if it has notes, otherwise the channel whose notes are mostly General MIDI drum notes). Defaults to `10`. The web app sends `auto`. |

**Response `201 Created`:**
```json
{
  "id": 3,
  "name": "My Drum Track",
  "filename": "my_track.mid",
  "duration_ms": 124350.5,
  "note_count": 842,
  "ppq": 480,
  "notes": [
    { "note_number": 36, "time_ms": 0.0,   "velocity": 100, "instrument": "Bass Drum" },
    { "note_number": 38, "time_ms": 500.0,  "velocity": 90,  "instrument": "Snare" },
    { "note_number": 42, "time_ms": 250.0,  "velocity": 72,  "instrument": "Hi-Hat" }
  ],
  "created_at": "2026-02-19T10:00:00Z"
}
```

**Error responses:**
- `400 Bad Request` — not a valid MIDI file, no notes on the chosen channel, or (`auto`) no drum part found
- `413 Content Too Large` — file exceeds `MAX_UPLOAD_SIZE_MB`

---

### GET /api/v1/tracks

List all uploaded tracks. Notes are not included in the list response.

**Response `200 OK`:**
```json
{
  "tracks": [
    {
      "id": 1,
      "name": "Track A",
      "filename": "track_a.mid",
      "duration_ms": 124350.5,
      "note_count": 842,
      "created_at": "2026-02-19T10:00:00Z",
      "stats": {
        "sessions": 7,
        "best_hit_rate": 0.96,
        "last_hit_rate": 0.92,
        "last_practiced_at": "2026-10-16T20:44:03Z",
        "recent_hit_rates": [0.83, 0.88, 0.88, 0.96, 0.92]
      }
    },
    {
      "id": 2,
      "name": "Track B",
      "filename": "track_b.mid",
      "duration_ms": 93200.0,
      "note_count": 621,
      "created_at": "2026-02-19T11:00:00Z"
    }
  ]
}
```

---

Every track carries `stats` (the user's practice on it):

| Field | Description |
|-------|-------------|
| `sessions` | Number of takes recorded |
| `best_hit_rate`, `last_hit_rate` | Best and most recent hit rate (0–1); `null` before the first take |
| `last_practiced_at` | Time of the most recent take; `null` before the first take |
| `recent_hit_rates` | Hit rates of up to the 10 latest takes, oldest first |

### GET /api/v1/tracks/:id

Fetch a single track with all its parsed notes and its `stats`. Used by the track and recording pages.

**Response `200 OK`:** Same shape as the POST /tracks response, plus `stats`.

**Error responses:**
- `404 Not Found` — track ID does not exist

---

### DELETE /api/v1/tracks/:id

Delete a track, all associated sessions and scores (via SQLite CASCADE), and the `.mid` file from disk.

**Response `204 No Content`**

**Error responses:**
- `404 Not Found` — track ID does not exist

---

## Sharing

Tracks can be shared through an unlisted link. Every track object carries `visibility`
(`"private"` | `"unlisted"`), `share_token` (null while private), `rights_basis` and
`source_track_id` (set on copies made from someone else's link). See
[sharing-plan.md](sharing-plan.md) for the design.

### PUT /api/v1/tracks/:id/sharing

Share a track by link, or update the rights declaration of an already-shared track (the token
is kept). Records an entry in `track_share_events` with the current terms version.

**Request:** `{ "rights_basis": "original" | "licensed" | "public_domain" | "transcription", "accept_terms": true }`

**Response `200 OK`:** `{ "visibility": "unlisted", "share_token": "…", "rights_basis": "original" }`

**Error responses:**
- `400 Bad Request` — invalid `rights_basis`, or `accept_terms` is not `true`
- `404 Not Found` — not the caller's track
- `409 Conflict` — the track is a copy of someone else's shared track

### POST /api/v1/tracks/:id/sharing/rotate

Replace the link token; the old link stops working. **Response `200 OK`:** `{ "visibility": "unlisted", "share_token": "…" }`. `409` if the track isn't shared.

### DELETE /api/v1/tracks/:id/sharing

Make the track private and clear its token. Existing copies are kept. **Response `204 No Content`**

### GET /api/v1/shared/:token

**No session required.** Metadata and the opening four bars of a shared track; the full note
list is never returned here. Sent with `X-Robots-Tag: noindex, nofollow`.

**Response `200 OK`:**
```json
{
  "name": "test_beat", "duration_ms": 8000, "note_count": 48, "bpm": 120, "time_sig_numerator": 4,
  "rights_basis": "original", "shared_at": "2026-10-17T15:05:00Z",
  "instruments": { "Bass Drum": 8, "Snare": 8, "Hi-Hat": 32 },
  "preview_notes": [ { "note_number": 36, "time_ms": 0, "velocity": 100, "instrument": "Bass Drum" } ],
  "preview_ms": 8000,
  "is_owner": false,
  "copied_track_id": null
}
```
`is_owner` and `copied_track_id` are only filled when the request carries a valid session.

**Error responses:** `404 Not Found` — unknown token, reset link, or track no longer shared.

### POST /api/v1/shared/:token/copy

Copy the shared track into the caller's library: notes and note mapping, but not the original
MIDI file, backing tracks, practice tempo or takes. Copying again returns the existing copy.

**Response `201 Created`** (or `200 OK` for the owner's own track): `{ "track_id": 7 }`

---

## Sessions

### POST /api/v1/sessions

Submit a completed recording session. The backend:
1. Loads target notes for the given `track_id`
2. Runs the closest-first alignment algorithm (per-note window of up to ±200ms, per instrument group — see `docs/architecture.md`)
3. Persists the session, raw hits, and computed scores
4. Returns the full result immediately

**Request body:**
```json
{
  "track_id": 3,
  "started_at": "2026-02-19T10:05:00.000Z",
  "ended_at":   "2026-02-19T10:07:04.350Z",
  "hits": [
    { "note_number": 36, "time_ms": 12.5,   "velocity": 110 },
    { "note_number": 38, "time_ms": 501.3,  "velocity": 88 },
    { "note_number": 42, "time_ms": 249.8,  "velocity": 72 }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `track_id` | integer | ID of the target track |
| `started_at` | string (ISO 8601) | UTC timestamp when recording started |
| `ended_at` | string (ISO 8601) | UTC timestamp when recording stopped |
| `hits[].note_number` | integer | MIDI note number (GM drum map) |
| `hits[].time_ms` | float | Milliseconds from the start of the track, in the track's original tempo. May be slightly negative (down to −1000) for a note played just before the first beat after a count-in. |
| `hits[].velocity` | integer | MIDI velocity 0–127 |

**Response `201 Created`:**
```json
{
  "session": {
    "id": 7,
    "track_id": 3,
    "duration_ms": 124350.5,
    "hit_count": 839,
    "started_at": "2026-02-19T10:05:00.000Z",
    "ended_at":   "2026-02-19T10:07:04.350Z",
    "created_at": "2026-02-19T10:07:05Z"
  },
  "score": {
    "hit_rate": 0.93,
    "miss_rate": 0.07,
    "ghost_rate": 0.04,
    "avg_offset_ms": -12.3,
    "std_offset_ms": 28.7,
    "by_instrument": [
      {
        "instrument": "Bass Drum",
        "target_count": 210,
        "matched_count": 198,
        "missed_count": 12,
        "ghost_count": 3,
        "hit_rate": 0.943,
        "avg_offset_ms": -8.1,
        "std_offset_ms": 22.5
      },
      {
        "instrument": "Snare",
        "target_count": 180,
        "matched_count": 172,
        "missed_count": 8,
        "ghost_count": 1,
        "hit_rate": 0.956,
        "avg_offset_ms": -15.2,
        "std_offset_ms": 31.0
      }
    ],
    "matched_hits": [
      {
        "target_time_ms": 500.0,
        "recorded_time_ms": 501.3,
        "offset_ms": 1.3,
        "note_number": 38,
        "instrument": "Snare"
      }
    ]
  }
}
```

**Score field reference:**

| Field | Description |
|-------|-------------|
| `hit_rate` | `matched_count / target_count` (0.0–1.0) |
| `miss_rate` | `missed_count / target_count` (0.0–1.0) |
| `ghost_rate` | `ghost_count / recorded_count` (0.0–1.0); 0 if no hits recorded |
| `avg_offset_ms` | Mean signed offset of all matched pairs. Negative = playing early, positive = playing late. |
| `std_offset_ms` | Standard deviation of offsets. Lower = more consistent timing. |
| `matched_hits[].offset_ms` | `recorded_time_ms - target_time_ms` |

**Error responses:**
- `400 Bad Request` — malformed JSON or missing required fields
- `404 Not Found` — `track_id` does not exist

---

### GET /api/v1/sessions

List sessions, optionally filtered by track.

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `track_id` | integer (optional) | Filter sessions for a specific track |

**Response `200 OK`:**
```json
{
  "sessions": [
    {
      "id": 7,
      "track_id": 3,
      "hit_count": 839,
      "hit_rate": 0.93,
      "avg_offset_ms": -12.3,
      "created_at": "2026-02-19T10:07:05Z"
    }
  ]
}
```

---

### GET /api/v1/sessions/:id

Retrieve the full result for a completed session. Used by the results page.

**Response `200 OK`:** Same shape as the POST /sessions response.

**Error responses:**
- `404 Not Found` — session ID does not exist

---

## Kit profiles

A kit profile maps the MIDI notes sent by the user's e-drum kit to instrument groups. The **active** profile is applied to recorded hits when a take is scored; a track's own note mapping (`/tracks/:id/mapping`) applies to the notes in the MIDI file. Without an active profile, hits use the track's mapping, as before kit profiles existed. Notes a profile doesn't list fall back to the General MIDI defaults.

### GET /api/v1/kits

**Response `200 OK`:**
```json
{
  "kits": [
    {
      "id": 1,
      "name": "Roland TD-17",
      "mappings": [{ "note_number": 26, "instrument": "Hi-Hat" }],
      "active": true,
      "created_at": "2026-10-17T08:00:00Z"
    }
  ]
}
```

### POST /api/v1/kits · PUT /api/v1/kits/:id

**Request body:** `{ "name": "Roland TD-17", "mappings": [{ "note_number": 26, "instrument": "Hi-Hat" }] }`

`name` is 1–60 characters; each `note_number` (0–127) may appear once; `instrument` is one of the instrument groups. Returns the saved kit (`201 Created` / `200 OK`). The user's first kit becomes active automatically.

### DELETE /api/v1/kits/:id

**Response `204 No Content`.** Deleting the active kit leaves no kit active.

### PUT /api/v1/kits/active

**Request body:** `{ "id": 1 }`, or `{ "id": null }` for standard General MIDI numbering. **Response `204 No Content`.**

**Error responses (all kit endpoints):** `400` for invalid input, `404` if the kit doesn't belong to the user.

---

## Stats

### GET /api/v1/stats

Practice statistics for the trailing 12 months.

**Query parameters:**
- `tz` (optional) — IANA time zone (e.g. `Europe/Zurich`) used for week and day boundaries; default UTC.

**Response `200 OK`:**
```json
{
  "weekly": [
    { "week_start": "2026-10-12", "session_count": 8, "minutes_played": 1.2, "tracks_practiced": 2 }
  ],
  "totals": { "sessions": 8, "minutes_played": 1.2, "tracks_practiced": 2 },
  "streak_days": 3,
  "practiced_today": true
}
```

`weekly` always has 52 entries ending with the current week. `streak_days` counts consecutive days with at least one take, ending today — or yesterday, so a streak isn't shown as lost before today's practice.

---

## Health

### GET /api/v1/health

`version` is the build version (the image tag in Docker builds, `dev` for local builds).

**Response `200 OK`:**
```json
{ "status": "ok", "version": "1.2.3" }
```

---

## Error Response Shape

All error responses use a consistent envelope:

```json
{
  "error": "human-readable error message"
}
```

---

## CORS

In development, Vite proxies `/api/*` to the Go backend, so no CORS headers are needed in the browser. The Go CORS middleware (`github.com/go-chi/cors`) is configured for production deployments where the frontend and backend may be on different origins.
