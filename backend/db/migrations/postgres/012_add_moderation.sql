-- Moderation of shared tracks: copyright/abuse reports, takedowns that reach
-- every copy, strikes against repeat infringers and a basic stay-down list.

-- origin_track_id is the track a copy was made from. Unlike source_track_id it
-- has no foreign key and is never cleared, so a takedown still reaches copies
-- after the uploader deletes the original.
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS origin_track_id INTEGER;
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS disabled_at     TIMESTAMPTZ;
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS disabled_reason TEXT;
UPDATE tracks SET origin_track_id = source_track_id WHERE source_track_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tracks_origin ON tracks(origin_track_id);

ALTER TABLE users ADD COLUMN IF NOT EXISTS strikes           INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS sharing_banned_at TIMESTAMPTZ;

-- Reports outlive the reported track: track_id is a plain number and the
-- track's name and owner are snapshotted when the report is filed.
CREATE TABLE IF NOT EXISTS track_reports (
    id               SERIAL      PRIMARY KEY,
    track_id         INTEGER     NOT NULL,
    track_name       TEXT        NOT NULL,
    owner_user_id    INTEGER     REFERENCES users(id) ON DELETE SET NULL,
    share_token      TEXT        NOT NULL,
    reporter_user_id INTEGER     REFERENCES users(id) ON DELETE SET NULL,
    reporter_name    TEXT        NOT NULL,
    reporter_email   TEXT        NOT NULL,
    kind             TEXT        NOT NULL, -- copyright | abuse | other
    work_identified  TEXT        NOT NULL DEFAULT '',
    details          TEXT        NOT NULL,
    good_faith       BOOLEAN     NOT NULL DEFAULT FALSE,
    ip_hash          TEXT,
    status           TEXT        NOT NULL DEFAULT 'open', -- open | actioned | rejected
    resolution_note  TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_track_reports_status ON track_reports(status, created_at);
CREATE INDEX IF NOT EXISTS idx_track_reports_track  ON track_reports(track_id);

-- Note fingerprints of removed tracks. Sharing a track with the same notes is
-- refused, so a takedown can't be undone by uploading the file again.
CREATE TABLE IF NOT EXISTS blocked_fingerprints (
    fingerprint TEXT        PRIMARY KEY,
    track_id    INTEGER     NOT NULL,
    reason      TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
