CREATE TABLE IF NOT EXISTS tracks (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT           NOT NULL,
    filename    TEXT           NOT NULL,
    file_path   TEXT           NOT NULL,
    file_size   BIGINT         NOT NULL,
    duration_ms DOUBLE PRECISION NOT NULL,
    note_count  INTEGER        NOT NULL,
    ppq         INTEGER        NOT NULL,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS track_notes (
    id          BIGSERIAL PRIMARY KEY,
    track_id    BIGINT  NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    note_number INTEGER NOT NULL,
    time_ms     DOUBLE PRECISION NOT NULL,
    velocity    INTEGER NOT NULL,
    instrument  TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_track_notes_track_id ON track_notes(track_id);
CREATE INDEX IF NOT EXISTS idx_track_notes_time     ON track_notes(track_id, time_ms);

CREATE TABLE IF NOT EXISTS sessions (
    id          BIGSERIAL PRIMARY KEY,
    track_id    BIGINT  NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL,
    ended_at    TIMESTAMPTZ NOT NULL,
    duration_ms DOUBLE PRECISION NOT NULL,
    hit_count   INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS session_hits (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT  NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    note_number INTEGER NOT NULL,
    time_ms     DOUBLE PRECISION NOT NULL,
    velocity    INTEGER NOT NULL,
    instrument  TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_session_hits_session ON session_hits(session_id);

CREATE TABLE IF NOT EXISTS session_scores (
    id            BIGSERIAL PRIMARY KEY,
    session_id    BIGINT  NOT NULL UNIQUE REFERENCES sessions(id) ON DELETE CASCADE,
    hit_rate      DOUBLE PRECISION NOT NULL,
    miss_rate     DOUBLE PRECISION NOT NULL,
    ghost_rate    DOUBLE PRECISION NOT NULL,
    avg_offset_ms DOUBLE PRECISION NOT NULL,
    std_offset_ms DOUBLE PRECISION NOT NULL,
    score_json    TEXT    NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS session_instrument_scores (
    id            BIGSERIAL PRIMARY KEY,
    session_id    BIGINT  NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    instrument    TEXT    NOT NULL,
    target_count  INTEGER NOT NULL,
    matched_count INTEGER NOT NULL,
    missed_count  INTEGER NOT NULL,
    ghost_count   INTEGER NOT NULL,
    hit_rate      DOUBLE PRECISION NOT NULL,
    avg_offset_ms DOUBLE PRECISION,
    std_offset_ms DOUBLE PRECISION
);

CREATE INDEX IF NOT EXISTS idx_instr_scores_session ON session_instrument_scores(session_id);

CREATE TABLE IF NOT EXISTS track_note_mappings (
    track_id    BIGINT  NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    note_number INTEGER NOT NULL,
    instrument  TEXT    NOT NULL,
    PRIMARY KEY (track_id, note_number)
);
