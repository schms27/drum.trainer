# Track Sharing — Implementation Plan

A user can share one of their drum tracks with others, either through an **unlisted link**
(anyone holding the link can see it) or by making it **public** (listed in an Explore
gallery). Recipients can add a copy to their own library and practise it like any
other track.

This plan also covers the copyright exposure that sharing creates and how to push
responsibility back onto the uploader as far as the law allows.

> Not legal advice. The copyright section describes the usual mechanisms platforms
> rely on; before opening public sharing to strangers, have the terms of use and the
> takedown process checked by someone qualified for the jurisdiction the service is
> operated from.

---

## What gets shared (and what doesn't)

| Item | Shared? | Why |
|---|---|---|
| Track metadata (name, BPM, time sig, duration, note count) | yes | needed for the landing page |
| Parsed drum notes (`track_notes`) | yes | this is the practice content |
| Owner's note mapping | yes (copied as the starting mapping) | makes the copy usable immediately |
| Original `.mid` file | **no** | may contain the full arrangement (melody, bass, chords); we only ever parsed channel 10. The owner can download their own upload (`GET /tracks/:id/file`) and export its drum part (`GET /tracks/:id/export.mid`), but copies of shared tracks and removed tracks can't be downloaded by anyone (issue #12) |
| Pilot tracks (audio) | **no** | audio is almost always a copyrighted sound recording — by far the highest-risk content. See "Pilot tracks" below |
| Owner's sessions / scores / practice BPM | no | personal data |

This scoping is the single most effective copyright measure in the plan: the platform
only ever redistributes a drum part as note events, never audio and never the full
MIDI arrangement.

---

## Core design decision: copy, don't reference

`practice_bpm` and the note mapping live on the `tracks` row, and `sessions.track_id`
cascades on delete. If recipients practised the owner's track directly, the owner
deleting it would wipe every recipient's history, and per-user settings would need to
move into a new table.

**Chosen approach:** "Add to my library" **clones** the track for the recipient
(track row + `track_notes` + `track_note_mappings`, no pilots), with a lineage column
`source_track_id`. Every existing endpoint keeps working unchanged because the
recipient owns the copy.

