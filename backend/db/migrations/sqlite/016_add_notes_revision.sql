-- Bumped every time a track's notes are edited. Each take records the
-- revision it was played against, so progress can tell takes before and after
-- an edit apart (NULL for takes recorded before this migration).
ALTER TABLE tracks ADD COLUMN notes_revision INTEGER NOT NULL DEFAULT 0;
ALTER TABLE sessions ADD COLUMN notes_revision INTEGER;
