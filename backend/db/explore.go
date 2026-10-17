package db

import (
	"strings"

	"github.com/simu/drum-trainer/backend/models"
)

// Explore sort orders.
const (
	SortNewest  = "newest"  // most recently listed first
	SortPopular = "popular" // most often added to libraries first
)

// ListPublicTracks returns one page of the Explore gallery: public tracks that
// moderation hasn't removed, optionally filtered by a case-insensitive name
// search. It also returns whether another page follows.
func ListPublicTracks(db *DB, query, sort string, limit, offset int) ([]models.PublicTrack, bool, error) {
	order := "t.listed_at DESC, t.id DESC"
	if sort == SortPopular {
		order = "copies DESC, t.listed_at DESC, t.id DESC"
	}
	pattern := "%" + escapeLike(strings.ToLower(strings.TrimSpace(query))) + "%"

	// Fetch one extra row to learn whether there is a next page.
	rows, err := db.Query(db.Rebind(`SELECT
		t.share_token, t.name, t.duration_ms, t.note_count, t.bpm, t.time_sig_numerator, t.rights_basis, t.listed_at,
		(SELECT COUNT(*) FROM tracks c WHERE c.origin_track_id = t.id) AS copies
		FROM tracks t
		WHERE t.visibility = ? AND t.disabled_at IS NULL AND t.share_token IS NOT NULL
		  AND LOWER(t.name) LIKE ? ESCAPE '!'
		ORDER BY `+order+`
		LIMIT ? OFFSET ?`),
		models.VisibilityPublic, pattern, limit+1, offset,
	)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	out := []models.PublicTrack{}
	for rows.Next() {
		var p models.PublicTrack
		var listed scanTime
		if err := rows.Scan(&p.ShareToken, &p.Name, &p.DurationMs, &p.NoteCount, &p.BPM, &p.TimeSigNumerator,
			&p.RightsBasis, &listed, &p.Copies); err != nil {
			return nil, false, err
		}
		p.ListedAt = listed.T
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	more := len(out) > limit
	if more {
		out = out[:limit]
	}
	return out, more, nil
}

// escapeLike escapes LIKE wildcards in s, using ! as the escape character
// (the same in SQLite and PostgreSQL, unlike backslash).
func escapeLike(s string) string {
	return strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(s)
}
