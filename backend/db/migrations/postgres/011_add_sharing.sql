-- Track sharing via unlisted links. A recipient copies a shared track into
-- their own library; source_track_id records where the copy came from.
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS visibility       TEXT NOT NULL DEFAULT 'private'; -- private | unlisted
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS share_token      TEXT;
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS shared_at        TIMESTAMPTZ;
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS rights_basis     TEXT; -- original | licensed | public_domain | transcription
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS rights_terms_ver TEXT;
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS source_track_id  INTEGER REFERENCES tracks(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tracks_share_token ON tracks(share_token);
CREATE INDEX IF NOT EXISTS idx_tracks_source ON tracks(source_track_id);

-- Audit trail of every sharing change and the rights attestation that came with it.
CREATE TABLE IF NOT EXISTS track_share_events (
    id           SERIAL      PRIMARY KEY,
    track_id     INTEGER     NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    user_id      INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action       TEXT        NOT NULL, -- share | rotate | unshare
    visibility   TEXT        NOT NULL,
    rights_basis TEXT,
    terms_ver    TEXT,
    ip_hash      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_track_share_events_track ON track_share_events(track_id);
