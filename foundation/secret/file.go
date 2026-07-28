// Package secret provides small, fail-closed helpers for deployment secrets.
package secret

import (
	"crypto/rand"
	"encoding/hex"
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

// Replace atomically replaces an existing owner-only secret file. It refuses
// symlinks and unsafe existing files so deployment reconciliation cannot turn
// a credential refresh into an arbitrary file write.
func Replace(path string, payload []byte) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("secret file path is required")
	}
	if len(payload) == 0 {
		return errors.New("secret payload is required")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect existing secret file: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("secret file must be an existing regular file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return errors.New("secret file permissions must be owner-only")
	}

	dir := filepath.Dir(path)
	parent, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("inspect secret directory: %w", err)
	}
	if !parent.IsDir() {
		return errors.New("secret parent is not a directory")
	}
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return fmt.Errorf("generate replacement file name: %w", err)
	}
	temporaryPath := filepath.Join(dir, "."+filepath.Base(path)+"."+hex.EncodeToString(random)+".tmp")
	file, err := os.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create replacement secret file: %w", err)
	}
	defer os.Remove(temporaryPath)
	if _, err := file.Write(payload); err != nil {
		file.Close()
		return fmt.Errorf("write replacement secret file: %w", err)
	}
	if _, err := file.Write([]byte("\n")); err != nil {
		file.Close()
		return fmt.Errorf("finalize replacement secret file: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync replacement secret file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close replacement secret file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace secret file: %w", err)
	}
	directory, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open secret directory for sync: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync secret directory: %w", err)
	}
	return nil
}
