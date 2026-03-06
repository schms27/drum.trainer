CREATE TABLE IF NOT EXISTS pilot_tracks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    filename    TEXT    NOT NULL,
    file_path   TEXT    NOT NULL,
    file_size   INTEGER NOT NULL,
    mime_type   TEXT    NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_pilot_tracks_track ON pilot_tracks(track_id);
