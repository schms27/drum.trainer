-- Kit profiles: how the pads of a user's e-drum kit map to instruments.
-- Applied to recorded hits; track_note_mappings applies to a file's notes.
CREATE TABLE IF NOT EXISTS kit_profiles (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_kit_profiles_user ON kit_profiles(user_id);

CREATE TABLE IF NOT EXISTS kit_profile_mappings (
    profile_id  INTEGER NOT NULL REFERENCES kit_profiles(id) ON DELETE CASCADE,
    note_number INTEGER NOT NULL,
    instrument  TEXT    NOT NULL,
    PRIMARY KEY (profile_id, note_number)
);

ALTER TABLE users ADD COLUMN active_kit_profile_id INTEGER REFERENCES kit_profiles(id) ON DELETE SET NULL;