Consequences:
- Owner un-sharing or deleting the original does not remove existing copies.
- A **takedown propagates along lineage**: disabling a track also disables every copy
  with that `source_track_id` (recursively, since copies can't be re-shared — see below).
- Copies **cannot be re-shared** in v1 (`source_track_id IS NOT NULL` → sharing
  refused). That keeps lineage one level deep and means only the original uploader
  ever makes the "I have the rights" attestation.

---

## Data model

### Migration `011_add_sharing.sql` (sqlite + postgres variants)

```sql
ALTER TABLE tracks ADD COLUMN visibility       TEXT NOT NULL DEFAULT 'private'; -- private | unlisted | public
ALTER TABLE tracks ADD COLUMN share_token      TEXT;           -- 22-char base62, 128 bits of entropy
ALTER TABLE tracks ADD COLUMN shared_at        TEXT;
ALTER TABLE tracks ADD COLUMN rights_basis     TEXT;           -- original | licensed | public_domain | transcription
ALTER TABLE tracks ADD COLUMN rights_terms_ver TEXT;           -- terms version the owner accepted
ALTER TABLE tracks ADD COLUMN source_track_id  INTEGER REFERENCES tracks(id) ON DELETE SET NULL;
ALTER TABLE tracks ADD COLUMN disabled_at      TEXT;           -- set by moderation; hides track + copies
ALTER TABLE tracks ADD COLUMN disabled_reason  TEXT;

CREATE UNIQUE INDEX idx_tracks_share_token ON tracks(share_token);
CREATE INDEX idx_tracks_public ON tracks(visibility, shared_at) ;
CREATE INDEX idx_tracks_source ON tracks(source_track_id);

-- Immutable audit trail of every visibility change + attestation.
CREATE TABLE track_share_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    visibility  TEXT    NOT NULL,
    rights_basis TEXT,
    terms_ver   TEXT,
    ip_hash     TEXT,               -- salted hash, not the raw IP
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Copyright / abuse reports. Reporter may be anonymous.
CREATE TABLE track_reports (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id         INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    reporter_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    reporter_name    TEXT,
    reporter_email   TEXT NOT NULL,
    kind             TEXT NOT NULL,  -- copyright | abuse | other
    work_identified  TEXT,           -- "which work is infringed"
    details          TEXT NOT NULL,
    good_faith       INTEGER NOT NULL DEFAULT 0,  -- sworn statement checkbox
    status           TEXT NOT NULL DEFAULT 'open', -- open | actioned | rejected
    resolution_note  TEXT,
    created_at       TEXT NOT NULL DEFAULT (datetime('now')),
    resolved_at      TEXT
);

ALTER TABLE users ADD COLUMN is_admin     INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN strikes      INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN sharing_banned_at TEXT;
```

`models.Track` gains `Visibility`, `ShareURL` (owner only), `SourceTrackID`,
`RightsBasis`, `Disabled bool`. All existing owner queries add
`AND disabled_at IS NULL` or surface the disabled state in the UI ("removed after a
copyright notice").

Admin bootstrap: `ADMIN_EMAILS` env var in `config.Config`; matching users get
`is_admin = 1` on login.

---

## API

All new routes live in `backend/api/sharing.go`; DB code in `backend/db/sharing.go`.

### Owner (auth required)

| Method | Path | Body / notes |
|---|---|---|
| `PUT` | `/tracks/{id}/sharing` | `{visibility, rights_basis, accept_terms: true}`. Refused if the track is a copy, disabled, or the user is sharing-banned. `public` additionally requires `rights_basis ≠ transcription` (see below). Generates `share_token` on first share; writes a `track_share_events` row. |
| `POST` | `/tracks/{id}/sharing/rotate` | New token; old link stops working. |
| `DELETE` | `/tracks/{id}/sharing` | Back to private, token cleared. |

### Recipient

| Method | Path | Auth | Notes |
|---|---|---|---|
| `GET` | `/shared/{token}` | none | Metadata + owner display name + note preview (first N bars, for a mini timeline). Never returns notes in full to anonymous users. `X-Robots-Tag: noindex` for unlisted. Rate-limited per IP. |
| `POST` | `/shared/{token}/copy` | required | Clones into caller's library, returns new track id. Idempotent per (user, source). |
| `GET` | `/explore?q=&sort=&page=` | none | `visibility='public' AND disabled_at IS NULL`, paginated. |
| `POST` | `/reports` | none | `{token or track_id, kind, email, work_identified, details, good_faith}`. Strict rate limit (e.g. 5/hour/IP). |

Requiring login to copy gives accountability, blocks anonymous scraping of full note
data, and keeps the anonymous surface small.

### Admin (auth + `is_admin`)

| Method | Path | Notes |
|---|---|---|
| `GET` | `/admin/reports?status=open` | Queue with track + owner info. |
| `POST` | `/admin/tracks/{id}/disable` | `{reason, report_id?, strike: bool}`. Sets `disabled_at` on the track **and all copies**, forces visibility private, optionally increments owner's strikes; 3 strikes → `sharing_banned_at`. Emails owner (later; log for now). |
| `POST` | `/admin/tracks/{id}/restore` | Counter-notice accepted / report rejected. |

---

## Frontend

- **Share dialog** (`components/tracks/ShareDialog.svelte`, opened from the track
  detail page): visibility radio (Private / Anyone with the link / Public), rights
  basis select, the attestation checkbox with a link to the terms, copy-link button,
  "reset link" button. Explains plainly that audio pilots and the original MIDI file are
  not shared.
- **Landing page** `routes/s/[token]/` — track info, mini timeline preview, "Add to my
  library" (→ login redirect with return URL if anonymous), "Report" link.
- **Explore** `routes/explore/` — searchable grid reusing `TrackCard`.
- **Report form** `routes/report/` — prefilled from the landing page; also reachable
  from the footer as "Copyright notice".
- **Static pages**: `routes/terms/`, `routes/copyright/` (takedown policy + contact),
  and an imprint/contact page.
- **Admin** `routes/admin/reports/` — minimal table, disable/restore buttons. Hidden
  unless `me.is_admin`.
- `TrackCard`: badges for "Shared" / "Public" / "Copy of … by …" / "Removed".

SPA caveat: `adapter-static` means link previews (Open Graph) in chats will be generic.
If rich previews matter later, nginx can route `/s/*` for bot user-agents to a tiny
backend-rendered HTML page with OG tags.

---

## Copyright: exposure and how to shift it

### Where the risk actually is

1. **Pilot audio** — recordings of commercial songs are copyrighted sound recordings,
   and redistributing them is unambiguous infringement. **Excluded from sharing.**
2. **Full MIDI arrangements** — a MIDI file of a song reproduces the composition
   (melody, harmony, lyrics events). Commercial MIDI sites have historically received
   publisher takedowns. **Excluded: we never serve the original file.**
3. **Drum part as note events** — what we would share. Rhythms and grooves by
   themselves get thin protection, but a note-for-note transcription of a specific
   song's drum part, labelled with the song's title, is still arguably a reproduction
   of part of the work. Low but not zero risk. Song titles in a public, searchable
   gallery increase it a lot, both because it attracts rights holders' automated
   searches and because curating or promoting such content weakens a "we didn't know"
   defence.
4. **Unlisted links** are lower risk than public listing: no discovery, no indexing,
   comparable to a user emailing a file.

### What "putting it on the user" can and can't do

Terms of use **can**:
- make the uploader warrant they own or have a licence for anything they share;
- grant the platform only the licence needed to host and display it;
- include an indemnity (the uploader covers costs caused by their infringing upload);
- reserve the right to remove content and terminate sharing for repeat infringers.

Terms of use **cannot** make the platform immune to rights holders. A rights holder
isn't bound by your terms and can go after the host directly. In practice, an
indemnity from a hobbyist is worth little. What protects the operator is the hosting
safe harbour, and that depends on **behaviour, not contract wording**:

- **EU (Digital Services Act, Art. 6 + 16):** a host isn't liable for user content as
  long as it has no actual knowledge of it and removes it promptly once notified. It
  has to offer an easy electronic notice mechanism, give the uploader a statement of
  reasons when removing, and publish a point of contact. Art. 17 of the DSM
  Copyright Directive (upload-filter duties) targets large content-sharing services;
  small new services get reduced obligations, but it's worth confirming this applies.
- **Switzerland (URG Art. 39d):** hosts whose service creates a *particular risk* of
  infringement must also prevent re-uploads of content they've taken down
  ("stay-down"). Not encouraging infringement (no song-title browsing, no audio) keeps
  the service outside that category.
- **US (DMCA §512), if US users are targeted:** register a designated agent with the
  Copyright Office, publish the takedown/counter-notice procedure, and keep a
  repeat-infringer policy.

### Concrete measures in this plan

| Measure | Where | Effect |
|---|---|---|
| Share only drum note events; no audio, no `.mid` file | scope table | removes the high-risk content entirely |
| Unlisted is the default share mode | Share dialog | lower discovery = lower exposure |
| `rights_basis` declaration + per-share attestation with terms version, timestamp, hashed IP | `track_share_events` | documented user responsibility; evidence for indemnity and for showing good faith |
| `transcription` tracks can be shared by link but **not listed publicly** | `PUT /sharing` validation | keeps the public gallery to originals, exercises, licensed and public-domain material |
| Explore shows no "trending songs"/artist pages, no editorial featuring | Explore | avoids knowledge/promotion of specific works |
| Report + copyright notice form, reachable without an account | `/report`, `/copyright` | notice-and-action mechanism required for the safe harbour |
| Admin queue; disable propagates to all copies; strike system and sharing ban | admin API | prompt removal, repeat-infringer policy, plus a basic stay-down |
| Re-sharing of copies blocked | `PUT /sharing` | one accountable uploader per lineage |
| Terms, takedown policy, imprint/contact pages | static routes | legally expected contact point; makes the policy enforceable |
| Rate limits on anonymous endpoints; full notes only for logged-in users | router | prevents bulk scraping, which would turn the service into a distribution mirror |

### Pilot tracks — possible later phase

If sharing audio is wanted later, allow it only for **unlisted** links, only with
`rights_basis = original` (the user's own click track or own recording), with a
separate attestation, and treat any copyright report on an audio file as immediate
disable with no counter-notice wait. Leave it out of v1.

---

## Phases

1. **Unlisted links (MVP):** migration, owner sharing endpoints, `/shared/{token}`,
   copy-to-library, Share dialog, landing page, terms page + attestation. No public
   listing yet, so the moderation surface stays small. *Done:* migration `011` adds only
   the phase-1 columns and `track_share_events`; the moderation columns, `track_reports`
   and the user flags arrive with phase 2. Sign-in also accepts `?next=` so recipients
   return to the share page.
2. **Moderation baseline:** report form, `/copyright` page, admin role and queue,
   disable/restore with lineage propagation, strikes. **Must ship before phase 3.**
   *Done*, with these differences from the design above: admins come from
   `ADMIN_EMAILS` on every request (no `is_admin` column, so removal takes effect
   at once); copies carry `origin_track_id` (no foreign key, never cleared) so a
   takedown still reaches them after the uploader deletes the original; removed
   tracks' note fingerprints go into `blocked_fingerprints` (basic stay-down);
   `track_reports` snapshots the track name and owner so reports outlive the
   track; the contact address for notices is `CONTACT_EMAIL`, served at
   `GET /api/v1/site`. Owners aren't emailed yet: they see the reason on the
   track page, and takedowns are logged.
3. **Public Explore:** `public` visibility, gallery with search, the
   no-transcriptions-public rule, badges. *Done:* `/explore` (no account needed) with name
   search and newest / most-added sorting; `listed_at` (migration 013) orders newly listed
   tracks first. The share dialog edits rights basis and visibility together, and every
   change is a fresh, logged attestation. Explore shows no owner identities, no artist or
   song pages and no editorial picks. Only public share pages may be indexed.
4. **Later:** owner notifications by email, OG previews for `/s/*`, optional audio
   sharing as described above, "update my copy" when the original changes.

## Tests

- DB: clone copies notes and mapping but not pilots or sessions; disable cascades to
  copies; token rotation invalidates the old token; copies can't be shared.
- API: anonymous `/shared/{token}` returns no full note list; disabled or private
  tracks return 404 (never 403, so private tracks can't be probed); `public` +
  `transcription` is rejected; sharing-banned user gets 403; non-admins can't reach
  `/admin/*`.
- Migrations on both SQLite and Postgres (existing `db_test.go` pattern).
