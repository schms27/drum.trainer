-- Tempo a take was played at. practice_bpm is the tempo the player chose;
-- track_bpm is the track's native tempo at the time, so the percentage stays
-- right if the track's tempo is edited later. Both are NULL for takes recorded
-- before this migration (tempo unknown).
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS practice_bpm DOUBLE PRECISION;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS track_bpm DOUBLE PRECISION;
