-- A user's own tags for organising their library. Names are unique per user,
-- ignoring case; colour is a key of models.TagColors. Tags are private: they
-- are never shared, and copies of shared tracks don't get them.
CREATE TABLE IF NOT EXISTS tags (
    id          SERIAL      PRIMARY KEY,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    color       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_user_name ON tags(user_id, lower(name));

CREATE TABLE IF NOT EXISTS track_tags (
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    tag_id      INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_track_tags_tag ON track_tags(tag_id);
