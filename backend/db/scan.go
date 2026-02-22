package db

import (
	"fmt"
	"time"
)

// scanTime implements sql.Scanner for timestamp columns that return time.Time
// (PostgreSQL) or string (SQLite).
type scanTime struct{ T time.Time }

func (s *scanTime) Scan(v any) error {
	switch val := v.(type) {
	case time.Time:
		s.T = val
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", val)
		if err != nil {
			t, err = time.Parse(time.RFC3339, val)
		}
		if err != nil {
			return fmt.Errorf("scanTime: cannot parse %q: %w", val, err)
		}
		s.T = t
	case []byte:
		return s.Scan(string(val))
	default:
		return fmt.Errorf("scanTime: unsupported type %T", v)
	}
	return nil
}
