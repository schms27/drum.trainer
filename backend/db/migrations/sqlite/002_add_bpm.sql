-- Stores the initial tempo of the uploaded MIDI file.
-- Existing rows default to 120 BPM.
ALTER TABLE tracks ADD COLUMN bpm REAL NOT NULL DEFAULT 120;
