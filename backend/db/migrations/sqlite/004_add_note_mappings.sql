-- Custom note-to-instrument overrides per track.
-- Created as a separate migration because this table was added to 001_initial.sql
-- after some databases were already initialised.
CREATE TABLE IF NOT EXISTS track_note_mappings (
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    note_number INTEGER NOT NULL,
    instrument  TEXT    NOT NULL,
    PRIMARY KEY (track_id, note_number)
);
