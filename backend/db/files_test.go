package db

import "testing"

func TestNormalizeFilePaths(t *testing.T) {
	db := openTestDB(t)

	user, err := UpsertUser(db, "github", "1", "a@example.com", "A", "")
	if err != nil {
		t.Fatal(err)
	}
	insert := func(path string) int64 {
		var id int64
		err := db.QueryRow(
			`INSERT INTO tracks (name, filename, file_path, file_size, duration_ms, note_count, ppq, user_id)
			 VALUES ('t', 't.mid', ?, 1, 1, 1, 480, ?) RETURNING id`, path, user.ID,
		).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	cases := map[int64]string{
		insert("/data/uploads/111.mid"): "111.mid",
		insert(`data\uploads\222.mid`):  "222.mid",
		insert("data/uploads/333.mid"):  "333.mid",
		insert("444.mid"):               "444.mid",
	}

	if err := NormalizeFilePaths(db); err != nil {
		t.Fatal(err)
	}
	for id, want := range cases {
		var got string
		if err := db.QueryRow(`SELECT file_path FROM tracks WHERE id = ?`, id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("track %d: file_path = %q, want %q", id, got, want)
		}
	}
}
