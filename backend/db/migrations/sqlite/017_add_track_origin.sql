-- Where a track came from: 'upload' (a MIDI file) or 'editor' (created from
-- scratch in the track editor, so it has no uploaded file).
ALTER TABLE tracks ADD COLUMN origin TEXT NOT NULL DEFAULT 'upload';
