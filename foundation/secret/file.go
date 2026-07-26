// Package secret provides small, fail-closed helpers for deployment secrets.
package secret

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Read returns a non-empty secret from a regular owner-only file. Secret
// material is deliberately never included in returned error messages.
func Read(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("secret file path is required")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect secret file: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, errors.New("secret file must be a regular file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, errors.New("secret file permissions must be owner-only")
	}
	value, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read secret file: %w", err)
	}
	value = []byte(strings.TrimSpace(string(value)))
	if len(value) == 0 {
		return nil, errors.New("secret file is empty")
	}
	return value, nil
}

// WriteOnce creates an owner-only secret file without replacing an existing
// credential. Callers must create the parent directory with their desired
// ownership before invoking this function.
func WriteOnce(path string, payload []byte) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("secret file path is required")
	}
	if len(payload) == 0 {
		return errors.New("secret payload is required")
	}
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("inspect secret directory: %w", err)
	}
	if !info.IsDir() {
		return errors.New("secret parent is not a directory")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create secret file: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(payload); err != nil {
		return fmt.Errorf("write secret file: %w", err)
	}
	if _, err := file.Write([]byte("\n")); err != nil {
		return fmt.Errorf("finalize secret file: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync secret file: %w", err)
	}
	return nil
}
