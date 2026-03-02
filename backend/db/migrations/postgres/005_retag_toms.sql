-- Split the single "Tom" instrument group into four pitch-based levels.
-- Applies to both stored track notes and recorded session hits.

UPDATE track_notes SET instrument = 'Tom (high)'  WHERE instrument = 'Tom' AND note_number IN (50);
UPDATE track_notes SET instrument = 'Tom (mid)'   WHERE instrument = 'Tom' AND note_number IN (47, 48);
UPDATE track_notes SET instrument = 'Tom (low)'   WHERE instrument = 'Tom' AND note_number IN (43, 45);
UPDATE track_notes SET instrument = 'Tom (floor)' WHERE instrument = 'Tom' AND note_number IN (41);

UPDATE session_hits SET instrument = 'Tom (high)'  WHERE instrument = 'Tom' AND note_number IN (50);
UPDATE session_hits SET instrument = 'Tom (mid)'   WHERE instrument = 'Tom' AND note_number IN (47, 48);
UPDATE session_hits SET instrument = 'Tom (low)'   WHERE instrument = 'Tom' AND note_number IN (43, 45);
UPDATE session_hits SET instrument = 'Tom (floor)' WHERE instrument = 'Tom' AND note_number IN (41);
