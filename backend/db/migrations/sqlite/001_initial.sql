PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tracks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    filename    TEXT    NOT NULL,
    file_path   TEXT    NOT NULL,
    file_size   INTEGER NOT NULL,
    duration_ms REAL    NOT NULL,
    note_count  INTEGER NOT NULL,
    ppq         INTEGER NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS track_notes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    note_number INTEGER NOT NULL,
    time_ms     REAL    NOT NULL,
    velocity    INTEGER NOT NULL,
    instrument  TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_track_notes_track_id ON track_notes(track_id);
CREATE INDEX IF NOT EXISTS idx_track_notes_time     ON track_notes(track_id, time_ms);

CREATE TABLE IF NOT EXISTS sessions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    started_at  TEXT    NOT NULL,
    ended_at    TEXT    NOT NULL,
    duration_ms REAL    NOT NULL,
    hit_count   INTEGER NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS session_hits (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id   INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    note_number  INTEGER NOT NULL,
    time_ms      REAL    NOT NULL,
    velocity     INTEGER NOT NULL,
    instrument   TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_session_hits_session ON session_hits(session_id);

CREATE TABLE IF NOT EXISTS session_scores (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      INTEGER NOT NULL UNIQUE REFERENCES sessions(id) ON DELETE CASCADE,
    hit_rate        REAL    NOT NULL,
    miss_rate       REAL    NOT NULL,
    ghost_rate      REAL    NOT NULL,
    avg_offset_ms   REAL    NOT NULL,
    std_offset_ms   REAL    NOT NULL,
    score_json      TEXT    NOT NULL,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS session_instrument_scores (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    instrument      TEXT    NOT NULL,
    target_count    INTEGER NOT NULL,
    matched_count   INTEGER NOT NULL,
    missed_count    INTEGER NOT NULL,
    ghost_count     INTEGER NOT NULL,
    hit_rate        REAL    NOT NULL,
    avg_offset_ms   REAL,
    std_offset_ms   REAL
);

CREATE INDEX IF NOT EXISTS idx_instr_scores_session ON session_instrument_scores(session_id);

CREATE TABLE IF NOT EXISTS track_note_mappings (
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    note_number INTEGER NOT NULL,
    instrument  TEXT    NOT NULL,
    PRIMARY KEY (track_id, note_number)
);
