package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
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

// UniqueFilename generates a collision-resistant storage filename while
// preserving the original file extension.
func UniqueFilename(original string) string {
	ext := filepath.Ext(original)
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
}
