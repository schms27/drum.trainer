-- Public tracks (visibility = 'public') are listed in Explore. listed_at is
-- when the track last became public, so newly listed tracks sort first even
-- if they were shared by link long before.
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS listed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_tracks_listed ON tracks(visibility, listed_at);
