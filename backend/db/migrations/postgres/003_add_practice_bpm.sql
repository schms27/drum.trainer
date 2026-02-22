-- Per-track practice tempo override set by the user.
-- NULL means "use the native track BPM".
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS practice_bpm DOUBLE PRECISION;
