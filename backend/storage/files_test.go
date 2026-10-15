package storage

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveStream(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	data := bytes.Repeat([]byte("x"), 100_000)
	n, err := s.SaveStream("a.wav", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(data)) {
		t.Fatalf("wrote %d bytes, want %d", n, len(data))
	}
	f, err := s.Open("a.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("saved content differs from input")
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestSaveStreamFailureLeavesNoFiles(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	r := io.MultiReader(strings.NewReader("partial"), failingReader{})
	if _, err := s.SaveStream("a.wav", r); err == nil {
		t.Fatal("expected error")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("expected empty dir, found %d entries", len(entries))
	}
}

func TestUniqueFilename(t *testing.T) {
	a, b := UniqueFilename("song.MP3"), UniqueFilename("song.MP3")
	if a == b {
		t.Fatal("expected distinct names")
	}
	if !strings.HasSuffix(a, ".mp3") || len(a) != 32+4 {
		t.Fatalf("unexpected name %q", a)
	}
	for _, name := range []string{"noext", `evil.\..\x`, "a.ph p", "a.verylongextension"} {
		if got := UniqueFilename(name); len(got) != 32 {
			t.Errorf("%q: expected no extension, got %q", name, got)
		}
	}
}

func TestPathStaysInsideStore(t *testing.T) {
	s := &Store{Dir: "/uploads"}
	for _, name := range []string{"../../etc/passwd", "/etc/passwd", "sub/x.mid"} {
		if got := s.path(name); filepath.Dir(got) != filepath.Clean("/uploads") {
			t.Errorf("%q resolved outside store: %q", name, got)
		}
	}
}
