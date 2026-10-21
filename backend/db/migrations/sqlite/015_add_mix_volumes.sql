-- Playback levels (0–1, NULL = not set) for the session page's mixer.
-- A track remembers its metronome level and each backing track its own level;
-- a kit profile holds the defaults used when a track has none saved.
ALTER TABLE tracks ADD COLUMN metronome_volume REAL;
ALTER TABLE pilot_tracks ADD COLUMN volume REAL;
ALTER TABLE kit_profiles ADD COLUMN metronome_volume REAL;
ALTER TABLE kit_profiles ADD COLUMN pilot_volume REAL;
