package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Store handles saving and reading uploaded MIDI files on disk.
type Store struct {
	Dir string
}

// NewStore creates a Store and ensures the upload directory exists.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir %q: %w", dir, err)
	}
	return &Store{Dir: dir}, nil
}

// Save writes data to a file named <filename> inside the store directory.
// Returns the absolute path of the saved file.
func (s *Store) Save(filename string, data []byte) (string, error) {
	path := filepath.Join(s.Dir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write file %q: %w", path, err)
	}
	return path, nil
}

// SaveStream copies r into a file named <filename> inside the store directory
// without buffering it in memory. The data is written to a temporary file
// first and renamed into place, so a failed copy never leaves a partial file.
// Returns the path of the saved file and the number of bytes written.
func (s *Store) SaveStream(filename string, r io.Reader) (string, int64, error) {
	path := filepath.Join(s.Dir, filename)
	tmp, err := os.CreateTemp(s.Dir, ".upload-*")
	if err != nil {
		return "", 0, fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	n, err := io.Copy(tmp, r)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmpName)
		return "", 0, fmt.Errorf("write file %q: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return "", 0, fmt.Errorf("rename to %q: %w", path, err)
	}
	return path, n, nil
}

// Read reads and returns the contents of the file at the given absolute path.
func (s *Store) Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}
	return data, nil
}

// Delete removes a file at the given absolute path. Safe to call if the file
// does not exist.
func (s *Store) Delete(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file %q: %w", path, err)
	}
	return nil
}

// UniqueFilename generates a random storage filename while preserving the
// original file extension (if it is a plain alphanumeric extension).
func UniqueFilename(original string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b) + safeExt(original)
}

// safeExt returns the lower-cased extension of name, or "" if it contains
// anything other than ASCII letters and digits or is unreasonably long.
func safeExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if len(ext) < 2 || len(ext) > 8 {
		return ""
	}
	for _, c := range ext[1:] {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			return ""
		}
	}
	return ext
}
