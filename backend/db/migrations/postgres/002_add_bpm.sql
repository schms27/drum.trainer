-- Stores the initial tempo of the uploaded MIDI file.
-- Existing rows default to 120 BPM.
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS bpm DOUBLE PRECISION NOT NULL DEFAULT 120;
