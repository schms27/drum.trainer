-- Playback levels (0–1, NULL = not set) for the session page's mixer.
-- A track remembers its metronome level and each backing track its own level;
-- a kit profile holds the defaults used when a track has none saved.
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS metronome_volume DOUBLE PRECISION;
ALTER TABLE pilot_tracks ADD COLUMN IF NOT EXISTS volume DOUBLE PRECISION;
ALTER TABLE kit_profiles ADD COLUMN IF NOT EXISTS metronome_volume DOUBLE PRECISION;
ALTER TABLE kit_profiles ADD COLUMN IF NOT EXISTS pilot_volume DOUBLE PRECISION;
