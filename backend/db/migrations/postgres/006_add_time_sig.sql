-- Stores the time signature numerator (beats per bar) parsed from the MIDI file.
-- Existing rows default to 4 (i.e. 4/4 time).
ALTER TABLE tracks ADD COLUMN time_sig_numerator INTEGER NOT NULL DEFAULT 4;
