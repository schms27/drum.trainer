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

// Store handles saving and reading uploaded files on disk. Files are kept
// flat in Dir and addressed by name only, so the database never stores
// absolute paths and UPLOAD_DIR can move without breaking existing rows.
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

// path resolves a stored file name to its location on disk. Only the base
// name is used, so a name can never point outside the store directory.
func (s *Store) path(name string) string {
	return filepath.Join(s.Dir, filepath.Base(name))
}

// Save writes data to a file named <name> inside the store directory.
func (s *Store) Save(name string, data []byte) error {
	path := s.path(name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	return nil
}

// SaveStream copies r into a file named <name> inside the store directory
// without buffering it in memory. The data is written to a temporary file
// first and renamed into place, so a failed copy never leaves a partial file.
// Returns the number of bytes written.
func (s *Store) SaveStream(name string, r io.Reader) (int64, error) {
	path := s.path(name)
	tmp, err := os.CreateTemp(s.Dir, ".upload-*")
	if err != nil {
		return 0, fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	n, err := io.Copy(tmp, r)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmpName)
		return 0, fmt.Errorf("write file %q: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return 0, fmt.Errorf("rename to %q: %w", path, err)
	}
	return n, nil
}

// Open opens the stored file <name> for reading.
func (s *Store) Open(name string) (*os.File, error) {
	return os.Open(s.path(name))
}

// Delete removes the stored file <name>. Safe to call if the file does not
// exist.
func (s *Store) Delete(name string) error {
	path := s.path(name)
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
