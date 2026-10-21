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

### GET /api/v1/tracks/:id/file · GET /api/v1/tracks/:id/export.mid

Downloads for the track's owner (`Content-Disposition: attachment`, `audio/midi`):

- `/file` returns the MIDI file exactly as uploaded, under its original filename.
- `/export.mid` renders the track's drum notes as a new type-0 MIDI file: channel 10 only, with the track's name, tempo and time signature, as `<name> (drums).mid`.

Copies of shared tracks and tracks removed by moderation can't be downloaded: both return `404`, like a track that doesn't exist or belongs to someone else. Shared tracks are only ever handed out as notes inside the app.

---

### PUT /api/v1/tracks/:id/mix

Saves the session page's mixer for a track. Levels are 0–1; `null` clears a level, so the active kit profile's default (or full level) applies. Backing tracks that don't belong to the track are ignored.

**Request body:** `{ "metronome_volume": 0.6, "pilots": [{ "id": 4, "volume": 0.8 }] }`

**Response `204 No Content`** · `400` level outside 0–1 · `404` track not found

Tracks return the saved level as `metronome_volume`, backing tracks as `volume` (both `null` when not set).

---

## Sharing

Tracks can be shared through an unlisted link. Every track object carries `visibility`
(`"private"` | `"unlisted"`), `share_token` (null while private), `rights_basis` and
`source_track_id` (set on copies made from someone else's link). See
[sharing-plan.md](sharing-plan.md) for the design.

### PUT /api/v1/tracks/:id/sharing

Share a track, or change how an already-shared track is shared (the token is kept). Each call
is a new attestation, recorded in `track_share_events` with the current terms version.

**Request:** `{ "visibility": "unlisted" | "public", "rights_basis": "original" | "licensed" | "public_domain" | "transcription", "accept_terms": true }`

`visibility` defaults to `"unlisted"` (only people with the link). `"public"` also lists the track
in Explore and is refused (`400`) for `transcription`.

**Response `200 OK`:** `{ "visibility": "public", "share_token": "…", "rights_basis": "original" }`

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
list is never returned here. Unlisted tracks are sent with `X-Robots-Tag: noindex, nofollow`;
public ones may be indexed.

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

### GET /api/v1/explore?q=&sort=newest|popular&page=1

**No session required.** One page (24 tracks) of public tracks that moderation hasn't removed.
`q` searches names (case-insensitive, up to 100 characters); `popular` sorts by how many people
added the track. Rate-limited to 60 per minute per IP.

**Response `200 OK`:**
```json
{
  "tracks": [
    { "share_token": "…", "name": "Funk Groove", "duration_ms": 8000, "note_count": 48, "bpm": 120,
      "time_sig_numerator": 4, "rights_basis": "original", "listed_at": "2026-10-17T15:40:54Z", "copies": 3 }
  ],
  "page": 1,
  "has_more": false
}
```

### POST /api/v1/shared/:token/copy

Copy the shared track into the caller's library: notes and note mapping, but not the original
MIDI file, backing tracks, practice tempo or takes. Copying again returns the existing copy.

**Response `201 Created`** (or `200 OK` for the owner's own track): `{ "track_id": 7 }`

---

## Reports and moderation

Track objects also carry `copied` (true for every copy, even after the original is deleted),
`disabled` and `disabled_reason`. A disabled track is returned with `notes: []`, can't be shared,
and `POST /sessions` for it returns `409`. `GET /auth/me` adds `sharing_banned` and `is_admin`.

### POST /api/v1/reports

**No session required.** Rate-limited to 5 per hour per IP, counting invalid submissions.

**Request:**
```json
{
  "token": "https://drums.example.com/s/abc…",
  "kind": "copyright",
  "reporter_name": "Jane Doe",
  "reporter_email": "jane@example.com",
  "work_identified": "Song X by Band Y",
  "details": "The shared track copies our drum part note for note.",
  "good_faith": true
}
```
`token` may be a bare token or a full share link. `kind` is `copyright`, `abuse` or `other`;
copyright reports need `work_identified` and `good_faith: true`.

**Response `201 Created`:** `{ "id": 1 }` · `400` invalid fields · `404` link not (or no longer) shared · `429` rate limit

### Admin endpoints

Only for users whose email is listed in `ADMIN_EMAILS`; everyone else gets `404`.

| Method | Path | Body | Effect |
|---|---|---|---|
| `GET` | `/api/v1/admin/reports?status=open\|actioned\|rejected\|all` | — | Reports with the track's state, copy count and owner (name, email, strikes, ban) |
| `POST` | `/api/v1/admin/reports/:id/reject` | `{ "note": "…" }` | Closes an open report without action |
| `POST` | `/api/v1/admin/tracks/:id/disable` | `{ "reason": "…", "strike": true }` | Disables the track and every copy, kills the link, blocks its note fingerprint from being shared again, closes its open reports. `reason` is shown to the owner and copy holders. A strike counts once per takedown; 3 strikes ban the owner from sharing and unshare all their tracks |
| `POST` | `/api/v1/admin/tracks/:id/restore` | `{ "note": "…", "revoke_strike": true }` | Undoes a takedown (tracks stay private), unblocks the fingerprint, marks its reports rejected; revoking a strike lifts a ban below the limit |

`disable` and `restore` return `{ "tracks_affected": 2, "owner_strikes": 1, "owner_banned": false }`.

### GET /api/v1/site

**No session required.** Public settings for the static frontend: `{ "contact_email": "…" }`
(from `CONTACT_EMAIL`, empty if unset).

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
| `practice_bpm` | float (optional) | Tempo the take was played at, 20–300. Stored with the session together with the track's native BPM; omitted means the tempo is unknown. |

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
    "created_at": "2026-02-19T10:07:05Z",
    "practice_bpm": 96,
    "track_bpm": 120,
    "tempo_percent": 80
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

List sessions, optionally filtered by track and tempo.

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `track_id` | integer (optional) | Filter sessions for a specific track |
| `tempo_min` | integer (optional) | Only takes played at at least this percentage of the track's native tempo |
| `tempo_max` | integer (optional) | Only takes played at at most this percentage. `tempo_min=100&tempo_max=100` gives takes at the original tempo. Takes without a recorded tempo never match a tempo bound. |

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
      "created_at": "2026-02-19T10:07:05Z",
      "practice_bpm": 96,
      "track_bpm": 120,
      "tempo_percent": 80
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

### PUT /api/v1/kits/:id/mix

Saves the kit profile's default playback levels, used on tracks without their own mix: `{ "metronome_volume": 0.5, "pilot_volume": 0.8 }`. Levels are 0–1, `null` clears one (full level). Returns the kit, which carries `metronome_volume` and `pilot_volume`.

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
